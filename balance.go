package gopayd

import (
	"context"
)

type Balance struct {
	Satoshis uint64 `json:"satoshis" db:"satoshis"`
}

type BalanceService interface {
	Balance(ctx context.Context) (*Balance, error)
}

type BalanceReader interface {
	Balance(ctx context.Context) (*Balance, error)
}
