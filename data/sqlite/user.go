package sqlite

import (
	"context"

	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/payd"
	"github.com/pkg/errors"
)

const (
	sqlCreateUser = `
		INSERT INTO users(name, is_owner, avatar_url, email, address, phone_number)
		VALUES(:name, 0, :avatar_url, :email, :address, :phone_number)
		RETURNING user_id
	`

	sqlGetUserByID = `
		SELECT u.user_id, u.name, u.avatar_url, u.email, u.phone_number, u.address, k.xprv 
		FROM users u
		JOIN keys k ON u.user_id = k.user_id
		WHERE k.user_id = :user_id
	`

	sqlDeleteUserByID = `
		DELETE FROM users
		WHERE user_id = :user_id
	`
)

func (s *sqliteStore) CreateUser(ctx context.Context, req payd.CreateUserArgs, pks payd.PrivateKeyService) (*payd.CreateUserResponse, error) {
	tx, err := s.newTx(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create new user for: %s", req.Name)
	}
	defer func() {
		_ = rollback(ctx, tx)
	}()
	var resp payd.CreateUserResponse
	err = tx.GetContext(ctx, &resp, sqlCreateUser, req.Name, req.Avatar, req.Email, req.Address, req.PhoneNumber) //:name, :avatar_url, :email, :address, :phone_number
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create new user: %s", req.Name)
	}
	if err := commit(ctx, tx); err != nil {
		return nil, errors.Wrapf(err, "failed to commit transaction when creating new user: %s", req.Name)
	}
	// Create a new xpriv for this new user
	err = pks.Create(ctx, "masterkey", resp.ID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new xkey")
	}
	return &resp, nil
}

func (s *sqliteStore) ReadUser(ctx context.Context, userID uint64) (*payd.User, error) {
	var data struct {
		ID          uint64 `db:"user_id"`
		Name        string `db:"name"`
		Email       string `db:"email"`
		PhoneNumber string `db:"phone_number"`
		Address     string `db:"address"`
		Avatar      string `db:"avatar_url"`
		Xprv        string `db:"xprv"`
	}

	if err := s.db.GetContext(ctx, &data, sqlGetUserByID, userID); err != nil {
		return nil, errors.Wrap(err, "failed to get wallet owner")
	}

	xPriv, err := bip32.NewKeyFromString(data.Xprv)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse key from database xpriv")
	}

	meta := make([]struct {
		Key   string `db:"key"`
		Value string `db:"value"`
	}, 0)
	if err := s.db.SelectContext(ctx, &meta, sqlOwnerMetaGet, userID); err != nil {
		return nil, errors.Wrap(err, "failed to get wallet owner extended info")
	}

	user := payd.User{
		ID:           data.ID,
		Name:         data.Name,
		Email:        data.Email,
		Avatar:       data.Avatar,
		Address:      data.Address,
		PhoneNumber:  data.PhoneNumber,
		ExtendedData: make(map[string]interface{}, 3),
		MasterKey:    xPriv,
	}

	for _, m := range meta {
		user.ExtendedData[m.Key] = m.Value
	}

	return &user, nil
}

func (s *sqliteStore) UpdateUser(ctx context.Context, ID uint64, d payd.User) (*payd.User, error) {
	return nil, nil
}

func (s *sqliteStore) DeleteUser(ctx context.Context, userID uint64) error {
	data := struct {
		ID uint64 `db:"user_id"`
	}{
		ID: userID,
	}
	_, err := s.db.NamedExec(sqlDeleteUserByID, data)
	if err != nil {
		return errors.Wrap(err, "failed to get wallet owner")
	}
	return nil
}
