package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/libsv/payd"
)

type p4 struct {
	c Client
}

// NewP4 returns a new p4 interface.
func NewP4(c Client) P4 {
	return &p4{c: c}
}

// PaymentRequest performs a payment request http request to the specified url.
func (p *p4) PaymentRequest(ctx context.Context, args payd.PayRequest) (*payd.PaymentRequestResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, args.PayToURL, nil)
	if err != nil {
		return nil, err
	}
	res, err := p.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code %d", res.StatusCode)
	}

	var payRec payd.PaymentRequestResponse
	if err = json.NewDecoder(res.Body).Decode(&payRec); err != nil {
		return nil, err
	}

	return &payRec, nil
}

// PaymentSend sends a payment http request to the specified url, with the provided payment packet.
func (p *p4) PaymentSend(ctx context.Context, args payd.PayRequest, req payd.PaymentSend) (*payd.PaymentACK, error) {
	bb, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, args.PayToURL, bytes.NewBuffer(bb))
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/json")

	resp, err := p.c.Do(request)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var ack payd.PaymentACK
	if err := json.NewDecoder(resp.Body).Decode(&ack); err != nil {
		return nil, err
	}

	return &ack, nil
}
