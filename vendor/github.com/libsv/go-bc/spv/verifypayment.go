package spv

import (
	"context"

	"github.com/libsv/go-bt/v2"
)

// VerifyPayment is a method for parsing a binary payment transaction and its corresponding ancestry in binary.
// It will return the paymentTx struct if all validations pass.
func (v *verifier) VerifyPayment(ctx context.Context, pTx *bt.Tx, ancestors []byte, opts ...VerifyOpt) (*bt.Tx, error) {
	vOpt := v.opts.clone()
	for _, opt := range opts {
		opt(vOpt)
	}
	ancestry, err := NewAncestryFromBytes(ancestors)
	if err != nil {
		return nil, err
	}
	ancestry.PaymentTx = pTx
	err = VerifyAncestors(ctx, ancestry, v, vOpt)
	if err != nil {
		return nil, err
	}
	return pTx, nil
}
