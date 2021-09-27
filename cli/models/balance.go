package models

import (
	"context"
	"strconv"
)

type Balance struct {
	Satoshis uint64 `json:"satoshis" yaml:"satoshis"`
}

type BalanceService interface {
	BalanceReader
}

type BalanceReader interface {
	Balance(ctx context.Context) (*Balance, error)
}

func (b Balance) Columns() []string {
	return []string{"Satoshis"}
}

func (b Balance) Rows() [][]string {
	return [][]string{{strconv.FormatUint(b.Satoshis, 10)}}
}
