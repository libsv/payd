package service

import (
	"context"

	"github.com/libsv/payd/cli/models"
)

type fundSvc struct {
	fStr models.FundStore
}

// NewFundService returns a new fund service.
func NewFundService(fStr models.FundStore) models.FundService {
	return &fundSvc{
		fStr: fStr,
	}
}

func (f *fundSvc) Add(ctx context.Context, args models.FundAddArgs) (models.Funds, error) {
	return f.fStr.Add(ctx, args)
}

func (f *fundSvc) Get(ctx context.Context, args models.FundGetArgs) (models.Funds, error) {
	return f.fStr.Get(ctx, args)
}

func (f *fundSvc) Spend(ctx context.Context, args models.FundSpendArgs) error {
	return f.fStr.Spend(ctx, args)
}
