package sqlite

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/libsv/go-payd/ppctl"
)

const (
	insertTransaction = `
	INSERT INTO transaction(txid, paymentID, txhex, createdAt)
	VALUES(:txid, :paymentID, :txhex, :createdAt)
	`

	insertTxo = `
	INSERT INTO txos(outpoint, txid, vout, keyname, derivationpath, lockingscript, satoshis,  createdat, modifiedat)
	VALUES(:outpoint, :txid, :vout, :keyname, :derivationPath, :lockingscript, :satoshis, :createdAt, :modifiedAt)
	`

	transactionByID = `
	SELECT txid, paymentID, txhex, createdAt
	FROM transactions
	WHERE txid = :txId
	`

	txosByTxID = `
	SELECT outpoint, txid, vout, alias, derivationpath, lockingscript, satoshis, 
				spentat, spendingtxid, createdat, modifiedat 
	FROM txos
	WHERE txid = :txId
	`

	updateInvoiceDate = `
		UPDATE invoices 
		SET paymentReceivedAt = :paymentReceivedAt
		WHERE paymentID = :paymentId
	`
)

type transaction struct {
	db *sqlx.DB
}

func NewTransaction(db *sqlx.DB) *transaction {
	return &transaction{db: db}
}

// Create can be implemented to store a Transaction and outputs in a datastore.
func (t *transaction) Create(ctx context.Context, req ppctl.CreateTransaction) (*ppctl.Transaction, error) {
	tx, err := t.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to start transaction when inserting transaction to db")
	}
	defer tx.Rollback()
	if err := handleNamedExec(tx, insertTransaction, req); err != nil {
		return nil, errors.Wrap(err, "failed to insert new transaction")
	}
	if err := handleNamedExec(tx, insertScriptKeys, req.Outputs); err != nil {
		return nil, errors.Wrap(err, "failed to insert transaction outputs")
	}
	// TODO - ideally I'd have this as a separate call but to keep in the same ATOMIC tx
	// the simplest thing is to update the invoice here.
	if err := handleNamedExec(tx, updateInvoice, map[string]interface{}{
		"paymentReceivedAt": time.Now().UTC(),
	}); err != nil {
		return nil, errors.Wrap(err, "failed to update invoice date")
	}
	var outTx *ppctl.Transaction
	if err := tx.Get(&outTx, transactionByID, req); err != nil {
		return nil, errors.Wrapf(err, "failed to get transaction for paymentID %s", req.PaymentID)
	}
	var outTxos []ppctl.Txo
	if err := tx.Get(&outTxos, txosByTxID, req); err != nil {
		return nil, errors.Wrapf(err, "failed to get transaction for paymentID %s", req.PaymentID)
	}
	outTx.Outputs = outTxos
	return outTx, errors.Wrapf(tx.Commit(),
		"failed to commit transaction when adding tx and outpurs for paymentID %s", req.PaymentID)
}
