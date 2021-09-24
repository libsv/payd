package http

import (
	"context"
	"net/http"

	"github.com/libsv/payd"
)

// Client interfaces the Do(*http.Request) function to allow for easy mocking.
type Client interface {
	Do(*http.Request) (*http.Response, error)
}

type P4 interface {
	PaymentRequest(ctx context.Context, req payd.PayRequest) (*payd.PaymentRequestResponse, error)
}
