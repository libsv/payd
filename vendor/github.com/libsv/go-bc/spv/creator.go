package spv

import (
	"context"

	"github.com/libsv/go-bt/v2"
	"github.com/pkg/errors"

	"github.com/libsv/go-bc"
)

// An TxAncestryCreator is an interface used to build the spv.TxAncestry data type for
// Simple Payment Verification (SPV).
//
// The implementation of an spv.TxStore and spv.MerkleProofStore which is supplied will depend
// on the client you are using.
type TxAncestryCreator interface {
	CreateTxAncestry(context.Context, *bt.Tx) (*AncestryJSON, error)
}

// TxStore interfaces the a tx store.
type TxStore interface {
	Tx(ctx context.Context, txID string) (*bt.Tx, error)
}

// MerkleProofStore interfaces a Merkle Proof store.
type MerkleProofStore interface {
	MerkleProof(ctx context.Context, txID string) (*bc.MerkleProof, error)
}

type creator struct {
	txc TxStore
	mpc MerkleProofStore
}

// NewEnvelopeCreator creates a new spv.Creator with the provided spv.TxStore and tx.MerkleProofStore.
// If either implementation is not provided, the setup will return an error.
func NewEnvelopeCreator(txc TxStore, mpc MerkleProofStore) (TxAncestryCreator, error) {
	if txc == nil {
		return nil, errors.New("an spv.TxStore implementation is required")
	}
	if mpc == nil {
		return nil, errors.New("an spv.MerkleProofStore implementation is required")
	}

	return &creator{txc: txc, mpc: mpc}, nil
}
