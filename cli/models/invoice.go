package models

import (
	"context"
	"strconv"
	"time"
)

// InvoiceService interfaces an invoice service.
type InvoiceService interface {
	Invoice(ctx context.Context, args InvoiceGetArgs) (*Invoice, error)
	Invoices(ctx context.Context) (Invoices, error)
	Create(ctx context.Context, req InvoiceCreateRequest) (*Invoice, error)
	Delete(ctx context.Context, args InvoiceDeleteArgs) error
}

// InvoiceReader interfaces reading invoices.
type InvoiceReader interface {
	Invoice(ctx context.Context, args InvoiceGetArgs) (*Invoice, error)
	Invoices(ctx context.Context) (Invoices, error)
}

// InvoiceWriter interfaces writing invoices.
type InvoiceWriter interface {
	Create(ctx context.Context, req InvoiceCreateRequest) (*Invoice, error)
	Delete(ctx context.Context, args InvoiceDeleteArgs) error
}

// InvoiceReaderWriter interfaces reading and writing invoices.
type InvoiceReaderWriter interface {
	InvoiceReader
	InvoiceWriter
}

// Invoice a payment invoice.
type Invoice struct {
	PaymentID         string     `json:"id" yaml:"id"`
	Satoshis          uint64     `json:"satoshis" yaml:"satoshis"`
	Reference         *string    `json:"reference" yaml:"reference"`
	Description       *string    `json:"description" yaml:"description"`
	State             string     `json:"state" yaml:"state"`
	ExpiresAt         *time.Time `json:"expiresAt" yaml:"expiresAt"`
	PaymentReceivedAt *time.Time `json:"paymentReceivedAt" yaml:"paymentReceivedAt"`
	RefundTo          *string    `json:"refundTo" yaml:"refundTo"`
	RefundedAt        *time.Time `json:"refundedAt" yaml:"refundedAt"`
	CreatedAt         *time.Time `json:"createdAt" yaml:"createdAt"`
	UpdatedAt         *time.Time `json:"updatedAt" yaml:"updatedAt"`
	DeletedAt         *time.Time `json:"deletedAt" yaml:"deletedAt"`
}

// Invoice a payment invoice.

// Invoices a slice of *model.Invoice.
type Invoices []*Invoice

// InvoiceGetArgs the args for getting an invoice.
type InvoiceGetArgs struct {
	ID string
}

// InvoiceCreateRequest the request for creating an invoice.
type InvoiceCreateRequest struct {
	Account     string     `json:"-"`
	Satoshis    uint64     `json:"satoshis"`
	Reference   *string    `json:"reference"`
	Description *string    `json:"description"`
	ExpiresAt   *time.Time `json:"expiresAt"`
}

// InvoiceDeleteArgs the args for deleted an invoice.
type InvoiceDeleteArgs struct {
	ID string
}

// Columns builds column headers.
func (ii Invoices) Columns() []string {
	return []string{"ID", "Satoshis", "State", "ReceivedAt", "Ref"}
}

// Columns builds column headers.
func (i *Invoice) Columns() []string {
	return Invoices{i}.Columns()
}

// Rows builds a series of rows.
func (ii Invoices) Rows() [][]string {
	rows := make([][]string, len(ii))
	for i, inv := range ii {
		rows[i] = inv.Row()
	}
	return rows
}

// Rows builds a series of rows.
func (i *Invoice) Rows() [][]string {
	return Invoices{i}.Rows()
}

// Row builds a row.
func (i *Invoice) Row() []string {
	var pra string
	var ref string
	if i.PaymentReceivedAt != nil {
		pra = i.PaymentReceivedAt.String()
	}
	if i.Reference != nil {
		ref = *i.Reference
		if len(ref) > 50 {
			ref = ref[:50] + "..."
		}
	}
	return []string{i.PaymentID, strconv.FormatUint(i.Satoshis, 10), i.State, pra, ref}
}
