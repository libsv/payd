package models

import (
	"context"
	"strconv"
	"time"

	gopayd "github.com/libsv/payd"
)

type FundAddArgs struct {
	TxHex   string `json:"tx"`
	Account string `json:"account"`
}

type FundGetArgs struct {
	Amount  uint64
	Account string
}

type FundSpendArgs struct {
	SpendingTx string `json:"spendingTx"`
	Account    string `json:"-"`
}

type FundsRequest struct {
	Fee Fee `json:"fee"`
}

type Fund struct {
	TxID          string     `json:"txId" yaml:"txId"`
	Vout          int        `json:"vout" yaml:"vout"`
	LockingScript string     `json:"lockingScript" yaml:"lockingScript"`
	Satoshis      uint64     `json:"satoshis" yaml:"satoshis"`
	SpentAt       *time.Time `json:"spentAt" yaml:"spentAt"`
	SpendingTxID  *string    `json:"spendingTxId" yaml:"spendingTxId"`
	CreatedAt     *time.Time `json:"createdAt" yaml:"createdAt"`
	ModifiedAt    *time.Time `json:"modifiedAt" yaml:"modifiedAt"`
}

type Funds []Fund

type FundService interface {
	Add(ctx context.Context, args FundAddArgs) (Funds, error)
	Get(ctx context.Context, args FundGetArgs) (Funds, error)
	GetAmount(ctx context.Context, req FundsRequest, args FundGetArgs) (*gopayd.FundsGetResponse, error)
	Spend(ctx context.Context, args FundSpendArgs) error
}

type FundStore interface {
	Add(ctx context.Context, args FundAddArgs) (Funds, error)
	Get(ctx context.Context, args FundGetArgs) (Funds, error)
	GetAmount(ctx context.Context, req FundsRequest, args FundGetArgs) (*gopayd.FundsGetResponse, error)
	Spend(ctx context.Context, args FundSpendArgs) error
}

func (ff Funds) Columns() []string {
	return []string{"TxID", "Vout", "Satoshis", "Spent"}
}

func (ff Funds) Rows() [][]string {
	rows := make([][]string, 0)
	for _, f := range ff {
		rows = append(rows, f.Row())
	}

	return rows
}

func (f Fund) Row() []string {
	spent := "N"
	if f.SpentAt != nil {
		spent = "Y"
	}
	return []string{
		f.TxID,
		strconv.Itoa(f.Vout),
		strconv.FormatUint(f.Satoshis, 10),
		spent,
	}
}
