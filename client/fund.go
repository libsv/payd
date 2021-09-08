package client

import (
	"context"

	"github.com/libsv/go-bt/v2"
	gopayd "github.com/libsv/payd"
)

type Fund struct {
	Outpoint      string `db:"outpoint" json:"-"`
	KeyName       string `db:"keyname" json:"-"`
	TxID          string `db:"txid" json:"txId"`
	Vout          int    `db:"vout" json:"vout"`
	LockingScript string `db:"lockingscript" json:"lockingScript"`
	SpendingTxID  string `db:"spendingtxid" json:"spendingTxId,omitempty"`
	Satoshis      uint64 `db:"satoshis" json:"satoshis"`
}

type FundsUnspentResponse struct {
	Balance uint64  `json:"balance"`
	Funds   []*Fund `json:"funds"`
}

type FundArgs struct {
	KeyName string `db:"keyname"`
}

type FundSeed struct {
	Amount float64 `json:"amount"`
}

type FundsCreate struct {
	TxID  string `db:"txid"`
	TxHex string `db:"txhex"`
	Funds []*Fund
}

type FundService interface {
	Seed(ctx context.Context, req FundSeed) (*gopayd.Transaction, error)
	FundsCreate(ctx context.Context, tx *bt.Tx) (*gopayd.Transaction, error)
	FundsUnspent(ctx context.Context) (*FundsUnspentResponse, error)
}

type FundReaderWriter interface {
	FundReader
	FundWriter
}

type FundReader interface {
	Funds(ctx context.Context, args FundArgs) ([]*Fund, error)
}

type FundWriter interface {
	FundsCreate(ctx context.Context, arg FundsCreate) (*gopayd.Transaction, error)
	FundSpend(ctx context.Context, args Fund) error
	FundsSpend(ctx context.Context, args []*Fund) error
}
