package service

import (
	"context"

	"github.com/libsv/payd/cli/models"
)

type paySvc struct {
	ps models.PayStore
}

// NewPayService creates a new pay service.
func NewPayService(ps models.PayStore) *paySvc {
	return &paySvc{
		ps: ps,
	}
}

// Request calls the http data store to POST a pay to url.
func (p *paySvc) Request(ctx context.Context, args models.SendPayload) error {
	return p.ps.Request(ctx, args)
}
