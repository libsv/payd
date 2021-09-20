package service

import (
	"context"
	"crypto/rand"
	"encoding/binary"

	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	gopayd "github.com/libsv/payd"
	"github.com/pkg/errors"
)

type signer struct {
	tStr gopayd.TxoReaderWriter
	pk   gopayd.PrivateKeyService
}

// NewSignerService returns a new signer service.
func NewSignerService(pk gopayd.PrivateKeyService, tStr gopayd.TxoReaderWriter) gopayd.SignerService {
	return &signer{
		pk:   pk,
		tStr: tStr,
	}
}

func (s *signer) FundAndSignTx(ctx context.Context, req gopayd.FundAndSignTxRequest) (*gopayd.SignTxResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	tx, err := bt.NewTxFromString(req.TxHex)
	if err != nil {
		return nil, err
	}
	fq := bt.NewFeeQuote().
		AddQuote(req.Fee.Standard.FeeType, &req.Fee.Standard).
		AddQuote(req.Fee.Data.FeeType, &req.Fee.Data)

	privKey, err := s.pk.PrivateKey(ctx, req.Account)
	if err != nil {
		return nil, err
	}

	if err := tx.FromInputs(ctx, fq, s.inputGetter(req.PaymentID, req.Account)); err != nil {
		return nil, errors.Wrap(err, "failed to fund transaction")
	}

	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return nil, err
	}
	seed := binary.LittleEndian.Uint64(b[:])

	path := bip32.DerivePath(seed)
	pubKey, err := privKey.DerivePublicKeyFromPath(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to derive key when create change output")
	}

	script, err := bscript.NewP2PKHFromPubKeyBytes(pubKey)
	if err != nil {
		return nil, err
	}

	if err := tx.Change(script, fq); err != nil {
		return nil, err
	}

	n, err := tx.Bip32SignAuto(ctx, &bt.LocalBip32SignerDeriver{MasterPrivateKey: privKey}, s.tStr.DerivationPath)
	if err != nil {
		return nil, err
	}
	if len(n) == 0 {
		return nil, errors.New("no inputs signed")
	}

	if err := s.tStr.StoreChange(ctx, gopayd.TxoCreate{
		TxID:           tx.TxID(),
		TxHex:          tx.String(),
		KeyName:        req.Account,
		LockingScript:  script.String(),
		DerivationPath: path,
		Satoshis:       tx.Outputs[len(tx.Outputs)-1].Satoshis,
		Vout:           uint64(len(tx.Outputs) - 1),
	}); err != nil {
		return nil, err
	}

	return &gopayd.SignTxResponse{SignedTx: tx.String()}, nil
}

func (s *signer) inputGetter(paymentID, account string) bt.InputGetterFunc {
	offset := 0

	return func(ctx context.Context) (*bt.Input, error) {
		txos, err := s.tStr.ReserveTxos(ctx, gopayd.TxoReserveArgs{
			Account:     account,
			ReservedFor: paymentID,
			Limit:       1,
			Offset:      offset,
		})
		if err != nil {
			return nil, err
		}
		offset++
		if len(txos) == 0 {
			return nil, bt.ErrNoInput
		}

		return bt.NewInputFrom(txos[0].TxID, txos[0].Vout, txos[0].LockingScript, txos[0].Satoshis)
	}
}
