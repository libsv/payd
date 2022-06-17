package service

import (
	"context"
	"fmt"
	"time"

	"github.com/libsv/payd/log"
	"github.com/pkg/errors"
	"github.com/theflyingcodr/lathos/errs"
	"golang.org/x/sync/errgroup"

	"github.com/libsv/go-bt/v2"
	"github.com/libsv/payd"
	"github.com/libsv/payd/config"
)

// paymentRequest enforces business rules.
type paymentRequest struct {
	cfg     *config.Wallet
	destSvc payd.DestinationsService
	feeFtr  payd.FeeQuoteFetcher
	feeWtr  payd.FeeQuoteWriter
	ownSvc  payd.OwnerStore
	l       log.Logger
}

//  NewPaymentRequest will setup a new paymentRequest service.
func NewPaymentRequest(cfg *config.Wallet, destSvc payd.DestinationsService, feeFtr payd.FeeQuoteFetcher, feeWtr payd.FeeQuoteWriter, ownSvc payd.OwnerStore,
	l log.Logger) *paymentRequest {
	return &paymentRequest{
		cfg:     cfg,
		destSvc: destSvc,
		feeFtr:  feeFtr,
		feeWtr:  feeWtr,
		ownSvc:  ownSvc,
		l:       l,
	}
}

// PaymentRequest will build and return a paymentRequest.
func (p *paymentRequest) PaymentRequest(ctx context.Context, args payd.PaymentRequestArgs) (*payd.PaymentRequestResponse, error) {
	p.l.Debugf("[payment request] hit for invoice %s", args.InvoiceID)
	if err := args.Validate(); err != nil {
		return nil, errors.WithStack(err)
	}

	var dd *payd.Destination
	var oo []payd.DPPOutput
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		p.l.Debugf("[payment request] getting destinations for invoice %s", args.InvoiceID)
		destinations, err := p.destSvc.Destinations(ctx, payd.DestinationsArgs{InvoiceID: args.InvoiceID})
		if err != nil {
			p.l.Debugf("[payment request] failed to get destinations for invoice %s: %s", args.InvoiceID, err)
			return errors.Wrapf(err, "failed to get destinations when building payment request '%s'", args.InvoiceID)
		}
		oo = make([]payd.DPPOutput, len(destinations.Outputs))
		for i, out := range destinations.Outputs {
			oo[i] = payd.DPPOutput{
				Amount:      out.Satoshis,
				Script:      out.LockingScript.String(),
				Description: "payment reference " + args.InvoiceID,
			}
		}
		dd = destinations

		if !dd.ExpiresAt.IsZero() && dd.ExpiresAt.Before(time.Now().UTC()) {
			p.l.Debugf("[payment request] payment expired for invoice %s", args.InvoiceID)
			return errs.NewErrUnprocessable("U102", "payment expired")
		}
		return nil
	})

	var owner *payd.User
	g.Go(func() (err error) {
		p.l.Debugf("[payment request] getting owner for invoice %s", args.InvoiceID)
		owner, err = p.ownSvc.Owner(ctx)
		if err != nil {
			return errors.Wrapf(err, "failed to get owner when building payment request '%s'", args.InvoiceID)
		}
		if owner.ExtendedData == nil {
			owner.ExtendedData = map[string]interface{}{}
		}
		// here we store paymentRef in extended data to allow some validation in payment flow
		owner.ExtendedData["paymentReference"] = args.InvoiceID
		return nil
	})
	var fees *bt.FeeQuote
	g.Go(func() error {
		p.l.Debugf("[payment request] getting fee quote for invoice %s", args.InvoiceID)
		fq, err := p.feeFtr.FeeQuote(ctx)
		if err != nil {
			return errors.Wrapf(err, "failed to get fees when getting payment request")
		}
		p.l.Debugf("[payment request] storing fee quote for invoice %s", args.InvoiceID)
		if err := p.feeWtr.FeeQuoteCreate(ctx, &payd.FeeQuoteCreateArgs{
			InvoiceID: args.InvoiceID,
			FeeQuote:  fq,
		}); err != nil {
			return errors.Wrapf(err, "failed to write fees when getting payment request")
		}
		fees = fq
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}
	p.l.Debugf("[payment request] return payment request for invoice %s", args.InvoiceID)
	return &payd.PaymentRequestResponse{
		Network:             string(p.cfg.Network),
		AncestryRequired:    dd.SPVRequired,
		Destinations:        payd.DPPDestination{Outputs: oo},
		Fee:                 fees,
		CreationTimestamp:   dd.CreatedAt,
		ExpirationTimestamp: dd.ExpiresAt,
		Memo:                fmt.Sprintf("invoice %s", args.InvoiceID),
		MerchantData: payd.User{
			Avatar:       owner.Avatar,
			Name:         owner.Name,
			Email:        owner.Email,
			Address:      owner.Address,
			ExtendedData: owner.ExtendedData,
		},
	}, nil
}
