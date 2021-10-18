package service

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/libsv/go-bc/spv"
	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/payd"
	"github.com/libsv/payd/config"
	"github.com/libsv/payd/data/http"
	"github.com/pkg/errors"
	"github.com/theflyingcodr/lathos/errs"
)

type pay struct {
	txoWtr  payd.TxoWriter
	txWtr   payd.TransactionWriter
	destWtr payd.DestinationsWriter
	p4      http.P4
	pk      payd.PrivateKeyService
	spvc    spv.EnvelopeCreator
	svrCfg  *config.Server
	seed    payd.SeedService
}

// NewPayService returns a pay service.
func NewPayService(txoWtr payd.TxoWriter, txWtr payd.TransactionWriter, destWtr payd.DestinationsWriter, p4 http.P4, pk payd.PrivateKeyService, spvc spv.EnvelopeCreator, svrCfg *config.Server, seed payd.SeedService) payd.PayService {
	return &pay{
		txoWtr:  txoWtr,
		txWtr:   txWtr,
		destWtr: destWtr,
		p4:      p4,
		pk:      pk,
		spvc:    spvc,
		svrCfg:  svrCfg,
		seed:    seed,
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

	// Retrieve private key and build change utxo in advance of making any calls, so that
	// if something internal goes wrong we don't make a premature request to the receiver's
	// p4 server, creating unneeded traffic.
	privKey, err := p.pk.PrivateKey(ctx, keyname)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve private key")
	}

	seed, err := p.seed.Uint64()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create seed for derivation path")
	}

	derivationPath := bip32.DerivePath(seed)
	pubKey, err := privKey.DerivePublicKeyFromPath(derivationPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to derive key when create change output")
	}

	changeLockingScript, err := bscript.NewP2PKHFromPubKeyBytes(pubKey)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to derived change locking script for seed %d, path %s", seed, derivationPath)
	}

	// Retrieve the payment request information from the receiver.
	payReq, err := p.p4.PaymentRequest(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to request payment for url %s", req.PayToURL)
	}

	tx := bt.NewTx()
	// Add funds to new tx.
	for _, out := range payReq.Destinations.Outputs {
		lockingScript, err := bscript.NewFromHexString(out.Script)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parsed script %s", out.Script)
		}
		if err = tx.AddP2PKHOutputFromScript(lockingScript, out.Amount); err != nil {
			return nil, errors.Wrapf(err, "failed to add locking script to tx for script %s, amount %d", out.Script, out.Amount)
		}
	}

	// Create a signer to map locking scripts with derivation paths.
	signer := &derivationSigner{
		pathMap:       make(map[*bscript.Script]string),
		masterPrivKey: privKey,
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
	if err = tx.Fund(ctx, payReq.Fee, func(ctx context.Context, deficit uint64) ([]*bt.UTXO, error) {
		utxos, err := p.txoWtr.UTXOReserve(ctx, payd.UTXOReserve{
			ReservedFor: req.PayToURL,
			Satoshis:    deficit,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to reserve utxos")
		}
		if len(utxos) == 0 {
			return nil, bt.ErrNoUTXO
		}
		var txos []*bt.UTXO
		for _, utxo := range utxos {
			txid, err := hex.DecodeString(utxo.TxID)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to decode txid %s for utxo", utxo.TxID)
			}
			lockingScript, err := bscript.NewFromHexString(utxo.LockingScript)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse locking script %s for utxo", utxo.LockingScript)
			}

			txos = append(txos, &bt.UTXO{
				TxID:          txid,
				Vout:          utxo.Vout,
				Satoshis:      utxo.Satoshis,
				LockingScript: lockingScript,
			})

			// Add the locking script and its derivation path to the signers map.
			signer.pathMap[lockingScript] = utxo.DerivationPath
		}
		return txos, nil
	}); err != nil {
		if ok := errors.Is(err, bt.ErrInsufficientFunds); ok {
			return nil, errs.NewErrUnprocessable("F001", bt.ErrInsufficientFunds.Error())
		}
		return nil, errors.Wrapf(err, "failed to fund tx for payment %s", req.PayToURL)
	}

	// Finalise the tx.
	if err = tx.Change(changeLockingScript, payReq.Fee); err != nil {
		return nil, errors.Wrap(err, "failed to set change")
	}

	if err = tx.SignAll(ctx, signer); err != nil {
		return nil, errors.Wrapf(err, "failed to sign tx %s", tx.String())
	}

	// Create the spv envelope for the tx.
	spvEnvelope, err := p.spvc.CreateEnvelope(ctx, tx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create spv envelope for tx %s", tx.String())
	}

	// Send the payment to the p4 server.
	ack, err := p.p4.PaymentSend(ctx, req, payd.PaymentSend{
		SPVEnvelope: spvEnvelope,
		ProofCallbacks: map[string]payd.ProofCallback{
			"https://" + p.svrCfg.Hostname + "/api/v1/proofs/" + spvEnvelope.TxID: {},
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
	paymentSent = true

	txCreate := payd.TransactionCreate{
		TxID:  spvEnvelope.TxID,
		TxHex: spvEnvelope.RawTx,
	}
	// Only insert change utxo if change exists.
	if changeLockingScript.Equals(tx.Outputs[tx.OutputCount()-1].LockingScript) {
		oo, err := p.destWtr.DestinationsCreate(ctx, payd.DestinationsCreateArgs{},
			[]payd.DestinationCreate{{
				Script:         changeLockingScript.String(),
				DerivationPath: derivationPath,
				Keyname:        keyname,
				Satoshis:       tx.Outputs[tx.OutputCount()-1].Satoshis,
			}})
		if err != nil {
			return nil, errors.Wrap(err, "failed to create destination for change output")
		}
		txCreate.Outputs = []*payd.TxoCreate{{
			TxID:          spvEnvelope.TxID,
			Outpoint:      fmt.Sprintf("%s%d", spvEnvelope.TxID, tx.OutputCount()-1),
			Vout:          uint64(tx.OutputCount() - 1),
			DestinationID: oo[0].ID,
		}}
	}

	// Create a tx in the data store with the sent tx's information.
	if err = p.txWtr.TransactionCreate(ctx, txCreate); err != nil {
		return nil, errors.Wrap(err, "failed to created transaction for change output")
	}

	// Mark the reserved utxos as spent.
	if err = p.txoWtr.UTXOSpend(ctx, payd.UTXOSpend{
		SpendingTxID: spvEnvelope.TxID,
		Reservation:  payReq.PaymentURL,
	}); err != nil {
		return nil, errors.Wrap(err, "failed to mark utxos as spent")
	}

	return ack, nil
}
