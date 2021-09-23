package models

import (
	"context"
	"fmt"

	"github.com/libsv/go-bc"
	"github.com/libsv/go-bt/v2"
)

// Regtest interfaces interactions with regtest.
type Regtest interface {
	RawTransaction(ctx context.Context, txID string) (*RawTxResponse, error)
	RawTransaction1(ctx context.Context, txID string) (*RawTx1Response, error)
	CreateRawTransaction(ctx context.Context, utxos []*bt.UTXO, outputs map[string]float64) (*CreateRawTxResponse, error)
	ListUnspent(ctx context.Context) (*ListUnspentResponse, error)
	SignRawTransaction(ctx context.Context, tx string) (*SignRawTxResponse, error)
	MerkleProof(ctx context.Context, blockHash, txID string) (*MerkleProofResponse, error)
}

// Request models a JSON RPC request.
type Request struct {
	ID      string        `json:"id"`
	JSONRpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type response struct {
	Error *Error `json:"error"`
	ID    string `json:"id"`
}

// Error models an error response.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type stringResponse struct {
	Result *string `json:"result"`
	response
}

// RawTxResponse models the response of `getrawtransaction <txid>`.
type RawTxResponse struct {
	stringResponse
}

// RawTx1Response models the response of `getrawtransaction <txid> 1`.
type RawTx1Response struct {
	Result struct {
		BlockHash string `json:"blockhash"`
	} `json:"result"`
	response
}

// MerkleProofResponse models the response of `getmerkleproof2 <bh> <txid>`.
type MerkleProofResponse struct {
	Result *bc.MerkleProof `json:"result"`
	response
}

type ListUnspentResponse struct {
	Result []*bt.UTXO `json:"result"`
	response
}

type CreateRawTxResponse struct {
	stringResponse
}

type SignRawTxResponse struct {
	Result struct {
		Hex      string `json:"hex"`
		Complete bool   `json:"complete"`
	} `json:"result"`
	response
}

func (e *Error) Error() string {
	return fmt.Sprintf("code (%d): %s", e.Code, e.Message)
}
