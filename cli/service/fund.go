package service

import (
	"context"

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

func NewFundService(rt models.Regtest, ps models.PaymentStore, sec spv.EnvelopeCreator) *fund {
	return &fund{
		rt:  rt,
		ps:  ps,
		sec: sec,
	}
}

func (f *fund) Fund(ctx context.Context, payReq models.PaymentRequest) (*models.PaymentAck, error) {
	opts := make(map[string]float64)
	for _, o := range payReq.Outputs {
		s, err := bscript.NewFromHexString(o.Script)
		if err != nil {
			return nil, err
		}

		pkh, err := s.PublicKeyHash()
		if err != nil {
			return nil, err
		}

		addr, err := bscript.NewAddressFromPublicKeyHash(pkh, false)
		if err != nil {
			return nil, err
		}

		opts[addr.AddressString] = float64(o.Amount) / 1000000000
	}

	resp, err := f.rt.ListUnspent(ctx)
	if err != nil {
		return nil, err
	}

	tx := bt.NewTx()
	if err := tx.Fund(ctx, payReq.Fee, func() bt.UTXOGetterFunc {
		idx := 0
		utxos := resp.Result
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
	if err := tx.ChangeToAddress("mk4aGx8uGR2U5Qku2zzngdEC9VH8zBqQ9K", payReq.Fee); err != nil {
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
		SPVEnvelope:    *spvEnvelope,
	})
}
