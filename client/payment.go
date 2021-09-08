package client

import (
	"context"

	gopayd "github.com/libsv/payd"
)

type CreatePayment struct {
	Satoshis  uint64 `json:"satoshis"`
	ServerURL string `json:"serverUrl"`
}

type PaymentService interface {
	CreatePayment(ctx context.Context, req CreatePayment) (*gopayd.PaymentACK, error)
}

type PaymentCreator interface {
	Invoice(ctx context.Context, serverURL string, req gopayd.InvoiceCreate) (*gopayd.Invoice, error)
	RequestPayment(ctx context.Context, serverURL string, req gopayd.PaymentRequestArgs) (*gopayd.PaymentRequest, error)
	SendPayment(ctx context.Context, endpoint string, req gopayd.CreatePayment) (*gopayd.PaymentACK, error)
}
