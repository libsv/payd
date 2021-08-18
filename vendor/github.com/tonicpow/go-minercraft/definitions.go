package minercraft

import (
	"crypto/sha256"
	"encoding/json"
	"strings"

	"github.com/bitcoinschema/go-bitcoin"
)

// Miner is a configuration per miner, including connection url, auth token, etc
type Miner struct {
	MinerID string `json:"miner_id,omitempty"`
	Name    string `json:"name,omitempty"`
	Token   string `json:"token,omitempty"`
	URL     string `json:"url"`
}

// JSONEnvelope is a standard response from the Merchant API requests
//
// Standard for serializing a JSON document in order to have consistency when ECDSA signing the document.
// Any changes to a document being signed and verified, however minor they may be, will cause the signature
// verification to fail since the document will be converted into a string before being
// (hashed and then) signed. With JSON documents, the format permits changes to be made without
// compromising the validity of the format (e.g. extra spaces, carriage returns, etc.).
//
// This spec describes a technique to ensure consistency of the data being signed by encapsulating the
// JSON data as a string in parent JSON object. That way, however the JSON is marshaled,
// the first element in the parent JSON, the payload, would remain the same and be
// signed/verified the way it is.
//
// Specs: https://github.com/bitcoin-sv-specs/brfc-misc/tree/master/jsonenvelope
type JSONEnvelope struct {
	Miner     *Miner `json:"miner"`     // Custom field for our internal Miner configuration
	Validated bool   `json:"validated"` // Custom field if the signature has been validated
	Payload   string `json:"payload"`
	Signature string `json:"signature"`
	PublicKey string `json:"publicKey"`
	Encoding  string `json:"encoding"`
	MimeType  string `json:"mimetype"`
}

// process will take the raw payload and process into a struct
// while also validating the signature vs payload
func (p *JSONEnvelope) process(miner *Miner, bodyContents []byte) error {

	// Set the miner on the response
	p.Miner = miner

	// Unmarshal the response
	var err error
	if err = json.Unmarshal(bodyContents, &p); err != nil {
		return err
	}

	// Do we have a payload?
	if len(p.Payload) > 0 {

		// Remove all escaped slashes from payload envelope
		// Also needed for signature validation since it was signed before escaping
		p.Payload = strings.Replace(p.Payload, "\\", "", -1)
	}

	// Verify using DER format
	p.Validated, err = validateSignature(p.Signature, p.PublicKey, p.Payload)
	return err
}

// validateSignature will check the data against the pubkey + signature
func validateSignature(signature, pubKey, data string) (bool, error) {
	// Only if we have a signature and pubkey
	if len(signature) > 0 && len(pubKey) > 0 {
		return bitcoin.VerifyMessageDER(sha256.Sum256([]byte(data)), pubKey, signature)
	}
	return false, nil
}
