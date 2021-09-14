package gopayd

import (
	"context"

	"github.com/libsv/go-bt/v2"
	validator "github.com/theflyingcodr/govalidator"
	"gopkg.in/guregu/null.v3"
)

type FundService interface {
	FundsAdd(ctx context.Context, req FundAddRequest) ([]*Txo, error)
	FundsGet(ctx context.Context, args FundsGetArgs) ([]Txo, error)
	FundsGetAmount(ctx context.Context, req FundsRequest, args FundsGetArgs) (*FundsGetResponse, error)
	FundsSpend(ctx context.Context, req FundsSpendReq, args FundsSpendArgs) error
}

type FundAddRequest struct {
	Tx      string `json:"tx"`
	Account string `json:"account"`
}

func (f FundAddRequest) Validate() error {
	v := validator.New()

	v.Validate("tx", validator.NotEmpty(f.Tx), validator.IsHex(f.Tx),
		func() error {
			_, err := bt.NewTxFromString(f.Tx)
			return err
		})

	v.Validate("account", validator.NotEmpty(f.Account))

	return v.Err()
}

type FundsGetArgs struct {
	Amount  uint64
	Account string
}

type FundsRequest struct {
	Fee struct {
		Data     bt.Fee `json:"data"`
		Standard bt.Fee `json:"standard"`
	} `json:"fee"`
}

type FundsSpendArgs struct {
	Account string
}

func (f FundsSpendArgs) Validate() error {
	return validator.New().Validate("account", validator.NotEmpty(f.Account)).Err()
}

type FundsSpendReq struct {
	SpendingTxID string `json:"spendingTxId"`
	SpendingTx   string `json:"spendingTx"`
	Txos         []Txo  `json:"-"`
}

func (f FundsSpendReq) Validate() error {
	return validator.New().Validate("spendingTx",
		validator.NotEmpty(f.SpendingTx),
		validator.IsHex(f.SpendingTx),
		func() error {
			_, err := bt.NewTxFromString(f.SpendingTx)
			return err
		}).Err()
}

type FundsGetResponse struct {
	Surplus uint64 `json:"surplus"`
	Funds   []Txo  `json:"funds"`
}

type StoreFundRequest struct {
	TxID      string      `db:"txid"`
	TxHex     string      `db:"txhex"`
	PaymentID null.String `db:"paymentID"`
	Txos      []*Txo
}

type FundStore interface {
	StoreFund(ctx context.Context, req StoreFundRequest) error
	Funds(ctx context.Context, args FundsGetArgs) ([]Txo, error)
	SpendFunds(ctx context.Context, req *FundsSpendReq, args FundsSpendArgs) error
}
