package models

import (
	"context"
	"fmt"

	"github.com/libsv/go-bc"
)

type Regtest interface {
	SendRawTransaction(ctx context.Context, txHex string) (*SendRawTransactionResponse, error)
	RawTransaction(ctx context.Context, txID string) (*RawTxResponse, error)
	RawTransaction1(ctx context.Context, txID string) (*RawTx1Response, error)
	MerkleProof(ctx context.Context, blockHash, txID string) (*MerkleProofResponse, error)
	SendToAddress(ctx context.Context, address string, amount float64) (*SendToAddressResponse, error)
	Generate(ctx context.Context, amount uint64) (*GenerateResponse, error)
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

// SendToAddressResponse models the response of `sendtoaddress <addr> <amount>`.
type SendToAddressResponse struct {
	stringResponse
}

// SendRawTransactionResponse models the response of `sendrawtransaction <txhex>`.
type SendRawTransactionResponse struct {
	stringResponse
}

// GenerateResponse models the response of `generate <n>`.
type GenerateResponse struct {
	Result []string `json:"result"`
	response
}

func (e *Error) Error() string {
	return fmt.Sprintf("code (%d): %s", e.Code, e.Message)
}
