package models

import "context"

// SendArgs struct contains the pay endpoint and pay to url.
type SendArgs struct {
	PayToURL    string `json:"payToUrl"`
	PayEndpoint string `json:"payEndpoint"`
}

// SendPayload is the expected shape for the /pay endpoint.
type SendPayload struct {
	PayToURL string `json:"payToUrl"`
}

// PayStore interface for a pay (not to be confused with payment).
type PayStore interface {
	Request(ctx context.Context, args SendArgs) error
}
