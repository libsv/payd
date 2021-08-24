package spv

import (
	"errors"

	"github.com/libsv/go-bc"
)

type verifier struct {
	// BlockHeaderChain will be set when an implementation returning a bc.BlockHeader type is provided.
	bhc bc.BlockHeaderChain
}

type creator struct {
	txc TxStore
	mpc MerkleProofStore
}

// NewVerifier creates a new spv.Verifer with the bc.BlockHeaderChain provided.
// If no BlockHeaderChain implementation is provided, the setup will return an error.
func NewVerifier(bhc bc.BlockHeaderChain) (Verifier, error) {
	if bhc == nil {
		return nil, errors.New("at least one blockchain header implementation should be returned")
	}

	return &verifier{bhc: bhc}, nil
}

// NewCreator creates a new spv.Creator with the provided spv.TxStore and tx.MerkleProofStore.
// If either implementation is not provided, the setup will return an error.
func NewCreator(txc TxStore, mpc MerkleProofStore) (Creator, error) {
	if txc == nil {
		return nil, errors.New("an spv.TxStore implementation is required")
	}
	if mpc == nil {
		return nil, errors.New("an spv.MerkleProofStore implementation is required")
	}

	return &creator{txc: txc, mpc: mpc}, nil
}
