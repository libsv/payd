package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/libsv/payd/cli/config"
	"github.com/libsv/payd/cli/models"
	"github.com/pkg/errors"
)

type invoice struct {
	c   models.HTTPClient
	cfg *config.Wallet
}

// NewInvoiceAPI returns a new invoice api.
func NewInvoiceAPI(c models.HTTPClient, cfg *config.Wallet) models.InvoiceReaderWriter {
	return &invoice{
		c:   c,
		cfg: cfg,
	}
}

func (i *invoice) Invoice(ctx context.Context, args models.InvoiceGetArgs) (*models.Invoice, error) {
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, i.cfg.URLFor("/api/v1/invoices/", args.ID), nil)
	if err != nil {
		return nil, err
	}
	r.Header.Add("Content-Type", "application/json")

	resp, err := i.c.Do(r)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if err := checkError(resp, http.StatusOK); err != nil {
		return nil, err
	}

	var invResp models.Invoice
	if err := json.NewDecoder(resp.Body).Decode(&invResp); err != nil {
		return nil, err
	}
	return &invResp, nil
}

func (i *invoice) Invoices(ctx context.Context) (models.Invoices, error) {
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, i.cfg.URLFor("/api/v1/invoices"), nil)
	if err != nil {
		return nil, err
	}
	r.Header.Add("Content-Type", "application/json")

	resp, err := i.c.Do(r)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if err := checkError(resp, http.StatusOK); err != nil {
		return nil, err
	}

	var invResp []*models.Invoice
	if err := json.NewDecoder(resp.Body).Decode(&invResp); err != nil {
		return nil, err
	}
	return invResp, nil
}

func (i *invoice) Create(ctx context.Context, req models.InvoiceCreateRequest) (*models.Invoice, error) {
	bb := &bytes.Buffer{}
	if err := json.NewEncoder(bb).Encode(req); err != nil {
		return nil, errors.Wrap(err, "failed to decode create invoice request")
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, i.cfg.URLFor("/api/v1/invoices"), bb)
	if err != nil {
		return nil, err
	}
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("X-Account", req.Account)

	resp, err := i.c.Do(r)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if err := checkError(resp, http.StatusCreated); err != nil {
		return nil, err
	}

	var createResp models.Invoice
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		return nil, err
	}

	return &createResp, nil
}

func (i *invoice) Delete(ctx context.Context, args models.InvoiceDeleteArgs) error {
	r, err := http.NewRequestWithContext(ctx, http.MethodDelete, i.cfg.URLFor("/api/v1/invoices/", args.ID), nil)
	if err != nil {
		return err
	}

	resp, err := i.c.Do(r)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	return checkError(resp, http.StatusNoContent)
}
