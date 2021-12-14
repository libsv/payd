package payd

import (
	"context"

	"github.com/libsv/go-bt/v2"
)

// FeeReader can be implemented to read fees.
type FeeReader interface {
	// Fees will return fees from a datastore.
	Fees(ctx context.Context, invoiceID string) (*bt.FeeQuote, error)
}

// FeeQuoteCreateArgs for store a fee quote.
type FeeQuoteCreateArgs struct {
	InvoiceID string
	FeeQuote  *bt.FeeQuote
}

// FeeWriter writes fees to a store.
type FeeWriter interface {
	FeesQuoteCreate(ctx context.Context, args *FeeQuoteCreateArgs) error
}
