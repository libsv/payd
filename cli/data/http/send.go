package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/libsv/payd/cli/config"
	"github.com/libsv/payd/cli/models"
)

type payHttp struct {
	c   models.HTTPClient
	cfg *config.Payd
}

// NewPayAPI creates an instace of pay api.
func NewPayAPI(c models.HTTPClient, cfg *config.Payd) models.PayStore {
	return &payHttp{c: c, cfg: cfg}
}

// Request performs a post a pay request to a payd instance.
func (p *payHttp) Request(ctx context.Context, args models.SendPayload) error {
	bb, err := json.Marshal(models.SendPayload{
		PayToURL: args.PayToURL,
	})
	if err != nil {
		return err
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, p.cfg.URLFor("/api/v1/pay"), bytes.NewBuffer(bb))
	if err != nil {
		return err
	}
	r.Header.Add("Content-Type", "application/json")

	resp, err := p.c.Do(r)
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if err := checkError(resp, http.StatusCreated); err != nil {
		return err
	}
	return nil
}
