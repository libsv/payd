package payd

import (
	"context"
	"time"

	"github.com/libsv/go-bt/v2"
	validator "github.com/theflyingcodr/govalidator"
)

// PaymentRequestArgs are used to create a new paymentRequest.
type PaymentRequestArgs struct {
	InvoiceID string `param:"invoiceID"`
}

// Validate will check that invoice arguments match expectations.
func (p *PaymentRequestArgs) Validate() error {
	return validator.New().
		Validate("invoiceID", validator.StrLength(p.InvoiceID, 1, 30)).
		Err()
}

// PaymentRequestResponse a payment request from dpp.
type PaymentRequestResponse struct {
	Network             string         `json:"network"`
	Destinations        DPPDestination `json:"destinations"`
	CreationTimestamp   time.Time      `json:"creationTimestamp"`
	ExpirationTimestamp time.Time      `json:"expirationTimestamp"`
	PaymentURL          string         `json:"paymentURL"`
	Memo                string         `json:"memo"`
	MerchantData        User           `json:"merchantData"`
	Fee                 *bt.FeeQuote   `json:"fees"`
	AncestryRequired    bool           `json:"ancestryRequired" example:"true"`
}

// PaymentRequestService will create and return a paymentRequest using the args provided.
type PaymentRequestService interface {
	PaymentRequest(ctx context.Context, args PaymentRequestArgs) (*PaymentRequestResponse, error)
}
