package service

import (
	"context"
	"net/url"

	"github.com/libsv/payd"
	"github.com/pkg/errors"
)

type payStrat struct {
	svcs map[string]payd.PayService
}

// NewPayStrategy returns a strategy based on url scheme.
func NewPayStrategy() payd.PayStrategy {
	return &payStrat{
		svcs: make(map[string]payd.PayService),
	}
}

// Register a strategy to the provided names.
func (p *payStrat) Register(svc payd.PayService, names ...string) payd.PayStrategy {
	for _, name := range names {
		p.svcs[name] = svc
	}

	return p
}

// Pay to a url.
func (p *payStrat) Pay(ctx context.Context, req payd.PayRequest) (*payd.PaymentACK, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	u, err := url.Parse(req.PayToURL)
	if err != nil {
		return nil, err
	}

	svc, ok := p.svcs[u.Scheme]
	if !ok {
		return nil, errors.New("invalid scheme" + u.Scheme)
	}

	return svc.Pay(ctx, req)
}
