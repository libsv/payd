package gopayd

import (
	"context"
)

// SignerService interfaces a signing service.
type SignerService interface {
	FundAndSignTx(ctx context.Context, req FundAndSignTxRequest) (*SignTxResponse, error)
}

// FundAndSignTxRequest the request for signing and funding a tx.
type FundAndSignTxRequest struct {
	TxHex   string `json:"tx"`
	Account string `json:"account"`
	Fee     Fee    `json:"fee"`
}

// SignTxResponse a signed tx response.
type SignTxResponse struct {
	SignedTx string `json:"signedTx"`
}
