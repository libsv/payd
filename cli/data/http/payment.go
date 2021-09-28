package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/libsv/payd/cli/models"
)

type payment struct {
	c models.HTTPClient
}

// NewPaymentAPI returns a new payment api.
func NewPaymentAPI(c models.HTTPClient) models.PaymentStore {
	return &payment{c: c}
}

// Request a payment request from a p4 server.
func (p *payment) Request(ctx context.Context, args models.PaymentRequestArgs) (*models.PaymentRequest, error) {
	endpoint := fmt.Sprintf("%s/api/v1/payment/%s", args.PayTo, args.ID)
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.c.Do(r)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if err := checkError(resp, http.StatusCreated); err != nil {
		return nil, err
	}

	var response models.PaymentRequest
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// Submit a payment to a p4 server.
func (p *payment) Submit(ctx context.Context, args models.PaymentSendArgs) (*models.PaymentAck, error) {
	bb, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, args.PaymentRequest.PaymentURL, bytes.NewBuffer(bb))
	if err != nil {
		return nil, err
	}
	r.Header.Add("Content-Type", "application/json")

	resp, err := p.c.Do(r)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if err := checkError(resp, http.StatusCreated); err != nil {
		return nil, err
	}

	var response models.PaymentAck
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}
