package ppctl

import (
	"context"
)

// PaymentRequest message used in BIP270.
// See https://github.com/moneybutton/bips/blob/master/bip-0270.mediawiki#paymentrequest
type PaymentRequest struct {
	// Network  Always set to "bitcoin" (but seems to be set to 'bitcoin-sv'
	// outside bip270 spec, see https://handcash.github.io/handcash-merchant-integration/#/merchant-payments)
	// {enum: bitcoin, bitcoin-sv, test}
	// Required.
	Network string `json:"network"`
	// Outputs Is an array of outputs. required, but can have zero elements.
	// Required.
	Outputs []*Output `json:"outputs"`
	// CreationTimestamp Unix timestamp (seconds since 1-Jan-1970 UTC) when the PaymentRequest was created.
	// Required.
	CreationTimestamp int64 `json:"creationTimestamp"`
	// ExpirationTimestamp Unix timestamp (UTC) after which the PaymentRequest should be considered invalid.
	// Optional.
	ExpirationTimestamp int64 `json:"expirationTimestamp,omitempty"`
	// PaymentURL secure HTTPS location where a Payment message (see below) will be sent to obtain a PaymentACK.
	// Maximum length is 4000 characters
	PaymentURL string `json:"paymentUrl"`
	// Memo Optional note that should be displayed to the customer, explaining what this PaymentRequest is for.
	// Maximum length is 50 characters.
	Memo string `json:"memo,omitempty"`
	// MerchantData contains arbitrary data that may be used by the payment host to identify the PaymentRequest.
	// May be omitted if the payment host does not need to associate Payments with PaymentRequest
	// or if they associate each PaymentRequest with a separate payment address.
	// Maximum length is 10000 characters.
	MerchantData *MerchantData `json:"merchantData,omitempty"`
}

// Output message used in BIP270.
// See https://github.com/moneybutton/bips/blob/master/bip-0270.mediawiki#output
type Output struct {
	// Amount is the number of satoshis to be paid.
	Amount uint64 `json:"amount"`
	// Script is a locking script where payment should be sent, formatted as a hexadecimal string.
	Script string `json:"script"`
	// Description, an optional description such as "tip" or "sales tax". Maximum length is 100 chars.
	Description string `json:"description"`
}

// MerchantData to be displayed to the user.
type MerchantData struct {
	// AvatarURL displays a canonical url to a merchants avatar.
	AvatarURL string `json:"avatarUrl,omitempty"`
	// MerchantName is a human readable string identifying the merchant.
	MerchantName string `json:"merchantName,omitempty"`
}

// PaymentRequestArgs are request arguments that can be passed to the service.
type PaymentRequestArgs struct {
	// PaymentID is an identifier for an invoice.
	PaymentID string
}

// PaymentRequestService can be implemented to enforce business rules
// and process in order to fulfill a PaymentRequest.
type PaymentRequestService interface {
	CreatePaymentRequest(ctx context.Context, args PaymentRequestArgs) (*PaymentRequest, error)
}
