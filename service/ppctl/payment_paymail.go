package ppctl

import (
	"context"

	"github.com/pkg/errors"
	validator "github.com/theflyingcodr/govalidator"
	gopaymail "github.com/tonicpow/go-paymail"

	gopayd "github.com/libsv/payd"
	"github.com/libsv/payd/config"
)

type paymentPaymailService struct {
	cfg   *config.Paymail
	pmSvc gopayd.PaymailWriter
}

// NewPaymailPaymentService will setup and return a new payment service that uses paymail.
func NewPaymailPaymentService(pmSvc gopayd.PaymailWriter, cfg *config.Paymail) *paymentPaymailService {
	return &paymentPaymailService{
		pmSvc: pmSvc,
		cfg:   cfg,
	}
}

// Send will submit a transaction via the paymail network.
func (p *paymentPaymailService) Send(ctx context.Context, args gopayd.CreatePaymentArgs, req gopayd.CreatePayment) error {
	if _, err := gopaymail.ValidateAndSanitisePaymail(p.cfg.Address, p.cfg.IsBeta); err != nil {
		// convert to known type for the global error handler.
		return validator.ErrValidation{
			"paymailAddress": []string{err.Error()},
		}
	}
	return errors.WithStack(p.pmSvc.Send(ctx, args, req))
}
