package regtest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/libsv/payd/cli/models"
	"github.com/pkg/errors"
)

// Default jsonrpc fields.
const (
	ID      = "spvclient"
	JSONRpc = "1.0"
)

// Bitcoin node method constants.
const (
	RequestGetRawTx       = "getrawtransaction"
	RequestSignRawTx      = "signrawtransaction"
	RequestGetMerkleProof = "getmerkleproof2"
	RequestGetNewAddress  = "getnewaddress"
	RequestListUnspent    = "listunspent"
	RequestGenerate       = "generate"
	RequestSendToAddress  = "sendtoaddress"
)

type regtest struct {
	c *http.Client
}

// NewRegtest returns a new regtest.
func NewRegtest(c *http.Client) models.Regtest {
	return &regtest{
		c: c,
	}
}

func (r *regtest) ListUnspent(ctx context.Context) (*models.ListUnspentResponse, error) {
	var resp models.ListUnspentResponse
	if err := r.performRPC(ctx, RequestListUnspent, &resp); err != nil {
		if resp.Error != nil {
			return nil, errors.Wrap(resp.Error, err.Error())
		}
		return nil, err
	}

	return &resp, nil
}

func (r *regtest) RawTransaction(ctx context.Context, txID string) (*models.RawTxResponse, error) {
	var resp models.RawTxResponse
	if err := r.performRPC(ctx, RequestGetRawTx, &resp, txID); err != nil {
		if resp.Error != nil {
			return nil, errors.Wrap(resp.Error, err.Error())
		}

		return nil, err
	}

	return &resp, nil
}

func (r *regtest) GetNewAddress(ctx context.Context) (*models.GetNewAddressResponse, error) {
	var resp models.GetNewAddressResponse
	if err := r.performRPC(ctx, RequestGetNewAddress, &resp); err != nil {
		if resp.Error != nil {
			return nil, errors.Wrap(resp.Error, err.Error())
		}

		return nil, err
	}

	return &resp, nil
}

func (r *regtest) RawTransaction1(ctx context.Context, txID string) (*models.RawTx1Response, error) {
	var resp models.RawTx1Response
	if err := r.performRPC(ctx, RequestGetRawTx, &resp, txID, 1); err != nil {
		if resp.Error != nil {
			return nil, errors.Wrap(resp.Error, err.Error())
		}

		return nil, err
	}

	return &resp, nil
}

func (r *regtest) SignRawTransaction(ctx context.Context, tx string) (*models.SignRawTxResponse, error) {
	var resp models.SignRawTxResponse
	if err := r.performRPC(ctx, RequestSignRawTx, &resp, tx, nil, nil, "ALL|FORKID"); err != nil {
		if resp.Error != nil {
			return nil, errors.Wrap(resp.Error, err.Error())
		}

		return nil, err
	}

	return &resp, nil
}

func (r *regtest) SendToAddress(ctx context.Context, address string, amount uint64) (*models.SendToAddressResponse, error) {
	var resp models.SendToAddressResponse
	if err := r.performRPC(ctx, RequestSendToAddress, &resp, address, amount); err != nil {
		if resp.Error != nil {
			return nil, errors.Wrap(resp.Error, err.Error())
		}
		return nil, err
	}

	return &resp, nil
}

func (r *regtest) MerkleProof(ctx context.Context, blockHash, txID string) (*models.MerkleProofResponse, error) {
	var resp models.MerkleProofResponse
	if err := r.performRPC(ctx, RequestGetMerkleProof, &resp, blockHash, txID); err != nil {
		if resp.Error != nil {
			return nil, errors.Wrap(resp.Error, err.Error())
		}

		return nil, err
	}

	return &resp, nil
}

func (r *regtest) Generate(ctx context.Context, amount uint64) (*models.GenerateResponse, error) {
	var resp models.GenerateResponse
	if err := r.performRPC(ctx, RequestGenerate, &resp, amount); err != nil {
		if resp.Error != nil {
			return nil, errors.Wrap(resp.Error, err.Error())
		}
		return nil, err
	}

	return &resp, nil
}

func (r *regtest) performRPC(ctx context.Context, method string, out interface{}, params ...interface{}) error {
	data, err := json.Marshal(&models.Request{
		ID:      ID,
		JSONRpc: JSONRpc,
		Method:  method,
		Params:  params,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"http://localhost:18332",
		bytes.NewReader(data),
	)
	if err != nil {
		return err
	}
	req.SetBasicAuth("bitcoin", "bitcoin")
	req.Header.Add("Content-Type", "text/plain")

	resp, err := r.c.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if err = json.NewDecoder(resp.Body).Decode(out); err != nil {
		return errors.Wrapf(err, "error decoding response")
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("status code %d not ok for request %s", resp.StatusCode, method)
	}

	return nil
}