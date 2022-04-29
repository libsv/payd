package spv

import (
	"context"

	"github.com/pkg/errors"
)

// VerifyPayment is a method for parsing a binary payment transaction and its corresponding ancestry in binary.
// It will return the paymentTx struct if all validations pass.
func (v *verifier) VerifyPayment(ctx context.Context, p *Payment, opts ...VerifyOpt) error {
	o := v.opts.clone()
	for _, opt := range opts {
		opt(o)
	}
	if o.proofs && v == nil {
		return errors.New("Merkle Proof Verifier is required when proofs is set")
	}

	aa, err := parseAncestry(p.Ancestry)
	if err != nil {
		return err
	}

	var paymentTxID [32]byte
	copy(paymentTxID[:], p.PaymentTx.TxIDBytes())
	aa[paymentTxID] = &ancestry{
		Tx: p.PaymentTx,
	}
	if o.fees {
		if o.feeQuote == nil {
			return ErrNoFeeQuoteSupplied
		}
		for i, input := range p.PaymentTx.Inputs {
			var inputID [32]byte
			copy(inputID[:], input.PreviousTxID())
			parent, ok := aa[inputID]
			if !ok {
				return errors.Wrapf(ErrNoFeeQuoteSupplied, "missing tx for input %d", i)
			}

			out := parent.Tx.OutputIdx(int(input.PreviousTxOutIndex))
			if out == nil {
				return ErrMissingOutput
			}

			input.PreviousTxSatoshis = out.Satoshis
		}
		ok, err := p.PaymentTx.IsFeePaidEnough(o.feeQuote)
		if err != nil {
			return err
		}
		if !ok {
			return ErrFeePaidNotEnough
		}
	}
	for _, a := range aa {
		inputsToCheck := make(map[[32]byte]*extendedInput)
		if len(a.Tx.Inputs) == 0 {
			return ErrNoTxInputsToVerify
		}
		for idx, input := range a.Tx.Inputs {
			var inputID [32]byte
			copy(inputID[:], input.PreviousTxID())
			inputsToCheck[inputID] = &extendedInput{
				input: input,
				vin:   idx,
			}
		}
		// if we have a proof, check it.
		if o.proofs {
			if a.Proof == nil {
				for inputID := range inputsToCheck {
					// check if we have that ancestry, if not validation fail.
					if aa[inputID] == nil {
						return ErrProofOrInputMissing
					}
				}
			} else {
				// check proof.
				response, err := v.VerifyMerkleProof(ctx, a.Proof)
				if response == nil {
					return ErrInvalidProof
				}
				if response.TxID != "" && response.TxID != a.Tx.TxID() {
					return ErrTxIDMismatch
				}
				if err != nil || !response.Valid {
					return ErrInvalidProof
				}
			}
		}
		if o.script {
			// otherwise check the inputs.
			for inputID, extendedInput := range inputsToCheck {
				input := extendedInput.input
				// check if we have that ancestry, if not validation fail.
				if aa[inputID] == nil {
					if a.Proof == nil && o.proofs {
						return ErrProofOrInputMissing
					}
					continue
				}
				if len(aa[inputID].Tx.Outputs) <= int(input.PreviousTxOutIndex) {
					return ErrInputRefsOutOfBoundsOutput
				}
				lockingScript := aa[inputID].Tx.Outputs[input.PreviousTxOutIndex].LockingScript
				unlockingScript := input.UnlockingScript
				if !verifyInputOutputPair(a.Tx, lockingScript, unlockingScript) {
					return ErrPaymentNotVerified
				}
			}
		}
	}
	return nil
}
