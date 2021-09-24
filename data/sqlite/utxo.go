package sqlite

import (
	"context"

	"github.com/libsv/payd"
	"github.com/pkg/errors"
)

const (
	sqlUTXOGet = `
	SELECT t.outpoint, t.tx_id, t.vout, d.locking_script, d.satoshis, d.derivation_path
	FROM txos t JOIN destinations d ON t.destination_id = d.destination_id
	WHERE reserved_for IS NULL and spent_at IS NULL and spending_txid IS NULL
	LIMIT 0,1
	`

	sqlUTXOReserve = `
	UPDATE txos
	SET reserved_for = $1, updated_at = DATETIME('now')
	WHERE outpoint = $2
	`

	sqlUTXOSpend = `
	UPDATE txos
	SET spent_at = DATETIME('now'), spending_txid = :spending_txid, updated_at = DATETIME('now')
	WHERE reserved_for = :reserved_for
	`
)

func (s *sqliteStore) UTXOReserve(ctx context.Context, req payd.UTXOReserve) ([]payd.UTXO, error) {
	tx, err := s.newTx(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error creation transaction to get utxos")
	}

	var utxos []payd.UTXO
	for total := uint64(0); total <= req.Satoshis; {
		var utxo payd.UTXO
		if err := tx.GetContext(ctx, &utxo, sqlUTXOGet); err != nil {
			return nil, errors.Wrap(err, "failed to get utxo")
		}
		result, err := tx.ExecContext(ctx, sqlUTXOReserve, req.ReservedFor, utxo.Outpoint)
		if err != nil {
			return nil, errors.Wrap(err, "failed to reserve utxo")
		}
		if err := handleExecRows(result); err != nil {
			return nil, errors.Wrap(err, "failed to handle update for reserving utxo")
		}

		utxos = append(utxos, utxo)
		total += utxo.Satoshis
	}

	return utxos, errors.Wrap(commit(ctx, tx), "error commiting utxo reservation")
}

func (s *sqliteStore) UTXOSpend(ctx context.Context, req payd.UTXOSpend) error {
	tx, err := s.newTx(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create transaction to spend utxos")
	}

	if err := handleNamedExec(tx, sqlUTXOSpend, req); err != nil {
		return errors.Wrap(err, "failed to mark utxos as spent")
	}

	return errors.Wrap(commit(ctx, tx), "failed to commit transaction for spending utxo")
}
