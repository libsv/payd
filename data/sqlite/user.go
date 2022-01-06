package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/payd"
	"github.com/pkg/errors"
)

func (s *sqliteStore) Create(ctx context.Context, user payd.User) (sql.Result, error) {
	sqlCreateUser := fmt.Sprintf(`
		INSERT INTO users(name, is_owner, handle, avatar_url, email, address, phone_number) 
		VALUES('%s', 0, '%s', '%s', '%s', '%s');
	`, user.Name, user.Avatar, user.Email, user.Address, user.PhoneNumber)
	res, err := s.db.ExecContext(ctx, sqlCreateUser)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get wallet owner")
	}
	return res, nil
}

func (s *sqliteStore) Read(ctx context.Context, handle string) (*payd.User, error) {
	user := payd.User{
		ExtendedData: make(map[string]interface{}),
	}

	sqlGetUserIDFromHandle := fmt.Sprintf(`
		SELECT user_id
		FROM paymail_handles
		WHERE (handle = "%s")
	`, handle)

	if err := s.db.GetContext(ctx, &user, sqlGetUserIDFromHandle); err != nil {
		return nil, errors.Wrap(err, "failed to get wallet owner")
	}

	sqlGetUserByID := fmt.Sprintf(`
		SELECT user_id, name, avatar_url, email, address, phone_number
		FROM users
		WHERE (user_id = %d)
	`, user.ID)

	if err := s.db.GetContext(ctx, &user, sqlGetUserByID); err != nil {
		return nil, errors.Wrap(err, "failed to get wallet owner")
	}

	sqlGetKeysByID := fmt.Sprintf(`
		SELECT user_id, xprv
		FROM keys
		WHERE (user_id = %d)
	`, user.ID)

	keys := struct {
		Xpriv string `db:"xpriv"`
	}{}

	if err := s.db.GetContext(ctx, &keys, sqlGetKeysByID); err != nil {
		return nil, errors.Wrap(err, "failed to get wallet owner")
	}

	xPriv, err := bip32.NewKeyFromString(keys.Xpriv)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse key from database xpriv")
	}
	pki, err := xPriv.DerivePublicKeyFromPath("0/0/0")
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse key from database xpriv")
	}

	meta := make([]struct {
		Key   string `db:"key"`
		Value string `db:"value"`
	}, 0)
	if err := s.db.SelectContext(ctx, &meta, sqlOwnerMetaGet, user.ID); err != nil {
		return nil, errors.Wrap(err, "failed to get wallet owner extended info")
	}

	for _, m := range meta {
		user.ExtendedData[m.Key] = m.Value
	}

	user.ExtendedData["pki"] = string(pki)

	return &user, nil
}

func (s *sqliteStore) Update(ctx context.Context, ID uint64, d payd.User) (*payd.User, error) {
	return nil, nil
}

func (s *sqliteStore) Delete(ctx context.Context, ID uint64) (*payd.User, error) {
	return nil, nil
}
