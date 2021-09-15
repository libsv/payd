package service

import (
	"context"

	"github.com/libsv/go-bc"
	"github.com/libsv/go-bc/spv"
	"github.com/libsv/payd/cli/models"
)

type mpSvc struct {
	rt models.Regtest
}

// NewMerkleProofStore returns a new merkle proof store.
func NewMerkleProofStore(rt models.Regtest) spv.MerkleProofStore {
	return &mpSvc{
		rt: rt,
	}
}

func (m *mpSvc) MerkleProof(ctx context.Context, txID string) (*bc.MerkleProof, error) {
	rtResp, err := m.rt.RawTransaction1(ctx, txID)
	if err != nil {
		return nil, err
	}

	mpResp, err := m.rt.MerkleProof(ctx, rtResp.Result.BlockHash, txID)
	if err != nil {
		return nil, err
	}
	if mpResp == nil {
		return nil, nil
	}

	return mpResp.Result, nil
}
