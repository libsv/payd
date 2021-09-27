package service

import (
	"context"

	"github.com/libsv/payd/cli/models"
)

type balance struct {
	rdr models.BalanceReader
}

func NewBalanceService(rdr models.BalanceReader) models.BalanceService {
	return &balance{rdr: rdr}
}

func (b *balance) Balance(ctx context.Context) (*models.Balance, error) {
	return b.rdr.Balance(ctx)
}
