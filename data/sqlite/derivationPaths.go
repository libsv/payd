package sqlite

import (
	"context"
	"errors"
	"fmt"

	gopayd "github.com/libsv/go-payd"
)

const (
	sqlReserveDerivPath = `
	INSERT INTO derivation_paths(paymentID, path, prefix, pathIndex)
	VALUES(:paymentID,
       :prefix || '/' || (SELECT (COALESCE(MAX(pathIndex)+1,0)) as idx FROM derivation_paths WHERE prefix = :prefix),
       :prefix,
       (SELECT (COALESCE(MAX(pathIndex)+1,0)) as idx FROM derivation_paths WHERE prefix = :prefix));
	`

	sqlDerivationPathByID = `
	SELECT ID, paymentID, path, prefix, pathIndex, createdAt
	FROM derivation_paths
	WHERE ID = :id
	`

	sqlDerivationPathExists = `
	SELECT EXISTS(SELECT id from derivation_paths where paymentID = :paymentID)
	`
)

// ReserveDerivationPath will create a derivation path for an invoice and
// return with the index incremented ready for use.
func (s *sqliteStore) DerivationPathCreate(ctx context.Context, req gopayd.DerivationPathCreate) (*gopayd.DerivationPath, error) {
	tx, err := s.newTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start tx for creating derivPath %w", err)
	}
	res, err := tx.NamedExec(sqlReserveDerivPath, req)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create derivation path %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to get rows affected when creating derivation path %w", err)
	}
	if rows <= 0 {
		tx.Rollback()
		return nil, errors.New("no rows affected when creating derivation path")
	}
	id, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to get lastInsertedID when creating derivation path %w", err)
	}
	var dp gopayd.DerivationPath
	if err := tx.Get(&dp, sqlDerivationPathByID, id); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to get derivation path with id %d when creating derivation path %w", id, err)
	}
	if err := commit(ctx, tx); err != nil {
		return nil, fmt.Errorf("failed to commit tx when creating derivation path %w", err)
	}
	return &dp, nil
}

// DerivationPath will return a derivationPath that matches the supplied args.
func (s *sqliteStore) DerivationPath(ctx context.Context, args gopayd.DerivationPathArgs) (*gopayd.DerivationPath, error) {
	var dp *gopayd.DerivationPath
	if err := s.db.GetContext(ctx, &dp, sqlDerivationPathByID, args.ID); err != nil {
		return nil, fmt.Errorf("failed to get derivationPath with id %d %w", args.ID, err)
	}
	return dp, nil
}

// DerivationPathExists will return a DerivationPathExists that matches the supplied args.
func (s *sqliteStore) DerivationPathExists(ctx context.Context, args gopayd.DerivationPathExistsArgs) (bool, error) {
	var found int
	if err := s.db.GetContext(ctx, &found, sqlDerivationPathExists, args.PaymentID); err != nil {
		return false, fmt.Errorf("failed to check derivationPath exists for paymentID %s %w", args.PaymentID, err)
	}
	return found > 0, nil
}
