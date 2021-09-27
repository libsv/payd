package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/libsv/go-bc"
	"github.com/libsv/payd"
)

const (
	sqlProofInsert = `
	INSERT INTO proofs(blockhash, tx_id, data)
	VALUES(:blockhash, :tx_id, :data)
	`

	sqlProofGet = `
	SELECT data
	FROM proofs
	WHERE tx_id = $1
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

// Proof will retrieve a proof.
func (s *sqliteStore) MerkleProof(ctx context.Context, txID string) (*bc.MerkleProof, error) {
	var data string
	if err := s.db.GetContext(ctx, &data, sqlProofGet, txID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to retrieve proof")
	}
	var proof bc.MerkleProof
	if err := json.Unmarshal([]byte(data), &proof); err != nil {
		return nil, errors.Wrap(err, "failed to process raw proof")
	}
	return &proof, nil
}
