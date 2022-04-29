package spv

import (
	"context"
	"fmt"

	"github.com/libsv/go-bt/v2"
	"github.com/pkg/errors"
)

// CreateTxAncestry builds and returns an spv.TxAncestry for the provided tx.
func (c *creator) CreateTxAncestry(ctx context.Context, tx *bt.Tx) (*AncestryJSON, error) {
	if len(tx.Inputs) == 0 {
		return nil, ErrNoTxInputs
	}

	ancestry := &AncestryJSON{
		TxID:    tx.TxID(),
		RawTx:   tx.String(),
		Parents: make(map[string]*AncestryJSON),
	}

	for _, input := range tx.Inputs {
		pTxID := input.PreviousTxIDStr()

		// If we already have added the tx to the parent ancestry, there's no point in
		// redoing the same work
		if _, ok := ancestry.Parents[pTxID]; ok {
			continue
		}

		// Build a *bt.Tx from its TxID and recursively call this function building
		// for inputs without proofs, until a parent with a Merkle Proof is found.
		pTx, err := c.txc.Tx(ctx, pTxID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get tx %s", pTxID)
		}
		if pTx == nil {
			return nil, fmt.Errorf("could not find tx %s", pTxID)
		}

		// Check the store for a Merkle Proof for the current input.
		mp, err := c.mpc.MerkleProof(ctx, pTxID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get merkle proof for tx %s", pTxID)
		}
		// If a Merkle Proof is found, create the ancestry and skip any further recursion
		if mp != nil {
			ancestry.Parents[pTxID] = &AncestryJSON{
				RawTx: pTx.String(),
				TxID:  pTxID,
				Proof: mp,
			}
			continue
		}

		pEnvelope, err := c.CreateTxAncestry(ctx, pTx)
		if err != nil {
			return nil, err
		}

		ancestry.Parents[pTxID] = pEnvelope
	}

	return ancestry, nil
}
