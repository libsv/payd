package sqlite

import (
	"context"
	"time"

	go_payd "github.com/libsv/go-payd"
	gopayd "github.com/libsv/go-payd"
	"github.com/pkg/errors"
)

// CompletePayment will store the tx and utxos as well as update the invoice as paid in a single transaction.
func (s *sqliteStore) CompletePayment(ctx context.Context, req gopayd.CreateTransaction) (*go_payd.Transaction, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to start transaction when inserting transaction to db")
	}
	defer tx.Rollback()
	resp, err := s.txCreateTransaction(tx, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to complete payment")
	}
	if _, err := s.txUpdateInvoicePaid(tx,
		gopayd.UpdateInvoiceArgs{PaymentID: req.PaymentID},
		gopayd.UpdateInvoice{PaymentReceivedAt: time.Now().UTC()}); err != nil {
		return nil, errors.Wrap(err, "failed to complete payment")
	}
	return resp, errors.Wrapf(tx.Commit(),
		"failed to commit transaction when adding tx and outputs for paymentID %s", req.PaymentID)
}
