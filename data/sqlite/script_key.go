package sqlite

import (
	"context"

	gopayd "github.com/libsv/go-payd"
	"github.com/pkg/errors"
)

const (
	sqlScriptKeyByScript = `
		SELECT lockingscript, keyname, path
		FROM script_keys as sk INNER JOIN derivation_paths dp on dp.ID = sk.derivationID
		WHERE lockingscript = :lockingscript
	`

	sqlScriptKeysInsert = `
		INSERT INTO script_keys(lockingscript, keyname, derivationID)
		VALUES(:lockingscript,:keyname,:derivationID)
	`
)

// ScriptKey will return a script key matching the supplied args.
func (s *sqliteStore) ScriptKey(ctx context.Context, args gopayd.ScriptKeyArgs) (*gopayd.ScriptKey, error) {
	var resp *gopayd.ScriptKey
	if err := s.db.GetContext(ctx, &resp, sqlScriptKeyByScript, args); err != nil {
		return nil, errors.Wrap(err, "failed to get script key")
	}
	return resp, nil
}

// CreateScriptKeys will add one or many script keys to the data store.
// These can then be used to sign the payment outputs to ensure they are valid.
func (s *sqliteStore) CreateScriptKeys(ctx context.Context, req []gopayd.CreateScriptKey) error {
	tx, err := s.newTx(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create transaction")
	}
	if err := handleNamedExec(tx, sqlScriptKeysInsert, req); err != nil {
		tx.Rollback()
		return errors.Wrap(err, "failed to insert script keys.")
	}
	//return errors.New("aggg")
	return errors.Wrap(commit(ctx, tx), "failed to commit transaction when creating script keys.")
}
