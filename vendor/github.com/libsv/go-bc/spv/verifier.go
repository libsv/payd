package spv

import (
	"context"

	"github.com/libsv/go-bt/v2"
	"github.com/pkg/errors"

	"github.com/libsv/go-bc"
)

type verifyOptions struct {
	// proofs validation
	proofs   bool
	script   bool
	fees     bool
	feeQuote *bt.FeeQuote
}

// clone will copy the verifyOptions to a new struct and return it.
func (v *verifyOptions) clone() *verifyOptions {
	return &verifyOptions{
		proofs:   v.proofs,
		fees:     v.fees,
		script:   v.script,
		feeQuote: v.feeQuote,
	}
}

// VerifyOpt defines a functional option that is used to modify behaviour of
// the payment verifier.
type VerifyOpt func(opts *verifyOptions)

// VerifyProofs will make the verifier validate the ancestry merkle proofs for each parent transaction.
func VerifyProofs() VerifyOpt {
	return func(opts *verifyOptions) {
		opts.proofs = true
	}
}

// NoVerifyProofs will switch off ancestry proof verification
// and rely on mAPI/node verification when the tx is broadcast.
func NoVerifyProofs() VerifyOpt {
	return func(opts *verifyOptions) {
		opts.proofs = false
	}
}

// VerifyFees will make the verifier check the transaction fees
// of the supplied transaction are enough based on the feeQuote
// provided.
//
// It is recommended to provide a fresh fee quote when calling the VerifyPayment
// method rather than loading fees when calling NewPaymentVerifier as fees can go out of date
// over the lifetime of the application and you may be supplying different feeQuotes
// to different consumers.
func VerifyFees(fees *bt.FeeQuote) VerifyOpt {
	return func(opts *verifyOptions) {
		opts.fees = true
		opts.feeQuote = fees
	}
}

// NoVerifyFees will switch off transaction fee verification and rely on
// mAPI / node verification when the transaction is broadcast.
func NoVerifyFees() VerifyOpt {
	return func(opts *verifyOptions) {
		opts.fees = false
		opts.feeQuote = nil
	}
}

// VerifyScript will ensure the scripts provided in the transaction are valid.
func VerifyScript() VerifyOpt {
	return func(opts *verifyOptions) {
		opts.script = true
	}
}

// NoVerifyScript will switch off script verification and rely on
// mAPI / node verification when the tx is broadcast.
func NoVerifyScript() VerifyOpt {
	return func(opts *verifyOptions) {
		opts.script = false
	}
}

// NoVerifySPV will turn off any spv validation for merkle proofs
// and script validation. This is a helper method that is equivalent to
// NoVerifyProofs && NoVerifyScripts.
func NoVerifySPV() VerifyOpt {
	return func(opts *verifyOptions) {
		opts.proofs = false
		opts.script = false
	}
}

// VerifySPV will turn on spv validation for merkle proofs
// and script validation. This is a helper method that is equivalent to
// VerifyProofs && VerifyScripts.
func VerifySPV() VerifyOpt {
	return func(opts *verifyOptions) {
		opts.proofs = true
		opts.script = true
	}
}

// A PaymentVerifier is an interface used to complete Simple Payment Verification (SPV)
// in conjunction with a Merkle Proof.
//
// The implementation of bc.BlockHeaderChain which is supplied will depend on the client
// you are using, some may return a HeaderJSON response others may return the blockhash.
type PaymentVerifier interface {
	VerifyPayment(ctx context.Context, p *Payment, opts ...VerifyOpt) error
	MerkleProofVerifier
}

// MerkleProofVerifier interfaces the verification of Merkle Proofs.
type MerkleProofVerifier interface {
	VerifyMerkleProof(context.Context, []byte) (*MerkleProofValidation, error)
	VerifyMerkleProofJSON(context.Context, *bc.MerkleProof) (bool, bool, error)
}

type verifier struct {
	// BlockHeaderChain will be set when an implementation returning a bc.BlockHeader type is provided.
	bhc  bc.BlockHeaderChain
	opts *verifyOptions
}

// NewPaymentVerifier creates a new spv.PaymentVerifer with the bc.BlockHeaderChain provided.
// If no BlockHeaderChain implementation is provided, the setup will return an error.
//
// opts control the global behaviour of the verifier and all options are enabled by default, they are:
// - ancestry verification (proofs checked etc)
// - fees checked, ensuring the root tx covers enough fees
// - script verification which checks the script is correct (not currently implemented).
func NewPaymentVerifier(bhc bc.BlockHeaderChain, opts ...VerifyOpt) (PaymentVerifier, error) {
	o := &verifyOptions{
		proofs: true,
		fees:   false,
		script: true,
	}
	for _, opt := range opts {
		opt(o)
	}
	if o.proofs && bhc == nil {
		return nil, errors.New("at least one blockchain header implementation should be returned")
	}
	return &verifier{bhc: bhc, opts: o}, nil
}

// NewMerkleProofVerifier creates a new spv.MerkleProofVerifer with the bc.BlockHeaderChain provided.
// If no BlockHeaderChain implementation is provided, the setup will return an error.
func NewMerkleProofVerifier(bhc bc.BlockHeaderChain) (MerkleProofVerifier, error) {
	return NewPaymentVerifier(bhc)
}
