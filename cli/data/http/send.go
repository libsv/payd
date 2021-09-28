package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/libsv/payd/cli/models"
)

type payHttp struct {
	c models.HTTPClient
}

func NewPayAPI(c models.HTTPClient) models.PayStore {
	return &payHttp{c: c}
}

func (p *payHttp) Request(ctx context.Context, args models.SendArgs) error {
	bb, err := json.Marshal(models.SendPayload{
		PayToURL: args.PayToURL,
	})
	if err != nil {
		return err
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, args.PayEndpoint, bytes.NewBuffer(bb))
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
