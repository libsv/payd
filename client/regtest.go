package client

import (
	"context"

	"github.com/libsv/payd/client/data/regtest/models"
)

type Regtest interface {
	RawTransaction(ctx context.Context, txID string) (*models.RawTxResponse, error)
	RawTransaction1(ctx context.Context, txID string) (*models.RawTx1Response, error)
	MerkleProof(ctx context.Context, blockHash, txID string) (*models.MerkleProofResponse, error)
	SendToAddress(ctx context.Context, address string, amount float64) (*models.SendToAddressResponse, error)
}
