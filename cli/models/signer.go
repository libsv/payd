package models

import (
	"context"

	gopayd "github.com/libsv/payd"
)

type Signer interface {
	FundAndSign(ctx context.Context, req gopayd.FundAndSignTxRequest) (*gopayd.SignTxResponse, error)
}
