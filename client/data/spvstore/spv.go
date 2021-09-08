package spvstore

import (
	"context"

	"github.com/libsv/go-bc"
	"github.com/libsv/go-bc/spv"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/payd/client"
)

type spvstore struct {
	rt client.Regtest
}

func NewSPVStore(rt client.Regtest) interface {
	spv.TxStore
	spv.MerkleProofStore
} {
	return &spvstore{rt: rt}
}

func (s *spvstore) Tx(ctx context.Context, txID string) (*bt.Tx, error) {
	rawTx, err := s.rt.RawTransaction(ctx, txID)
	if err != nil {
		return nil, err
	}
	return bt.NewTxFromString(*rawTx.Result)
}

func (s *spvstore) MerkleProof(ctx context.Context, txID string) (*bc.MerkleProof, error) {
	rawTx, err := s.rt.RawTransaction1(ctx, txID)
	if err != nil {
		return nil, err
	}

	mp, err := s.rt.MerkleProof(ctx, rawTx.Result.BlockHash, txID)
	if err != nil {
		return nil, err
	}
	if mp == nil {
		return nil, nil
	}

	return mp.Result, nil
}
