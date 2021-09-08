package ppctl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	gopayd "github.com/libsv/payd"
	"github.com/libsv/payd/config"
	"github.com/pkg/errors"
)

type ppctl struct {
	c   *http.Client
	cfg *config.Ppctl
}

// NewPPCTL returns a new PPCTL.
func NewPPCTL(c *http.Client, cfg *config.Ppctl) *ppctl {
	return &ppctl{
		c:   c,
		cfg: cfg,
	}
}

// Invoice creates an invoice.
func (p *ppctl) Invoice(ctx context.Context, req gopayd.InvoiceCreate) (*gopayd.Invoice, error) {
	url, err := url.Parse(p.cfg.URL)
	if err != nil {
		return nil, err
	}
	url.Path = path.Join(url.Path, "/api/v1/invoices")

	data, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "error marshalling invoice request")
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewBuffer(data))
	if err != nil {
		return nil, errors.Wrap(err, "error creating invoice request")
	}
	r.Header.Add("Content-Type", "application/json")

	resp, err := p.c.Do(r)
	if err != nil {
		return nil, errors.Wrap(err, "error performing invoice create request")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var invoice gopayd.Invoice
	if err = json.NewDecoder(resp.Body).Decode(&invoice); err != nil {
		return nil, errors.Wrap(err, "error decoding payment request response")
	}

	return &invoice, nil
}

// RequestPayment created a payment request.
func (p *ppctl) RequestPayment(ctx context.Context, args gopayd.PaymentRequestArgs) (*gopayd.PaymentRequest, error) {
	url, err := url.Parse(p.cfg.URL)
	if err != nil {
		return nil, err
	}
	url.Path = path.Join(url.Path, "/api/v1/payment/", args.PaymentID)

	r, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creating payment request request")
	}
	r.Header.Add("Content-Type", "application/json")

	resp, err := p.c.Do(r)
	if err != nil {
		return nil, errors.Wrap(err, "error performing payment request request")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var paymentReq gopayd.PaymentRequest
	if err = json.NewDecoder(resp.Body).Decode(&paymentReq); err != nil {
		return nil, errors.Wrap(err, "error decoding payment request response")
	}

	return &paymentReq, nil
}

// SendPayment sends and completes a payment request.
func (p *ppctl) SendPayment(ctx context.Context, paymentID string, req gopayd.CreatePayment) (*gopayd.PaymentACK, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "error marshalling invoice request")
	}

	url, err := url.Parse(p.cfg.URL)
	if err != nil {
		return nil, err
	}
	url.Path = path.Join(url.Path, "/api/v1/payment/", paymentID)

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewBuffer(data))
	if err != nil {
		return nil, errors.Wrap(err, "error creating payment request request")
	}
	r.Header.Add("Content-Type", "application/json")

	resp, err := p.c.Do(r)
	if err != nil {
		return nil, errors.Wrap(err, "error performing payment request request")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusCreated {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "error reading error body in payment request response")
		}

		return nil, fmt.Errorf("error requesting payment %s", string(body))
	}

	var paymentAck gopayd.PaymentACK
	if err = json.NewDecoder(resp.Body).Decode(&paymentAck); err != nil {
		return nil, errors.Wrap(err, "error decoding payment request response")
	}

	return &paymentAck, nil
}

// TxStatus retrieves the status of a tx.
func (p *ppctl) TxStatus(ctx context.Context, txID string) (*gopayd.TxStatus, error) {
	url, err := url.Parse(p.cfg.URL)
	if err != nil {
		return nil, err
	}

	url.Path = path.Join(url.Path, "/api/v1/txstatus/", txID)

	r, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creating payment request request")
	}
	r.Header.Add("Content-Type", "application/json")

	resp, err := p.c.Do(r)
	if err != nil {
		return nil, errors.Wrap(err, "error performing payment request request")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var txStatus gopayd.TxStatus
	if err = json.NewDecoder(resp.Body).Decode(&txStatus); err != nil {
		return nil, errors.Wrap(err, "error decoding payment request response")
	}

	return &txStatus, nil
}
