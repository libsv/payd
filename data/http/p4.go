package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/libsv/payd"
)

type p4 struct {
	c Client
}

func NewP4(c Client) P4 {
	return &p4{c: c}
}

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
