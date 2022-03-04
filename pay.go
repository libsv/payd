package payd

import (
	"context"
	"net/url"

	"github.com/libsv/go-dpp"
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

// DPPOutput an output matching what a dpp server expects.
type DPPOutput struct {
	Amount      uint64 `json:"amount"`
	Script      string `json:"script"`
	Description string `json:"description"`
}

// DPPDestination defines a dpp payment destination object.
type DPPDestination struct {
	Outputs []DPPOutput `json:"outputs"`
}

// MerchantData dpp from a dpp server.
type MerchantData struct {
	Avatar           string                 `json:"avatar"`
	Name             string                 `json:"name"`
	Email            string                 `json:"email"`
	Address          string                 `json:"address"`
	PaymentReference string                 `json:"paymentReference"`
	ExtendedData     map[string]interface{} `json:"extendedData"`
}

// PaymentACK message used in BIP270.
// See https://github.com/moneybutton/bips/blob/master/bip-0270.mediawiki#paymentack
type PaymentACK struct {
	Payment PaymentCreate `json:"payment"`
	Memo    string        `json:"memo,omitempty"`
	// A number indicating why the transaction was not accepted. 0 or undefined indicates no error.
	// A 1 or any other positive integer indicates an error. The errors are left undefined for now;
	// it is recommended only to use “1” and to fill the memo with a textual explanation about why
	// the transaction was not accepted until further numbers are defined and standardised.
	Error int `json:"error,omitempty"`
}

// PayStrategy for registering different payment strategies.
type PayStrategy interface {
	PayService
	Register(svc PayService, names ...string) PayStrategy
}

// PayService for sending payments to another wallet.
type PayService interface {
	Pay(ctx context.Context, req PayRequest) (*dpp.PaymentACK, error)
}

// PayWriter will send a payment to another wallet or dpp server.
type PayWriter interface {
	Pay(ctx context.Context, req PayRequest) error
}
