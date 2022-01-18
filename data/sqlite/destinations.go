package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/pkg/errors"
	lathos "github.com/theflyingcodr/lathos/errs"

	"github.com/libsv/payd"
	"github.com/libsv/payd/internal"
)

const (
	sqlDestinationCreateFromMasterKey = `
	INSERT INTO destinations (key_name, locking_script, derivation_path, satoshis, state)
	VALUES("masterkey", :locking_script, :derivation_path, :satoshis, 'pending')
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

	sqlDestinationsByInvoiceID = `
	SELECT d.destination_id, d.locking_script, d.derivation_path, d.satoshis, d.state
	FROM destinations as d INNER JOIN destination_invoice as di ON d.destination_id = di.destination_id
	WHERE di.invoice_id = :invoice_id 
	`
)

func (s *sqliteStore) DestinationsCreate(ctx context.Context, args payd.DestinationsCreateArgs, req []payd.DestinationCreate) ([]payd.Output, error) {
	tx, err := s.newTx(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to setup sql transaction when adding destinations for invoice %s", args.InvoiceID.ValueOrZero())
	}
	defer func() {
		_ = rollback(ctx, tx)
	}()
	if err := handleNamedExec(tx, sqlDestinationCreateFromMasterKey, req); err != nil {
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
	var dd []payd.Output
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
		dd = append(dd, payd.Output{
			ID:             id,
			LockingScript:  internal.StringToScript(lockingScript),
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

type dbOutput struct {
	ID uint64 `db:"destination_id"`
	// LockingScript is the P2PKH locking script used.
	LockingScript string `db:"locking_script"`
	Satoshis      uint64 `json:"satoshis" db:"satoshis"`
	// DerivationPath is the deterministic path for this destination.
	DerivationPath string `db:"derivation_path"`
	// State will indicate if this destination is still waiting on a tx to fulfil it (pending)
	// has been paid to in a tx (received) or has been deleted.
	State string `db:"state"  enums:"pending,received,deleted"`
}

// Destinations will return a set of destination outputs for a specific invoiceID.
func (s *sqliteStore) Destinations(ctx context.Context, args payd.DestinationsArgs) ([]payd.Output, error) {
	var oo []dbOutput
	if err := s.db.SelectContext(ctx, &oo, sqlDestinationsByInvoiceID, args.InvoiceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, lathos.NewErrNotFound("N0002", fmt.Sprintf("destinations with invoiceID %s not found", args.InvoiceID))
		}
		return nil, errors.Wrapf(err, "failed to get destinations with invoiceID %s", args.InvoiceID)
	}
	if len(oo) == 0 {
		return nil, lathos.NewErrNotFound("N0002", fmt.Sprintf("destinations with invoiceID %s not found", args.InvoiceID))
	}
	outs := make([]payd.Output, len(oo))
	for i := 0; i < len(oo); i++ {
		s, _ := bscript.NewFromHexString(oo[i].LockingScript)
		outs[i] = payd.Output{
			ID:             oo[i].ID,
			LockingScript:  s,
			Satoshis:       oo[i].Satoshis,
			DerivationPath: oo[i].DerivationPath,
			State:          oo[i].State,
		}
	}
	return outs, nil
}
