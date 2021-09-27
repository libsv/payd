package service

import (
	"context"

	"github.com/pkg/errors"

	"github.com/libsv/payd"
)

type balance struct {
	store payd.BalanceReader
}

// NewBalance will setup and return the current balance of the wallet.
func NewBalance(store payd.BalanceReader) *balance {
	return &balance{store: store}
}

// Balance will return the current wallet balance.
func (b *balance) Balance(ctx context.Context) (*payd.Balance, error) {
	resp, err := b.store.Balance(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get balance")
	}
	return resp, nil
}
