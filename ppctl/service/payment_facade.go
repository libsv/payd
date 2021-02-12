package service

import (
	"context"

	"github.com/libsv/go-payd/config"
	"github.com/libsv/go-payd/ppctl"
)

// paymentFacade is a layer on top of the payment services of which we currently support:
// * wallet payments, that are handled by the wallet and transmitted to the network
// * paymail payments, that use the paymail protocol for making the payments.
type paymentFacade struct {
	cfg   *config.Paymail
	pwSvc ppctl.PaymentService
	pmSvc ppctl.PaymentService
}

// NewPaymentFacade will create and return a new facade to determine between payments to use.
func NewPaymentFacade(cfg *config.Paymail, pwSvc ppctl.PaymentService, pmSvc ppctl.PaymentService) *paymentFacade {
	return &paymentFacade{cfg: cfg, pwSvc: pwSvc, pmSvc: pmSvc}
}

// Create will setup a new payment and return the result.
func (p *paymentFacade) Create(ctx context.Context, args ppctl.CreatePaymentArgs, req ppctl.CreatePayment) (*ppctl.PaymentACK, error) {
	if p.cfg.UsePaymail {
		return p.pmSvc.Create(ctx, args, req)
	}
	return p.pwSvc.Create(ctx, args, req)
}
