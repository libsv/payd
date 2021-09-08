package sqlite

import (
	"context"
	"database/sql"

	gopayd "github.com/libsv/payd"
	"github.com/libsv/payd/client"
	"github.com/pkg/errors"
	"github.com/theflyingcodr/lathos/errs"
)

const (
	sqlFundGet = `
	SELECT txid, vout, lockingscript, satoshis, keyname
	FROM txos
	WHERE spentat IS NULL
	AND keyname = $1
	ORDER BY createdAt ASC
	`

	sqlFundTxCreate = `
	INSERT OR IGNORE INTO transactions(txid, paymentID, txhex)
	VALUES(:txid, NULL, :txhex)
	`

	sqlFundTxoInsert = `
	INSERT INTO txos (outpoint, keyname, txid, vout, lockingscript, spendingtxid, satoshis)
	VALUES (NULL, :keyname, :txid, :vout, :lockingscript, NULL, :satoshis)
	`

	sqlFundSpend = `
	UPDATE txos SET spentat = DATETIME('now'), spendingtxid = :spendingtxid, modifiedat = DATETIME('now')
	WHERE lockingscript = :lockingscript AND keyname = :keyname AND txid = :txid AND vout = :vout
	`
)

func (s *sqliteStore) Funds(ctx context.Context, args client.FundArgs) ([]*client.Fund, error) {
	var funds []*client.Fund
	if err := s.db.SelectContext(ctx, &funds, sqlFundGet, args.KeyName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.NewErrNotFound("N104", "unable to find unspent funds")
		}
		return nil, errors.Wrap(err, "failed to read funds")
	}
	return funds, nil
}

func (s *sqliteStore) FundSpend(ctx context.Context, args client.Fund) error {
	return s.FundsSpend(ctx, []*client.Fund{&args})
}

func (s *sqliteStore) FundsSpend(ctx context.Context, args []*client.Fund) error {
	tx, err := s.newTx(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create context when spending fund")
	}
	defer func() {
		_ = rollback(ctx, tx)
	}()

	for _, arg := range args {
		if err := handleNamedExec(tx, sqlFundSpend, arg); err != nil {
			return errors.Wrap(err, "failed to update funds")
		}
	}
	return errors.Wrap(commit(ctx, tx), "failed to commit transaction when updating txos")
}

func (s *sqliteStore) FundsCreate(ctx context.Context, args client.FundsCreate) (*gopayd.Transaction, error) {
	tx, err := s.newTx(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to start transaction when inserting transaction to db")
	}
	defer func() {
		_ = rollback(ctx, tx)
	}()
	if err := handleNamedExec(tx, sqlFundTxCreate, args); err != nil {
		return nil, errors.Wrap(err, "failed to insert new transaction")
	}
	if err := handleNamedExec(tx, sqlFundTxoInsert, args.Funds); err != nil {
		return nil, errors.Wrap(err, "failed to insert new transaction")
	}

	var outTx gopayd.Transaction
	if err := tx.Get(&outTx, sqlTransactionByID, args.TxID); err != nil {
		return nil, errors.Wrapf(err, "failed to get stored transaction for fund %s", args.TxID)
	}
	var outTxos []gopayd.Txo
	if err := tx.Select(&outTxos, sqlTxosByTxID, args.TxID); err != nil {
		return nil, errors.Wrapf(err, "failed to get stored transaction outputs for fund %s", args.TxID)
	}
	outTx.Outputs = outTxos
	return &outTx, errors.Wrapf(commit(ctx, tx),
		"failed to commit transaction when adding funds for %s", args.TxID)

}
