package payd

import (
	"context"
	"net/url"
	"time"

	"github.com/libsv/go-bt/v2"
	validator "github.com/theflyingcodr/govalidator"
)

// PayRequest a request for making a payment.
type PayRequest struct {
	PayToURL string `json:"payToURL"`
}

// Validate validates the request.
func (p PayRequest) Validate() error {
	return validator.New().Validate("payToURL", func() error {
		_, err := url.Parse(p.PayToURL)
		return err
	}).Err()
}

// P4Output an output matching what a p4 server expects.
type P4Output struct {
	Amount      uint64 `json:"amount"`
	Script      string `json:"script"`
	Description string `json:"description"`
}

// PaymentRequestResponse a payment request from p4.
type PaymentRequestResponse struct {
	Network             string     `json:"network"`
	Outputs             []P4Output `json:"outputs"`
	CreationTimestamp   time.Time  `json:"creationTimestamp"`
	ExpirationTimestamp time.Time  `json:"expirationTimestamp"`
	PaymentURL          string     `json:"paymentURL"`
	Memo                string     `json:"memo"`
	MerchantData        struct {
		Avatar           string            `json:"avatar"`
		Name             string            `json:"name"`
		Email            string            `json:"email"`
		Address          string            `json:"address"`
		PaymentReference string            `json:"paymentReference"`
		ExtendedData     map[string]string `json:"extendedData"`
	} `json:"merchantData"`
	Fee *bt.FeeQuote `json:"fee"`
}

// PaymentACK an ack response from P4.
type PaymentACK struct {
	Payment Payment `json:"payment"`
	Memo    string  `json:"memo"`
}

// PayService for sending payments to another wallet.
type PayService interface {
	Pay(ctx context.Context, req PayRequest) (*PaymentACK, error)
}
