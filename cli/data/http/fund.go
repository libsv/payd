package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	gopayd "github.com/libsv/payd"
	"github.com/libsv/payd/cli/models"
)

type fundHttp struct {
	c models.HTTPClient
}

// NewFundAPI returns a new fund api.
func NewFundAPI(c models.HTTPClient) models.FundStore {
	return &fundHttp{c: c}
}

func (p *fundHttp) Add(ctx context.Context, args models.FundAddArgs) (models.Funds, error) {
	bb, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost:8443/api/v1/funds", bytes.NewBuffer(bb))
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

	if err = checkError(resp, http.StatusCreated); err != nil {
		return nil, err
	}

	var funds models.Funds
	if err = json.NewDecoder(resp.Body).Decode(&funds); err != nil {
		return nil, err
	}

	return funds, nil
}

func (p *fundHttp) Get(ctx context.Context, args models.FundGetArgs) (models.Funds, error) {
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8443/api/v1/funds", nil)
	if err != nil {
		return nil, err
	}
	r.Header.Add("X-Account", args.Account)

	resp, err := p.c.Do(r)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if err = checkError(resp, http.StatusOK); err != nil {
		return nil, err
	}

	var funds models.Funds
	if err := json.NewDecoder(resp.Body).Decode(&funds); err != nil {
		return nil, err
	}

	return funds, nil
}

func (p *fundHttp) GetAmount(ctx context.Context, req models.FundsRequest, args models.FundGetArgs) (*gopayd.FundsGetResponse, error) {
	bb, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("http://localhost:8443/api/v1/funds/%d", args.Amount)
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, bytes.NewBuffer(bb))
	if err != nil {
		return nil, err
	}
	r.Header.Add("X-Account", args.Account)

	resp, err := p.c.Do(r)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if err = checkError(resp, http.StatusOK); err != nil {
		return nil, err
	}

	var funds *gopayd.FundsGetResponse
	if err := json.NewDecoder(resp.Body).Decode(&funds); err != nil {
		return nil, err
	}

	return funds, nil
}

func (p *fundHttp) Spend(ctx context.Context, args models.FundSpendArgs) error {
	bb, err := json.Marshal(args)
	if err != nil {
		return err
	}
	r, err := http.NewRequestWithContext(ctx, http.MethodPut, "http://localhost:8443/api/v1/funds/spend", bytes.NewBuffer(bb))
	if err != nil {
		return err
	}
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("x-account", args.Account)

	resp, err := p.c.Do(r)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if err = checkError(resp, http.StatusNoContent); err != nil {
		return err
	}

	return nil
}
