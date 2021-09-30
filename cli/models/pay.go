package models

import "context"

// PayRequest is the expected shape for the /pay endpoint.
type PayRequest struct {
	PayToURL string `json:"payToUrl"`
}

// PayService interfaces a with a pay service.
type PayService interface {
	PayStore
}

// PayStore interface for a pay (not to be confused with payment).
type PayStore interface {
	Pay(ctx context.Context, args PayRequest) (*PaymentACK, error)
}
