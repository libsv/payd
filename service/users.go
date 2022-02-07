package service

import (
	"context"

	"github.com/libsv/payd"
	"github.com/pkg/errors"
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

// CreateUser allows you to add user data to the payd instance, and it will return the same data plus a user_id to confirm acceptance.
func (u *users) CreateUser(ctx context.Context, user payd.CreateUserArgs) (*payd.User, error) {
	// Check for a valid set of data
	if user.Name == "" || user.Email == "" {
		return nil, errors.New("Please specify a name and email address for the user")
	}
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

// ReadUser will return the  user associated with a particular user_id of the wallet.
func (u *users) ReadUser(ctx context.Context, userID uint64) (*payd.User, error) {
	return u.str.ReadUser(ctx, userID)
}

// UpdateUser is not required for MVP, not implemented.
func (u *users) UpdateUser(ctx context.Context, userID uint64, d payd.User) (*payd.User, error) {
	return nil, nil
}

// DeleteUser is not required for MVP, is implemented but not attached to any endpoints.
func (u *users) DeleteUser(ctx context.Context, userID uint64) error {
	return u.str.DeleteUser(ctx, userID)
}
