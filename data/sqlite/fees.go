package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/libsv/go-bt/v2"
	"github.com/libsv/payd"
	"github.com/pkg/errors"
	lathos "github.com/theflyingcodr/lathos/errs"
)

const sqlInsertFees = `
	INSERT INTO fee_rates(invoice_id, fee_json, expires_at)
	VALUES (:invoice_id, :fee_json, :expires_at)
	ON CONFLICT(invoice_id) DO UPDATE SET fee_json=:fee_json, expires_at=:expires_at
`

const sqlSelectFees = `
	SELECT fee_json, expires_at FROM fee_rates
	WHERE invoice_id = :invoice_id
`

func (s *sqliteStore) FeesQuoteCreate(ctx context.Context, args *payd.FeeQuoteCreateArgs) error {
	tx, err := s.newTx(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to start transaction when inserting fees to db")
	}
	defer func() {
		_ = rollback(ctx, tx)
	}()
	bb, err := json.Marshal(args.FeeQuote)
	if err != nil {
		return errors.Wrap(err, "failed to convert fee quote into json")
	}
	if err := handleNamedExec(tx, sqlInsertFees, struct {
		InvoiceID string    `db:"invoice_id"`
		FeeJSON   string    `db:"fee_json"`
		ExpiresAt time.Time `db:"expires_at"`
	}{
		InvoiceID: args.InvoiceID,
		FeeJSON:   string(bb),
		ExpiresAt: args.FeeQuote.Expiry(),
	}); err != nil {
		return errors.Wrap(err, "failed to insert fees into db")
	}
	return errors.Wrapf(commit(ctx, tx), "failed to commit transaction when inserting fee rate for invoice '%s'", args.InvoiceID)
}

func (s *sqliteStore) Fees(ctx context.Context, invoiceID string) (*bt.FeeQuote, error) {
	var row struct {
		FeeJSON   string `db:"fee_json"`
		ExpiresAt string `db:"expires_at"`
	}
	if err := s.db.GetContext(ctx, &row, sqlSelectFees, invoiceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, lathos.NewErrNotFoundf("N0002", "cannot find fee quote for invoiceID %s", invoiceID)
		}
		return nil, errors.Wrapf(err, "failed to get fee quote for invoiceID %s", invoiceID)
	}
	var fq bt.FeeQuote
	if err := json.Unmarshal([]byte(row.FeeJSON), &fq); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal fee quote")
	}
	expiredAt, err := time.Parse(time.RFC3339, row.ExpiresAt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse expiry timestamp")
	}
	fq.UpdateExpiry(expiredAt)
	return &fq, nil
}
