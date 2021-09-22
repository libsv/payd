package models

import (
	"context"
	"strconv"
	"time"

	"github.com/libsv/go-bc/spv"
	"github.com/libsv/go-bt/v2"
)

// PaymentService interfaces a service for payments.
type PaymentService interface {
	Request(ctx context.Context, args PaymentRequestArgs) (*PaymentRequest, error)
	Send(ctx context.Context, args PaymentSendArgs) (*PaymentAck, error)
}

// PaymentStore interfaces a store for payments.
type PaymentStore interface {
	Request(ctx context.Context, args PaymentRequestArgs) (*PaymentRequest, error)
	Submit(ctx context.Context, args PaymentSendArgs) (*PaymentAck, error)
}

// PaymentRequestArgs the args for requesting a payment.
type PaymentRequestArgs struct {
	ID    string
	PayTo string `json:"-" yaml:"-"`
}

// PaymentSendArgs the args for sending a payment.
type PaymentSendArgs struct {
	Account        string         `json:"-" yaml:"-"`
	PaymentRequest PaymentRequest `json:"-" yaml:"-"`
	Tx             string         `json:"-" yaml:"-"`

	Transaction  string       `json:"transaction" yaml:"transaction"`
	Memo         string       `json:"memo" yaml:"memo"`
	MerchantData MerchantData `json:"merchantData" yaml:"merchantData"`
	SPVEnvelope  spv.Envelope `json:"spvEnvelope" yaml:"spvEnvelope"`
}

// MerchantData merchant data.
type MerchantData struct {
	Avatar           string                 `json:"avatar" yaml:"avatar"`
	Name             string                 `json:"name" yaml:"name"`
	Email            string                 `json:"email" yaml:"email"`
	Address          string                 `json:"address" yaml:"address"`
	PaymentReference string                 `json:"paymentReference" yaml:"paymentReference"`
	ExtendedData     map[string]interface{} `json:"extendedData" yaml:"extendedData"`
}

// PaymentRequest a payment request.
type PaymentRequest struct {
	Network string `json:"network" yaml:"network"`
	Outputs []struct {
		Amount      uint64 `json:"amount" yaml:"amount"`
		Script      string `json:"script" yaml:"script"`
		Description string `json:"description" yaml:"description"`
	} `json:"outputs" yaml:"outputs"`
	CreatedAt    time.Time    `json:"creationTimestamp" yaml:"createdAt"`
	ExpiresAt    time.Time    `json:"expirationTimestamp" yaml:"expiresAt"`
	PaymentURL   string       `json:"paymentURL" yaml:"paymentURL"`
	Memo         string       `json:"memo" yaml:"memo"`
	MerchantData MerchantData `json:"merchantData" yaml:"merchantData"`
	Fee          *bt.FeeQuote `json:"fee" yaml:"fee"`
}

// PaymentAck an acknowledgement of a payment.
type PaymentAck struct {
	Payment *PaymentSendArgs `json:"payment" yaml:"payment"`
	Memo    *string          `json:"memo" yaml:"memo"`
	Error   *int             `json:"error" yaml:"error"`
}

// Columns builds column headers.
func (p PaymentRequest) Columns() []string {
	return []string{
		"Network", "Merchant", "PayToURL", "CreatedAt", "NumOutputs",
	}
}

// Rows builds a series of rows.
func (p PaymentRequest) Rows() [][]string {
	return [][]string{p.Row()}
}

// Row builds a row.
func (p PaymentRequest) Row() []string {
	return []string{
		p.Network,
		p.MerchantData.Name,
		p.PaymentURL,
		p.CreatedAt.String(),
		strconv.FormatInt(int64(len(p.Outputs)), 10),
	}
}

// Columns builds column headers.
func (p PaymentAck) Columns() []string {
	return []string{"TxID", "Merchant", "Payment Reference"}
}

// Rows builds a series of rows.
func (p PaymentAck) Rows() [][]string {
	tx, _ := bt.NewTxFromString(p.Payment.Transaction)
	return [][]string{{
		tx.TxID(), p.Payment.MerchantData.Name, p.Payment.MerchantData.PaymentReference,
	}}
}
