package service

import (
	"context"
	"encoding/hex"

	"github.com/libsv/go-bc/spv"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/payd/cli/models"
)

type fund struct {
	rt  models.Regtest
	ps  models.PaymentStore
	sec spv.EnvelopeCreator
}

// NewFundService creates a new fund service.
func NewFundService(rt models.Regtest, ps models.PaymentStore, sec spv.EnvelopeCreator) *fund {
	return &fund{
		rt:  rt,
		ps:  ps,
		sec: sec,
	}
}

// Fund a wallet with a given amount of satoshis.
func (f *fund) Fund(ctx context.Context, payReq models.PaymentRequest) (*models.PaymentACK, error) {
	tx := bt.NewTx()
	for _, o := range payReq.Outputs {
		s, err := bscript.NewFromHexString(o.Script)
		if err != nil {
			return nil, err
		}

		if err := tx.AddP2PKHOutputFromScript(s, o.Amount); err != nil {
			return nil, err
		}
	}

	resp, err := f.rt.ListUnspent(ctx)
	if err != nil {
		return nil, err
	}

	utxos := make([]*bt.UTXO, len(resp.Result))
	for i, utxo := range resp.Result {
		script, err := bscript.NewFromHexString(utxo.ScriptPubKey)
		if err != nil {
			return nil, err
		}
		txid, err := hex.DecodeString(utxo.TxID)
		if err != nil {
			return nil, err
		}
		utxos[i] = &bt.UTXO{
			LockingScript: script,
			Vout:          utxo.Vout,
			Satoshis:      uint64(utxo.Amount * 100000000),
			TxID:          txid,
		}
	}

	if err := tx.Fund(ctx, payReq.Fee, func() bt.UTXOGetterFunc {
		idx := 0
		return func(ctx context.Context, deficit uint64) ([]*bt.UTXO, error) {
			if idx == len(utxos) {
				return nil, bt.ErrNoUTXO
			}
			for !utxos[idx].LockingScript.IsP2PKH() {
				idx++
			}
			defer func() { idx++ }()
			return utxos[idx : idx+1], nil
		}
	}()); err != nil {
		return nil, err
	}

	addressResp, err := f.rt.GetNewAddress(ctx)
	if err != nil {
		return nil, err
	}
	if err := tx.ChangeToAddress(*addressResp.Result, payReq.Fee); err != nil {
		return nil, err
	}

	signedResp, err := f.rt.SignRawTransaction(ctx, tx.String())
	if err != nil {
		return nil, err
	}

	signedTx, err := bt.NewTxFromString(signedResp.Result.Hex)
	if err != nil {
		return nil, err
	}

	spvEnvelope, err := f.sec.CreateEnvelope(ctx, signedTx)
	if err != nil {
		return nil, err
	}

	return f.ps.Submit(ctx, models.PaymentSendArgs{
		Transaction:    signedTx.String(),
		PaymentRequest: payReq,
		MerchantData:   payReq.MerchantData,
		Memo:           payReq.Memo,
		SPVEnvelope:    spvEnvelope,
	})
}

func (f *fund) FundedTx(ctx context.Context, payReq models.PaymentRequest) (*bt.Tx, error) {
	tx := bt.NewTx()
	for _, o := range payReq.Outputs {
		s, err := bscript.NewFromHexString(o.Script)
		if err != nil {
			return nil, err
		}

		if err := tx.AddP2PKHOutputFromScript(s, o.Amount); err != nil {
			return nil, err
		}
	}

	resp, err := f.rt.ListUnspent(ctx)
	if err != nil {
		return nil, err
	}

	utxos := make([]*bt.UTXO, len(resp.Result))
	for i, utxo := range resp.Result {
		script, err := bscript.NewFromHexString(utxo.ScriptPubKey)
		if err != nil {
			return nil, err
		}
		txid, err := hex.DecodeString(utxo.TxID)
		if err != nil {
			return nil, err
		}
		utxos[i] = &bt.UTXO{
			LockingScript: script,
			Vout:          utxo.Vout,
			Satoshis:      uint64(utxo.Amount * 100000000),
			TxID:          txid,
		}
	}

	if err := tx.Fund(ctx, payReq.Fee, func() bt.UTXOGetterFunc {
		idx := 0
		return func(ctx context.Context, deficit uint64) ([]*bt.UTXO, error) {
			if idx == len(utxos) {
				return nil, bt.ErrNoUTXO
			}
			for !utxos[idx].LockingScript.IsP2PKH() {
				idx++
			}
			defer func() { idx++ }()
			return utxos[idx : idx+1], nil
		}
	}()); err != nil {
		return nil, err
	}

	addressResp, err := f.rt.GetNewAddress(ctx)
	if err != nil {
		return nil, err
	}
	if err := tx.ChangeToAddress(*addressResp.Result, payReq.Fee); err != nil {
		return nil, err
	}

	signedResp, err := f.rt.SignRawTransaction(ctx, tx.String())
	if err != nil {
		return nil, err
	}

	signedTx, err := bt.NewTxFromString(signedResp.Result.Hex)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}
