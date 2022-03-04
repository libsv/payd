package service

import (
	"context"
	"fmt"
	"time"

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
}

//  NewPaymentRequest will setup a new paymentRequest service.
func NewPaymentRequest(cfg *config.Wallet, destSvc payd.DestinationsService, feeFtr payd.FeeQuoteFetcher, feeWtr payd.FeeQuoteWriter, ownSvc payd.OwnerStore) *paymentRequest {
	return &paymentRequest{
		cfg:     cfg,
		destSvc: destSvc,
		feeFtr:  feeFtr,
		feeWtr:  feeWtr,
		ownSvc:  ownSvc,
	}
}

// PaymentRequest will build and return a paymentRequest.
func (p *paymentRequest) PaymentRequest(ctx context.Context, args payd.PaymentRequestArgs) (*payd.PaymentRequestResponse, error) {
	if err := args.Validate(); err != nil {
		return nil, errors.WithStack(err)
	}

	var dd *payd.Destination
	var oo []payd.DPPOutput
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		destinations, err := p.destSvc.Destinations(ctx, payd.DestinationsArgs{InvoiceID: args.InvoiceID})
		if err != nil {
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
			return errs.NewErrUnprocessable("U102", "payment expired")
		}
		return nil
	})

	var owner *payd.User
	g.Go(func() (err error) {
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
		fq, err := p.feeFtr.FeeQuote(ctx)
		if err != nil {
			return errors.Wrapf(err, "failed to get fees when getting payment request")
		}
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
	return &payd.PaymentRequestResponse{
		Network:             string(p.cfg.Network),
		SPVRequired:         dd.SPVRequired,
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
