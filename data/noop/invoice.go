package noop

import (
	"context"

	"github.com/libsv/payd"
	"gopkg.in/guregu/null.v3"
)

// invoice is a no-op invoice that returns some stubbed data.
type invoice struct {
}

// NewInvoice will return a new instance of a noop invoice.
func NewInvoice() *invoice {
	return &invoice{}
}

// Invoice will return an invoice that matches the provided args.
func (i *invoice) Invoice(ctx context.Context, args payd.InvoiceArgs) (*payd.Invoice, error) {
	return &payd.Invoice{
		ID:       args.InvoiceID,
		Satoshis: 10000,
	}, nil
}

// Invoice will return an invoice that matches the provided args.
func (i *invoice) Invoices(ctx context.Context) ([]payd.Invoice, error) {
	return []payd.Invoice{{
		ID:       "noop-abc123",
		Satoshis: 10000,
	}}, nil
}

// Create will persist a new Invoice in the data store.
func (i *invoice) InvoiceCreate(ctx context.Context, req payd.InvoiceCreate) (*payd.Invoice, error) {
	return &payd.Invoice{
		ID:       req.Reference.ValueOrZero(),
		Satoshis: req.Satoshis,
	}, nil
}

// Update will update an invoice and return the result.
func (i *invoice) InvoiceUpdate(ctx context.Context, args payd.InvoiceUpdateArgs, req payd.InvoiceUpdatePaid) (*payd.Invoice, error) {
	return &payd.Invoice{
		ID:                args.InvoiceID,
		Satoshis:          10000,
		PaymentReceivedAt: null.TimeFrom(req.PaymentReceivedAt),
	}, nil
}

// Delete does nothing.
func (i *invoice) InvoiceDelete(ctx context.Context, args payd.InvoiceArgs) error {
	return nil
}
