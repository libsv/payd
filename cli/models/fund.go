package models

import (
	"context"
	"strconv"
	"time"
)

// FundAddArgs the args for adding a fund.
type FundAddArgs struct {
	TxHex   string `json:"tx"`
	Account string `json:"account"`
}

// FundGetArgs the args for getting a fund.
type FundGetArgs struct {
	Account string
}

// FundSpendArgs the args for spending a fund.
type FundSpendArgs struct {
	SpendingTx string `json:"spendingTx"`
	Account    string `json:"-"`
}

// Fund a fund.
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

// Funds a slice of models.Fund.
type Funds []Fund

// FundService interfaces a fund service.
type FundService interface {
	Add(ctx context.Context, args FundAddArgs) (Funds, error)
	Get(ctx context.Context, args FundGetArgs) (Funds, error)
	Spend(ctx context.Context, args FundSpendArgs) error
}

// FundStore interfaces a fund store.
type FundStore interface {
	Add(ctx context.Context, args FundAddArgs) (Funds, error)
	Get(ctx context.Context, args FundGetArgs) (Funds, error)
	Spend(ctx context.Context, args FundSpendArgs) error
}

// Columns builds column headers.
func (ff Funds) Columns() []string {
	return []string{"TxID", "Vout", "Satoshis", "Spent"}
}

// Rows builds a series of rows.
func (ff Funds) Rows() [][]string {
	rows := make([][]string, 0)
	for _, f := range ff {
		rows = append(rows, f.Row())
	}

	return rows
}

// Row builds a row.
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
