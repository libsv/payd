package service

import (
	"context"

	"github.com/libsv/payd"
)

type users struct {
	str payd.UserStore
	pks payd.PrivateKeyService
}

// NewUsersService returns a new owner service.
func NewUsersService(str payd.UserStore, pks payd.PrivateKeyService) payd.UserService {
	return &users{
		str: str,
		pks: pks,
	}
}

// Owner will return the current owner of the wallet.
func (u *users) CreateUser(ctx context.Context, user payd.CreateUserArgs) (*payd.User, error) {
	resp, err := u.str.CreateUser(ctx, user, u.pks)
	if err != nil {
		return nil, err
	}
	usr := &payd.User{
		ID:           resp.ID,
		Name:         user.Name,
		Email:        user.Email,
		Avatar:       user.Avatar,
		Address:      user.Address,
		PhoneNumber:  user.PhoneNumber,
		ExtendedData: user.ExtendedData,
	}
	return usr, nil
}

// Read will return the current owner of the wallet.
func (u *users) ReadUser(ctx context.Context, userID uint64) (*payd.User, error) {
	return u.str.ReadUser(ctx, userID)
}

// Update will return the current owner of the wallet.
func (u *users) UpdateUser(ctx context.Context, userID uint64, d payd.User) (*payd.User, error) {
	return nil, nil
}

// Delete will return the current owner of the wallet.
func (u *users) DeleteUser(ctx context.Context, userID uint64) error {
	return u.str.DeleteUser(ctx, userID)
}
