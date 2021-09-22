package models

import (
	"context"
)

type FundAndSignTxRequest struct {
	Tx string
}

type SignTxResponse struct {
	SignedTx string
}

// Signer interfaces the signing of a tx.
type Signer interface {
	FundAndSign(ctx context.Context, req FundAndSignTxRequest) (*SignTxResponse, error)
}
