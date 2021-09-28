package service

import (
	"context"

	"github.com/libsv/payd/cli/models"
)

type paySvc struct {
	ps models.PayStore
}

func NewPayService(ps models.PayStore) *paySvc {
	return &paySvc{
		ps: ps,
	}
}

func (p *paySvc) Request(ctx context.Context, args models.SendArgs) error {
	return p.ps.Request(ctx, args)
}
