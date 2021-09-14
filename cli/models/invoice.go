package models

import (
	"context"
	"strconv"
	"time"
)

type InvoiceService interface {
	Invoice(ctx context.Context, args InvoiceGetArgs) (*Invoice, error)
	Invoices(ctx context.Context) (Invoices, error)
	Create(ctx context.Context, req InvoiceCreateRequest) (*Invoice, error)
	Delete(ctx context.Context, args InvoiceDeleteArgs) error
}

type InvoiceReader interface {
	Invoice(ctx context.Context, args InvoiceGetArgs) (*Invoice, error)
	Invoices(ctx context.Context) (Invoices, error)
}

type InvoiceWriter interface {
	Create(ctx context.Context, req InvoiceCreateRequest) (*Invoice, error)
	Delete(ctx context.Context, args InvoiceDeleteArgs) error
}

type InvoiceReaderWriter interface {
	InvoiceReader
	InvoiceWriter
}

type Invoice struct {
	PaymentID         string     `json:"paymentID" yaml:"paymentID"`
	Satoshis          uint64     `json:"satoshis" yaml:"satoshis"`
	PaymentReceivedAt *time.Time `json:"paymentReceivedAt" yaml:"paymentReceivedAt"`
	RefundTo          *string    `json:"refundTo" yaml:"refundTo"`
}

type Invoices []*Invoice

type InvoiceGetArgs struct {
	ID string
}

type InvoiceCreateRequest struct {
	Satoshis uint64 `json:"satoshis"`
}

type InvoiceDeleteArgs struct {
	ID string
}

func (ii Invoices) Columns() []string {
	return []string{"ID", "Satoshis", "ReceivedAt", "RefundTo"}
}

func (i *Invoice) Columns() []string {
	return Invoices{i}.Columns()
}

func (ii Invoices) Rows() [][]string {
	rows := make([][]string, len(ii))
	for i, inv := range ii {
		rows[i] = inv.Row()
	}
	return rows
}

func (i *Invoice) Rows() [][]string {
	return Invoices{i}.Rows()
}

func (i *Invoice) Row() []string {
	var t string
	var r string
	if i.PaymentReceivedAt != nil {
		t = i.PaymentReceivedAt.String()
	}
	if i.RefundTo != nil {
		r = *i.RefundTo
	}
	return []string{i.PaymentID, strconv.FormatUint(i.Satoshis, 10), t, r}
}
