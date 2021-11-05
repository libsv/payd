package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/payd"
	"github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	lathos "github.com/theflyingcodr/lathos/errs"
	"gopkg.in/guregu/null.v3"
)

const (
	sqlTransactionCreate = `
		INSERT INTO transactions(tx_id, tx_hex)
		VALUES(:tx_id, :tx_hex)
	`

	sqlTransactionInvoiceCreate = `
		INSERT INTO transaction_invoice(tx_id, invoice_id)
		VALUES(:tx_id, :invoice_id)
	`

	sqlTxoCreate = `
		INSERT INTO txos(outpoint, destination_id, tx_id, vout)
		VALUES(:outpoint,:destination_id, :tx_id, :vout)
	`

	sqlDestinationSetReceived = `
		UPDATE destinations
		SET state = 'received', updated_at = ?
		WHERE destination_id IN(?)
	`

	sqlTransactionUpdateState = `
		UPDATE transactions
		SET state = ?
		WHERE tx_id = ?
	`

	sqlInvoiceSetPaid = `
	UPDATE invoices 
	SET payment_received_at = :timestamp, refund_to=:refundto state = 'paid', updated_at = :timestamp
	WHERE invoice_id = :invoice_id
	`

	sqlTransactionGet = `
	SELECT tx_hex
	FROM transactions
	WHERE tx_id=$1
	`
)

// TransactionCreate will store a transaction and its txos in the data base.
func (s *sqliteStore) TransactionCreate(ctx context.Context, req payd.TransactionCreate) error {
	tx, err := s.newTx(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to start transaction when inserting transaction to db")
	}
	defer func() {
		_ = rollback(ctx, tx)
	}()
	timestamp := time.Now().UTC()
	// insert tx and utxos
	if err := handleNamedExec(tx, sqlTransactionCreate, req); err != nil {
		var sqlErr sqlite3.Error
		if ok := errors.As(err, sqlErr); ok {
			if sqlErr.Code == sqlite3.ErrConstraint {
				return lathos.NewErrDuplicate("D001", "transaction has already been stored")
			}
		}
		return errors.Wrap(err, "failed to insert new transaction")
	}

	if err = s.insertOutputs(ctx, tx, timestamp, req); err != nil {
		return err
	}

	// If no invoice id was provided, end here as this is a change tx.
	if req.InvoiceID == "" {
		return errors.Wrapf(commit(ctx, tx),
			"failed to commit transaction when adding tx and outputs for tx '%s'", req.TxID)
	}

	invUpdate := struct {
		Timestamp time.Time   `db:"timestamp"`
		InvoiceID string      `db:"invoice_id"`
		RefundTo  null.String `db:"refundto"`
	}{
		Timestamp: timestamp,
		InvoiceID: req.InvoiceID,
		RefundTo:  req.RefundTo,
	}
	if err = handleNamedExec(tx, sqlTransactionInvoiceCreate, req); err != nil {
		return errors.Wrapf(err, "failed to create invoice mapping for tx %s invoice %s", req.TxID, req.InvoiceID)
	}

	if err = handleNamedExec(tx, sqlInvoiceSetPaid, invUpdate); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return lathos.NewErrNotFound("N0007", fmt.Sprintf("invoiceID '%s' not found when updating payment received info", req.InvoiceID))
		}
	}

	return errors.Wrapf(commit(ctx, tx),
		"failed to commit transaction when adding tx and outputs for tx '%s'", req.TxID)
}

func (s *sqliteStore) insertOutputs(ctx context.Context, tx *sqlx.Tx, timestamp time.Time, req payd.TransactionCreate) error {
	// Only write outputs if they exist. This can happen in the case of a tx being funded an
	// exact amount, causing no change output to be created.
	if req.Outputs == nil || len(req.Outputs) == 0 {
		return nil
	}

	if err := handleNamedExec(tx, sqlTxoCreate, req.Outputs); err != nil {
		return errors.Wrap(err, "failed to insert transaction outputs")
	}

	ll := make([]uint64, 0, len(req.Outputs))
	for _, d := range req.Outputs {
		ll = append(ll, d.DestinationID)
	}
	query, sqlArgs, err := sqlx.In(sqlDestinationSetReceived, timestamp, ll)
	if err != nil {
		return errors.Wrap(err, "failed to create sql for updating destination state")
	}
	result, err := tx.ExecContext(ctx, query, sqlArgs...)
	if err != nil {
		return errors.Wrap(err, "failed to update destinations state to received")
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to update destinations state to received")
	}
	if rows <= 0 {
		return errors.Wrap(err, "failed to update destinations state to received")
	}

	return nil
}

// TransactionUpdateState will update a transactions internal state.
func (s *sqliteStore) TransactionUpdateState(ctx context.Context, args payd.TransactionArgs, req payd.TransactionStateUpdate) error {
	tx, err := s.newTx(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to start transaction when updating transaction state")
	}
	defer func() {
		_ = rollback(ctx, tx)
	}()
	result, err := tx.Exec(sqlTransactionUpdateState, req.State, args.TxID)
	if err != nil {
		return errors.Wrapf(err, "failed to update transactionId '%s' state to '%s'", args.TxID, req.State)
	}
	if err := handleExecRows(result); err != nil {
		return errors.Wrapf(err, "failed to update transactionId '%s' state to '%s'", args.TxID, req.State)
	}
	return errors.Wrapf(commit(ctx, tx),
		"failed to commit transaction when updating transactionId '%s' state to '%s'", args.TxID, req.State)
}

// Tx returns a tx from the internal store.
func (s *sqliteStore) Tx(ctx context.Context, txID string) (*bt.Tx, error) {
	var txhex struct {
		TxHex string `db:"tx_hex"`
	}
	if err := s.db.GetContext(ctx, &txhex, sqlTransactionGet, txID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, lathos.NewErrNotFound("T001", fmt.Sprintf("tx '%s' not in store", txID))
		}
		return nil, errors.Wrapf(err, "failed to retrieve transaction for id %s", txID)
	}

	return bt.NewTxFromString(txhex.TxHex)
}
