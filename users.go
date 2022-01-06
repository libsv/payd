package payd

import (
	"context"
	"database/sql"
)

// User information on wallet users.
type User struct {
	ID           uint64                 `json:"id" db:"user_id"`
	Name         string                 `json:"name" db:"name"`
	Email        string                 `json:"email" db:"email"`
	Avatar       string                 `json:"avatar" db:"avatar_url"`
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

// UserService interfaces with users.
type UserService interface {
	Create(ctx context.Context, user User) (sql.Result, error)
	Read(ctx context.Context, handle string) (*User, error)
	Update(ctx context.Context, ID uint64, d User) (*User, error)
	Delete(ctx context.Context, ID uint64) (*User, error)
}

// UserStore interfaces with a user store.
type UserStore interface {
	Create(ctx context.Context, user User) (sql.Result, error)
	Read(ctx context.Context, handle string) (*User, error)
	Update(ctx context.Context, ID uint64, user User) (*User, error)
	Delete(ctx context.Context, ID uint64) (*User, error)
}
