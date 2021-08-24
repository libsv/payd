package spv

import (
	"context"

	"github.com/libsv/go-bc"
	"github.com/libsv/go-bt/v2"
)

// A Creator is an interface used to build the spv.Envelope data type for
// Simple Payment Verification (SPV).
//
// The implementation of an spv.TxStore and spv.MerkleProofStore which is supplied will depend
// on the client you are using.
type Creator interface {
	CreateEnvelope(context.Context, *bt.Tx) (*Envelope, error)
}

// A Verifier is an interface used to complete Simple Payment Verification (SPV)
// in conjunction with a Merkle Proof.
//
// The implementation of bc.BlockHeaderChain which is supplied will depend on the client
// you are using, some may return a HeaderJSON response others may return the blockhash.
type Verifier interface {
	EnvelopeVerifier
	MerkleProofVerifier
}

// EnvelopeVerifier interfaces the verification of SPV Envelopes
type EnvelopeVerifier interface {
	VerifyPayment(context.Context, *Envelope) (bool, error)
}

// MerkleProofVerifier interfaces the verification of Merkle Proofs
type MerkleProofVerifier interface {
	VerifyMerkleProof(context.Context, []byte) (bool, bool, error)
	VerifyMerkleProofJSON(context.Context, *bc.MerkleProof) (bool, bool, error)
}

// TxStore interfaces the a tx store
type TxStore interface {
	Tx(ctx context.Context, txID string) (*bt.Tx, error)
}

// MerkleProofStore interfaces a Merkle Proof store
type MerkleProofStore interface {
	MerkleProof(ctx context.Context, txID string) (*bc.MerkleProof, error)
}
