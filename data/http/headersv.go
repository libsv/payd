package http

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/libsv/go-bc"
	"github.com/libsv/payd/data"
	"github.com/pkg/errors"
)

type hsvConnection struct {
	client data.Client
	host   string
}

// NewHeaderSVConnection returns a bc.BlockHeaderChain using a header client.
func NewHeaderSVConnection(client data.Client, host string) bc.BlockHeaderChain {
	return &hsvConnection{
		client: client,
		host:   host,
	}
}

// BlockHeader returns the header for the provided blockhash.
func (h *hsvConnection) BlockHeader(ctx context.Context, blockHash string) (*bc.BlockHeader, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/v1/chain/header/%s", h.host, blockHash),
		nil,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating request for chain/header/%s", blockHash)
	}
	req.Header.Add("Content-Type", "application/octet-stream")
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse error message body")
		}

		return nil, fmt.Errorf("block header request: unexpected status code %d\nresponse body:\n%s", resp.StatusCode, body)
	}

	headerBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	bh, err := bc.NewBlockHeaderFromBytes(headerBytes)
	if err != nil {
		return nil, err
	}
	return bh, nil
}
