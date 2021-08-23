package sqlite

import (
	"context"

	"github.com/pkg/errors"

	gopayd "github.com/libsv/payd"
)

const (
	sqlTxoCreate = `
	INSERT INTO txos (keyname, derivationpath, lockingscript, satoshis)
	VALUES(:keyname, :derivationpath, :lockingscript, :satoshis)
	`
)

// TxosCreate will store a txo created during payment requests.
func (s *sqliteStore) TxoCreate(ctx context.Context, req gopayd.TxoCreate) error {
	return s.TxosCreate(ctx, []*gopayd.TxoCreate{
		&req,
	})
}

// TxosCreate will store txos created during payment requests.
func (s *sqliteStore) TxosCreate(ctx context.Context, req []*gopayd.TxoCreate) error {
	tx, err := s.newTx(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create context when creating a txo")
	}
	defer func() {
		_ = rollback(ctx, tx)
	}()
	if err := handleNamedExec(tx, sqlTxoCreate, req); err != nil {
		return errors.Wrap(err, "failed to insert script keys.")
	}
	return errors.Wrap(commit(ctx, tx), "failed to commit transaction when creating txos.")
}
