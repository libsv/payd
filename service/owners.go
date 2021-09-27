package service

import (
	"context"

	"github.com/libsv/payd"
)

type owner struct {
	str payd.OwnerStore
}

// NewOwnerService returns a new owner service.
func NewOwnerService(str payd.OwnerStore) payd.OwnerService {
	return &owner{
		str: str,
	}
}

// Owner will return the current owner of the wallet.
func (o *owner) Owner(ctx context.Context) (*payd.User, error) {
	return o.str.Owner(ctx)
}
