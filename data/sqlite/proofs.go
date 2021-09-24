package sqlite

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/libsv/payd"
)

const (
	sqlProofInsert = `
	INSERT INTO proofs(blockhash, tx_id, data)
	VALUES(:blockhash, :tx_id, :data)
	`
)

// ProofsCreate will insert a proof to the database.
func (s *sqliteStore) ProofCreate(ctx context.Context, req payd.ProofWrapper) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.WithStack(err)
	}
	bb, err := json.Marshal(req.CallbackPayload)
	if err != nil {
		return errors.WithStack(err)
	}
	dbProof := struct {
		Blockhash string `db:"blockhash"`
		TxID      string `db:"tx_id"`
		Data      string `db:"data"`
	}{
		Blockhash: req.BlockHash,
		TxID:      req.CallbackTxID,
		Data:      string(bb),
	}
	res, err := tx.NamedExecContext(ctx, sqlProofInsert, dbProof)
	if err != nil {
		return errors.Wrapf(err, "failed to proof for txid %s and blockhash '%s'", req.CallbackTxID, req.BlockHash)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected when creating proof")
	}
	if rows <= 0 {
		return errors.Wrap(err, "no rows affected when creating proof")
	}
	return errors.WithStack(tx.Commit())
}

func (s *sqliteStore) Proof(ctx context.Context, req payd.ProofRequest) (*payd.ProofWrapper, error) {
	return nil, nil
}
