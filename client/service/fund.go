package service

import (
	"context"

	"github.com/libsv/go-bk/chaincfg"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	gopayd "github.com/libsv/payd"
	"github.com/libsv/payd/client"
	"github.com/pkg/errors"
)

const keyname = `client`

type fund struct {
	rt    client.Regtest
	fWtr  client.FundWriter
	pkSvc gopayd.PrivateKeyService
}

func NewFundService(rt client.Regtest, fWtr client.FundWriter, pkSvc gopayd.PrivateKeyService) *fund {
	return &fund{
		rt:    rt,
		fWtr:  fWtr,
		pkSvc: pkSvc,
	}
}

func (f *fund) Seed(ctx context.Context, req client.FundSeed) (*gopayd.Transaction, error) {
	pk, err := f.pkSvc.PrivateKey(ctx, "client")
	if err != nil {
		return nil, errors.Wrap(err, "error getting client private key")
	}

	staResp, err := f.rt.SendToAddress(ctx, pk.Address(&chaincfg.Params{}), req.Amount)
	if err != nil {
		return nil, err
	}

	rawTx, err := f.rt.RawTransaction(ctx, *staResp.Result)
	if err != nil {
		return nil, err
	}

	tx, err := bt.NewTxFromString(*rawTx.Result)
	if err != nil {
		return nil, err
	}

	return f.FundsCreate(ctx, tx)
}

func (f *fund) FundsCreate(ctx context.Context, tx *bt.Tx) (*gopayd.Transaction, error) {
	pk, err := f.pkSvc.PrivateKey(ctx, "client")
	if err != nil {
		return nil, errors.Wrap(err, "error getting client private key")
	}
	addr := pk.Address(&chaincfg.Params{})
	script, err := bscript.NewP2PKHFromAddress(addr)
	if err != nil {
		return nil, err
	}

	txID := tx.TxID()
	txos := make([]*client.Fund, 0)
	for i, o := range tx.Outputs {
		if !script.Equals(o.LockingScript) {
			continue
		}
		txos = append(txos, &client.Fund{
			TxID:          txID,
			Vout:          i,
			KeyName:       "client",
			Satoshis:      o.Satoshis,
			LockingScript: o.LockingScriptHexString(),
		})
	}

	return f.fWtr.FundsCreate(ctx, client.FundsCreate{
		TxID:  txID,
		TxHex: tx.String(),
		Funds: txos,
	})
}
