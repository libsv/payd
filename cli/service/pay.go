package service

import (
	"context"

	"github.com/libsv/payd/cli/models"
)

type paySvc struct {
	ps models.PayStore
}

// NewPayService creates a new pay service.
func NewPayService(ps models.PayStore) models.PayStore {
	return &paySvc{
		ps: ps,
	}
}

// Request calls the http data store to POST a pay to url.
func (p *paySvc) Pay(ctx context.Context, args models.PayRequest) (*models.PaymentACK, error) {
	return p.ps.Pay(ctx, args)
}
