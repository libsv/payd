package models

import "context"

// SendPayload is the expected shape for the /pay endpoint.
type SendPayload struct {
	PayToURL string `json:"payToUrl"`
}

// PayStore interface for a pay (not to be confused with payment).
type PayStore interface {
	Request(ctx context.Context, args SendPayload) error
}
