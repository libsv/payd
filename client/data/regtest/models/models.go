package models

import (
	"fmt"

	"github.com/libsv/go-bc"
)

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

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type stringResponse struct {
	Result *string `json:"result"`
	response
}

type RawTxResponse struct {
	stringResponse
}

type RawTx1Response struct {
	Result struct {
		BlockHash string `json:"blockhash"`
	} `json:"result"`
	response
}

type MerkleProofResponse struct {
	Result *bc.MerkleProof `json:"result"`
	response
}

type SendToAddressResponse struct {
	stringResponse
}

type GenerateResponse struct {
	Result []string `json:"result"`
	response
}

func (e *Error) Error() string {
	return fmt.Sprintf("code (%d): %s", e.Code, e.Message)
}
