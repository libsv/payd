package payd

import (
	"context"

	"github.com/libsv/go-bc/spv"
)

// EnvelopeArgs identify where an envelope is being paid to.
type EnvelopeArgs struct {
	PayToURL string `json:"payToURL"`
}

// EnvelopeService will create an spv envelope from a paymentRequest.
type EnvelopeService interface {
	Envelope(ctx context.Context, args EnvelopeArgs, req PaymentRequestResponse) (*spv.Envelope, error)
}
