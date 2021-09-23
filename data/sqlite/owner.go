package sqlite

import (
	"context"

	gopayd "github.com/libsv/payd"
	"github.com/pkg/errors"
)

const (
	sqlOwnerGet = `
	SELECT name, avatar, email, address, phoneNumber
	FROM owners
	`
	sqlOwnerMetaGet = `
	SELECT key, value FROM owner_meta where owner_name = $1
	`
)

func (s *sqliteStore) Owner(ctx context.Context) (*gopayd.Owner, error) {
	owner := gopayd.Owner{
		ExtendedData: make(map[string]string),
	}

	if err := s.db.GetContext(ctx, &owner, sqlOwnerGet); err != nil {
		return nil, errors.Wrap(err, "failed to get owner")
	}

	meta := []struct {
		Key   string `db:"key"`
		Value string `db:"value"`
	}{}
	if err := s.db.SelectContext(ctx, &meta, sqlOwnerMetaGet, owner.Name); err != nil {
		return nil, errors.Wrap(err, "failed to get owner extended info")
	}

	for _, m := range meta {
		owner.ExtendedData[m.Key] = m.Value
	}

	return &owner, nil
}
