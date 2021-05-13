package gopayd

import (
	"context"
	"errors"

	"github.com/libsv/go-bt"
	validator "github.com/theflyingcodr/govalidator"
	"gopkg.in/guregu/null.v3"
)

// CreatePayment is a Payment message used in BIP270.
// See https://github.com/moneybutton/bips/blob/master/bip-0270.mediawiki#payment
type CreatePayment struct {
	// Transaction is a valid, signed Bitcoin transaction that fully
	// pays the PaymentRequest.
	// The transaction is hex-encoded and must NOT be prefixed with "0x".
	Transaction string `json:"transaction"`
	// MerchantData is copied from PaymentDetails.merchantData.
	// Payment hosts may use invoice numbers or any other data they require to match Payments to PaymentRequests.
	// Note that malicious clients may modify the merchantData, so should be authenticated
	// in some way (for example, signed with a payment host-only key).
	// Maximum length is 10000 characters.
	MerchantData null.String `json:"merchantData,omitempty"`
	// RefundTo is a paymail to send a refund to should a refund be necessary.
	// Maximum length is 100 characters
	// TODO - we're not handling paymail in V1 so this will always be empty
	RefundTo null.String `json:"refundTo,omitempty"`
	// Memo is a plain-text note from the customer to the payment host.
	Memo string `json:"memo,omitempty"`
}

// Validate will ensure the users request is correct.
func (c CreatePayment) Validate() validator.ErrValidation {
	v := validator.New().
		Validate("transaction",
			func() error {
				if _, err := bt.NewTxFromString(c.Transaction); err != nil {
					return errors.New("not a valid bitcoin transaction")
				}
				return nil
			},
		)
	if c.MerchantData.Valid {
		v = v.Validate("merchantData", validator.Length(c.MerchantData.String, 0, 10000))
	}
	if c.RefundTo.Valid {
		v = v.Validate("refundTo", validator.Length(c.RefundTo.String, 0, 100))
	}
	return v
}

// PaymentACK message used in BIP270.
// See https://github.com/moneybutton/bips/blob/master/bip-0270.mediawiki#paymentack
type PaymentACK struct {
	Payment *CreatePayment `json:"payment"`
	Memo    string         `json:"memo,omitempty"`
	// A number indicating why the transaction was not accepted. 0 or undefined indicates no error.
	// A 1 or any other positive integer indicates an error. The errors are left undefined for now;
	// it is recommended only to use “1” and to fill the memo with a textual explanation about why
	// the transaction was not accepted until further numbers are defined and standardised.
	Error int `json:"error,omitempty"`
	// TODO: check anypay https://docs.anypayinc.com/pay-protocol/overview
	// and consider deleting success field because seems they don't actually
	// return it, tried sending a rubbish tx and didn't get back success: "false"
	// - only got back proper bip270 paymentack response (like above)
	Success string `json:"success,omitempty"` // 'true' or 'false' string
}

// CreatePaymentArgs identifies the paymentID used for the payment.
type CreatePaymentArgs struct {
	PaymentID string
}

// PaymentService enforces business rules when creating payments.
type PaymentService interface {
	CreatePayment(ctx context.Context, args CreatePaymentArgs, req CreatePayment) (*PaymentACK, error)
}

// PaymentWriter reads payment info from a data store.
type PaymentWriter interface {
	// CompletePayment when implemented can store the tx and utxos as well as update the invoice as paid.
	StoreUtxos(ctx context.Context, req CreateTransaction) (*Transaction, error)
}

// PaymentSender will broadcast a payment to a network.
type PaymentSender interface {
	Send(ctx context.Context, args CreatePaymentArgs, req CreatePayment) error
}
