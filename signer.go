package gopayd

import (
	"context"
)

type SignerService interface {
	FundAndSignTx(ctx context.Context, req FundAndSignTxRequest) (*SignTxResponse, error)
}

type FundAndSignTxRequest struct {
	TxHex   string `json:"tx"`
	Account string `json:"account"`
	Fee     Fee    `json:"fee"`
}

type SignTxResponse struct {
	SignedTx string `json:"signedTx"`
}
