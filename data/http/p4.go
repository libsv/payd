package http

import (
	"context"

	"github.com/libsv/payd"
)

type p4 struct {
	c Client
}

func NewP4(c Client) P4 {
	return &p4{c: c}
}

func (p *p4) PaymentRequest(ctx context.Context, args payd.PayRequest) (interface{}, error) {
	return nil, nil
}
