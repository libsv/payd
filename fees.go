package payd

import (
	"context"

	"github.com/libsv/go-bt/v2"
)

// FeeQuoteFetcher fetch a new fee quote.
type FeeQuoteFetcher interface {
	// FeeQuote return fees from a fee quoter.
	FeeQuote(ctx context.Context) (*bt.FeeQuote, error)
}

// FeeQuoteReader can be implemented to read fees from a datastore.
type FeeQuoteReader interface {
	// FeeQuote will return fees from a datastore.
	FeeQuote(ctx context.Context, invoiceID string) (*bt.FeeQuote, error)
}

// FeeQuoteCreateArgs for store a fee quote.
type FeeQuoteCreateArgs struct {
	InvoiceID string
	FeeQuote  *bt.FeeQuote
}

// FeeQuoteWriter writes fee quotes to a store.
type FeeQuoteWriter interface {
	FeeQuoteCreate(ctx context.Context, args *FeeQuoteCreateArgs) error
}
