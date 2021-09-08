package payd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	gopayd "github.com/libsv/payd"
	"github.com/pkg/errors"
)

type ppctl struct {
	c *http.Client
}

func NewPPCTL(c *http.Client) *ppctl {
	return &ppctl{
		c: c,
	}
}

func (p *ppctl) Invoice(ctx context.Context, serverURL string, req gopayd.InvoiceCreate) (*gopayd.Invoice, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "error marshalling invoice request")
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, serverURL+"/api/v1/invoices", bytes.NewBuffer(data))
	if err != nil {
		return nil, errors.Wrap(err, "error creating invoice request")
	}
	r.Header.Add("Content-Type", "application/json")

	resp, err := p.c.Do(r)
	if err != nil {
		return nil, errors.Wrap(err, "error performing invoice create request")
	}
	defer resp.Body.Close()

	var invoice gopayd.Invoice
	if err = json.NewDecoder(resp.Body).Decode(&invoice); err != nil {
		return nil, errors.Wrap(err, "error decoding payment request response")
	}

	return &invoice, nil
}

func (p *ppctl) RequestPayment(ctx context.Context, serverURL string, args gopayd.PaymentRequestArgs) (*gopayd.PaymentRequest, error) {
	endpoint := fmt.Sprintf("%s/api/v1/payment/%s", serverURL, args.PaymentID)
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creating payment request request")
	}
	r.Header.Add("Content-Type", "application/json")

	resp, err := p.c.Do(r)
	if err != nil {
		return nil, errors.Wrap(err, "error performing payment request request")
	}
	defer resp.Body.Close()

	var paymentReq gopayd.PaymentRequest
	if err = json.NewDecoder(resp.Body).Decode(&paymentReq); err != nil {
		return nil, errors.Wrap(err, "error decoding payment request response")
	}

	return &paymentReq, nil
}

func (p *ppctl) SendPayment(ctx context.Context, endpoint string, req gopayd.CreatePayment) (*gopayd.PaymentACK, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "error marshalling invoice request")
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(data))
	if err != nil {
		return nil, errors.Wrap(err, "error creating payment request request")
	}
	r.Header.Add("Content-Type", "application/json")

	resp, err := p.c.Do(r)
	if err != nil {
		return nil, errors.Wrap(err, "error performing payment request request")
	}
	defer resp.Body.Close()

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
