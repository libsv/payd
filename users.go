package payd

import "context"

// User information on wallet users.
type User struct {
	ID           uint64                 `json:"id" db:"user_id"`
	Name         string                 `json:"name" db:"name"`
	Avatar       string                 `json:"avatar" db:"avatar_url"`
	Email        string                 `json:"email" db:"email"`
	Address      string                 `json:"address" db:"address"`
	PhoneNumber  string                 `json:"phoneNumber" db:"phone_number"`
	ExtendedData map[string]interface{} `json:"extendedData"`
}

// OwnerService interfaces with owners.
type OwnerService interface {
	Owner(ctx context.Context) (*User, error)
}

// OwnerStore interfaces with an owner store.
type OwnerStore interface {
	Owner(ctx context.Context) (*User, error)
}
