package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/libsv/go-payd/bip270"
)

const (
	InsertTransaction = `
	INSERT INTO transaction(txid, txhex, createdAt)
	VALUES(:txid, :txhex, :createdAt)
	`

	InsertTxo = `
	INSERT INTO txos(outpoint, instance, txid, vout, alias, derivationpath, scriptpubkey, satoshis, reservedat, spentat, spendingtxid, createdat, modifiedat)
	VALUES(:outpoint, :instance, :txid, :vout, :alias, :derivationPath, :scriptPubKey, :satoshis, :reservedAt, :spentAt, :spendingTxID, :createdAt, :modifiedAt)
	`

	TransactionByID = `
	SELECT txid, txhex, createdAt
	FROM transactions
	WHERE txid = :txID
	`

	TxosByTxID = `
	SELECT outpoint, instance, txid, vout, alias, derivationpath, scriptpubkey, satoshis, 
				reservedat, spentat, spendingtxid, createdat, modifiedat 
	FROM txos
	WHERE txid = :txID
	`
)

type transaction struct {
	db *sqlx.DB
}

func NewTransaction(db *sqlx.DB) *transaction {
	return &transaction{db: db}
}

// Create can be implemented to store a Transaction in a datastore.
func (t *transaction) Create(ctx context.Context, args bip270.CreateTxArgs, req *bip270.Output) (*bip270.Tx, error) {
	tx, err := t.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to start transaction when inserting transaction to db")
	}
	defer tx.Rollback()
	created := time.Now().UTC()
	if err := handleExec(tx, InsertTransaction, map[string]interface{}{
		"txid":      req.GetTxID(),
		"txhex":     req.ToString(),
		"createdAt": created,
	}); err != nil {
		return nil, errors.Wrap(err, "failed to insert new transaction")
	}
	// TODO - do we store all outputs, even those sent as change back to the payee
	// or do we just store those meant for us?
	for i, txo := range req.GetOutputs() {
		sqlArgs := map[string]interface{}{
			"outpoint":       fmt.Sprintf("%s%d", req.GetTxID(), i),
			"txid":           req.GetTxID(),
			"instance":       1, // TODO - should be auto updated
			"vout":           i,
			"alias":          nil, // TODO - is this keyname?
			"derivationPath": args.DerivationPath,
			"scriptPubKey":   txo.LockingScript.GetPublicKeyHash(),
			"satoshis":       txo.Satoshis,
			"reservedAt":     created, // TODO - is this correct?
			"spentAt":        nil,     // TODO - do we need this?
			"spendingTxID":   nil,     // TODO - is this correct?
			"createdAt":      created,
			"modifiedAt":     created,
		}
		if err := handleExec(tx, queries.InsertTxo, sqlArgs); err != nil {
			return nil, errors.Wrap(err, "failed to insert new transaction")
		}
	}
	var outTx *bip270.Tx
	tx.Get(&outTx, que)

	return nil, nil
}
