package service

import (
	"context"

	"github.com/libsv/go-bc/spv"
	"github.com/libsv/payd/cli/models"
)

type paymentSvc struct {
	ps   models.PaymentStore
	spvb spv.EnvelopeCreator
}

// NewPaymentService returns a new payment service.
func NewPaymentService(ps models.PaymentStore, spvb spv.EnvelopeCreator) models.PaymentService {
	return &paymentSvc{
		ps:   ps,
		spvb: spvb,
	}
}

func (p *paymentSvc) Request(ctx context.Context, args models.PaymentRequestArgs) (*models.PaymentRequest, error) {
	return p.ps.Request(ctx, args)
}

func (p *paymentSvc) Send(ctx context.Context, args models.PaymentSendArgs) (*models.PaymentACK, error) {
	return p.ps.Submit(ctx, args)
}
