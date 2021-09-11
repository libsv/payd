package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/libsv/payd/cli/models"
)

type paymentHttp struct {
	c models.HTTPClient
}

func NewPaymentAPI(c models.HTTPClient) models.PaymentStore {
	return &paymentHttp{c: c}
}

func (p *paymentHttp) Request(ctx context.Context, args models.PaymentRequestArgs) (*models.PaymentRequest, error) {
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8443/api/v1/payment/"+args.ID, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.c.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response models.PaymentRequest
	if json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (p *paymentHttp) Submit(ctx context.Context) error {
	return nil
}
