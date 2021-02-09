package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/libsv/go-bt"
	"github.com/pkg/errors"

	"github.com/libsv/go-payd/bip270"
	"github.com/libsv/go-payd/bip270/data/sqlite/queries"
)

type transaction struct {
	db *sqlx.DB
}

func NewTransaction(db *sqlx.DB) *transaction {
	return &transaction{db: db}
}

// Create can be implemented to store a Transaction in a datastore.
func (t *transaction) Create(ctx context.Context, args bip270.CreateTxArgs, req *bt.Tx) (*bip270.Tx, error) {
	tx, err := t.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to start transaction when inserting transaction to db")
	}
	defer tx.Rollback()
	created := time.Now().UTC()
	if err := handleExec(tx, queries.InsertTransaction, map[string]interface{}{
		"txid":      req.GetTxID(),
		"txhex":     req.ToString(),
		"createdAt": created,
	}); err != nil {
		return nil, errors.Wrap(err, "failed to insert new transaction")
	}
	// TODO - where does this data come from?
	for i, txo := range req.Outputs {
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
			"spendingTxID":   req.GetTxID(),
			"createdAt":      created,
			"modifiedAt":     created,
		}
		if err := handleExec(tx, queries.InsertTxo, sqlArgs); err != nil {
			return nil, errors.Wrap(err, "failed to insert new transaction")
		}
	}
	return nil, nil
}
