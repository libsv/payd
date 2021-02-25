package minercraft

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

/*
Example Transaction Submission (submitted in the body of the request)
{
  "rawtx":        "[transaction_hex_string]",
  "callBackUrl":  "https://your.service.callback/endpoint",
  "callBackToken" : <channel token>,
  "merkleProof" : true,
  "dsCheck" : true,
  "callBackEncryption" : <parameter>
}
*/

// Transaction is the body contents in the submit transaction request
type Transaction struct {
	RawTx              string `json:"rawtx"`
	CallBackURL        string `json:"callBackUrl,omitempty"`
	CallBackToken      string `json:"callBackToken,omitempty"`
	MerkleProof        string `json:"merkleProof,omitempty"`
	DsCheck            string `json:"dsCheck,omitempty"`
	CallBackEncryption string `json:"callBackEncryption,omitempty"`
}

/*
Example submit tx response from Merchant API:

{
  "payload": "{\"apiVersion\":\"0.1.0\",\"timestamp\":\"2020-01-15T11:40:29.826Z\",\"txid\":\"6bdbcfab0526d30e8d68279f79dff61fb4026ace8b7b32789af016336e54f2f0\",\"returnResult\":\"success\",\"resultDescription\":\"\",\"minerId\":\"03fcfcfcd0841b0a6ed2057fa8ed404788de47ceb3390c53e79c4ecd1e05819031\",\"currentHighestBlockHash\":\"71a7374389afaec80fcabbbf08dcd82d392cf68c9a13fe29da1a0c853facef01\",\"currentHighestBlockHeight\":207,\"txSecondMempoolExpiry\":0}",
  "signature": "3045022100f65ae83b20bc60e7a5f0e9c1bd9aceb2b26962ad0ee35472264e83e059f4b9be022010ca2334ff088d6e085eb3c2118306e61ec97781e8e1544e75224533dcc32379",
  "publicKey": "03fcfcfcd0841b0a6ed2057fa8ed404788de47ceb3390c53e79c4ecd1e05819031",
  "encoding": "UTF-8",
  "mimetype": "application/json"
}
*/

// SubmitTransactionResponse is the raw response from the Merchant API request
//
// Specs: https://github.com/bitcoin-sv-specs/brfc-merchantapi/tree/v1.2-beta#Submit-transaction
type SubmitTransactionResponse struct {
	JSONEnvelope
	Results *SubmissionPayload `json:"results"` // Custom field for unmarshalled payload data
}

/*
Example SubmitTransactionResponse.Payload (unmarshalled):

{
  "apiVersion": "1.2.3",
  "timestamp": "2020-01-15T11:40:29.826Z",
  "txid": "6bdbcfab0526d30e8d68279f79dff61fb4026ace8b7b32789af016336e54f2f0",
  "returnResult": "success",
  "resultDescription": "",
  "minerId": "03fcfcfcd0841b0a6ed2057fa8ed404788de47ceb3390c53e79c4ecd1e05819031",
  "currentHighestBlockHash": "71a7374389afaec80fcabbbf08dcd82d392cf68c9a13fe29da1a0c853facef01",
  "currentHighestBlockHeight": 207,
  "txSecondMempoolExpiry": 0,
  "conflictedWith": ""
}
*/

// SubmissionPayload is the unmarshalled version of the payload envelope
type SubmissionPayload struct {
	APIVersion                string `json:"apiVersion"`
	Timestamp                 string `json:"timestamp"`
	TxID                      string `json:"txid"`
	ReturnResult              string `json:"returnResult"`
	ResultDescription         string `json:"resultDescription"`
	MinerID                   string `json:"minerId"`
	CurrentHighestBlockHash   string `json:"currentHighestBlockHash"`
	ConflictedWith            string `json:"conflictedWith"`
	CurrentHighestBlockHeight int64  `json:"currentHighestBlockHeight"`
	TxSecondMempoolExpiry     int64  `json:"txSecondMempoolExpiry"`
}

// SubmitTransaction will fire a Merchant API request to submit a given transaction
//
// This endpoint is used to send a raw transaction to a miner for inclusion in the next block
// that the miner creates. It returns a JSONEnvelope with a payload that contains the response to the
// transaction submission. The purpose of the envelope is to ensure strict consistency in the
// message content for the purpose of signing responses.
//
// Specs: https://github.com/bitcoin-sv-specs/brfc-merchantapi/tree/v1.2-beta#Submit-transaction
func (c *Client) SubmitTransaction(miner *Miner, tx *Transaction) (*SubmitTransactionResponse, error) {

	// Make sure we have a valid miner
	if miner == nil {
		return nil, errors.New("miner was nil")
	}

	// Make the HTTP request
	result := submitTransaction(context.Background(), c, miner, tx)
	if result.Response.Error != nil {
		return nil, result.Response.Error
	}

	// Parse the response
	response, err := result.parseSubmission()
	if err != nil {
		return nil, err
	}

	// Valid query?
	if response.Results == nil || len(response.Results.ReturnResult) == 0 {
		return nil, errors.New("failed getting submission response from: " + miner.Name)
	}

	// Return the fully parsed response
	return &response, nil
}

// parseSubmission will convert the HTTP response into a struct and also unmarshal the payload JSON data
func (i *internalResult) parseSubmission() (response SubmitTransactionResponse, err error) {

	// Process the initial response payload
	if err = response.process(i.Miner, i.Response.BodyContents); err != nil {
		return
	}

	// If we have a valid payload
	if len(response.Payload) > 0 {
		err = json.Unmarshal([]byte(response.Payload), &response.Results)
	}
	return
}

// submitTransaction will fire the HTTP request to submit a transaction
func submitTransaction(ctx context.Context, client *Client, miner *Miner, tx *Transaction) (result *internalResult) {
	result = &internalResult{Miner: miner}
	data, _ := json.Marshal(tx) // Ignoring error - if it fails, the submission would also fail
	result.Response = httpRequest(ctx, client, &httpPayload{
		Method: http.MethodPost,
		URL:    miner.URL + routeSubmitTx,
		Token:  miner.Token,
		Data:   data,
	})
	return
}
