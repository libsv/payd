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
	// ID is a unique identifier for an invoice and can be used
	// to lookup a single invoice.
	ID string `json:"id" db:"invoice_id"`
	// Reference is an identifier that can be used to link the
	// PayD invoice with an external system.
	Reference null.String `json:"reference" db:"payment_reference"`
	// Description is an optional text field that can have some further info
	// like 'invoice for oranges'.
	Description null.String `json:"description" db:"description"`
	// Satoshis is the total amount this invoice is to pay.
	Satoshis uint64 `json:"satoshis" db:"satoshis"`
	// ExpiresAt is an optional param that can be passed to set an expiration
	// date on an invoice, after which, payments will not be accepted.
	ExpiresAt null.Time `json:"expiresAt" db:"expires_at"`
	// PaymentReceivedAt will be set when this invoice has been paid and
	// states when the payment was received in UTC time.
	PaymentReceivedAt null.Time `json:"paymentReceivedAt" db:"payment_received_at"`
	// RefundTo is an optional paymail address that can be used to refund the
	// customer if required.
	RefundTo null.String `json:"refundTo" db:"refund_to"`
	// RefundedAt if this payment has been refunded, this date will be set
	// to the UTC time of the refund.
	RefundedAt null.Time `json:"refundedAt" db:"refunded_at"`
	// State is the current status of the invoice.
	State string `json:"state" db:"state" enums:"pending,paid,refunded,deleted"`
	// CreatedAt is the UTC time the invoice was created.
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	// UpdatedAt is the UTC time the invoice was updated.
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
	// DeletedAt is the date the invoice was removed.
	DeletedAt null.Time `json:"deletedAt,omitempty" db:"deleted_at"`
}

// InvoiceCreate is used to create a new invoice.
type InvoiceCreate struct {
	InvoiceID string `json:"-" db:"invoice_id"`
	// Satoshis is the total amount this invoice is to pay.
	Satoshis uint64 `json:"satoshis" db:"satoshis"`
	// Reference is an identifier that can be used to link the
	// payd invoice with an external system.
	// MaxLength is 32 characters.
	Reference null.String `json:"reference" db:"payment_reference"`
	// Description is an optional text field that can have some further info
	// like 'invoice for oranges'.
	// MaxLength is 1024 characters.
	Description null.String `json:"description" db:"description"`
	// ExpiresAt is an optional param that can be passed to set an expiration
	// date on an invoice, after which, payments will not be accepted.
	ExpiresAt null.Time `json:"expiresAt" db:"expires_at"`
}

// Validate will check that InvoiceCreate params match expectations.
func (i InvoiceCreate) Validate() validator.ErrValidation {
	return validator.New().
		Validate("satoshis", validator.MinUInt64(i.Satoshis, bt.DustLimit)).
		Validate("description", validator.Length(i.Description.ValueOrZero(), 0, 1024)).
		Validate("paymentReference", validator.Length(i.Reference.ValueOrZero(), 0, 32))
}

// InvoiceUpdatePaid can be used to update an invoice after it has been created.
type InvoiceUpdatePaid struct {
	PaymentReceivedAt time.Time `db:"payment_received_at"`
}

// InvoiceUpdateRefunded can be used to update an invoice state to refunded.
type InvoiceUpdateRefunded struct {
	// RefundTo will set an invoice as refunded.
	RefundTo null.String `db:"refund_to"`
	// RefundedAt if this payment has been refunded, this date will be set
	// to the UTC time of the refund.
	RefundedAt null.Time `json:"refundedAt" db:"refunded_at"`
}

// InvoiceUpdateArgs are used to identify the invoice to update.
type InvoiceUpdateArgs struct {
	InvoiceID string
}

// InvoiceArgs contains argument/s to return a single invoice.
type InvoiceArgs struct {
	InvoiceID string `param:"invoiceID" db:"invoice_id"`
}

// Validate will check that invoice arguments match expectations.
func (i *InvoiceArgs) Validate() validator.ErrValidation {
	return validator.New().Validate("invoiceID", validator.Length(i.InvoiceID, 1, 30))
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
	Update(ctx context.Context, args InvoiceUpdateArgs, req InvoiceUpdatePaid) (*Invoice, error)
	// Delete will remove an invoice from the data store, depending on implementation this could
	// be a hard or soft delete.
	Delete(ctx context.Context, args InvoiceArgs) error
}

// InvoiceReader defines a data store used to read invoice data.
type InvoiceReader interface {
	// Invoice will return an invoice that matches the provided args.
	Invoice(ctx context.Context, args InvoiceArgs) (*Invoice, error)
	// Invoices returns all currently stored invoices TODO: update to support search args
	Invoices(ctx context.Context) ([]Invoice, error)
}
