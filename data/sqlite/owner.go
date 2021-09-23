package sqlite

import (
	"context"

	gopayd "github.com/libsv/payd"
	"github.com/pkg/errors"
)

const (
	sqlOwnerGet = `
	SELECT user_id, name, avatar_url, email, address, phone_number
	FROM users
	WHERE is_owner = 1
	`
	sqlOwnerMetaGet = `
	SELECT key, value FROM users_meta where user_id = $1
	`
)

// Owner will return the owner of the wallet.
func (s *sqliteStore) Owner(ctx context.Context) (*gopayd.User, error) {
	owner := gopayd.User{
		ExtendedData: make(map[string]string),
	}

	if err := s.db.GetContext(ctx, &owner, sqlOwnerGet); err != nil {
		return nil, errors.Wrap(err, "failed to get wallet owner")
	}

	meta := make([]struct {
		Key   string `db:"key"`
		Value string `db:"value"`
	}, 0)
	if err := s.db.SelectContext(ctx, &meta, sqlOwnerMetaGet, owner.ID); err != nil {
		return nil, errors.Wrap(err, "failed to get wallet owner extended info")
	}

	for _, m := range meta {
		owner.ExtendedData[m.Key] = m.Value
	}

	return &owner, nil
}
