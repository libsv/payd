package gopayd

import (
	"context"

	"github.com/libsv/go-bt/v2"
)

type FeeReader interface {
	Fees(ctx context.Context) (*bt.FeeQuote, error)
}
