package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/libsv/payd"
	"github.com/pkg/errors"
)

const (
	sqlUTXOGet = `
	SELECT t.outpoint, t.tx_id, t.vout, d.locking_script, d.satoshis, d.derivation_path
	FROM txos t 
	    INNER JOIN destinations d ON t.destination_id = d.destination_id 
		INNER JOIN transactions tx on t.tx_id = tx.tx_id
	WHERE reserved_for IS NULL 
	  AND spent_at IS NULL 
	  AND spending_txid IS NULL
	  AND tx.state = 'broadcast'
	LIMIT 0,1
	`

	sqlUTXOReserve = `
	UPDATE txos
	SET reserved_for = $1, updated_at = $2
	WHERE outpoint = $3
	`

	sqlUTXOUnreserve = `
	UPDATE txos
	SET reserved_for = NULL, updated_at = $1
	WHERE reserved_for = $2 AND spent_at IS NULL and spending_txid IS NULL
	`

	sqlUTXOSpend = `
	UPDATE txos
	SET spent_at = :timestamp, spending_txid = :spending_txid, updated_at = :timestamp
	WHERE reserved_for = :reserved_for
	`
)

// UTXOReserve queries the db for utxos and marks them as reserved, returning any retrieved utxo.
func (s *sqliteStore) UTXOReserve(ctx context.Context, req payd.UTXOReserve) ([]payd.UTXO, error) {
	tx, err := s.newTx(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error creation transaction to get utxos")
	}
	defer func() {
		_ = rollback(ctx, tx)
	}()
	timestamp := time.Now().UTC()
	var utxos []payd.UTXO
	for total := uint64(0); total <= req.Satoshis; {
		var utxo payd.UTXO
		if err := tx.GetContext(ctx, &utxo, sqlUTXOGet); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return []payd.UTXO{}, nil
			}
			return nil, errors.Wrap(err, "failed to get utxo")
		}
		result, err := tx.ExecContext(ctx, sqlUTXOReserve, req.ReservedFor, timestamp, utxo.Outpoint)
		if err != nil {
			return nil, errors.Wrap(err, "failed to reserve utxo")
		}
		if err := handleExecRows(result); err != nil {
			return nil, errors.Wrap(err, "failed to handle update for reserving utxo")
		}

		utxos = append(utxos, utxo)
		total += utxo.Satoshis
	}

	return utxos, errors.Wrap(commit(ctx, tx), "error committing utxo reservation")
}

// UTXOUnreserve unmarks the reservation from matching reservations that haven't been spent.
func (s *sqliteStore) UTXOUnreserve(ctx context.Context, req payd.UTXOUnreserve) error {
	tx, err := s.newTx(ctx)
	if err != nil {
		return errors.Wrap(err, "error creation transaction to get utxos")
	}
	defer func() {
		_ = rollback(ctx, tx)
	}()
	if _, err = tx.ExecContext(ctx, sqlUTXOUnreserve, time.Now().UTC(), req.ReservedFor); err != nil {
		return errors.Wrap(err, "failed to unreserve utxos")
	}

	return errors.Wrap(commit(ctx, tx), "failed to commit transaction to unreserve utxos")
}

// UTXOSpend spends txs matching the provided reservation.
func (s *sqliteStore) UTXOSpend(ctx context.Context, req payd.UTXOSpend) error {
	tx, err := s.newTx(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create transaction to spend utxos")
	}
	defer func() {
		_ = rollback(ctx, tx)
	}()
	req.Timestamp = time.Now().UTC()
	if err := handleNamedExec(tx, sqlUTXOSpend, req); err != nil {
		return errors.Wrap(err, "failed to mark utxos as spent")
	}

	return errors.Wrap(commit(ctx, tx), "failed to commit transaction for spending utxo")
}
