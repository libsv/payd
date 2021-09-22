package gopayd

import "context"

type OwnerRequest struct {
	Name string `db:"name"`
}

type Owner struct {
	Name         string            `json:"name" db:"name"`
	Avatar       string            `json:"avatar" db:"avatar"`
	Email        string            `json:"email" db:"email"`
	Address      string            `json:"address" db:"address"`
	PhoneNumber  string            `json:"phoneNumber" db:"phoneNumber"`
	ExtendedData map[string]string `json:"extendedData"`
}

type OwnerService interface {
	Owner(ctx context.Context) (*Owner, error)
}

type OwnerStore interface {
	Owner(ctx context.Context) (*Owner, error)
}
