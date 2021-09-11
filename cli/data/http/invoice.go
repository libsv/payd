package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/libsv/payd/cli/models"
	"github.com/pkg/errors"
)

type invoice struct {
	c models.HTTPClient
}

func NewInvoiceAPI(c models.HTTPClient) models.InvoiceReaderWriter {
	return &invoice{
		c: c,
	}
}

func (i *invoice) Invoice(ctx context.Context, args models.InvoiceGetArgs) (*models.Invoice, error) {
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8443/api/v1/invoices/"+args.ID, nil)
	if err != nil {
		return nil, err
	}
	r.Header.Add("Content-Type", "application/json")

	resp, err := i.c.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var invResp models.Invoice
	if err := json.NewDecoder(resp.Body).Decode(&invResp); err != nil {
		return nil, err
	}
	return &invResp, nil
}

func (i *invoice) Invoices(ctx context.Context) (models.Invoices, error) {
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8443/api/v1/invoices", nil)
	if err != nil {
		return nil, err
	}
	r.Header.Add("Content-Type", "application/json")

	resp, err := i.c.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost:8443/api/v1/invoices", bb)
	if err != nil {
		return nil, err
	}
	r.Header.Add("Content-Type", "application/json")

	resp, err := i.c.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var createResp models.Invoice
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		return nil, err
	}

	return &createResp, nil
}

func (i *invoice) Delete(ctx context.Context, args models.InvoiceDeleteArgs) error {
	r, err := http.NewRequestWithContext(ctx, http.MethodDelete, "http://localhost:8443/api/v1/invoices/"+args.ID, nil)
	if err != nil {
		return err
	}

	resp, err := i.c.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return errors.New("failed to delete")
	}

	return nil
}
