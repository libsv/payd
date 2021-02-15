package sqlite

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func handleExec(tx sqlx.Execer, sql string, args interface{}) error {
	res, err := tx.Exec(sql, args)
	if err != nil {
		return errors.Wrap(err, "failed to run exec")
	}
	return handleExecRows(res)
}

func handleNamedExec(tx db, sql string, args interface{}) error {
	res, err := tx.NamedExec(sql, args)
	if err != nil {
		return errors.Wrap(err, "failed to run exec")
	}
	return handleExecRows(res)
}

func handleExecRows(res sql.Result) error {
	ra, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to read rows affected")
	}
	if ra <= 0 {
		return errors.Wrap(err, "exec did not affect rows")
	}
	return nil
}

type db interface {
	NamedExec(query string, arg interface{}) (sql.Result, error)
	Get(dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}
