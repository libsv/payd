package models

import (
	"context"

	gopayd "github.com/libsv/payd"
)

// Signer interfaces the signing of a tx.
type Signer interface {
	FundAndSign(ctx context.Context, req gopayd.FundAndSignTxRequest) (*gopayd.SignTxResponse, error)
}
