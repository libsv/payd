package models

import (
	"fmt"

	"github.com/libsv/go-bc"
)

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

// GenerateResponse models the response of `generate <n>`.
type GenerateResponse struct {
	Result []string `json:"result"`
	response
}

func (e *Error) Error() string {
	return fmt.Sprintf("code (%d): %s", e.Code, e.Message)
}
