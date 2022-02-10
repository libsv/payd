package payd

import (
	"context"

	"github.com/libsv/go-bk/bip32"
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
	MasterKey    *bip32.ExtendedKey     `json:"-"`
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
	CreateUser(context.Context, CreateUserArgs) (*User, error)
	ReadUser(context.Context, uint64) (*User, error)
	UpdateUser(context.Context, uint64, User) (*User, error)
	DeleteUser(context.Context, uint64) error
}

// UserStore interfaces with a user store.
type UserStore interface {
	CreateUser(context.Context, CreateUserArgs, PrivateKeyService) (*CreateUserResponse, error)
	ReadUser(context.Context, uint64) (*User, error)
	UpdateUser(context.Context, uint64, User) (*User, error)
	DeleteUser(context.Context, uint64) error
}

// CreateUserArgs is what we expect to be sent to create a new user in the payd user store.
type CreateUserArgs struct {
	Name         string                 `json:"name"`
	Email        string                 `json:"email"`
	Avatar       string                 `json:"avatar"`
	Address      string                 `json:"address"`
	PhoneNumber  string                 `json:"phoneNumber"`
	ExtendedData map[string]interface{} `json:"extendedData"`
}

// CreateUserResponse is what we expect to receive once a new user is created.
type CreateUserResponse struct {
	ID uint64 `json:"userId" db:"user_id"`
}
