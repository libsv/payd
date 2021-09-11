package models

import (
	"context"

	"github.com/libsv/go-bt/v2"
)

type PaymentService interface {
	Request(ctx context.Context, args PaymentRequestArgs) (*PaymentRequest, error)
	Send(ctx context.Context) error
}

type PaymentStore interface {
	Request(ctx context.Context, args PaymentRequestArgs) (*PaymentRequest, error)
	Submit(ctx context.Context) error
}

type PaymentRequestArgs struct {
	ID string
}

type Output struct {
	Amount      uint64 `json:"amount"`
	Script      string `json:"script"`
	Description string `json:"description"`
}

type MerchantData struct {
	Avatar           string                 `json:"avatar"`
	Name             string                 `json:"name"`
	Email            string                 `json:"email"`
	Address          string                 `json:"address"`
	PaymentReference string                 `json:"paymentReference"`
	ExtendedData     map[string]interface{} `json:"extendedData"`
}

type Fee struct {
	Data     bt.FeeUnit `json:"data"`
	Standard bt.FeeUnit `json:"standard"`
}

type PaymentRequest struct {
	Network      string       `json:"network"`
	Outputs      []Output     `json:"outputs"`
	CreatedAt    uint64       `json:"creationTimestamp"`
	ExpiresAt    uint64       `json:"expirationTimestamp"`
	PaymentURL   string       `json:"paymentURL"`
	Memo         string       `json:"memo"`
	MerchantData MerchantData `json:"merchantData"`
	Fee          Fee          `json:"fee"`
}
