package spv

import (
	"context"

	"github.com/libsv/go-bt/v2"
	"github.com/pkg/errors"
)

// VerifyPayment verifies whether or not the txs supplied via the supplied spv.Envelope are valid.
// This method also accepts functional options which will override the options set in the verifier upon creation.
// Any options passed here persist only for that call and do not override the main options. This gives flexibility
// in that you can setup a verifier with sensible defaults and override them conditionally by providing
// options here to enable or disable checks as required.
//
//  // with functional options to override verifier defaults.
//  tx, err := VerifyPayment(ctx, env, spv.VerifyProofs(), spv.VerifyFees(fees))
//
//  // without functional options to use verifier defaults.
//  tx, err := VerifyPayment(ctx, env)
//
func (v *verifier) VerifyPayment(ctx context.Context, initialPayment *Envelope, opts ...VerifyOpt) (*bt.Tx, error) {
	if initialPayment == nil {
		return nil, ErrNilInitialPayment
	}
	vOpt := v.opts.clone()
	for _, opt := range opts {
		opt(vOpt)
	}
	if vOpt.proofs && v.bhc == nil {
		return nil, errors.New("at least one blockchain header implementation should be set in the verifier if validating proofs")
	}
	// parse initial tx, fail fast if it isn't a valid tx.
	tx, err := bt.NewTxFromString(initialPayment.RawTx)
	if err != nil {
		return nil, err
	}

	// verify tx fees
	if vOpt.fees {
		if err := v.verifyFees(initialPayment, tx, vOpt); err != nil {
			return nil, err
		}
	}
	// if we are validating proofs or scripts carry out the validation.
	if vOpt.requiresEnvelope() {
		// The tip tx is the transaction we're trying to verify, and it should not have a supplied
		// Merkle Proof.
		if initialPayment.IsAnchored() {
			return nil, ErrTipTxConfirmed
		}
		if err := v.verifyTxs(ctx, initialPayment, vOpt); err != nil {
			return nil, err
		}
	}
	return tx, nil
}

// verifyFees takes the initial payment and iterates the immediate parents in order to gather
// the satoshis used for each input of the initialPayment tx.
//
// If there are no parents the method will fail, also, if there are no fees the method will fail.
func (v *verifier) verifyFees(initialPayment *Envelope, tx *bt.Tx, opts *verifyOptions) error {
	if !initialPayment.HasParents() {
		return ErrCannotCalculateFeePaid
	}
	if opts.feeQuote == nil {
		return ErrNoFeeQuoteSupplied
	}
	for _, input := range tx.Inputs {
		parent, err := initialPayment.ParentTx(input.PreviousTxIDStr())
		if err != nil {
			return errors.Wrapf(err, "tx %s failed to get parent tx", tx.TxID())
		}
		out := parent.OutputIdx(int(input.PreviousTxOutIndex))
		if out == nil {
			return ErrMissingOutput
		}
		input.PreviousTxSatoshis = out.Satoshis
	}
	ok, err := tx.IsFeePaidEnough(opts.feeQuote)
	if err != nil {
		return err
	}
	if !ok {
		return ErrFeePaidNotEnough
	}
	return nil
}

func (v *verifier) verifyTxs(ctx context.Context, payment *Envelope, opts *verifyOptions) error {
	// If at the beginning or middle of the tx chain and tx is unconfirmed, fail and error.
	if opts.proofs && !payment.IsAnchored() && !payment.HasParents() {
		return errors.Wrapf(ErrNoConfirmedTransaction, "tx %s has no confirmed/anchored tx", payment.TxID)
	}

	// Recurse back to the anchor transactions of the transaction chain and verify forward towards
	// the tip transaction. This way, we check that the first transactions in the chain are anchored
	// to the blockchain through a valid Merkle Proof.
	for parentTxID, parent := range payment.Parents {
		if parent.TxID == "" {
			parent.TxID = parentTxID
		}
		if err := v.verifyTxs(ctx, parent, opts); err != nil {
			return err
		}
	}

	// If a Merkle Proof is provided, assume we are at the anchor/beginning of the tx chain.
	// Verify and return the result.
	if payment.IsAnchored() || !payment.HasParents() {
		if opts.proofs {
			return v.verifyTxAnchor(ctx, payment)
		}
		return nil
	}

	tx, err := bt.NewTxFromString(payment.RawTx)
	if err != nil {
		return err
	}
	// We must verify the tx or else we can not know if any of it's child txs are valid.
	if opts.script {
		return v.verifyUnconfirmedTx(tx, payment)
	}
	return nil
}

func (v *verifier) verifyTxAnchor(ctx context.Context, payment *Envelope) error {
	proofTxID := payment.Proof.TxOrID
	if len(proofTxID) != 64 {
		proofTx, err := bt.NewTxFromString(payment.Proof.TxOrID)
		if err != nil {
			return err
		}

		proofTxID = proofTx.TxID()
	}

	// If the txid of the Merkle Proof doesn't match the txid provided in the spv.Envelope,
	// fail and error
	if proofTxID != payment.TxID {
		return errors.Wrapf(ErrTxIDMismatch, "envelope tx id %s does not match proof %s", payment.TxID, proofTxID)
	}

	valid, _, err := v.VerifyMerkleProofJSON(ctx, payment.Proof)
	if err != nil {
		return err
	}
	if !valid {
		return errors.Wrapf(ErrInvalidProof, "envelope tx id %s has invalid proof %s", payment.TxID, proofTxID)
	}
	return nil
}

func (v *verifier) verifyUnconfirmedTx(tx *bt.Tx, payment *Envelope) error {
	// If no tx inputs have been provided, fail and error
	if len(tx.Inputs) == 0 {
		return errors.Wrapf(ErrNoTxInputsToVerify, "tx %s has no inputs to verify", tx.TxID())
	}

	for _, input := range tx.Inputs {
		parent, err := payment.ParentTx(input.PreviousTxIDStr())
		if err != nil {
			return errors.Wrapf(err, "tx %s missing parent", tx.TxID())
		}
		// If the input is indexing an output that is out of bounds, fail and error
		if int(input.PreviousTxOutIndex) > len(parent.Outputs)-1 {
			return errors.Wrapf(ErrInputRefsOutOfBoundsOutput,
				"input %s is referring out of bounds output %d", input.PreviousTxIDStr(), input.PreviousTxOutIndex)
		}

		output := parent.OutputIdx(int(input.PreviousTxOutIndex))
		// TODO: verify script using input and previous output
		_ = output
	}

	return nil
}
