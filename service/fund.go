package service

import (
	"context"
	"errors"

	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	gopayd "github.com/libsv/payd"
	"gopkg.in/guregu/null.v3"
)

type fund struct {
	pk   gopayd.PrivateKeyService
	fStr gopayd.FundStore
}

// NewFundService returns a new fund service.
func NewFundService(pk gopayd.PrivateKeyService, fStr gopayd.FundStore) gopayd.FundService {
	return &fund{
		pk:   pk,
		fStr: fStr,
	}
}

func (f *fund) FundsAdd(ctx context.Context, req gopayd.FundAddRequest) ([]*gopayd.Txo, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	tx, err := bt.NewTxFromString(req.Tx)
	if err != nil {
		return nil, err
	}
	pk, err := f.pk.PrivateKey(ctx, req.Account)
	if err != nil {
		return nil, err
	}

	ecPubKey, err := pk.ECPubKey()
	if err != nil {
		return nil, err
	}

	ls, err := bscript.NewP2PKHFromPubKeyEC(ecPubKey)
	if err != nil {
		return nil, err
	}

	txID := tx.TxID()
	txos := make([]*gopayd.Txo, 0)
	for i, o := range tx.Outputs {
		if !ls.Equals(o.LockingScript) {
			continue
		}

		txos = append(txos, &gopayd.Txo{
			TxID:          txID,
			Vout:          i,
			KeyName:       null.StringFrom(req.Account),
			Satoshis:      o.Satoshis,
			LockingScript: o.LockingScriptHexString(),
		})
	}

	if err := f.fStr.StoreFund(ctx, gopayd.StoreFundRequest{
		TxID:  txID,
		TxHex: tx.String(),
		Txos:  txos,
	}); err != nil {
		return nil, err
	}

	return txos, nil
}

func (f *fund) FundsGet(ctx context.Context, args gopayd.FundsGetArgs) ([]gopayd.Txo, error) {
	if args.Account == "" {
		return nil, errors.New("account header needed")
	}
	txos, err := f.fStr.Funds(ctx, args)
	if err != nil {
		return nil, err
	}

	return txos, nil
}

func (f *fund) FundsSpend(ctx context.Context, req gopayd.FundsSpendReq, args gopayd.FundsSpendArgs) error {
	if err := req.Validate(); err != nil {
		return err
	}
	if err := args.Validate(); err != nil {
		return err
	}

	if args.Account == "" {
		return errors.New("account header needed")
	}
	tx, err := bt.NewTxFromString(req.SpendingTx)
	if err != nil {
		return err
	}

	txID := tx.TxID()
	txos := make([]gopayd.Txo, tx.InputCount())
	for i, input := range tx.Inputs {
		txos[i] = gopayd.Txo{
			TxID:         input.PreviousTxIDStr(),
			Vout:         int(input.PreviousTxOutIndex),
			SpendingTxID: null.StringFrom(txID),
			KeyName:      null.StringFrom(args.Account),
		}
	}
	req.Txos = txos
	req.SpendingTxID = txID
	return f.fStr.SpendFunds(ctx, &req, args)
}
