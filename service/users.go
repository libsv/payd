package service

import (
	"context"
	"database/sql"

	"github.com/libsv/payd"
)

type users struct {
	str payd.UserStore
}

// NewUsersService returns a new owner service.
func NewUsersService(str payd.UserStore) payd.UserService {
	return &users{
		str: str,
	}
}

// Owner will return the current owner of the wallet.
func (u *users) Create(ctx context.Context, user payd.User) (sql.Result, error) {
	return nil, nil
}

// Read will return the current owner of the wallet.
func (u *users) Read(ctx context.Context, handle string) (*payd.User, error) {
	return u.str.Read(ctx, handle)
}

// Update will return the current owner of the wallet.
func (u *users) Update(ctx context.Context, ID uint64, d payd.User) (*payd.User, error) {
	return nil, nil
}

// Delete will return the current owner of the wallet.
func (u *users) Delete(ctx context.Context, ID uint64) (*payd.User, error) {
	return nil, nil
}
