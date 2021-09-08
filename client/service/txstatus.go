package service

import (
	"context"

	gopayd "github.com/libsv/payd"
	"github.com/libsv/payd/client"
)

type txstatus struct {
	str client.TxStatusStore
}

// NewTxStatusService returns a new service for txstatus.
func NewTxStatusService(str client.TxStatusStore) *txstatus {
	return &txstatus{str: str}
}

func (t *txstatus) Status(ctx context.Context, args client.TxStatusArgs) (*gopayd.TxStatus, error) {
	return t.str.TxStatus(ctx, args.TxID)
}
