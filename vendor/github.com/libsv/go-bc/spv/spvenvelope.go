package spv

import (
	"context"
	"fmt"

	"github.com/libsv/go-bc"
	"github.com/libsv/go-bt/v2"
	"github.com/pkg/errors"
)

var (
	// ErrNoTxInputs returns if an envelope is attempted to be created from a transaction that has no inputs
	ErrNoTxInputs = errors.New("provided tx has no inputs to build envelope from")

	// ErrPaymentNotVerified returns if a transaction in the tree provided was missed during verification
	ErrPaymentNotVerified = errors.New("a tx was missed during validation")

	// ErrTipTxConfirmed returns if the tip transaction is already confirmed
	ErrTipTxConfirmed = errors.New("tip transaction must be unconfirmed")

	// ErrNoConfirmedTransaction returns if a path from tip to beginning/anchor contains no confirmed transcation
	ErrNoConfirmedTransaction = errors.New("not confirmed/anchored tx(s) provided")

	// ErrTxIDMismatch returns if they key value pair of a transactions input has a mismatch in txID
	ErrTxIDMismatch = errors.New("input and proof ID mismatch")

	// ErrNotAllInputsSupplied returns if an unconfirmed transaction in envelope contains inputs which are not
	// present in the parent envelope
	ErrNotAllInputsSupplied = errors.New("a tx input missing in parent envelope")

	// ErrNoTxInputsToVerify returns if a transaction has no inputs
	ErrNoTxInputsToVerify = errors.New("a tx has no inputs to verify")

	// ErrNilInitialPayment returns if a transaction has no inputs
	ErrNilInitialPayment = errors.New("initial payment cannot be nil")

	// ErrInputRefsOutOfBoundsOutput returns if a transaction has no inputs
	ErrInputRefsOutOfBoundsOutput = errors.New("tx input index into output is out of bounds")
)

// Envelope is a struct which contains all information needed for a transaction to be verified.
//
// spec at https://tsc.bitcoinassociation.net/standards/spv-envelope/
type Envelope struct {
	TxID          string               `json:"txid,omitempty"`
	RawTx         string               `json:"rawTx,omitempty"`
	Proof         *bc.MerkleProof      `json:"proof,omitempty"`
	MapiResponses []bc.MapiCallback    `json:"mapiResponses,omitempty"`
	Parents       map[string]*Envelope `json:"parents,omitempty"`
}

func (s *spvclient) CreateEnvelope(tx *bt.Tx) (*Envelope, error) {
	if len(tx.Inputs) == 0 {
		return nil, ErrNoTxInputs
	}

	envelope := &Envelope{
		TxID:    tx.TxID(),
		RawTx:   tx.String(),
		Parents: make(map[string]*Envelope),
	}

	for _, input := range tx.Inputs {
		pTxID := input.PreviousTxIDStr()

		// If we already have added the tx to the parent envelope, there's no point in
		// redoing the same work
		if _, ok := envelope.Parents[pTxID]; ok {
			continue
		}

		mp, err := s.mpg.MerkleProof(pTxID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get merkle proof for tx %s", pTxID)
		}
		if mp != nil {
			envelope.Parents[pTxID] = &Envelope{
				TxID:  pTxID,
				Proof: mp,
			}

			// Skip getting the tx data as we have everything we need for verifying the current tx.
			continue
		}

		pTx, err := s.txg.Tx(pTxID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get tx %s", pTxID)
		}
		if pTx == nil {
			return nil, fmt.Errorf("could not find tx %s", pTxID)
		}

		pEnvelope, err := s.CreateEnvelope(pTx)
		if err != nil {
			return nil, err
		}

		envelope.Parents[pTxID] = pEnvelope
	}

	return envelope, nil
}

// VerifyPayment verifies whether or not the txs supplied via the supplied spv.Envelope are valid
func (s *spvclient) VerifyPayment(ctx context.Context, initialPayment *Envelope) (bool, error) {
	if initialPayment == nil {
		return false, ErrNilInitialPayment
	}

	// The tip tx is the transaction we're trying to verify, and it should not have a supplied
	// Merkle Proof.
	if initialPayment.IsAnchored() {
		return false, ErrTipTxConfirmed
	}

	valid, err := s.verifyTxs(ctx, initialPayment)
	if err != nil {
		return false, err
	}

	return valid, nil
}

func (s *spvclient) verifyTxs(ctx context.Context, payment *Envelope) (bool, error) {
	// If at the beginning or middle of the tx chain and tx is unconfirmed, fail and error.
	if !payment.IsAnchored() && (payment.Parents == nil || len(payment.Parents) == 0) {
		return false, ErrNoConfirmedTransaction
	}

	// Recurse back to the anchor transactions of the transaction chain and verify forward towards
	// the tip transaction. This way, we check that the first transactions in the chain are anchored
	// to the blockchain through a valid Merkle Proof.
	for parentTxID, parent := range payment.Parents {
		if parent.TxID == "" {
			parent.TxID = parentTxID
		}

		valid, err := s.verifyTxs(ctx, parent)
		if err != nil {
			return false, err
		}
		if !valid {
			return false, nil
		}
	}

	// If a Merkle Proof is provided, assume we are at the anchor/beginning of the tx chain.
	// Verify and return the result.
	if payment.IsAnchored() {
		return s.verifyTxAnchor(ctx, payment)
	}

	tx, err := bt.NewTxFromString(payment.RawTx)
	if err != nil {
		return false, err
	}

	// We must verify the tx or else we can not know if any of it's child txs are valid.
	return s.verifyUnconfirmedTx(tx, payment)
}

func (s *spvclient) verifyTxAnchor(ctx context.Context, payment *Envelope) (bool, error) {
	proofTxID := payment.Proof.TxOrID
	if len(proofTxID) != 64 {
		proofTx, err := bt.NewTxFromString(payment.Proof.TxOrID)
		if err != nil {
			return false, err
		}

		proofTxID = proofTx.TxID()
	}

	// If the txid of the Merkle Proof doesn't match the txid provided in the spv.Envelope,
	// fail and error
	if proofTxID != payment.TxID {
		return false, ErrTxIDMismatch
	}

	valid, _, err := s.VerifyMerkleProofJSON(ctx, payment.Proof)
	if err != nil {
		return false, err
	}

	return valid, nil
}

func (s *spvclient) verifyUnconfirmedTx(tx *bt.Tx, payment *Envelope) (bool, error) {
	// If no tx inputs have been provided, fail and error
	if len(tx.Inputs) == 0 {
		return false, ErrNoTxInputsToVerify
	}

	for _, input := range tx.Inputs {
		parent, ok := payment.Parents[input.PreviousTxIDStr()]
		if !ok {
			return false, ErrNotAllInputsSupplied
		}

		parentTx, err := bt.NewTxFromString(parent.RawTx)
		if err != nil {
			return false, err
		}

		// If the input is indexing an output that is out of bounds, fail and error
		if int(input.PreviousTxOutIndex) > len(parentTx.Outputs)-1 {
			return false, ErrInputRefsOutOfBoundsOutput
		}

		output := parentTx.Outputs[int(input.PreviousTxOutIndex)]

		// TODO: verify script using input and previous output
		_ = output
	}

	return true, nil
}

// IsAnchored returns true if the envelope is the anchor tx.
func (s *Envelope) IsAnchored() bool {
	return s.Proof != nil
}
