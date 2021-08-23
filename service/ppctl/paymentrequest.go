package ppctl

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	validator "github.com/theflyingcodr/govalidator"
	"github.com/theflyingcodr/lathos/errs"

	gopayd "github.com/libsv/payd"
	"github.com/libsv/payd/config"
)

type paymentRequest struct {
	walletCfg *config.Wallet
	envCfg    *config.Server
	outputter gopayd.PaymentRequestOutputer
	store     gopayd.PaymentRequestReaderWriter
}

// NewPaymentRequest will setup and return a new PaymentRequest service that will generate outputs
// using the provided outputter which is defined in server config.
func NewPaymentRequest(walletCfg *config.Wallet,
	envCfg *config.Server,
	outputter gopayd.PaymentRequestOutputer,
	store gopayd.PaymentRequestReaderWriter) *paymentRequest {
	return &paymentRequest{
		walletCfg: walletCfg,
		envCfg:    envCfg,
		store:     store,
		outputter: outputter,
	}
}

// CreatePaymentRequest handles setting up a new PaymentRequest response and can use and optional existing paymentID.
func (p *paymentRequest) CreatePaymentRequest(ctx context.Context, args gopayd.PaymentRequestArgs) (*gopayd.PaymentRequest, error) {
	if err := validator.New().
		Validate("paymentID", validator.NotEmpty(args.PaymentID)).
		Validate("hostname", validator.NotEmpty(p.envCfg)); err.Err() != nil {
		return nil, err
	}
	inv, err := p.store.Invoice(ctx, gopayd.InvoiceArgs{PaymentID: args.PaymentID})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get invoice when creating payment request")
	}
	if !inv.PaymentReceivedAt.IsZero() {
		return nil, errs.NewErrDuplicate("D103", fmt.Sprintf("payment already received for paymentId %s", args.PaymentID))
	}
	oo, err := p.outputter.CreateOutputs(ctx, gopayd.OutputsCreate{
		Satoshis:     inv.Satoshis,
		Denomination: 0,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate outputs for paymentID %s", args.PaymentID)
	}
	return &gopayd.PaymentRequest{
		Network:             p.walletCfg.Network,
		Outputs:             oo,
		CreationTimestamp:   time.Now().UTC().Unix(),
		ExpirationTimestamp: time.Now().Add(24 * time.Hour).UTC().Unix(),
		PaymentURL:          fmt.Sprintf("http://%s/payment/%s", p.envCfg.Hostname, args.PaymentID),
		Memo:                fmt.Sprintf("invoice %s", args.PaymentID),
		MerchantData: &gopayd.MerchantData{
			AvatarURL:    p.walletCfg.MerchantAvatarURL,
			MerchantName: p.walletCfg.MerchantName,
		},
	}, nil
}
