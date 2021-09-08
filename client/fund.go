package client

import (
	"context"

	"github.com/libsv/go-bt/v2"
	gopayd "github.com/libsv/payd"
)

// Fund defines a fund.
type Fund struct {
	Outpoint      string `db:"outpoint" json:"-"`
	KeyName       string `db:"keyname" json:"-"`
	TxID          string `db:"txid" json:"txId"`
	Vout          int    `db:"vout" json:"vout"`
	LockingScript string `db:"lockingscript" json:"lockingScript"`
	SpendingTxID  string `db:"spendingtxid" json:"spendingTxId,omitempty"`
	Satoshis      uint64 `db:"satoshis" json:"satoshis"`
}

// FundsUnspentResponse defines the response for requesting to view unspent funds.
type FundsUnspentResponse struct {
	Balance uint64  `json:"balance"`
	Funds   []*Fund `json:"funds"`
}

// FundArgs defines the arguments for retrieving a fund.
type FundArgs struct {
	KeyName string `db:"keyname"`
}

// FundSeed defines the request for seeding a wallet with funds.
type FundSeed struct {
	Amount float64 `json:"amount"`
}

// FundsCreate defines the request for creating a fund.
type FundsCreate struct {
	TxID  string `db:"txid"`
	TxHex string `db:"txhex"`
	Funds []*Fund
}

// FundService interfaces a fund service.
type FundService interface {
	Seed(ctx context.Context, req FundSeed) (*gopayd.Transaction, error)
	FundsCreate(ctx context.Context, tx *bt.Tx) (*gopayd.Transaction, error)
	FundsUnspent(ctx context.Context) (*FundsUnspentResponse, error)
}

// FundReaderWriter interfaces a fund store.
type FundReaderWriter interface {
	FundReader
	FundWriter
}

// FundReader interfaces reading a fund from a store.
type FundReader interface {
	Funds(ctx context.Context, args FundArgs) ([]*Fund, error)
}

// FundWriter interfaces writing a fund to a store.
type FundWriter interface {
	FundsCreate(ctx context.Context, arg FundsCreate) (*gopayd.Transaction, error)
	FundSpend(ctx context.Context, args Fund) error
	FundsSpend(ctx context.Context, args []*Fund) error
}
