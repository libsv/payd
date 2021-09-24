package payd

import "context"

type PayRequest struct {
	PayToURL string `json:"payToURL"`
}

type PayService interface {
	Pay(ctx context.Context, req PayRequest) error
}
