package ppctl

import (
	"context"

	"github.com/pkg/errors"
	validator "github.com/theflyingcodr/govalidator"
	gopaymail "github.com/tonicpow/go-paymail"

	gopayd "github.com/libsv/payd"
	"github.com/libsv/payd/config"
)

type paymailOutputs struct {
	cfg    *config.Paymail
	rdrwtr gopayd.PaymailReaderWriter
}

// NewPaymailOutputs will setup and return a new paymailOutputs service that implements a paymentRequestOutputer.
func NewPaymailOutputs(cfg *config.Paymail, rdrwtr gopayd.PaymailReaderWriter) *paymailOutputs {
	return &paymailOutputs{
		cfg:    cfg,
		rdrwtr: rdrwtr,
	}
}

// CreateOutputs will generate paymail outputs for the current server paymail address.
func (p *paymailOutputs) CreateOutputs(ctx context.Context, satoshis uint64, _ gopayd.PaymentRequestArgs) ([]*gopayd.Output, error) {
	addr, err := gopaymail.ValidateAndSanitisePaymail(p.cfg.Address, p.cfg.IsBeta)
	if err != nil {
		// convert to known type for the global error handler.
		return nil, validator.ErrValidation{
			"paymailAddress": []string{err.Error()},
		}
	}
	oo, err := p.rdrwtr.OutputsCreate(ctx, gopayd.P2POutputCreateArgs{
		Domain: addr.Domain,
		Alias:  addr.Alias,
	}, gopayd.P2PPayment{Satoshis: satoshis})
	return oo, errors.WithStack(err)
}
