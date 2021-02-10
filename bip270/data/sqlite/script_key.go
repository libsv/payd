package sqlite

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/libsv/go-payd/bip270"
)

const (
	insertScriptKeys = `
	INSERT INTO script_keys(lockingscript, keyname, derivationpath)
	VALUES(:lockingscript,:keyname,:derivationpath)
	`

	scriptKeyByScript = `
	SELECT id, lockingscript, keyname, derivationpath
	FROM script_keys
	WHERE lockingscript = :lockingscript
	`
)

type scriptKey struct {
	db *sqlx.DB
}

func NewScriptKey(db *sqlx.DB) *scriptKey {
	return &scriptKey{db: db}
}

// Create will add one or many script keys to the data store.
// These can then be used to sign the payment outputs to ensure they are valid.
func (s *scriptKey) Create(ctx context.Context, req []bip270.CreateScriptKey) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create transaction")
	}
	defer tx.Rollback()
	res, err := tx.NamedExec(insertScriptKeys, req)
	if err != nil {
		return errors.Wrap(err, "failed to insert script keys.")
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to insert script keys - couldn't get rows affected.")
	}
	if rows <= 0 {
		return errors.Wrap(err, "failed to insert script keys - no rows affected.")
	}
	return errors.Wrap(tx.Commit(), "failed to commit transaction when creating script keys.")
}

// ScriptKey will return a script key matching the supplied args.
func (s *scriptKey) ScriptKey(ctx context.Context, args bip270.ScriptKeyArgs) (*bip270.ScriptKey, error) {
	var resp *bip270.ScriptKey
	if err := s.db.Get(&resp, scriptKeyByScript, args); err != nil {
		return nil, errors.Wrap(err, "failed to get script key")
	}
	return resp, nil
}
