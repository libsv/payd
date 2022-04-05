package service

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/libsv/go-bc/spv"
	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-dpp"
	"github.com/pkg/errors"
	"github.com/theflyingcodr/lathos/errs"

	"github.com/libsv/payd"
)

type envelopes struct {
	pkSvc   payd.PrivateKeyService
	destWtr payd.DestinationsWriter
	txoWtr  payd.TxoWriter
	txWtr   payd.TransactionWriter
	seedSvc payd.SeedService
	spvc    spv.EnvelopeCreator
}

// NewEnvelopes will setup and return an Envelope service, used to create spv envelopes.
func NewEnvelopes(pkSvc payd.PrivateKeyService, destWtr payd.DestinationsWriter, txWtr payd.TransactionWriter, txoWtr payd.TxoWriter, seedSvc payd.SeedService, spvc spv.EnvelopeCreator) *envelopes {
	return &envelopes{
		pkSvc:   pkSvc,
		destWtr: destWtr,
		txoWtr:  txoWtr,
		txWtr:   txWtr,
		seedSvc: seedSvc,
		spvc:    spvc,
	}
}

// Envelope will create and return a new Envelope.
func (e *envelopes) Envelope(ctx context.Context, args payd.EnvelopeArgs, req dpp.PaymentRequest) (*spv.Envelope, error) {
	// Retrieve private key and build change utxo in advance of making any calls, so that
	// if something internal goes wrong we don't make a premature request to the receiver's
	// dpp server, creating unneeded traffic.
	keyname := "masterkey"
	userID := uint64(1)
	privKey, err := e.pkSvc.PrivateKey(ctx, keyname, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve private key")
	}

	tx := bt.NewTx()
	// Add funds to new tx.
	for _, out := range req.Destinations.Outputs {
		if err = tx.AddP2PKHOutputFromScript(out.LockingScript, out.Amount); err != nil {
			return nil, errors.Wrapf(err, "failed to add locking script to tx for script %s, amount %d", out.LockingScript.String(), out.Amount)
		}
	}

	// Create a signer to map locking scripts with derivation paths.
	signer := &derivationSigner{
		pathMap:       make(map[*bscript.Script]string),
		masterPrivKey: privKey,
	}
	if err = tx.Fund(ctx, req.FeeRate, func(ctx context.Context, deficit uint64) ([]*bt.UTXO, error) {
		utxos, err := e.txoWtr.UTXOReserve(ctx, payd.UTXOReserve{
			ReservedFor: args.PayToURL,
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
		return nil, errors.Wrapf(err, "failed to fund tx for payment %s", args.PayToURL)
	}

	changeOutput, err := e.changeScript(privKey)
	if err != nil {
		return nil, err
	}
	// Finalise the tx.
	if err = tx.Change(changeOutput.LockingScript, req.FeeRate); err != nil {
		return nil, errors.Wrap(err, "failed to set change")
	}

	if err = tx.UnlockAll(ctx, signer); err != nil {
		return nil, errors.Wrapf(err, "failed to sign tx %s", tx.String())
	}

	spvEnvelope := &spv.Envelope{
		RawTx: tx.String(),
		TxID:  tx.TxID(),
	}

	// Create the spv envelope for the tx.
	s, err := e.spvc.CreateEnvelope(ctx, tx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create spv envelope for tx %s", tx.String())
	}
	spvEnvelope = s

	txCreate := payd.TransactionCreate{
		TxID:  tx.TxID(),
		TxHex: tx.String(),
	}
	// Only insert change utxo if change exists.
	if changeOutput.LockingScript.Equals(tx.Outputs[tx.OutputCount()-1].LockingScript) {
		oo, err := e.destWtr.DestinationsCreate(ctx, payd.DestinationsCreateArgs{},
			[]payd.DestinationCreate{{
				Script:         changeOutput.LockingScript.String(),
				DerivationPath: changeOutput.DerivationPath,
				UserID:         userID,
				Satoshis:       tx.Outputs[tx.OutputCount()-1].Satoshis,
			}})
		if err != nil {
			return nil, errors.Wrap(err, "failed to create destination for change output")
		}
		txCreate.Outputs = []*payd.TxoCreate{{
			TxID:          txCreate.TxID,
			Outpoint:      fmt.Sprintf("%s%d", txCreate.TxID, tx.OutputCount()-1),
			Vout:          uint64(tx.OutputCount() - 1),
			DestinationID: oo[0].ID,
		}}
	}

	// Create a tx in the data store with the sent tx's information.
	if err = e.txWtr.TransactionCreate(ctx, txCreate); err != nil {
		return nil, errors.Wrap(err, "failed to create transaction for change output")
	}

	// Mark the reserved utxos as spent.
	if err = e.txoWtr.UTXOSpend(ctx, payd.UTXOSpend{
		SpendingTxID: txCreate.TxID,
		Reservation:  args.PayToURL,
	}); err != nil {
		return nil, errors.Wrap(err, "failed to mark utxos as spent")
	}

	return spvEnvelope, nil
}

// changeScript will create and return a change locking script.
func (e *envelopes) changeScript(privKey *bip32.ExtendedKey) (*payd.Output, error) {
	seed, err := e.seedSvc.Uint64()
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
	return &payd.Output{
		LockingScript:  changeLockingScript,
		DerivationPath: derivationPath}, nil
}
