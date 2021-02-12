package sqlite

import (
	"context"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"

	"github.com/libsv/go-payd/wallet"
)

const (
	keyByName = `
	SELECT name, xprv, createdAt
	FROM keys
	WHERE name = :name
	`

	createKey = `
	INSERT INTO keys(name, xprv)
	VALUES(:name, :xprv)
	`
)

// keys implements bip270.KeyStorer and is used to get and store private keys.
type keys struct {
	db *sqlx.DB
}

// NewKeys will setup and return a new keys store.
func NewKeys(db *sqlx.DB) *keys {
	return &keys{
		db: db,
	}
}

// Key will return a key by name from the datastore.
// If not found an error will be returned.
func (k *keys) PrivateKey(ctx context.Context, args wallet.KeyArgs) (*wallet.PrivateKey, error) {
	var resp wallet.PrivateKey
	if err := k.db.Get(&resp, keyByName, args.Name); err != nil {
		return nil, errors.Wrapf(err, "failed to get key named %s from datastore", args.Name)
	}
	return &resp, nil
}

// Create will create and return a new key in the database.
func (k *keys) Create(ctx context.Context, req wallet.PrivateKey) (*wallet.PrivateKey, error) {
	tx, err := k.db.BeginTxx(ctx, nil)
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
	var resp *wallet.PrivateKey
	if err := tx.Get(resp, keyByName, req); err != nil {
		return nil, errors.Wrapf(err, "failed to get key named %s from datastore", req.Name)
	}
	return nil, errors.Wrap(tx.Commit(), "failed to commit create key tx")
}
