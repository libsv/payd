package sqlite

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	gopayd "github.com/libsv/go-payd"
	"github.com/libsv/go-payd/sqlite/queries"
)

type keys struct {
	db *sqlx.DB
}

func NewKeys(db *sqlx.DB) *keys {
	return &keys{
		db: db,
	}
}

// Create will create a new key in the database.
func (k *keys) Create(ctx context.Context, req gopayd.Key) (*gopayd.Key, error) {
	tx, err := k.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin tx when creating key")
	}
	defer tx.Rollback()
	res, err := tx.NamedExec(queries.CreateKey, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add key named '%s'", req.Name)
	}
	if  res.RowsAffected()
	return nil, errors.Wrap(tx.Commit(), "failed to commit create key tx")
}
