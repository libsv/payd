package bc

import (
	"encoding/json"
)

// MapiCallback is the body contents posted to the provided callback url from Merchant API.
type MapiCallback struct {
	CallbackPayload string `json:"callbackPayload"`
	APIVersion      string `json:"apiVersion"`
	Timestamp       string `json:"timestamp"`
	MinerID         string `json:"minerId"`
	BlockHash       string `json:"blockHash"`
	BlockHeight     uint64 `json:"blockHeight"`
	CallbackTxID    string `json:"callbackTxId"`
	CallbackReason  string `json:"callbackReason"`
}

// NewMapiCallbackFromBytes is a glorified json unmarshaller, but might be more sophisticated in future.
func NewMapiCallbackFromBytes(b []byte) (*MapiCallback, error) {
	var mapiCallback MapiCallback
	err := json.Unmarshal(b, &mapiCallback)
	if err != nil {
		return nil, err
	}
	// TODO check signature is valid.
	return &mapiCallback, nil
}

// Bytes converts the MapiCallback struct into a binary format.
// We are not going to parse anything out but rather take the whole json object as a binary blob.
// The reason behind this approach is that the whole callback is signed by the mapi server,
// so if a single byte is out of place the signature will be invalid.
func (mcb *MapiCallback) Bytes() ([]byte, error) {
	return json.Marshal(mcb)
}
