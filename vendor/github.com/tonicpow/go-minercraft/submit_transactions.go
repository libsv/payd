package minercraft

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// Reference: https://github.com/bitcoin-sv-specs/brfc-merchantapi#5-submit-multiple-transactions

type (

	// RawSubmitTransactionsResponse is the response returned from mapi where payload is a string.
	RawSubmitTransactionsResponse struct {
		Encoding  string `json:"encoding"`
		MimeType  string `json:"mimetype"`
		Payload   string `json:"payload"`
		PublicKey string `json:"publicKey"`
		Signature string `json:"signature"`
	}

	// SubmitTransactionsResponse is the formatted response which converts payload string to payloads.
	SubmitTransactionsResponse struct {
		Encoding  string     `json:"encoding"`
		MimeType  string     `json:"mimetype"`
		Payload   TxsPayload `json:"payload"`
		PublicKey string     `json:"publicKey"`
		Signature string     `json:"signature"`
	}

	// TxsPayload is the structure of the json payload string in the MapiResponse.
	TxsPayload struct {
		APIVersion                string    `json:"apiVersion"`
		CurrentHighestBlockHash   string    `json:"currentHighestBlockHash"`
		CurrentHighestBlockHeight int       `json:"currentHighestBlockHeight"`
		FailureCount              int       `json:"failureCount"`
		MinerID                   string    `json:"minerId"`
		Timestamp                 time.Time `json:"timestamp"`
		Txs                       []Tx      `json:"txs"`
		TxSecondMempoolExpiry     int       `json:"txSecondMempoolExpiry"`
	}

	// Tx is the transaction format in the mapi txs response.
	Tx struct {
		ConflictedWith    []ConflictedWith `json:"conflictedWith,omitempty"`
		ResultDescription string           `json:"resultDescription"`
		ReturnResult      string           `json:"returnResult"`
		TxID              string           `json:"txid"`
	}
)

// SubmitTransactions is used for submitting batched transactions
//
// Reference: https://github.com/bitcoin-sv-specs/brfc-merchantapi#5-submit-multiple-transactions
func (c *Client) SubmitTransactions(ctx context.Context, miner *Miner, txs []Transaction) (*SubmitTransactionsResponse, error) {
	if miner == nil {
		return nil, errors.New("miner was nil")
	}

	if len(txs) <= 0 {
		return nil, errors.New("no transactions")
	}

	data, err := json.Marshal(txs)
	if err != nil {
		return nil, err
	}

	response := httpRequest(ctx, c, &httpPayload{
		Method: http.MethodPost,
		URL:    miner.URL + routeSubmitTxs,
		Token:  miner.Token,
		Data:   data,
	})

	if response.Error != nil {
		return nil, err
	}

	var raw RawSubmitTransactionsResponse
	if err = json.Unmarshal(
		response.BodyContents, &raw,
	); err != nil {
		return nil, err
	}

	result := &SubmitTransactionsResponse{
		Signature: raw.Signature,
		PublicKey: raw.PublicKey,
		Encoding:  raw.Encoding,
		MimeType:  raw.MimeType,
	}

	if err = json.Unmarshal(
		[]byte(raw.Payload), &result.Payload,
	); err != nil {
		return nil, err
	}

	return result, err
}
