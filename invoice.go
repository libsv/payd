package gopayd

import (
	"context"
	"time"

	"gopkg.in/guregu/null.v3"
)

// Invoice stores information related to a payment.
type Invoice struct {
	PaymentID         string    `json:"paymentID" db:"paymentID"`
	Satoshis          uint64    `json:"satoshis" db:"satoshis"`
	PaymentReceivedAt null.Time `json:"paymentReceivedAt" db:"paymentReceivedAt"`
}

// CreateInvoice is used to create a new invoice.
type CreateInvoice struct {
	// PaymentID is the unique identifier for a payment.
	PaymentID string `db:"paymentId"`
	Satoshis  uint64 `db:"satoshis"`
}

// UpdateInvoice can be used to update an invoice after it has been created.
type UpdateInvoice struct {
	PaymentReceivedAt time.Time `db:"paymentReceviedAt"`
}

// UpdateInvoiceArgs are used to identify the invoice to update.
type UpdateInvoiceArgs struct {
	PaymentID string
}

// InvoiceArgs contains argument/s to return a single invoice.
type InvoiceArgs struct {
	PaymentID string `db:"paymentId"`
}

// InvoiceReaderWriter can be implemented to support storing and retrieval of invoices.
type InvoiceReaderWriter interface {
	InvoiceWriter
	InvoiceReader
}

type InvoiceWriter interface {
	// Create will persist a new Invoice in the data store.
	Create(ctx context.Context, req CreateInvoice) (*Invoice, error)
	// Update will update an invoice matching the provided args with the requested changes.
	Update(ctx context.Context, args UpdateInvoiceArgs, req UpdateInvoice) (*Invoice, error)
}

type InvoiceReader interface {
	// Invoice will return an invoice that matches the provided args.
	Invoice(ctx context.Context, args InvoiceArgs) (*Invoice, error)
}
