package sqlite

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/libsv/go-payd/bip270"
)

type transactionOutput struct {
	db *sqlx.DB
}

func NewTransactionOutput(db *sqlx.DB) *transactionOutput {
	return &transaction{db: db}
}

func (t *transactionOutput) CreateOutputs(ctx context.Context, req []*bip270.Output) error {
	// store the outputs in a temp table
	// can look these up and mark as paid / delete when the customer pays

}
