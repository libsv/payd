package spv

import (
	"context"

	"github.com/libsv/go-bc"
	"github.com/libsv/go-bt/v2/bscript/interpreter"
	"github.com/pkg/errors"
)

// A PaymentVerifier is an interface used to complete Simple Payment Verification (SPV)
// in conjunction with a Merkle Proof.
//
// The implementation of bc.BlockHeaderChain which is supplied will depend on the client
// you are using, some may return a HeaderJSON response others may return the blockhash.
type PaymentVerifier interface {
	VerifyPayment(context.Context, *Envelope) (bool, error)
	MerkleProofVerifier
}

// MerkleProofVerifier interfaces the verification of Merkle Proofs
type MerkleProofVerifier interface {
	VerifyMerkleProof(context.Context, []byte) (bool, bool, error)
	VerifyMerkleProofJSON(context.Context, *bc.MerkleProof) (bool, bool, error)
}

type verifier struct {
	// BlockHeaderChain will be set when an implementation returning a bc.BlockHeader type is provided.
	bhc bc.BlockHeaderChain

	eng interpreter.Engine
}

// NewPaymentVerifier creates a new spv.PaymentVerifer with the bc.BlockHeaderChain provided.
// If no BlockHeaderChain implementation is provided, the setup will return an error.
func NewPaymentVerifier(bhc bc.BlockHeaderChain, eng interpreter.Engine) (PaymentVerifier, error) {
	if bhc == nil {
		return nil, errors.New("at least one blockchain header implementation should be returned")
	}
	if eng == nil {
		return nil, errors.New("at least one engine implementation should be provided")
	}

	return &verifier{bhc: bhc, eng: eng}, nil
}

// NewMerkleProofVerifier creates a new spv.MerkleProofVerifer with the bc.BlockHeaderChain provided.
// If no BlockHeaderChain implementation is provided, the setup will return an error.
func NewMerkleProofVerifier(bhc bc.BlockHeaderChain) (MerkleProofVerifier, error) {
	if bhc == nil {
		return nil, errors.New("at least one blockchain header implementation should be returned")
	}
	return &verifier{bhc: bhc}, nil
}
