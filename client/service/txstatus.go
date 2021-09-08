package service

import (
	"context"

	gopayd "github.com/libsv/payd"
	"github.com/libsv/payd/client"
)

type txstatus struct {
	str client.TxStatusStore
}

func NewTxStatusService(str client.TxStatusStore) *txstatus {
	return &txstatus{str: str}
}

func (t *txstatus) Status(ctx context.Context, req client.TxStatusReq, args client.TxStatusArgs) (*gopayd.TxStatus, error) {
	return t.str.TxStatus(ctx, req.ServerURL, args.TxID)
}
