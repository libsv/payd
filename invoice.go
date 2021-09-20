package gopayd

import (
	"context"
	"time"

	"github.com/libsv/go-bt"
	validator "github.com/theflyingcodr/govalidator"
	"gopkg.in/guregu/null.v3"
)

// Invoice identifies a single payment request from this payd wallet,
// it states the amount, id and optional refund address. This indicate
// we are requesting n satoshis in payment.
type Invoice struct {
	// InvoiceID is a unique identifier for an invoice and can be used
	// to lookup a single invoice.
	InvoiceID string `json:"invoiceID" db:"invoice_id"`
	// PaymentReference is an identifier that can be used to link the
	// PayD invoice with an external system.
	PaymentReference null.String `json:"paymentReference" db:"payment_reference"`
	// Description is an optional text field that can have some further info
	// like 'invoice for oranges'.
	Description null.String `json:"description" db:"description"`
	// Satoshis is the total amount this invoice is to pay.
	Satoshis uint64 `json:"satoshis" db:"satoshis"`
	// PaymentReceivedAt will be set when this invoice has been paid and
	// states when the payment was received in UTC time.
	PaymentReceivedAt null.Time `json:"paymentReceivedAt" db:"payment_received_at"`
	// RefundTo is an optional paymail address that can be used to refund the
	// customer if required.
	RefundTo null.String `json:"refundTo" db:"refund_to"`
}

// InvoiceCreate is used to create a new invoice.
type InvoiceCreate struct {
	// Satoshis is the total amount this invoice is to pay.
	Satoshis uint64 `json:"satoshis" db:"satoshis"`
	// PaymentReference is an identifier that can be used to link the
	// payd invoice with an external system.
	// MaxLength is 32 characters.
	PaymentReference null.String `json:"paymentReference" db:"payment_reference"`
	// Description is an optional text field that can have some further info
	// like 'invoice for oranges'.
	// MaxLength is 1024 characters.
	Description null.String `json:"description" db:"description"`
}

// Validate will check that InvoiceCreate params match expectations.
func (i InvoiceCreate) Validate() validator.ErrValidation {
	return validator.New().
		Validate("satoshis", validator.MinUInt64(i.Satoshis, bt.DustLimit)).
		Validate("description", validator.Length(i.Description.ValueOrZero(), 0, 1024)).
		Validate("paymentReference", validator.Length(i.PaymentReference.ValueOrZero(), 0, 32))
}

// InvoiceUpdate can be used to update an invoice after it has been created.
type InvoiceUpdate struct {
	PaymentReceivedAt time.Time   `db:"paymentReceviedAt"`
	RefundTo          null.String `db:"refundTo"`
}

// InvoiceUpdateArgs are used to identify the invoice to update.
type InvoiceUpdateArgs struct {
	PaymentID string
}

// InvoiceArgs contains argument/s to return a single invoice.
type InvoiceArgs struct {
	PaymentID string `param:"paymentID" db:"paymentID"`
}

// Validate will check that invoice arguments match expectations.
func (i *InvoiceArgs) Validate() validator.ErrValidation {
	return validator.New().Validate("paymentID", validator.Length(i.PaymentID, 1, 30))
}

// InvoiceService defines a service for managing invoices.
type InvoiceService interface {
	Invoice(ctx context.Context, args InvoiceArgs) (*Invoice, error)
	Invoices(ctx context.Context) ([]Invoice, error)
	Create(ctx context.Context, req InvoiceCreate) (*Invoice, error)
	Delete(ctx context.Context, args InvoiceArgs) error
}

// InvoiceReaderWriter can be implemented to support storing and retrieval of invoices.
type InvoiceReaderWriter interface {
	InvoiceWriter
	InvoiceReader
}

// InvoiceWriter defines a data store used to write invoice data.
type InvoiceWriter interface {
	// Create will persist a new Invoice in the data store.
	Create(ctx context.Context, req InvoiceCreate) (*Invoice, error)
	// Update will update an invoice matching the provided args with the requested changes.
	Update(ctx context.Context, args InvoiceUpdateArgs, req InvoiceUpdate) (*Invoice, error)
	Delete(ctx context.Context, args InvoiceArgs) error
}

// InvoiceReader defines a data store used to read invoice data.
type InvoiceReader interface {
	// Invoice will return an invoice that matches the provided args.
	Invoice(ctx context.Context, args InvoiceArgs) (*Invoice, error)
	// Invoices returns all currently stored invoices TODO: update to support search args
	Invoices(ctx context.Context) ([]Invoice, error)
}
