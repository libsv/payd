package spv

import (
	"encoding/hex"

	"github.com/libsv/go-bt/v2"
	"github.com/pkg/errors"

	"github.com/libsv/go-bc"
)

// AncestryJSON is a struct which contains all information needed for a transaction to be verified.
// this contains all ancestors for the transaction allowing proofs etc to be verified.
//
// NOTE: this is the JSON format of the Ancestry but in a nested format (in comparison) with
// the flat structure that the TSC uses. This allows verification to become a lot easier and
// use a recursive function.
type AncestryJSON struct {
	TxID          string                   `json:"txid,omitempty"`
	RawTx         string                   `json:"rawTx,omitempty"`
	Proof         *bc.MerkleProof          `json:"proof,omitempty"`
	MapiResponses []bc.MapiCallback        `json:"mapiResponses,omitempty"`
	Parents       map[string]*AncestryJSON `json:"parents,omitempty"`
}

// IsAnchored returns true if the ancestry has a merkle proof.
func (e *AncestryJSON) IsAnchored() bool {
	return e.Proof != nil
}

// HasParents returns true if this ancestry has immediate parents.
func (e *AncestryJSON) HasParents() bool {
	return e.Parents != nil && len(e.Parents) > 0
}

// ParentTx will return a parent if found and convert the rawTx to a bt.TX, otherwise a ErrNotAllInputsSupplied error is returned.
func (e *AncestryJSON) ParentTx(txID string) (*bt.Tx, error) {
	env, ok := e.Parents[txID]
	if !ok {
		return nil, errors.Wrapf(ErrNotAllInputsSupplied, "expected parent tx %s is missing", txID)
	}
	return bt.NewTxFromString(env.RawTx)
}

// Bytes takes a TxAncestry struct and returns the serialised binary format.
func (e *AncestryJSON) Bytes() ([]byte, error) {
	ancestryBinary := make([]byte, 0)
	ancestryBinary = append(ancestryBinary, 1) // Binary format version 1
	binary, err := serialiseInputs(e.Parents)
	if err != nil {
		return nil, err
	}
	ancestryBinary = append(ancestryBinary, binary...)
	return ancestryBinary, nil
}

func serialiseInputs(parents map[string]*AncestryJSON) ([]byte, error) {
	binary := make([]byte, 0)
	for _, input := range parents {
		currentTx, err := hex.DecodeString(input.RawTx)
		if err != nil {
			return nil, err
		}
		dataLength := bt.VarInt(uint64(len(currentTx)))
		binary = append(binary, flagTx)                // first data will always be a rawTx.
		binary = append(binary, dataLength.Bytes()...) // of this length.
		binary = append(binary, currentTx...)          // the data.
		if input.MapiResponses != nil && len(input.MapiResponses) > 0 {
			binary = append(binary, flagMapi) // next data will be a mapi response.
			numMapis := bt.VarInt(uint64(len(input.MapiResponses)))
			binary = append(binary, numMapis.Bytes()...) // number of mapi reponses which follow
			for _, mapiResponse := range input.MapiResponses {
				mapiR, err := mapiResponse.Bytes()
				if err != nil {
					return nil, err
				}
				dataLength := bt.VarInt(uint64(len(mapiR)))
				binary = append(binary, dataLength.Bytes()...) // of this length.
				binary = append(binary, mapiR...)              // the data.
			}
		}
		if input.Proof != nil {
			proof, err := input.Proof.Bytes()
			if err != nil {
				return nil, errors.Wrap(err, "Failed to serialise this input's proof struct")
			}
			proofLength := bt.VarInt(uint64(len(proof)))
			binary = append(binary, flagProof)              // it's going to be a proof.
			binary = append(binary, proofLength.Bytes()...) // of this length.
			binary = append(binary, proof...)               // the data.
		} else if input.HasParents() {
			parentsBinary, err := serialiseInputs(input.Parents)
			if err != nil {
				return nil, err
			}
			binary = append(binary, parentsBinary...)
		}
	}
	return binary, nil
}
