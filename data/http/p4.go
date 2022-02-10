package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/libsv/go-p4"
	"github.com/libsv/payd"
	"github.com/libsv/payd/data"
	"github.com/theflyingcodr/lathos/errs"
)

type p4Client struct {
	c data.Client
}

// NewP4 returns a new p4 interface.
func NewP4(c data.Client) P4 {
	return &p4Client{c: c}
}

// PaymentRequest performs a payment request http request to the specified url.
func (p *p4Client) PaymentRequest(ctx context.Context, args payd.PayRequest) (*p4.PaymentRequest, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, args.PayToURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, p.handleErr(resp)
	}

	var payRec p4.PaymentRequest
	if err = json.NewDecoder(resp.Body).Decode(&payRec); err != nil {
		return nil, err
	}

	return &payRec, nil
}

// PaymentSend sends a payment http request to the specified url, with the provided payment packet.
func (p *p4Client) PaymentSend(ctx context.Context, args payd.PayRequest, req p4.Payment) (*p4.PaymentACK, error) {
	bb, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, args.PayToURL, bytes.NewBuffer(bb))
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/json")

	resp, err := p.c.Do(request)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusCreated {
		return nil, p.handleErr(resp)
	}

	var ack p4.PaymentACK
	if err := json.NewDecoder(resp.Body).Decode(&ack); err != nil {
		return nil, err
	}

	return &ack, nil
}

func (p *p4Client) handleErr(resp *http.Response) error {
	errResp := &struct {
		ID      string `json:"id"`
		Code    string `json:"code"`
		Title   string `json:"title"`
		Message string `json:"message"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return errs.NewErrNotAuthenticated(errResp.Code, errResp.Message)
	case http.StatusForbidden:
		return errs.NewErrNotAuthorised(errResp.Code, errResp.Message)
	case http.StatusNotFound:
		return errs.NewErrNotFound(errResp.Code, errResp.Message)
	case http.StatusConflict:
		return errs.NewErrDuplicate(errResp.Code, errResp.Message)
	case http.StatusUnprocessableEntity:
		return errs.NewErrUnprocessable(errResp.Code, errResp.Message)
	}

	return errs.NewErrInternal(errors.New(errResp.Message), resp.Status)
}
