package ppctl

import (
	"context"

	go_payd "github.com/libsv/go-payd"
	"github.com/libsv/go-payd/config"
)

// paymentFacade is a layer on top of the payment services of which we currently support:
// * wallet payments, that are handled by the wallet and transmitted to the network
// * paymail payments, that use the paymail protocol for making the payments.
type paymentFacade struct {
	cfg   *config.Paymail
	pwSvc go_payd.PaymentService
	pmSvc go_payd.PaymentService
}

// NewPaymentFacade will create and return a new facade to determine between payments to use.
func NewPaymentFacade(cfg *config.Paymail, pwSvc go_payd.PaymentService, pmSvc go_payd.PaymentService) *paymentFacade {
	return &paymentFacade{cfg: cfg, pwSvc: pwSvc, pmSvc: pmSvc}
}

// Create will setup a new payment and return the result.
func (p *paymentFacade) CreatePayment(ctx context.Context, args go_payd.CreatePaymentArgs, req go_payd.CreatePayment) (*go_payd.PaymentACK, error) {

	if p.cfg.UsePaymail {
		return p.pmSvc.CreatePayment(ctx, args, req)
	}
	return p.pwSvc.CreatePayment(ctx, args, req)
}
