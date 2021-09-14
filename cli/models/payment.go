package models

import (
	"context"
	"strconv"
	"time"

	"github.com/libsv/go-bc/spv"
	"github.com/libsv/go-bt/v2"
)

type PaymentService interface {
	Request(ctx context.Context, args PaymentRequestArgs) (*PaymentRequest, error)
	Send(ctx context.Context, args PaymentSendArgs) (*PaymentAck, error)
}

type PaymentStore interface {
	Request(ctx context.Context, args PaymentRequestArgs) (*PaymentRequest, error)
	Submit(ctx context.Context, args PaymentSendArgs) (*PaymentAck, error)
}

type PaymentRequestArgs struct {
	ID string
}

type PaymentSendArgs struct {
	PaymentRequest PaymentRequest `json:"-" yaml:"-"`
	Transaction    string         `json:"transaction" yaml:"transaction"`
	Memo           string         `json:"memo" yaml:"memo"`
	MerchantData   MerchantData   `json:"merchantData" yaml:"merchantData"`
	SPVEnvelope    spv.Envelope   `json:"spvEnvelope" yaml:"spvEnvelope"`
}

type Output struct {
	Amount      uint64 `json:"amount" yaml:"amount"`
	Script      string `json:"script" yaml:"script"`
	Description string `json:"description" yaml:"description"`
}

type MerchantData struct {
	Avatar           string                 `json:"avatar" yaml:"avatar"`
	Name             string                 `json:"name" yaml:"name"`
	Email            string                 `json:"email" yaml:"email"`
	Address          string                 `json:"address" yaml:"address"`
	PaymentReference string                 `json:"paymentReference" yaml:"paymentReference"`
	ExtendedData     map[string]interface{} `json:"extendedData" yaml:"extendedData"`
}

type Fee struct {
	Data     bt.Fee `json:"data" yaml:"data"`
	Standard bt.Fee `json:"standard" yaml:"standard"`
}

type PaymentRequest struct {
	Network      string       `json:"network" yaml:"network"`
	Outputs      []Output     `json:"outputs" yaml:"outputs"`
	CreatedAt    int64        `json:"creationTimestamp" yaml:"createdAt"`
	ExpiresAt    uint64       `json:"expirationTimestamp" yaml:"expiresAt"`
	PaymentURL   string       `json:"paymentURL" yaml:"paymentURL"`
	Memo         string       `json:"memo" yaml:"memo"`
	MerchantData MerchantData `json:"merchantData" yaml:"merchantData"`
	Fee          Fee          `json:"fee" yaml:"fee"`
}

type PaymentAck struct {
	Payment *PaymentSendArgs `json:"payment" yaml:"payment"`
	Memo    *string          `json:"memo" yaml:"memo"`
	Error   *int             `json:"error" yaml:"error"`
}

func (p PaymentRequest) Columns() []string {
	return []string{
		"Network", "Merchant", "PayToURL", "CreatedAt", "NumOutputs",
	}
}

func (p PaymentRequest) Rows() [][]string {
	return [][]string{p.Row()}
}

func (p PaymentRequest) Row() []string {
	t := time.Unix(p.CreatedAt, 0)
	return []string{
		p.Network,
		p.MerchantData.Name,
		p.PaymentURL,
		t.String(),
		strconv.FormatInt(int64(len(p.Outputs)), 10),
	}
}

func (p PaymentAck) Columns() []string {
	return []string{"TxID", "Merchant", "Payment Reference"}
}

func (p PaymentAck) Rows() [][]string {
	tx, _ := bt.NewTxFromString(p.Payment.Transaction)
	return [][]string{{
		tx.TxID(), p.Payment.MerchantData.Name, p.Payment.MerchantData.PaymentReference,
	}}
}
