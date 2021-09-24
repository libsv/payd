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
	"github.com/libsv/payd/data/http"
	"github.com/pkg/errors"
)

type pay struct {
	tWtr payd.TxoWriter
	p4   http.P4
	pk   payd.PrivateKeyService
	spvb spv.EnvelopeCreator
}

type lsdpm struct {
	m   map[*bscript.Script]string
	mpk *bip32.ExtendedKey
}

func NewPayService(tWtr payd.TxoWriter, p4 http.P4, pk payd.PrivateKeyService, spvb spv.EnvelopeCreator) payd.PayService {
	return &pay{tWtr: tWtr, p4: p4, pk: pk}
}

func (l lsdpm) Signer(ctx context.Context, script *bscript.Script) (bt.Signer, error) {
	path, ok := l.m[script]
	if !ok {
		return nil, errors.New("derivation path does not exist for script")
	}
	extPrivKey, err := l.mpk.DeriveChildFromPath(path)
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
	seed := binary.LittleEndian.Uint64(b[:])

	path := bip32.DerivePath(seed)
	pubKey, err := privKey.DerivePublicKeyFromPath(path)
	if err != nil {
		return errors.Wrap(err, "failed to derive key when create change output")
	}

	script, err := bscript.NewP2PKHFromPubKeyBytes(pubKey)
	if err != nil {
		return err
	}

	// Get request from p4
	payRec, err := p.p4.PaymentRequest(ctx, req)
	if err != nil {
		return err
	}

	tx := bt.NewTx()

	// Reserve utxos
	var totalOutputs uint64
	for _, out := range payRec.Outputs {
		s, err := bscript.NewFromHexString(out.Script)
		if err != nil {
			return err
		}
		totalOutputs += out.Amount
		tx.AddP2PKHOutputFromScript(s, out.Amount)
	}

	dpTracker := &lsdpm{
		m:   make(map[*bscript.Script]string),
		mpk: privKey,
	}

	if err = tx.Fund(ctx, payRec.Fee, func(ctx context.Context, deficit uint64) ([]*bt.UTXO, error) {
		utxos, err := p.tWtr.UTXOReserve(ctx, payd.UTXOReserve{
			ReservedFor: req.PayToURL,
			Satoshis:    deficit,
		})
		if err != nil {
			return nil, err
		}
		var txos []*bt.UTXO
		for _, utxo := range utxos {

			s, err := bscript.NewFromHexString(utxo.LockingScript)
			if err != nil {
				return nil, err
			}
			txid, err := hex.DecodeString(utxo.TxID)
			if err != nil {
				return nil, err
			}

			txos = append(txos, &bt.UTXO{
				TxID:          txid,
				Vout:          utxo.Vout,
				LockingScript: s,
				Satoshis:      utxo.Satoshis,
			})

			dpTracker.m[s] = utxo.DerivationPath
		}
		return txos, nil
	}); err != nil {
		return err
	}

	if err := tx.Change(script, payRec.Fee); err != nil {
		return err
	}

	if err := tx.SignAll(ctx, dpTracker); err != nil {
		return err
	}
	fmt.Println(tx.String())

	// Build spv envelope

	// Send request

	// Spend utxos?
	return nil
}
