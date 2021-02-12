package sqlite

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/libsv/go-payd/ppctl"
)

const (
	createInvoice = `
	INSERT INTO invoices(paymentID, satoshis)
	VALUES(:paymentId, :satoshis)
	`

	invoiceByPayID = `
	SELECT paymentID, satoshis, paymentReceivedAt
	FROM invoices
	WHERE paymentID = :paymentId
	`

	updateInvoice = `
	UPDATE invoices
	SET paymentReceivedAt = :paymentReceivedAt
	WHERE paymentID = :paymentId
	`
)

type invoice struct {
	db *sqlx.DB
}

func NewInvoice(db *sqlx.DB) *invoice {
	return &invoice{db: db}
}

// Invoice will return an invoice that matches the provided args.
func (i *invoice) Invoice(ctx context.Context, args ppctl.InvoiceArgs) (*ppctl.Invoice, error) {
	var resp *ppctl.Invoice
	if err := i.db.GetContext(ctx, &resp, invoiceByPayID, args); err != nil {
		return nil, errors.Wrapf(err, "failed to get new invoice with paymentID %s after creation", args.PaymentID)
	}
	return resp, nil
}

// Create will persist a new Invoice in the data store.
func (i *invoice) Create(ctx context.Context, req ppctl.CreateInvoice) (*ppctl.Invoice, error) {
	tx, err := i.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create new invoice with paymentID %s", req.PaymentID)
	}
	defer tx.Rollback()
	if err := handleNamedExec(tx, createInvoice, req); err != nil {
		return nil, errors.Wrap(err, "failed to insert invoice for ")
	}
	var resp *ppctl.Invoice
	if err := tx.Get(&resp, invoiceByPayID, req); err != nil {
		return nil, errors.Wrapf(err, "failed to get new invoice with paymentID %s after creation", req.PaymentID)
	}
	if err := tx.Commit(); err != nil {
		return nil, errors.Wrapf(err, "failed to commit transaction when creating invoice with paymentID %s", req.PaymentID)
	}
	return resp, nil
}

// Update will update an invoice and return the result.
func (i *invoice) Update(ctx context.Context, args ppctl.UpdateInvoiceArgs, req ppctl.UpdateInvoice) (*ppctl.Invoice, error) {
	tx, err := i.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update invoice with paymentID %s", args.PaymentID)
	}
	defer tx.Rollback()
	if err := handleNamedExec(tx, updateInvoice, req); err != nil {
		return nil, errors.Wrapf(err, "failed to update invoice for paymentID %s", args.PaymentID)
	}
	var resp *ppctl.Invoice
	if err := tx.Get(&resp, invoiceByPayID, req); err != nil {
		return nil, errors.Wrapf(err, "failed to get invoice with paymentID %s after update", args.PaymentID)
	}
	if err := tx.Commit(); err != nil {
		return nil, errors.Wrapf(err, "failed to commit transaction when updating invoice with paymentID %s", args.PaymentID)
	}
	return resp, nil
}
