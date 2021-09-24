package service

import (
	"context"

	"github.com/libsv/payd"
)

type pay struct {
	tWtr payd.TxoWriter
}

func NewPayService(tWtr payd.TxoWriter) payd.PayService {
	return &pay{tWtr: tWtr}
}

func (p *pay) Pay(ctx context.Context, req payd.PayRequest) error {
	// Get request from p4

	// Reserve utxos

	// Send request

	// Spend utxos?
	return nil
}
