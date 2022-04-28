package payd

import (
	"context"

	"github.com/libsv/go-bc/spv"
	"github.com/libsv/go-dpp"
	validator "github.com/theflyingcodr/govalidator"
)

// EnvelopeArgs identify where an envelope is being paid to.
type EnvelopeArgs struct {
	PayToURL string `json:"payToURL"`
}

// Validate will ensure that the args supplied are valid.
func (e EnvelopeArgs) Validate() error {
	return validator.New().
		Validate("payToURL", validator.NotEmpty(e.PayToURL))
}

// AncestryService will create an spv envelope from a paymentRequest.
// TODO - rename to AncestryCreate
type AncestryService interface {
	AncestryCreate(ctx context.Context, args EnvelopeArgs, req dpp.PaymentRequest) (*spv.AncestryJSON, error)
}
