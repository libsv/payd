package gopayd

import (
	"context"

	"github.com/libsv/go-bt/v2"
	validator "github.com/theflyingcodr/govalidator"
	"gopkg.in/guregu/null.v3"
)

// FundService interfaces funds.
type FundService interface {
	FundsAdd(ctx context.Context, req FundAddRequest) ([]*Txo, error)
	FundsGet(ctx context.Context, args FundsGetArgs) ([]Txo, error)
	FundsSpend(ctx context.Context, req FundsSpendReq, args FundsSpendArgs) error
}

// FundStore interfaces a fund store.
type FundStore interface {
	StoreFund(ctx context.Context, req StoreFundRequest) error
	Funds(ctx context.Context, args FundsGetArgs) ([]Txo, error)
	SpendFunds(ctx context.Context, req *FundsSpendReq, args FundsSpendArgs) error
}

// FundAddRequest the request for adding funds.
type FundAddRequest struct {
	Tx      string `json:"tx"`
	Account string `json:"account"`
}

// Validate validates.
func (f FundAddRequest) Validate() error {
	v := validator.New().
		Validate("tx", validator.NotEmpty(f.Tx), validator.IsHex(f.Tx),
			func() error {
				_, err := bt.NewTxFromString(f.Tx)
				return err
			}).
		Validate("account", validator.NotEmpty(f.Account))

	return v.Err()
}

// FundsGetArgs the args for getting funds.
type FundsGetArgs struct {
	Account string
}

// FundsSpendArgs the args for spending funds.
type FundsSpendArgs struct {
	Account string
}

// Validate validates.
func (f FundsSpendArgs) Validate() error {
	return validator.New().Validate("account", validator.NotEmpty(f.Account)).Err()
}

// FundsSpendReq the request for spending funds.
type FundsSpendReq struct {
	SpendingTxID string `json:"spendingTxId"`
	SpendingTx   string `json:"spendingTx"`
	Txos         []Txo  `json:"-"`
}

// Validate validates.
func (f FundsSpendReq) Validate() error {
	return validator.New().Validate("spendingTx",
		validator.NotEmpty(f.SpendingTx),
		validator.IsHex(f.SpendingTx),
		func() error {
			_, err := bt.NewTxFromString(f.SpendingTx)
			return err
		}).Err()
}

// StoreFundRequest the request for storing a fund in the db.
type StoreFundRequest struct {
	TxID      string      `db:"txid"`
	TxHex     string      `db:"txhex"`
	PaymentID null.String `db:"paymentID"`
	Txos      []*Txo
}
