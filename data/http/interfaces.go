package http

import (
	"context"
	"net/http"

	"github.com/libsv/go-p4"
	"github.com/libsv/payd"
)

// Client interfaces the Do(*http.Request) function to allow for easy mocking.
type Client interface {
	Do(*http.Request) (*http.Response, error)
}

// P4 interfaces interactions with a p4 server.
type P4 interface {
	PaymentRequest(ctx context.Context, req payd.PayRequest) (*p4.PaymentRequest, error)
	PaymentSend(ctx context.Context, args payd.PayRequest, req p4.Payment) (*p4.PaymentACK, error)
}
