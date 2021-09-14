package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	gopayd "github.com/libsv/payd"
	"github.com/libsv/payd/cli/models"
)

type signer struct {
	c models.HTTPClient
}

func NewSignerAPI(c models.HTTPClient) models.Signer {
	return &signer{c: c}
}

func (s *signer) FundAndSign(ctx context.Context, req gopayd.FundAndSignTxRequest) (*gopayd.SignTxResponse, error) {
	bb, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost:8443/api/v1/fundandsign", bytes.NewBuffer(bb))
	if err != nil {
		return nil, err
	}
	r.Header.Add("Content-Type", "application/json")

	resp, err := s.c.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkError(resp, http.StatusOK); err != nil {
		return nil, err
	}

	var response *gopayd.SignTxResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response, nil
}
