package spv

import (
	"context"

	"github.com/libsv/go-bc"
	"github.com/libsv/go-bt/v2"
)

// An Client is a struct used to specify interfaces
// used to complete Simple Payment Verification (SPV)
// in conjunction with a Merkle Proof.
//
// The implementation of BlockHeaderChain which is supplied will depend on the client
// you are using, some may return a HeaderJSON response others may return the blockhash.
type Client interface {
	EnvelopeHandler
	MerkleProofVerifier
}

// EnvelopeHandler interfaces the handling (creation and verification) of SPV Envelopes
type EnvelopeHandler interface {
	EnvelopeCreator
	EnvelopeVerifier
}

// EnvelopeCreator interfaces the creation of SPV Envelopes
type EnvelopeCreator interface {
	CreateEnvelope(context.Context, *bt.Tx) (*Envelope, error)
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
