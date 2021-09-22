package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/libsv/payd/cli/config"
	"github.com/libsv/payd/cli/models"
)

type signer struct {
	c   models.HTTPClient
	cfg *config.Wallet
}

// NewSignerAPI returns a new signer api.
func NewSignerAPI(c models.HTTPClient, cfg *config.Wallet) models.Signer {
	return &signer{
		c:   c,
		cfg: cfg,
	}
}

func (s *signer) FundAndSign(ctx context.Context, req models.FundAndSignTxRequest) (*models.SignTxResponse, error) {
	bb, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.URLFor("/api/v1/fundandsign"), bytes.NewBuffer(bb))
	if err != nil {
		return nil, err
	}
	r.Header.Add("Content-Type", "application/json")

	resp, err := s.c.Do(r)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if err := checkError(resp, http.StatusOK); err != nil {
		return nil, err
	}

	var response *models.SignTxResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response, nil
}
