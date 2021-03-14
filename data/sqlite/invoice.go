package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	gopayd "github.com/libsv/payd"
	"github.com/pkg/errors"
	"github.com/theflyingcodr/lathos"
)

const (
	sqlCreateInvoice = `
	INSERT INTO invoices(paymentID, satoshis)
	VALUES(:paymentId, :satoshis)
	`

	sqlInvoiceByPayID = `
	SELECT paymentId,satoshis,paymentReceivedAt
	FROM invoices
	WHERE paymentId = :paymentId
	`

	sqlInvoiceUpdate = `
		UPDATE invoices 
		SET paymentReceivedAt = :paymentReceivedAt, refundTo = :refundTo
		WHERE paymentID = :paymentID
	`
)

// Invoice will return an invoice that matches the provided args.
func (s *sqliteStore) Invoice(ctx context.Context, args gopayd.InvoiceArgs) (*gopayd.Invoice, error) {
	var resp gopayd.Invoice
	if err := s.db.GetContext(ctx, &resp, sqlInvoiceByPayID, args.PaymentID); err != nil {
		if err == sql.ErrNoRows {
			return nil, lathos.NewErrNotFound("N0001", fmt.Sprintf("invoice with paymentID %s not found", args.PaymentID))
		}
		return nil, errors.Wrapf(err, "failed to get new invoice with paymentID %s after creation", args.PaymentID)
	}
	return &resp, nil
}

// Create will persist a new Invoice in the data store.
func (s *sqliteStore) Create(ctx context.Context, req gopayd.CreateInvoice) (*gopayd.Invoice, error) {
	tx, err := s.newTx(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create new invoice with paymentID %s", req.PaymentID)
	}
	defer tx.Rollback()
	if err := handleNamedExec(tx, sqlCreateInvoice, req); err != nil {
		return nil, errors.Wrap(err, "failed to insert invoice for ")
	}
	var resp *gopayd.Invoice
	if err := tx.Get(&resp, sqlInvoiceByPayID, req); err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "failed to get new invoice with paymentID %s after creation", req.PaymentID)
	}
	if err := commit(ctx, tx); err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "failed to commit transaction when creating invoice with paymentID %s", req.PaymentID)
	}
	return resp, nil
}

// Update will update an invoice to mark it paid and return the result.
func (s *sqliteStore) Update(ctx context.Context, args gopayd.UpdateInvoiceArgs, req gopayd.UpdateInvoice) (*gopayd.Invoice, error) {
	tx, err := s.newTx(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update invoice with paymentID %s", args.PaymentID)
	}
	resp, err := s.txUpdateInvoicePaid(tx, args, req)
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrap(err, "failed to update invoice")
	}
	if err := commit(ctx, tx); err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(err, "failed to commit transaction when updating invoice with paymentID %s", args.PaymentID)
	}
	return resp, nil
}

// txUpdateInvoicePaid takes a db object / transaction and adds a transaction to the data store
// along with utxos, returning the updated invoice.
// This method can be used with other methods in the store allowing
// multiple methods to be ran in the same db transaction.
func (s *sqliteStore) txUpdateInvoicePaid(tx db, args gopayd.UpdateInvoiceArgs, req gopayd.UpdateInvoice) (*gopayd.Invoice, error) {
	req.PaymentReceivedAt = time.Now().UTC()
	if err := handleNamedExec(tx, sqlInvoiceUpdate, req); err != nil {
		return nil, errors.Wrapf(err, "failed to update invoice for paymentID %s", args.PaymentID)
	}
	var resp *gopayd.Invoice
	if err := tx.Get(&resp, sqlInvoiceByPayID, req); err != nil {
		return nil, errors.Wrapf(err, "failed to get invoice with paymentID %s after update", args.PaymentID)
	}
	return resp, nil
}
