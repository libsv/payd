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
	sqlTxoCreate = `
	INSERT INTO txos (keyname, derivationpath, lockingscript, satoshis, vout, txid)
	VALUES(:keyname, :derivationpath, :lockingscript, :satoshis, :vout, :txid)
	`

	sqlPartialTxoCreate = `
	INSERT INTO txos (keyname, derivationpath, lockingscript, satoshis, paymentID)
	VALUES(:keyname, :derivationpath, :lockingscript, :satoshis, :paymentid)
	`

	sqlPartialTxo = `
	SELECT keyname, derivationpath, lockingscript, satoshis, createdat, modifiedat
	FROM txos
	WHERE lockingscript = $1 AND satoshis = $2 AND keyname = $3 AND outpoint IS NULL
	`

	sqlPartialTxoByPaymentID = `
	SELECT keyname, derivationpath, lockingscript, satoshis, createdat, modifiedat
	FROM txos
	WHERE paymentID = :paymentID
	`

	sqlTxoUpdate = `
	UPDATE txos SET outpoint = :outpoint, vout = :vout, txid = :txid, modifiedat = DATETIME('now')
	WHERE outpoint IS NULL AND lockingscript = :lockingscript AND keyname = :keyname AND satoshis = :satoshis
	`

	sqlUnreservedTxos = `
	SELECT txid, vout, lockingscript, satoshis, keyname
	FROM txos
	WHERE spentat IS NULL
	AND reservedFor IS NULL
	AND keyname = $1
	AND txid IS NOT NULL
	AND ackReceivedAt IS NOT NULL
	ORDER BY createdAt ASC
	LIMIT $2,$3
	`

	sqlTxoReserve = `
	UPDATE txos SET reservedFor=:reservedfor
	WHERE keyname = :keyname AND txid = :txid AND vout=:vout AND spentat IS NULL AND reservedFor IS NULL
	`

	sqlTxoSpend = `
	UPDATE txos
	SET spentat = DATETIME('now'),
		spendingTxId = :spendingTxId,
		modifiedAt = DATETIME('now')
	WHERE txid = :txid AND keyname = :keyname AND vout = :vout
	`
)

// PartialTxoCreate will store a txo created during payment requests.
func (s *sqliteStore) PartialTxoCreate(ctx context.Context, req gopayd.PartialTxoCreate) error {
	return s.PartialTxosCreate(ctx, []*gopayd.PartialTxoCreate{
		&req,
	})
}

// PartialTxosCreate will store txos created during payment requests.
func (s *sqliteStore) PartialTxosCreate(ctx context.Context, req []*gopayd.PartialTxoCreate) error {
	tx, err := s.newTx(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create context when creating a txo")
	}
	defer func() {
		_ = rollback(ctx, tx)
	}()
	if err := handleNamedExec(tx, sqlPartialTxoCreate, req); err != nil {
		return errors.Wrap(err, "failed to insert script keys.")
	}
	return errors.Wrap(commit(ctx, tx), "failed to commit transaction when creating txos.")
}

// PartialTxo will return a txo that has been stored but not yet assigned to a transaction.
func (s *sqliteStore) PartialTxo(ctx context.Context, args gopayd.UnspentTxoArgs) (*gopayd.UnspentTxo, error) {
	var txo gopayd.UnspentTxo
	if err := s.db.GetContext(ctx, &txo, sqlPartialTxo, args.LockingScript, args.Satoshis, args.Keyname); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.NewErrNotFound("N104",
				fmt.Sprintf("unable to find txo with script '%s', value '%d' and keyname '%s'",
					args.LockingScript, args.Satoshis, args.Keyname))
		}
		return nil, errors.Wrap(err, "failed to read partialTxo")
	}
	return &txo, nil
}

// PartialTxo will return a txo that has been stored but not yet assigned to a transaction.
func (s *sqliteStore) PartialTxoByPaymentID(ctx context.Context, args gopayd.InvoiceArgs) ([]gopayd.UnspentTxo, error) {
	var txos []gopayd.UnspentTxo
	if err := s.db.SelectContext(ctx, &txos, sqlPartialTxoByPaymentID, args.PaymentID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.NewErrNotFound("N104",
				fmt.Sprintf("unable to find txos for payment '%s'", args.PaymentID))
		}
		return nil, errors.Wrap(err, "failed to read partialTxo")
	}
	return txos, nil
}

func (s *sqliteStore) TxoCreate(ctx context.Context, args gopayd.TxoCreate) error {
	return s.TxosCreate(ctx, []gopayd.TxoCreate{args})
}

func (s *sqliteStore) TxosCreate(ctx context.Context, args []gopayd.TxoCreate) error {
	tx, err := s.newTx(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to start transaction when inserting txos into db")
	}
	for _, txo := range args {
		if err = handleNamedExec(tx, sqlTxoCreate, txo); err != nil {
			return err
		}
	}

	return errors.Wrap(commit(ctx, tx), "failed to commit tx when adding txos")
}

func (s *sqliteStore) ReserveTxos(ctx context.Context, args gopayd.TxoReserveArgs) ([]gopayd.Txo, error) {
	tx, err := s.newTx(ctx)
	if err != nil {
		return nil, err
	}
	var txos []gopayd.Txo
	if err := tx.SelectContext(ctx, &txos, sqlUnreservedTxos, args.Account, args.Offset, args.Limit); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.NewErrNotFound("N104", "unable to find unspent funds")
		}
		return nil, errors.Wrap(err, "failed to read funds")
	}
	fmt.Println(txos)
	for _, txo := range txos {
		if err = handleNamedExec(tx, sqlTxoReserve, struct {
			TxID        string `db:"txid"`
			Vout        uint32 `db:"vout"`
			KeyName     string `db:"keyname"`
			ReservedFor string `db:"reservedfor"`
		}{
			TxID:        txo.TxID,
			Vout:        txo.Vout,
			KeyName:     args.Account,
			ReservedFor: args.ReservedFor,
		}); err != nil {
			return nil, errors.Wrap(err, "failed to reserve txo for payment")
		}
	}

	return txos, errors.Wrap(commit(ctx, tx), "failed to commit reserving txs")
}

func (s *sqliteStore) UnspentTxos(ctx context.Context, args gopayd.UnspentTxoArgs) ([]gopayd.Txo, error) {
	var txos []gopayd.Txo
	if err := s.db.SelectContext(ctx, &txos, sqlUnreservedTxos, args.Keyname); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.NewErrNotFound("N104", "unable to find unspent funds")
		}
		return nil, errors.Wrap(err, "failed to read funds")
	}
	return txos, nil
}

func (s *sqliteStore) DerivationPath(ctx context.Context, ls string) (string, error) {
	var path string
	if err := s.db.GetContext(ctx, &path, "select derivationpath from txos where lockingscript = $1", ls); err != nil {
		return "", err
	}

	return path, nil
}
