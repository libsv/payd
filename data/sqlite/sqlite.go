package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/theflyingcodr/lathos"
)

type sqliteStore struct {
	db *sqlx.DB
}

func NewSQLiteStore(db *sqlx.DB) *sqliteStore {
	return &sqliteStore{db: db}
}

func (s *sqliteStore) newTx(ctx context.Context) (*sqlx.Tx, error) {
	ctxx := TxFromContext(ctx)
	if ctxx != nil {
		if ctxx.Tx == nil {
			t, err := s.db.BeginTxx(ctx, nil)
			if err != nil {
				return nil, err
			}
			ctxx.Tx = t
		}
		return ctxx.Tx, nil
	}
	return s.db.BeginTxx(ctx, nil)
}

// commit
func commit(ctx context.Context, tx *sqlx.Tx) error {
	ctxx := TxFromContext(ctx)
	if ctxx != nil {
		if ctxx.Tx != nil {
			return nil
		}
	}
	return tx.Commit()
}

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

func dbErr(err error, errCode, message string) error {
	if err == nil {
		return err
	}
	if err == sql.ErrNoRows {
		return lathos.NewErrNotFound(errCode, message)
	}
	return errors.WithMessage(err, message)
}

func dbErrf(err error, errCode, format string, args ...interface{}) error {
	if err == nil {
		return err
	}
	if err == sql.ErrNoRows {
		return lathos.NewErrNotFound(errCode, fmt.Sprintf(format, args...))
	}
	return errors.WithMessage(err, fmt.Sprintf(format, args...))
}

type db interface {
	NamedExec(query string, arg interface{}) (sql.Result, error)
	Get(dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
}

type execKey int

var exec execKey

type Tx struct {
	*sqlx.Tx
}

func WithTxContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, exec, &Tx{})
}

func TxFromContext(ctx context.Context) *Tx {
	if tx, ok := ctx.Value(exec).(*Tx); ok {
		return tx
	}
	return nil
}

type SQLiteTransacter struct {
}

func (t *SQLiteTransacter) WithTx(ctx context.Context) context.Context {
	return WithTxContext(ctx)
}
func (t *SQLiteTransacter) Commit(ctx context.Context) error {
	tx := TxFromContext(ctx)
	if tx.Tx != nil {
		return tx.Tx.Commit()
	}
	return nil
}
