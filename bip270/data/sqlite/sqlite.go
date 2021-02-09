package sqlite

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func handleExec(tx sqlx.Execer, sql string, args interface{}) error {
	res, err := tx.Exec(sql, args)
	if err != nil {
		return errors.Wrap(err, "failed to run exec")
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to read rows affected")
	}
	if ra <= 0 {
		return errors.Wrap(err, "exec did not affect rows")
	}
	return nil
}
