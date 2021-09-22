package service

import (
	"context"

	"github.com/libsv/payd/cli/models"
)

type invoiceSvc struct {
	irw models.InvoiceReaderWriter
}

// NewInvoiceService returns a new invoice service.
func NewInvoiceService(irw models.InvoiceReaderWriter) models.InvoiceService {
	return &invoiceSvc{irw: irw}
}

func (i *invoiceSvc) Invoice(ctx context.Context, args models.InvoiceGetArgs) (*models.Invoice, error) {
	return i.irw.Invoice(ctx, args)
}

func (i *invoiceSvc) Invoices(ctx context.Context) (models.Invoices, error) {
	return i.irw.Invoices(ctx)
}

func (i *invoiceSvc) Create(ctx context.Context, req models.InvoiceCreateRequest) (*models.Invoice, error) {
	return i.irw.Create(ctx, req)
}

func (i *invoiceSvc) Delete(ctx context.Context, args models.InvoiceDeleteArgs) error {
	return i.irw.Delete(ctx, args)
}
