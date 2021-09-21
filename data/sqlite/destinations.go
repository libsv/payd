package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	lathos "github.com/theflyingcodr/lathos/errs"

	gopayd "github.com/libsv/payd"
)

const (
	sqlDestinationCreate = `
	INSERT INTO destinations (key_name, locking_script, derivation_path, satoshis, state)
	VALUES(:key_name, :locking_script, :derivation_path, :satoshis, 'pending')
	`

	sqlDestinationInvoiceCreate = `
	INSERT INTO destination_invoice(destination_id, invoice_id)
	VALUES(:destination_id, :invoice_id)
	`

	sqlDestinationsByScripts = `
	SELECT destination_id, locking_script, derivation_path, satoshis, state
	FROM destinations
	WHERE locking_script IN(?)
	`
)

func (s *sqliteStore) DestinationsCreate(ctx context.Context, args gopayd.DestinationsCreateArgs, req []gopayd.DestinationCreate) ([]gopayd.Output, error) {
	tx, err := s.newTx(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to setup sql transaction when adding destinations for invoice %s", args.InvoiceID.ValueOrZero())
	}
	defer func() {
		_ = rollback(ctx, tx)
	}()
	if err := handleNamedExec(tx, sqlDestinationCreate, req); err != nil {
		return nil, errors.Wrapf(err, "failed to insert payment destinations for invoiceID '%s'", args.InvoiceID.ValueOrZero())
	}
	ll := make([]string, 0, len(req))
	for _, d := range req {
		ll = append(ll, d.Script)
	}
	query, sqlArgs, err := sqlx.In(sqlDestinationsByScripts, ll)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create sql for getting destinations after creation")
	}
	query = tx.Rebind(query)
	rows, err := tx.Query(query, sqlArgs...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, lathos.NewErrNotFound("N0004", "destinations not found, did the create fail?")
		}
	}
	defer func() {
		_ = rows.Close()
	}()
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to scan destination rows")
	}
	fmt.Println(sqlArgs)
	var dd []gopayd.Output
	for rows.Next() {
		var (
			id             uint64
			lockingScript  string
			derivationPath string
			satoshis       uint64
			state          string
		)
		if err := rows.Scan(&id, &lockingScript, &derivationPath, &satoshis, &state); err != nil {
			return nil, errors.Wrap(err, "failed to scan destination row")
		}
		dd = append(dd, gopayd.Output{
			ID:             id,
			LockingScript:  lockingScript,
			Satoshis:       satoshis,
			DerivationPath: derivationPath,
			State:          state,
		})
	}

	// no invoice just return
	if args.InvoiceID.IsZero() {
		return dd, errors.Wrapf(commit(ctx, tx), "failed to commit transaction when creating payment destinations")
	}

	// add destinations and invoice reference
	destInv := make([]struct {
		DestinationID uint64 `db:"destination_id"`
		InvoiceID     string `db:"invoice_id"`
	}, 0)
	for _, d := range dd {
		destInv = append(destInv, struct {
			DestinationID uint64 `db:"destination_id"`
			InvoiceID     string `db:"invoice_id"`
		}{
			DestinationID: d.ID,
			InvoiceID:     args.InvoiceID.ValueOrZero(),
		})
	}
	if err := handleNamedExec(tx, sqlDestinationInvoiceCreate, destInv); err != nil {
		return nil, errors.Wrapf(err, "failed to insert payment destinations for invoiceID '%s'", args.InvoiceID.ValueOrZero())
	}
	return dd, errors.Wrapf(commit(ctx, tx), "failed to commit transaction when creating payment destinations")
}
