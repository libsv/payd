package service

import (
	"context"

	"github.com/pkg/errors"
	"github.com/theflyingcodr/lathos/errs"

	"github.com/libsv/go-bk/chaincfg"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/sighash"
	gopayd "github.com/libsv/payd"
)

type signer struct {
	fStr gopayd.FundStore
	pk   gopayd.PrivateKeyService
}

// NewSignerService returns a new signer service.
func NewSignerService(pk gopayd.PrivateKeyService, fStr gopayd.FundStore) gopayd.SignerService {
	return &signer{
		pk:   pk,
		fStr: fStr,
	}
}

func (s *signer) FundAndSignTx(ctx context.Context, req gopayd.FundAndSignTxRequest) (*gopayd.SignTxResponse, error) {
	tx, err := bt.NewTxFromString(req.TxHex)
	if err != nil {
		return nil, err
	}
	ff, err := s.fStr.Funds(ctx, gopayd.FundsGetArgs{Account: req.Account})
	if err != nil {
		return nil, err
	}

	fq := bt.NewFeeQuote().
		AddQuote(req.Fee.Standard.FeeType, &req.Fee.Standard).
		AddQuote(req.Fee.Data.FeeType, &req.Fee.Standard)

	pk, err := s.pk.PrivateKey(ctx, req.Account)
	if err != nil {
		return nil, err
	}
	epk, err := pk.ECPrivKey()
	if err != nil {
		return nil, err
	}

	localSigner := &bt.LocalSigner{PrivateKey: epk}

	feesPaid, err := tx.IsFeePaidEnough(fq)
	if err != nil {
		return nil, err
	}
	for i := 0; !feesPaid && i < len(ff); i++ {
		f := ff[i]
		if err := tx.From(f.TxID, uint32(f.Vout), f.LockingScript, f.Satoshis); err != nil {
			return nil, err
		}
		if err := tx.Sign(ctx, localSigner, uint32(len(tx.Inputs)-1), sighash.AllForkID); err != nil {
			return nil, err
		}

		feesPaid, err = tx.IsFeePaidEnough(fq)
		if err != nil {
			return nil, err
		}
	}
	if !feesPaid {
		return nil, errs.NewErrUnprocessable("F01", "not enough funds")
	}

	if err := tx.ChangeToAddress(pk.Address(&chaincfg.Params{}), fq); err != nil {
		return nil, err
	}

	n, err := tx.SignAuto(ctx, &bt.LocalSigner{PrivateKey: epk})
	if err != nil {
		return nil, err
	}
	if len(n) == 0 {
		return nil, errors.New("no inputs were signed")
	}

	return &gopayd.SignTxResponse{SignedTx: tx.String()}, nil
}
