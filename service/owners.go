package service

import (
	"context"

	gopayd "github.com/libsv/payd"
)

type owner struct {
	str gopayd.OwnerStore
}

func NewOwnerService(str gopayd.OwnerStore) *owner {
	return &owner{
		str: str,
	}
}

func (o *owner) Owner(ctx context.Context) (*gopayd.Owner, error) {
	return o.str.Owner(ctx)
}
