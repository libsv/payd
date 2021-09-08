package client

import (
	"context"

	gopayd "github.com/libsv/payd"
)

type TxStatusReq struct {
	ServerURL string `json:"serverUrl"`
}

type TxStatusArgs struct {
	TxID string `param:"txid"`
}

type TxStatusService interface {
	Status(ctx context.Context, req TxStatusReq, args TxStatusArgs) (*gopayd.TxStatus, error)
}

type TxStatusStore interface {
	TxStatus(ctx context.Context, url, txID string) (*gopayd.TxStatus, error)
}
