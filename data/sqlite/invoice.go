package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/libsv/payd"
	"github.com/pkg/errors"
	lathos "github.com/theflyingcodr/lathos/errs"
)

const (
	sqlCreateInvoice = `
	INSERT INTO invoices(invoice_id, satoshis, description, payment_reference, expires_at, state)
	VALUES(:invoice_id, :satoshis, :description, :payment_reference, :expires_at, 'pending')
	`

	sqlInvoiceByID = `
	SELECT invoice_id, satoshis, description, payment_reference, payment_received_at, expires_at, state, refund_to, refunded_at, created_at, updated_at, deleted_at
	FROM invoices
	WHERE invoice_id = :invoice_id
	`

	sqlInvoices = `
	SELECT invoice_id, satoshis, description, payment_reference, payment_received_at, expires_at, state, refund_to, refunded_at, created_at, updated_at, deleted_at
	FROM invoices
	WHERE state != 'deleted'
	`

	// TODO - sort updates when working on rest of Invoice API.
	sqlInvoiceUpdate = `
		UPDATE invoices 
		SET paymentReceivedAt = :paymentReceivedAt, refundTo = :refundTo
		WHERE invoice_id = :invoice_id AND state = :state
	`

	sqlInvoiceDelete = `
	UPDATE invoices
	SET deleted_at = :deleted_at, state = 'deleted'
	WHERE invoice_id = :invoice_id
	`
)

// Invoice will return an invoice that matches the provided args.
func (s *sqliteStore) Invoice(ctx context.Context, args payd.InvoiceArgs) (*payd.Invoice, error) {
	var resp payd.Invoice
	if err := s.db.GetContext(ctx, &resp, sqlInvoiceByID, args.InvoiceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, lathos.NewErrNotFound("N0001", fmt.Sprintf("invoice with invoiceID %s not found", args.InvoiceID))
		}
		return nil, errors.Wrapf(err, "failed to get invoice with invoiceID %s", args.InvoiceID)
	}
	return &resp, nil
}

// Invoice will return an invoice that matches the provided args.
func (s *sqliteStore) Invoices(ctx context.Context) ([]payd.Invoice, error) {
	var resp []payd.Invoice
	if err := s.db.SelectContext(ctx, &resp, sqlInvoices); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, lathos.NewErrNotFound("N0002", "no invoices found")
		}
		return nil, errors.Wrapf(err, "failed to get invoices")
	}
	return resp, nil
}

// Create will persist a new Invoice in the data store.
func (s *sqliteStore) InvoiceCreate(ctx context.Context, req payd.InvoiceCreate) (*payd.Invoice, error) {
	tx, err := s.newTx(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create new invoice with invoiceID %s", req.InvoiceID)
	}
	defer func() {
		_ = rollback(ctx, tx)
	}()
	if err := handleNamedExec(tx, sqlCreateInvoice, req); err != nil {
		return nil, errors.Wrap(err, "failed to insert invoice for ")
	}
	var resp payd.Invoice
	if err := tx.Get(&resp, sqlInvoiceByID, req.InvoiceID); err != nil {
		return nil, errors.Wrapf(err, "failed to get new invoice with invoiceID %s after creation", req.InvoiceID)
	}
	if err := commit(ctx, tx); err != nil {
		return nil, errors.Wrapf(err, "failed to commit transaction when creating invoice with invoiceID %s", req.InvoiceID)
	}
	return &resp, nil
}

// Update will update an invoice to mark it paid and return the result.
func (s *sqliteStore) InvoiceUpdate(ctx context.Context, args payd.InvoiceUpdateArgs, req payd.InvoiceUpdatePaid) (*payd.Invoice, error) {
	tx, err := s.newTx(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update invoice with invoiceID %s", args.InvoiceID)
	}
	defer func() {
		_ = rollback(ctx, tx)
	}()
	req.PaymentReceivedAt = time.Now().UTC()
	resp, err := s.txUpdateInvoicePaid(tx, args, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update invoice")
	}
	if err := commit(ctx, tx); err != nil {
		return nil, errors.Wrapf(err, "failed to commit transaction when updating invoice with invoiceID %s", args.InvoiceID)
	}
	return resp, nil
}

func (s *sqliteStore) InvoiceDelete(ctx context.Context, args payd.InvoiceArgs) error {
	tx, err := s.newTx(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to delete invoice with paymentID %s", args.InvoiceID)
	}
	defer func() {
		_ = rollback(ctx, tx)
	}()
	if _, err := s.Invoice(ctx, args); err != nil {
		return errors.WithMessagef(err, "failed to find key with id %s to delete", args.InvoiceID)
	}
	// invoice delete DTO
	delInv := struct {
		DeletedAt time.Time `db:"deleted_at"`
		InvoiceID string    `db:"invoice_id"`
	}{
		DeletedAt: time.Now().UTC(),
		InvoiceID: args.InvoiceID,
	}
	if err := handleNamedExec(tx, sqlInvoiceDelete, delInv); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return lathos.NewErrNotFound("N0003", fmt.Sprintf("invoice with ID %s not found", args.InvoiceID))
		}
		return errors.Wrapf(err, "failed to delete invoice for invoiceID %s", args.InvoiceID)
	}
	if err := commit(ctx, tx); err != nil {
		return errors.Wrapf(err, "failed to commit transaction when deleting invoice with invoiceID %s", args.InvoiceID)
	}
	return nil
}

// txUpdateInvoicePaid takes a db object / transaction and adds a transaction to the data store
// along with utxos, returning the updated invoice.
// This method can be used with other methods in the store allowing
// multiple methods to be ran in the same db transaction.
func (s *sqliteStore) txUpdateInvoicePaid(tx db, args payd.InvoiceUpdateArgs, req payd.InvoiceUpdatePaid) (*payd.Invoice, error) {
	req.PaymentReceivedAt = time.Now().UTC()
	if err := handleNamedExec(tx, sqlInvoiceUpdate, map[string]interface{}{
		"paymentReceivedAt": req.PaymentReceivedAt,
		//	"refundTo":          req.RefundTo,
		"paymentID": args.InvoiceID,
	}); err != nil {
		return nil, errors.Wrapf(err, "failed to update invoice for invoiceID %s", args.InvoiceID)
	}
	var resp payd.Invoice
	if err := tx.Get(&resp, sqlInvoiceByID, args.InvoiceID); err != nil {
		return nil, errors.Wrapf(err, "failed to get invoice with invoiceID %s after update", args.InvoiceID)
	}
	return &resp, nil
}
