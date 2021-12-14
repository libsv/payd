package spv

import (
	"encoding/hex"
	"fmt"

	"github.com/libsv/go-bt/v2"
	"github.com/pkg/errors"

	"github.com/libsv/go-bc"
)

// Envelope is a struct which contains all information needed for a transaction to be verified.
//
type Envelope struct {
	TxID          string               `json:"txid,omitempty"`
	RawTx         string               `json:"rawTx,omitempty"`
	Proof         *bc.MerkleProof      `json:"proof,omitempty"`
	MapiResponses []bc.MapiCallback    `json:"mapiResponses,omitempty"`
	Parents       map[string]*Envelope `json:"parents,omitempty"`
}

// IsAnchored returns true if the envelope is the anchor tx.
func (e *Envelope) IsAnchored() bool {
	return e.Proof != nil
}

// HasParents returns true if this envelope has immediate parents.
func (e *Envelope) HasParents() bool {
	return e.Parents != nil && len(e.Parents) > 0
}

// ParentTx will return a parent if found and convert the rawTx to a bt.TX, otherwise a ErrNotAllInputsSupplied error is returned.
func (e *Envelope) ParentTx(txID string) (*bt.Tx, error) {
	env, ok := e.Parents[txID]
	if !ok {
		return nil, errors.Wrapf(ErrNotAllInputsSupplied, "expected parent tx %s is missing", txID)
	}
	return bt.NewTxFromString(env.RawTx)
}

// Bytes takes an spvEnvelope struct and returns the serialised binary format.
func (e *Envelope) Bytes() ([]byte, error) {
	flake := make([]byte, 0)

	// Binary format version 1
	flake = append(flake, 1)

	initialTx := map[string]*Envelope{
		e.TxID: {
			TxID:          e.TxID,
			RawTx:         e.RawTx,
			Proof:         e.Proof,
			MapiResponses: e.MapiResponses,
			Parents:       e.Parents,
		},
	}

	err := serializeParents(initialTx, &flake, true)
	if err != nil {
		fmt.Println(err)
	}
	return flake, nil
}

func serializeParents(parents map[string]*Envelope, flake *[]byte, root bool) error {
	for _, input := range parents {
		currentTx, err := hex.DecodeString(input.RawTx)
		if err != nil {
			fmt.Print(err)
		}
		dataLength := bt.VarInt(uint64(len(currentTx)))
		if !root {
			*flake = append(*flake, flagTx) // first data will always be a rawTx.
		}
		*flake = append(*flake, dataLength.Bytes()...) // of this length.
		*flake = append(*flake, currentTx...)          // the data.
		if input.MapiResponses != nil && len(input.MapiResponses) > 0 {
			*flake = append(*flake, flagMapi) // next data will be a mapi response.
			numMapis := bt.VarInt(uint64(len(input.MapiResponses)))
			*flake = append(*flake, numMapis.Bytes()...) // number of mapi reponses which follow
			for _, mapiResponse := range input.MapiResponses {
				mapiR, err := mapiResponse.Bytes()
				if err != nil {
					return err
				}
				dataLength := bt.VarInt(uint64(len(mapiR)))
				*flake = append(*flake, dataLength.Bytes()...) // of this length.
				*flake = append(*flake, mapiR...)              // the data.
			}
		}
		if input.Proof != nil {
			proof, err := input.Proof.Bytes()
			if err != nil {
				return errors.Wrap(err, "Failed to serialise this input's proof struct")
			}
			proofLength := bt.VarInt(uint64(len(proof)))
			*flake = append(*flake, flagProof)              // it's going to be a proof.
			*flake = append(*flake, proofLength.Bytes()...) // of this length.
			*flake = append(*flake, proof...)               // the data.
		} else if input.HasParents() {
			err = serializeParents(input.Parents, flake, false)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
