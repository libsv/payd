package service

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/libsv/payd"
	"github.com/libsv/payd/config"
)

// paymentRequest enforces business rules.
type paymentRequest struct {
	cfg     *config.Wallet
	destSvc payd.DestinationsService
	ownSvc  payd.OwnerStore
}

//  NewPaymentRequest will setup a new paymentRequest service.
func NewPaymentRequest(cfg *config.Wallet, destSvc payd.DestinationsService, ownSvc payd.OwnerStore) *paymentRequest {
	return &paymentRequest{
		cfg:     cfg,
		destSvc: destSvc,
		ownSvc:  ownSvc,
	}
}

// PaymentRequest will build and return a paymentRequest.
func (p *paymentRequest) PaymentRequest(ctx context.Context, args payd.PaymentRequestArgs) (*payd.PaymentRequestResponse, error) {
	if err := args.Validate(); err != nil {
		return nil, errors.WithStack(err)
	}

	dd, err := p.destSvc.Destinations(ctx, payd.DestinationsArgs{InvoiceID: args.PaymentID})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get destinations when building payment request '%s'", args.PaymentID)
	}
	owner, err := p.ownSvc.Owner(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get owner when building payment request '%s'", args.PaymentID)
	}
	if owner.ExtendedData == nil {
		owner.ExtendedData = map[string]interface{}{}
	}
	// here we store paymentRef in extended data to allow some validation in payment flow
	owner.ExtendedData["paymentReference"] = args.PaymentID
	oo := make([]payd.P4Output, len(dd.Outputs))
	for i, out := range dd.Outputs {
		oo[i] = payd.P4Output{
			Amount:      out.Satoshis,
			Script:      out.LockingScript.String(),
			Description: "payment reference " + args.PaymentID,
		}
	}
	return &payd.PaymentRequestResponse{
		Network:             string(p.cfg.Network),
		SPVRequired:         dd.SPVRequired,
		Destinations:        payd.P4Destination{Outputs: oo},
		Fee:                 dd.Fees,
		CreationTimestamp:   dd.CreatedAt,
		ExpirationTimestamp: dd.ExpiresAt,
		Memo:                fmt.Sprintf("invoice %s", args.PaymentID),
		MerchantData: payd.User{
			Avatar:       owner.Avatar,
			Name:         owner.Name,
			Email:        owner.Email,
			Address:      owner.Address,
			ExtendedData: owner.ExtendedData,
		},
	}, nil
}
