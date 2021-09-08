package client

import (
	"context"

	gopayd "github.com/libsv/payd"
)

// TxStatusArgs contain the args for requesting a tx status.
type TxStatusArgs struct {
	TxID string `param:"txid"`
}

// TxStatusService interfaces with a txstatus service.
type TxStatusService interface {
	Status(ctx context.Context, args TxStatusArgs) (*gopayd.TxStatus, error)
}

// TxStatusStore interfaces with a tx status store.
type TxStatusStore interface {
	TxStatus(ctx context.Context, txID string) (*gopayd.TxStatus, error)
}
