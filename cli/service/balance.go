package service

import (
	"context"

	"github.com/libsv/payd/cli/models"
)

type balance struct {
	rdr models.BalanceReader
}

// NewBalanceService returns a balance service.
func NewBalanceService(rdr models.BalanceReader) models.BalanceService {
	return &balance{rdr: rdr}
}

// Balance request to view a balance.
func (b *balance) Balance(ctx context.Context) (*models.Balance, error) {
	return b.rdr.Balance(ctx)
}
