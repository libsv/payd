package models

import (
	"context"
	"strconv"
)

// Balance holds balance informaiton.
type Balance struct {
	Satoshis uint64 `json:"satoshis" yaml:"satoshis"`
}

// BalanceService interfaces with balances.
type BalanceService interface {
	BalanceReader
}

// BalanceReader reads balances from a store.
type BalanceReader interface {
	Balance(ctx context.Context) (*Balance, error)
}

// Columns builds column headers.
func (b Balance) Columns() []string {
	return []string{"Satoshis"}
}

// Rows builds a series of rows.
func (b Balance) Rows() [][]string {
	return [][]string{{strconv.FormatUint(b.Satoshis, 10)}}
}
