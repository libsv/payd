package spv

import (
	"context"

	"github.com/libsv/go-bk/crypto"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"

	"github.com/libsv/go-bc"
)

const (
	flagTx    = byte(1)
	flagProof = byte(2)
	flagMapi  = byte(3)
)

// Ancestry is a payment and its ancestors.
type Ancestry struct {
	PaymentTx *bt.Tx
	Ancestors map[[32]byte]*Ancestor
}

// Ancestor is an internal struct for validating transactions with their ancestors.
type Ancestor struct {
	Tx            *bt.Tx
	Proof         []byte
	MapiResponses []*bc.MapiCallback
}

// binaryChunk is a clear way to pass around chunks while keeping their type explicit.
type binaryChunk struct {
	ContentType byte
	Data        []byte
}

type extendedInput struct {
	input *bt.Input
	vin   int
}

// NewAncestryFromBytes creates a new struct from the bytes of a txContext.
func NewAncestryFromBytes(b []byte) (*Ancestry, error) {
	offset := uint64(1)
	total := uint64(len(b))

	l, size := bt.NewVarIntFromBytes(b[offset:])
	offset += uint64(size)
	paymentTx, err := bt.NewTxFromBytes(b[offset : offset+uint64(l)])
	if err != nil {
		return nil, err
	}
	ancestry := &Ancestry{
		PaymentTx: paymentTx,
		Ancestors: make(map[[32]byte]*Ancestor),
	}
	offset += uint64(l)

	var TxID [32]byte

	if total == offset {
		return nil, ErrCannotCalculateFeePaid
	}

	// You're not allowed to just have payment tx with a proof.
	if b[offset] != 1 {
		return nil, ErrTipTxConfirmed
	}

	for total > offset {
		chunk, size := parseChunk(b, offset)
		offset += size
		switch chunk.ContentType {
		case flagTx:
			hash := crypto.Sha256d(chunk.Data)
			copy(TxID[:], bt.ReverseBytes(hash)) // fixed size array from slice.
			tx, err := bt.NewTxFromBytes(chunk.Data)
			if err != nil {
				return nil, err
			}
			if len(tx.Inputs) == 0 {
				return nil, ErrNoTxInputsToVerify
			}
			ancestry.Ancestors[TxID] = &Ancestor{
				Tx: tx,
			}
		case flagProof:
			ancestry.Ancestors[TxID].Proof = chunk.Data
		case flagMapi:
			callBacks, err := parseMapiCallbacks(chunk.Data)
			if err != nil {
				return nil, err
			}
			ancestry.Ancestors[TxID].MapiResponses = callBacks
		default:
			continue
		}
	}
	return ancestry, nil
}

func parseChunk(b []byte, start uint64) (binaryChunk, uint64) {
	offset := start
	typeOfNextData := b[offset]
	offset++
	l, size := bt.NewVarIntFromBytes(b[offset:])
	offset += uint64(size)
	chunk := binaryChunk{
		ContentType: typeOfNextData,
		Data:        b[offset : offset+uint64(l)],
	}
	offset += uint64(l)
	return chunk, offset - start
}

func parseMapiCallbacks(b []byte) ([]*bc.MapiCallback, error) {
	if len(b) == 0 {
		return nil, ErrTriedToParseZeroBytes
	}
	var internalOffset uint64
	allBinary := uint64(len(b))
	numOfMapiResponses := b[internalOffset]
	if numOfMapiResponses == 0 && len(b) == 1 {
		return nil, ErrTriedToParseZeroBytes
	}
	internalOffset++

	var responses = [][]byte{}
	for allBinary > internalOffset {
		l, size := bt.NewVarIntFromBytes(b[internalOffset:])
		internalOffset += uint64(size)
		response := b[internalOffset : internalOffset+uint64(l)]
		internalOffset += uint64(l)
		responses = append(responses, response)
	}

	mapiResponses := make([]*bc.MapiCallback, 0)
	for _, response := range responses {
		mapiResponse, err := bc.NewMapiCallbackFromBytes(response)
		if err != nil {
			return nil, err
		}
		mapiResponses = append(mapiResponses, mapiResponse)
	}
	return mapiResponses, nil
}

// VerifyAncestryBinary will verify a slice of bytes which is a binary spv envelope.
func VerifyAncestryBinary(binaryData []byte, mpv MerkleProofVerifier, opts ...VerifyOpt) (bool, error) {
	o := &verifyOptions{
		proofs: true,
		script: true,
		fees:   false,
	}
	for _, opt := range opts {
		opt(o)
	}
	if binaryData[0] != 1 { // the first byte is the version number.
		return false, ErrUnsupporredVersion
	}
	ancestry, err := NewAncestryFromBytes(binaryData)
	if err != nil {
		return false, err
	}
	err = VerifyAncestors(ancestry, mpv, o)
	if err != nil {
		return false, err
	}
	return true, nil
}

// VerifyAncestors will run through the map of Ancestors and check each input of each transaction to verify it.
// Only if there is no Proof attached.
func VerifyAncestors(ancestry *Ancestry, mpv MerkleProofVerifier, opts *verifyOptions) error {
	ancestors := ancestry.Ancestors
	var paymentTxID [32]byte
	copy(paymentTxID[:], ancestry.PaymentTx.TxIDBytes())
	ancestors[paymentTxID] = &Ancestor{
		Tx: ancestry.PaymentTx,
	}
	for _, ancestor := range ancestors {
		inputsToCheck := make(map[[32]byte]*extendedInput)
		if len(ancestor.Tx.Inputs) == 0 {
			return ErrNoTxInputsToVerify
		}
		for idx, input := range ancestor.Tx.Inputs {
			var inputID [32]byte
			copy(inputID[:], input.PreviousTxID())
			inputsToCheck[inputID] = &extendedInput{
				input: input,
				vin:   idx,
			}
		}
		// if we have a proof, check it.
		if opts.proofs {
			if ancestor.Proof == nil {
				for inputID := range inputsToCheck {
					// check if we have that ancestor, if not validation fail.
					if ancestry.Ancestors[inputID] == nil {
						return ErrProofOrInputMissing
					}
				}
			} else {
				// check proof.
				response, err := mpv.VerifyMerkleProof(context.Background(), ancestor.Proof)
				if response == nil {
					return ErrInvalidProof
				}
				if response.TxID != "" && response.TxID != ancestor.Tx.TxID() {
					return ErrTxIDMismatch
				}
				if err != nil || !response.Valid {
					return ErrInvalidProof
				}
			}
		}
		if opts.script {
			// otherwise check the inputs.
			for inputID, extendedInput := range inputsToCheck {
				input := extendedInput.input
				// check if we have that ancestor, if not validation fail.
				if ancestry.Ancestors[inputID] == nil {
					if ancestor.Proof == nil && opts.proofs {
						return ErrProofOrInputMissing
					}
					continue
				}
				if len(ancestry.Ancestors[inputID].Tx.Outputs) <= int(input.PreviousTxOutIndex) {
					return ErrInputRefsOutOfBoundsOutput
				}
				lockingScript := ancestry.Ancestors[inputID].Tx.Outputs[input.PreviousTxOutIndex].LockingScript
				unlockingScript := input.UnlockingScript
				if !verifyInputOutputPair(ancestor.Tx, lockingScript, unlockingScript) {
					return ErrPaymentNotVerified
				}
			}
		}
		if opts.fees {
			if opts.feeQuote == nil {
				return ErrNoFeeQuoteSupplied
			}
			// no need to check fees for transactions we have proofs for
			if ancestor.Proof == nil {
				// add satoshi amounts to all inputs which correspond to outputs we have
				for inputID, extendedInput := range inputsToCheck {
					if ancestry.Ancestors[inputID] == nil {
						return ErrCannotCalculateFeePaid
					}
					sats := ancestry.Ancestors[inputID].Tx.Outputs[extendedInput.input.PreviousTxOutIndex].Satoshis
					ancestor.Tx.Inputs[extendedInput.vin].PreviousTxSatoshis = sats
				}
				// check the fees
				ok, err := ancestor.Tx.IsFeePaidEnough(opts.feeQuote)
				if err != nil || !ok {
					return ErrFeePaidNotEnough
				}
			}
		}
	}
	return nil
}

func verifyInputOutputPair(tx *bt.Tx, lock *bscript.Script, unlock *bscript.Script) bool {
	// TODO script interpreter.
	return true
}
