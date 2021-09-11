package service

import (
	"context"

	"github.com/libsv/go-bc"
	"github.com/libsv/go-bc/spv"
	"github.com/libsv/go-bt/v2"
)

type spvEnvelopeBuilder struct {
}

func (s *spvEnvelopeBuilder) MerkleProof(ctx context.Context, txID string) (*bc.MerkleProof, error) {
	panic("not implemented") // TODO: Implement
}

func (s *spvEnvelopeBuilder) Tx(ctx context.Context, txID string) (*bt.Tx, error) {
	panic("not implemented") // TODO: Implement
}

func NewSPVEnvelopeBuilder() interface {
	spv.MerkleProofStore
	spv.TxStore
} {
	return &spvEnvelopeBuilder{}
}
