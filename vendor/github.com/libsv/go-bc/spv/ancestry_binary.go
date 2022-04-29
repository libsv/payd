package spv

import (
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

// Payment is a payment and its ancestry.
type Payment struct {
	PaymentTx *bt.Tx
	Ancestry  []byte
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

type ancestry struct {
	Tx            *bt.Tx
	Proof         []byte
	MapiResponses []*bc.MapiCallback
}

// parseAncestry creates a new struct from the bytes of a txContext.
func parseAncestry(b []byte) (map[[32]byte]*ancestry, error) {

	if b[0] != 1 { // the first byte is the version number.
		return nil, ErrUnsupporredVersion
	}
	offset := uint64(1)
	total := uint64(len(b))
	aa := make(map[[32]byte]*ancestry)

	var TxID [32]byte

	if total == offset {
		return nil, ErrCannotCalculateFeePaid
	}

	// first Data must be a Tx
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
			aa[TxID] = &ancestry{
				Tx: tx,
			}
		case flagProof:
			aa[TxID].Proof = chunk.Data
		case flagMapi:
			callBacks, err := parseMapiCallbacks(chunk.Data)
			if err != nil {
				return nil, err
			}
			aa[TxID].MapiResponses = callBacks
		default:
			continue
		}
	}
	return aa, nil
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

func verifyInputOutputPair(tx *bt.Tx, lock *bscript.Script, unlock *bscript.Script) bool {
	// TODO script interpreter.
	return true
}
