package http

import (
	"context"

	"github.com/libsv/go-dpp"
	"github.com/libsv/payd"
)

// DPP interfaces interactions with a dpp-proxy.
type DPP interface {
	PaymentRequest(ctx context.Context, req payd.PayRequest) (*dpp.PaymentRequest, error)
	PaymentSend(ctx context.Context, args payd.PayRequest, req dpp.Payment) (*dpp.PaymentACK, error)
}
