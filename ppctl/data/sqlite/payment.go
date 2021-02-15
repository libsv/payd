package sqlite

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/libsv/go-payd/ppctl"
)

type payment struct {
	db *sqlx.DB
}

func NewPayment(db *sqlx.DB) *payment {
	return &payment{db: db}
}

func (p *payment) ScriptKey(ctx context.Context, args ppctl.ScriptKeyArgs) (*ppctl.ScriptKey, error) {
	return scriptKeys(ctx, p.db, args)
}
