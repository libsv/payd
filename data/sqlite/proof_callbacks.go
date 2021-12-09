package sqlite

import (
	"context"

	"github.com/pkg/errors"

	"github.com/libsv/go-p4"
	"github.com/libsv/payd"
)

const (
	sqlCallbackURLInsert = `
	INSERT INTO proof_callbacks(invoice_id, url, token, state)
	VALUES(:invoice_id,:url,:token,'pending')
	`
)

type proofCallbackDTO struct {
	InvoiceID string `db:"invoice_id"`
	URL       string `db:"url"`
	Token     string `db:"token"`
}

// ProofCallBacksCreate can be implemented to store merkle proof callback urls for an invoice.
func (s *sqliteStore) ProofCallBacksCreate(ctx context.Context, args payd.ProofCallbackArgs, req map[string]p4.ProofCallback) error {
	if len(req) == 0 {
		// nothing to store
		return nil
	}
	tx, err := s.newTx(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to create new callback urls for invoiceID %s", args.InvoiceID)
	}
	defer func() {
		_ = rollback(ctx, tx)
	}()

	cc := make([]proofCallbackDTO, 0, len(req))
	for url, val := range req {
		cc = append(cc, proofCallbackDTO{
			InvoiceID: args.InvoiceID,
			URL:       url,
			Token:     val.Token,
		})
	}
	if err := handleNamedExec(tx, sqlCallbackURLInsert, cc); err != nil {
		return errors.Wrapf(err, "failed to insert callback urls for invoiceID %s", args.InvoiceID)
	}
	if err := commit(ctx, tx); err != nil {
		return errors.Wrapf(err, "failed to commit transaction when creating callback urls with invoiceID %s", args.InvoiceID)
	}
	return nil
}
