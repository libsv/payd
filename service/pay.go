package service

import (
	"context"
	"crypto/rand"
	"encoding/binary"
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
)

type pay struct {
	txoWtr payd.TxoWriter
	txWtr  payd.TransactionWriter
	p4     http.P4
	pk     payd.PrivateKeyService
	spvc   spv.EnvelopeCreator
	svrCfg *config.Server
}

func NewPayService(txoWtr payd.TxoWriter, txWtr payd.TransactionWriter, p4 http.P4, pk payd.PrivateKeyService, spvc spv.EnvelopeCreator, svrCfg *config.Server) payd.PayService {
	return &pay{
		txoWtr: txoWtr,
		txWtr:  txWtr,
		p4:     p4,
		pk:     pk,
		spvc:   spvc,
		svrCfg: svrCfg,
	}
}

type derivationSigner struct {
	pathMap       map[*bscript.Script]string
	masterPrivKey *bip32.ExtendedKey
}

func (l derivationSigner) Signer(ctx context.Context, script *bscript.Script) (bt.Signer, error) {
	path, ok := l.pathMap[script]
	if !ok {
		return nil, errors.New("derivation path does not exist for script")
	}
	extPrivKey, err := l.masterPrivKey.DeriveChildFromPath(path)
	if err != nil {
		return nil, err
	}

	privKey, err := extPrivKey.ECPrivKey()
	if err != nil {
		return nil, err
	}

	return &bt.LocalSigner{
		PrivateKey: privKey,
	}, nil
}

func (p *pay) Pay(ctx context.Context, req payd.PayRequest) error {
	privKey, err := p.pk.PrivateKey(ctx, "masterkey")
	if err != nil {
		return err
	}

	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return err
	}

	derivationPath := bip32.DerivePath(binary.LittleEndian.Uint64(b[:]))
	pubKey, err := privKey.DerivePublicKeyFromPath(derivationPath)
	if err != nil {
		return errors.Wrap(err, "failed to derive key when create change output")
	}

	changeLockingScript, err := bscript.NewP2PKHFromPubKeyBytes(pubKey)
	if err != nil {
		return err
	}

	payReq, err := p.p4.PaymentRequest(ctx, req)
	if err != nil {
		return err
	}

	tx := bt.NewTx()
	var totalOutputs uint64
	for _, out := range payReq.Outputs {
		lockingScript, err := bscript.NewFromHexString(out.Script)
		if err != nil {
			return err
		}
		totalOutputs += out.Amount
		tx.AddP2PKHOutputFromScript(lockingScript, out.Amount)
	}

	signer := &derivationSigner{
		pathMap:       make(map[*bscript.Script]string),
		masterPrivKey: privKey,
	}

	if err = tx.Fund(ctx, payReq.Fee, func(ctx context.Context, deficit uint64) ([]*bt.UTXO, error) {
		utxos, err := p.txoWtr.UTXOReserve(ctx, payd.UTXOReserve{
			ReservedFor: req.PayToURL,
			Satoshis:    deficit,
		})
		if err != nil {
			return nil, err
		}
		var txos []*bt.UTXO
		for _, utxo := range utxos {
			txid, err := hex.DecodeString(utxo.TxID)
			if err != nil {
				return nil, err
			}
			lockingScript, err := bscript.NewFromHexString(utxo.LockingScript)
			if err != nil {
				return nil, err
			}

			txos = append(txos, &bt.UTXO{
				TxID:          txid,
				Vout:          utxo.Vout,
				Satoshis:      utxo.Satoshis,
				LockingScript: lockingScript,
			})

			signer.pathMap[lockingScript] = utxo.DerivationPath
		}
		return txos, nil
	}); err != nil {
		return err
	}

	if err := tx.Change(changeLockingScript, payReq.Fee); err != nil {
		return err
	}

	if err := tx.SignAll(ctx, signer); err != nil {
		return err
	}

	spvEnvelope, err := p.spvc.CreateEnvelope(ctx, tx)
	if err != nil {
		return err
	}

	if err := p.p4.PaymentSend(ctx, req, payd.PaymentSend{
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
	}); err != nil {
		return err
	}

	// Only insert change tx if it exists
	if changeLockingScript.Equals(tx.Outputs[tx.OutputCount()-1].LockingScript) {
		if err := p.txWtr.TransactionChangeCreate(ctx, payd.TransactionCreate{
			TxID:  spvEnvelope.TxID,
			TxHex: spvEnvelope.RawTx,
			Outputs: []*payd.TxoCreate{{
				Outpoint: fmt.Sprintf("%s%d", spvEnvelope.TxID, tx.OutputCount()-1),
				TxID:     spvEnvelope.TxID,
				Vout:     uint64(tx.OutputCount() - 1),
			}},
		}, payd.DestinationCreate{
			Script:         changeLockingScript.String(),
			DerivationPath: derivationPath,
			Keyname:        "masterkey",
			Satoshis:       tx.Outputs[tx.OutputCount()-1].Satoshis,
		}); err != nil {
			return err
		}
	}

	return p.txoWtr.UTXOSpend(ctx, payd.UTXOSpend{
		SpendingTxID: spvEnvelope.TxID,
		Reservation:  payReq.PaymentURL,
	})
}
