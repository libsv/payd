package gopayd

import "context"

// Owner data about the wallet owner.
type Owner struct {
	Name         string            `json:"name" db:"name"`
	Avatar       string            `json:"avatar" db:"avatar"`
	Email        string            `json:"email" db:"email"`
	Address      string            `json:"address" db:"address"`
	PhoneNumber  string            `json:"phoneNumber" db:"phoneNumber"`
	ExtendedData map[string]string `json:"extendedData"`
}

// OwnerService interfaces with owners.
type OwnerService interface {
	Owner(ctx context.Context) (*Owner, error)
}

// OwnerStore interfaces with an owner store.
type OwnerStore interface {
	Owner(ctx context.Context) (*Owner, error)
}
