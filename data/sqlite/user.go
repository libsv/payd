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

	sqlGetUserMetaByID = `
		SELECT key, value FROM users_meta where user_id = :user_id
	`

	sqlCreateUserMeta = `INSERT INTO users_meta(user_id, key, value) VALUES (:user_id, :key, :value)`
)

// userMeta is the struct for meta data table.
type userMeta struct {
	UserID uint64      `db:"user_id"`
	Key    string      `db:"key"`
	Value  interface{} `db:"value"`
}

// CreateUser creates a new user in the system.
func (s *sqliteStore) CreateUser(ctx context.Context, req payd.CreateUserArgs, pks payd.PrivateKeyService) (*payd.CreateUserResponse, error) {
	tx, err := s.newTx(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create new user for: %s", req.Name)
	}
	defer func() {
		_ = rollback(ctx, tx)
	}()

	var resp payd.CreateUserResponse
	if err = tx.GetContext(ctx, &resp, sqlCreateUser, req.Name, req.Avatar, req.Email, req.Address, req.PhoneNumber); err != nil {
		return nil, errors.Wrapf(err, "failed to create new user: %s", req.Name)
	}

	meta := make([]userMeta, 0)
	for k, v := range req.ExtendedData {
		meta = append(meta, userMeta{
			UserID: resp.ID,
			Key:    k,
			Value:  v,
		})
	}
	_, err = tx.NamedExec(sqlCreateUserMeta, meta)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create new user meta data: %s", req.Name)
	}

	if err := commit(ctx, tx); err != nil {
		return nil, errors.Wrapf(err, "failed to commit transaction when creating new user: %s", req.Name)
	}

	// Create a new xpriv for this new user
	if err = pks.Create(ctx, "masterkey", resp.ID); err != nil {
		return nil, errors.Wrap(err, "failed to create new xkey")
	}
	return &resp, nil
}

func (s *sqliteStore) ReadUser(ctx context.Context, userID uint64) (*payd.User, error) {
	var data struct {
		ID          uint64 `db:"user_id"`
		Name        string `db:"name"`
		Email       string `db:"email"`
		Avatar      string `db:"avatar_url"`
		Address     string `db:"address"`
		PhoneNumber string `db:"phone_number"`
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

	if err := s.db.SelectContext(ctx, &meta, sqlGetUserMetaByID, userID); err != nil {
		return nil, errors.Wrap(err, "failed to get wallet owner extended info")
	}

	user := payd.User{
		ID:           data.ID,
		Name:         data.Name,
		Email:        data.Email,
		Avatar:       data.Avatar,
		Address:      data.Address,
		PhoneNumber:  data.PhoneNumber,
		ExtendedData: make(map[string]interface{}),
		MasterKey:    xPriv,
	}

	for _, v := range meta {
		user.ExtendedData[v.Key] = v.Value
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
