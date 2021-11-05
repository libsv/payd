package service

import (
	"context"
	"fmt"

	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/pkg/errors"

	"github.com/libsv/payd"
	"github.com/libsv/payd/config"
	"github.com/libsv/payd/data/http"
)

type pay struct {
	txoWtr  payd.TxoWriter
	txWtr   payd.TransactionWriter
	destWtr payd.DestinationsWriter
	p4      http.P4
	spvc    payd.EnvelopeService
	svrCfg  *config.Server
}

// NewPayService returns a pay service.
func NewPayService(txoWtr payd.TxoWriter, txWtr payd.TransactionWriter, destWtr payd.DestinationsWriter, p4 http.P4, spvc payd.EnvelopeService, svrCfg *config.Server) payd.PayService {
	return &pay{
		txoWtr:  txoWtr,
		txWtr:   txWtr,
		destWtr: destWtr,
		p4:      p4,
		spvc:    spvc,
		svrCfg:  svrCfg,
	}
}

type derivationSigner struct {
	pathMap       map[*bscript.Script]string
	masterPrivKey *bip32.ExtendedKey
}

// Signer returns a signer configured for a provided *bscript.Script.
func (l derivationSigner) Signer(ctx context.Context, script *bscript.Script) (bt.Signer, error) {
	path, ok := l.pathMap[script]
	if !ok {
		return nil, errors.New("derivation path does not exist for script")
	}
	extPrivKey, err := l.masterPrivKey.DeriveChildFromPath(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to derive child from path %s for locking script %s", path, script.String())
	}

	privKey, err := extPrivKey.ECPrivKey()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create ec private key for script %s", script.String())
	}

	return &bt.LocalSigner{
		PrivateKey: privKey,
	}, nil
}

// Pay takes a pay-to url and performs a payment procedure, ultimately sending money to the
// url.
func (p *pay) Pay(ctx context.Context, req payd.PayRequest) (*payd.PaymentACK, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Retrieve the payment request information from the receiver.
	payReq, err := p.p4.PaymentRequest(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to request payment for url %s", req.PayToURL)
	}

	// Defer unreserve func so in the event of an error before sending the payment, reserved funds are freed up.
	var paymentSent bool
	defer func() {
		if paymentSent {
			return
		}
		_ = p.txoWtr.UTXOUnreserve(ctx, payd.UTXOUnreserve{
			ReservedFor: req.PayToURL,
		})
	}()

	env, err := p.spvc.Envelope(ctx, payd.EnvelopeArgs{PayToURL: req.PayToURL}, *payReq)
	if err != nil {
		return nil, errors.Wrapf(err, "envelope creation failed for '%s'", req.PayToURL)
	}
	tx, err := bt.NewTxFromString(env.SPVEnvelope.RawTx)
	if err != nil {
		return nil, err
	}
	// Send the payment to the p4 server.
	ack, err := p.p4.PaymentSend(ctx, req, payd.PaymentSend{
		SPVEnvelope: env.SPVEnvelope,
		ProofCallbacks: map[string]payd.ProofCallback{
			"https://" + p.svrCfg.Hostname + "/api/v1/proofs/" + env.SPVEnvelope.TxID: {},
		},
		MerchantData: payd.User{
			Name:         payReq.MerchantData.Name,
			Email:        payReq.MerchantData.Email,
			Avatar:       payReq.MerchantData.Avatar,
			Address:      payReq.MerchantData.Address,
			ExtendedData: payReq.MerchantData.ExtendedData,
		},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to send payment %s", req.PayToURL)
	}
	if ack.Error > 0 {
		return nil, fmt.Errorf("failed to send payment. Code '%d' reason '%s'", ack.Error, ack.Memo)
	}
	paymentSent = true

	txCreate := payd.TransactionCreate{
		TxID:  env.SPVEnvelope.TxID,
		TxHex: env.SPVEnvelope.RawTx,
	}
	// Only insert change utxo if change exists.
	if env.Change.LockingScript != nil && env.Change.LockingScript.Equals(tx.Outputs[tx.OutputCount()-1].LockingScript) {
		oo, err := p.destWtr.DestinationsCreate(ctx, payd.DestinationsCreateArgs{},
			[]payd.DestinationCreate{{
				Script:         env.Change.LockingScript.String(),
				DerivationPath: env.Change.DerivationPath,
				Keyname:        keyname,
				Satoshis:       tx.Outputs[tx.OutputCount()-1].Satoshis,
			}})
		if err != nil {
			return nil, errors.Wrap(err, "failed to create destination for change output")
		}
		txCreate.Outputs = []*payd.TxoCreate{{
			TxID:          env.SPVEnvelope.TxID,
			Outpoint:      fmt.Sprintf("%s%d", env.SPVEnvelope.TxID, tx.OutputCount()-1),
			Vout:          uint64(tx.OutputCount() - 1),
			DestinationID: oo[0].ID,
		}}
	}

	// Create a tx in the data store with the sent tx's information.
	if err = p.txWtr.TransactionCreate(ctx, txCreate); err != nil {
		return nil, errors.Wrap(err, "failed to create transaction for change output")
	}

	// Mark the reserved utxos as spent.
	if err = p.txoWtr.UTXOSpend(ctx, payd.UTXOSpend{
		SpendingTxID: env.SPVEnvelope.TxID,
		Reservation:  payReq.PaymentURL,
	}); err != nil {
		return nil, errors.Wrap(err, "failed to mark utxos as spent")
	}

	return ack, nil
}
