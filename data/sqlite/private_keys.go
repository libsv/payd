package sqlite

import (
	"context"
	"database/sql"

	"github.com/libsv/payd"

	// test here.
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

const (
	keyByName = `
	SELECT name, xprv, createdAt
	FROM keys
	WHERE name = :name
	`

	createKey = `
	INSERT INTO keys(user_id, name, xprv)
	VALUES(1, :name, :xprv)
	`
)

// Key will return a key by name from the datastore.
// If not found an error will be returned.
func (s *sqliteStore) PrivateKey(ctx context.Context, args payd.KeyArgs) (*payd.PrivateKey, error) {
	var resp payd.PrivateKey
	if err := s.db.Get(&resp, keyByName, args.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "failed to get key named %s from datastore", args.Name)
	}
	return &resp, nil
}

// PrivateKeyCreate will create and return a new key in the database.
func (s *sqliteStore) PrivateKeyCreate(ctx context.Context, req payd.PrivateKey) (*payd.PrivateKey, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin tx when creating key")
	}
	defer tx.Rollback()
	res, err := tx.NamedExec(createKey, req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to add key named '%s'", req.Name)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get rows affected when creating private key")
	}
	if rows <= 0 {
		return nil, errors.Wrap(err, "no rows affected when creating private key")
	}
	var resp payd.PrivateKey
	if err := tx.Get(&resp, keyByName, req.Name); err != nil {
		return nil, errors.Wrapf(err, "failed to get key named %s from datastore", req.Name)
	}
	return &resp, errors.Wrap(tx.Commit(), "failed to commit create key tx")
}
