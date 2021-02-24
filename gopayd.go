package gopayd

import (
	"context"
)

type Transacter interface {
	WithTx(ctx context.Context) context.Context
	Commit(ctx context.Context) error
}
