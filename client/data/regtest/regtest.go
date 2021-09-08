package regtest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/libsv/payd/client/data/regtest/models"
	"github.com/libsv/payd/config"
	"github.com/pkg/errors"
)

// Default jsonrpc fields.
const (
	ID      = "spvclient"
	JsonRpc = "1.0"
)

// Bitcoin node method constants.
const (
	RequestGetRawTx       = "getrawtransaction"
	RequestGetMerkleProof = "getmerkleproof2"
	RequestSendToAddress  = "sendtoaddress"
	RequestGenerate       = "generate"
)

type regtest struct {
	cfg *config.Regtest
	c   *http.Client
}

// NewRegtest returns a new regtest.
func NewRegtest(cfg *config.Regtest, c *http.Client) *regtest {
	return &regtest{
		cfg: cfg,
		c:   c,
	}
}

func (r *regtest) RawTransaction(ctx context.Context, txID string) (*models.RawTxResponse, error) {
	req, err := r.buildRequest(ctx, RequestGetRawTx, txID)
	if err != nil {
		return nil, err
	}

	resp, err := r.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var rawTxResp models.RawTxResponse
	if err = json.NewDecoder(resp.Body).Decode(&rawTxResp); err != nil {
		return nil, errors.Wrapf(err, "error decoding rawtx response for tx %s", txID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Wrapf(rawTxResp.Error, "status code not ok for tx %s", txID)
	}

	return &rawTxResp, nil
}

func (r *regtest) RawTransaction1(ctx context.Context, txID string) (*models.RawTx1Response, error) {
	req, err := r.buildRequest(ctx, RequestGetRawTx, txID, 1)
	if err != nil {
		return nil, err
	}

	resp, err := r.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var rawTxResp models.RawTx1Response
	if err = json.NewDecoder(resp.Body).Decode(&rawTxResp); err != nil {
		return nil, errors.Wrapf(err, "error decoding rawtx response for tx %s", txID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Wrapf(rawTxResp.Error, "status code not ok for tx %s", txID)
	}

	return &rawTxResp, nil
}

func (r *regtest) MerkleProof(ctx context.Context, blockHash, txID string) (*models.MerkleProofResponse, error) {
	req, err := r.buildRequest(ctx, RequestGetMerkleProof, blockHash, txID)
	if err != nil {
		return nil, err
	}

	resp, err := r.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var mpResp models.MerkleProofResponse
	if err = json.NewDecoder(resp.Body).Decode(&mpResp); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}

	return &mpResp, nil
}

func (r *regtest) SendToAddress(ctx context.Context, address string, amount float64) (*models.SendToAddressResponse, error) {
	req, err := r.buildRequest(ctx, RequestSendToAddress, address, amount)
	if err != nil {
		return nil, err
	}

	resp, err := r.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var staResp models.SendToAddressResponse
	if err = json.NewDecoder(resp.Body).Decode(&staResp); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Wrap(staResp.Error, "could not send to address")
	}

	return &staResp, nil
}

func (r *regtest) Generate(ctx context.Context, amount uint64) (*models.GenerateResponse, error) {
	req, err := r.buildRequest(ctx, RequestGenerate, amount)
	if err != nil {
		return nil, err
	}
	resp, err := r.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var genResp models.GenerateResponse
	if err = json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Wrap(genResp.Error, "could not send to address")
	}

	return nil, nil
}

func (r *regtest) buildRequest(ctx context.Context, method string, params ...interface{}) (*http.Request, error) {
	data, err := json.Marshal(&models.Request{
		ID:      ID,
		JSONRpc: JsonRpc,
		Method:  method,
		Params:  params,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"http://"+r.cfg.Host+r.cfg.Port,
		bytes.NewReader(data),
	)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(r.cfg.Username, r.cfg.Password)
	req.Header.Add("Content-Type", "text/plain")

	return req, nil
}
