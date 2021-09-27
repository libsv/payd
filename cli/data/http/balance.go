package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/libsv/payd/cli/config"
	"github.com/libsv/payd/cli/models"
)

type balance struct {
	c   models.HTTPClient
	cfg *config.Payd
}

func NewBalanceAPI(c models.HTTPClient, cfg *config.Payd) models.BalanceReader {
	return &balance{
		c:   c,
		cfg: cfg,
	}
}

func (b *balance) Balance(ctx context.Context) (*models.Balance, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.cfg.URLFor("/api/v1/balance"), nil)
	if err != nil {
		return nil, err
	}

	resp, err := b.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var balance models.Balance
	if err := json.NewDecoder(resp.Body).Decode(&balance); err != nil {
		return nil, err
	}

	return &balance, nil
}
