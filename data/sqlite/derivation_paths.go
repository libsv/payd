package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
	"github.com/theflyingcodr/lathos/errs"

	gopayd "github.com/libsv/payd"
)

const (
	sqlDerivationCounter = `
	SELECT pathcounter FROM keys WHERE name = :name
	`
	sqlDerivationIncrement = `
	UPDATE keys set pathCounter = pathCounter + $1
	WHERE name = $2
	`
)

// DerivationCounter will return the current derivation counter for a private key.
func (s *sqliteStore) DerivationCounter(ctx context.Context, args gopayd.DerivationCounterArgs) (uint64, error) {
	tx, err := s.newTx(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to begin tx when getting keycounter")
	}
	defer func() {
		_ = rollback(ctx, tx)
	}()
	var counter uint64
	if err := tx.GetContext(ctx, &counter, sqlDerivationCounter, args.Key); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errs.NewErrNotFound(102, fmt.Sprintf("no derivation counter found for key '%s', does it exist", args.Key))
		}
		return 0, errors.Wrapf(err, "failed to get derivationCounter for key '%s', does it exist?", args.Key)
	}
	return counter, errors.Wrap(commit(ctx, tx), "failed to commit tx when getting derivation counter")
}

// IncrementKeyCounter will increment a private key derivation counter but the amount requested by offset.
func (s *sqliteStore) IncrementKeyCounter(ctx context.Context, args gopayd.DerivationIncrementArgs) error {
	tx, err := s.newTx(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to begin tx when incrementing counter")
	}
	defer func() {
		_ = rollback(ctx, tx)
	}()
	if err := handleNamedExec(tx, sqlDerivationIncrement, args); err != nil {
		return errors.Wrap(err, "failed to update derivation path")
	}
	return errors.Wrap(tx.Commit(), "failed to commit increment counter")
}
