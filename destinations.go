package gopayd

import (
	"context"

	validator "github.com/theflyingcodr/govalidator"
)

type DestinationArgs struct {
	PaymentID string
}

func (d DestinationArgs) Validate() error {
	return validator.New().Validate("paymentID", validator.NotEmpty(d.PaymentID)).Err()
}

type DestinationService interface {
	Destinations(ctx context.Context, args DestinationArgs) ([]*Output, error)
}
