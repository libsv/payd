package service

import (
	"context"

	"github.com/libsv/go-bc/spv"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/payd/cli/models"
)

type txSvc struct {
	rt models.Regtest
}

// NewTxService returns a tx service.
func NewTxService(rt models.Regtest) spv.TxStore {
	return &txSvc{
		rt: rt,
	}
}

func (t *txSvc) Tx(ctx context.Context, txID string) (*bt.Tx, error) {
	resp, err := t.rt.RawTransaction(ctx, txID)
	if err != nil {
		return nil, err
	}

	return bt.NewTxFromString(*resp.Result)
}
