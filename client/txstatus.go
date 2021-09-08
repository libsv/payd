package client

import (
	"context"

	gopayd "github.com/libsv/payd"
)

// TxStatusReq contains the request body for requesting a tx status.
type TxStatusReq struct {
	ServerURL string `json:"serverUrl"`
}

// TxStatusArgs contain the args for requesting a tx status.
type TxStatusArgs struct {
	TxID string `param:"txid"`
}

// TxStatusService interfaces with a txstatus service.
type TxStatusService interface {
	Status(ctx context.Context, req TxStatusReq, args TxStatusArgs) (*gopayd.TxStatus, error)
}

// TxStatusStore interfaces with a tx status store.
type TxStatusStore interface {
	TxStatus(ctx context.Context, url, txID string) (*gopayd.TxStatus, error)
}
