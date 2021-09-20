package service

import (
	"context"

	gopayd "github.com/libsv/payd"
)

type destination struct {
	tRdr gopayd.TxoReader
}

func NewDestinationService(tRdr gopayd.TxoReader) gopayd.DestinationService {
	return &destination{tRdr: tRdr}
}

func (d *destination) Destinations(ctx context.Context, args gopayd.DestinationArgs) ([]*gopayd.Output, error) {
	if err := args.Validate(); err != nil {
		return nil, err
	}

	txos, err := d.tRdr.PartialTxoByPaymentID(ctx, gopayd.InvoiceArgs{
		PaymentID: args.PaymentID,
	})
	if err != nil {
		return nil, err
	}

	outputs := make([]*gopayd.Output, len(txos))
	for i, txo := range txos {
		outputs[i] = &gopayd.Output{
			Amount: txo.Satoshis,
			Script: txo.LockingScript,
		}
	}

	return outputs, nil
}
