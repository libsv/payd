package client

import (
	"context"

	gopayd "github.com/libsv/payd"
)

// CreatePayment defines the body for requesting to create a payment.
type CreatePayment struct {
	Satoshis  uint64 `json:"satoshis"`
	ServerURL string `json:"serverUrl"`
}

// PaymentService interfaces with a payment service.
type PaymentService interface {
	CreatePayment(ctx context.Context, req CreatePayment) (*gopayd.PaymentACK, error)
}

// PaymentCreator interfaces creating a payment.
type PaymentCreator interface {
	Invoice(ctx context.Context, serverURL string, req gopayd.InvoiceCreate) (*gopayd.Invoice, error)
	RequestPayment(ctx context.Context, serverURL string, req gopayd.PaymentRequestArgs) (*gopayd.PaymentRequest, error)
	SendPayment(ctx context.Context, endpoint string, req gopayd.CreatePayment) (*gopayd.PaymentACK, error)
}
