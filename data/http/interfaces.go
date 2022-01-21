package http

import (
	"context"

	"github.com/libsv/go-p4"
	"github.com/libsv/payd"
)

// P4 interfaces interactions with a p4 server.
type P4 interface {
	PaymentRequest(ctx context.Context, req payd.PayRequest) (*p4.PaymentRequest, error)
	PaymentSend(ctx context.Context, args payd.PayRequest, req p4.Payment) (*p4.PaymentACK, error)
}
