package service

import (
	"context"

	"github.com/libsv/payd/cli/models"
)

type paymentSvc struct {
	ps models.PaymentStore
}

func NewPaymentService(ps models.PaymentStore) models.PaymentService {
	return &paymentSvc{
		ps: ps,
	}
}

func (p *paymentSvc) Request(ctx context.Context, args models.PaymentRequestArgs) (*models.PaymentRequest, error) {
	return p.ps.Request(ctx, args)
}

func (p *paymentSvc) Send(ctx context.Context) error {
	panic("not implemented") // TODO: Implement
}
