package service

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/guregu/null.v3"

	gopayd "github.com/libsv/payd"
	"github.com/libsv/payd/config"

	"github.com/speps/go-hashids"
)

type destinationCreator interface {
	// DestinationsCreate will split satoshis into multiple denominations and store
	// as denominations waiting to be fulfilled in a tx.
	DestinationsCreate(ctx context.Context, req gopayd.DestinationsCreate) (*gopayd.Destination, error)
}

// invoice represents a purchase order system or other such system that a merchant would use
// to receive orders from customers.
// This could be a Pos system or online retailer etc.
// The invoice system would create an invoice / PO and then the protocol
// server would be sent this invoice for lookup.
// This invoicing system is separate to the protocol server itself but added here
// as a very basic example.
type invoice struct {
	store      gopayd.InvoiceReaderWriter
	destSvc    destinationCreator
	cfg        *config.Server
	transacter gopayd.Transacter
}

// NewInvoice will setup and return a new invoice service.
func NewInvoice(cfg *config.Server, store gopayd.InvoiceReaderWriter, destSvc destinationCreator, transacter gopayd.Transacter) *invoice {
	return &invoice{
		cfg:        cfg,
		store:      store,
		destSvc:    destSvc,
		transacter: transacter,
	}
}

// Invoice will return an invoice by paymentID.
func (i *invoice) Invoice(ctx context.Context, args gopayd.InvoiceArgs) (*gopayd.Invoice, error) {
	if err := args.Validate(); err != nil {
		return nil, err
	}
	inv, err := i.store.Invoice(ctx, args)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to get invoice with id %s", args.InvoiceID)
	}
	return inv, err
}

// Invoices will return all currently stored invoices.
func (i *invoice) Invoices(ctx context.Context) ([]gopayd.Invoice, error) {
	ii, err := i.store.Invoices(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get invoices")
	}
	return ii, nil
}

// Create will add a new invoice to the system.
func (i *invoice) Create(ctx context.Context, req gopayd.InvoiceCreate) (*gopayd.Invoice, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	hd := hashids.NewData()
	hd.Alphabet = hashids.DefaultAlphabet
	hd.Salt = fmt.Sprintf("%s:%d:%s:%s", i.cfg.Hostname, req.Satoshis, req.Reference.ValueOrZero(), req.ExpiresAt.ValueOrZero())
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	id, err := h.Encode([]int{time.Now().Nanosecond()})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	req.InvoiceID = id
	ctx = i.transacter.WithTx(ctx)
	defer func() {
		_ = i.transacter.Rollback(ctx)
	}()
	inv, err := i.store.InvoiceCreate(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// TODO - this could be an async call - though this ensures it all completes.
	if _, err := i.destSvc.DestinationsCreate(ctx, gopayd.DestinationsCreate{
		InvoiceID: null.StringFrom(req.InvoiceID),
		Satoshis:  req.Satoshis,
	}); err != nil {
		return nil, errors.Wrapf(err, "failed to create payment destinations for invoice")
	}
	return inv, errors.WithStack(i.transacter.Commit(ctx))
}

// Delete will permanently remove an invoice from the system.
func (i *invoice) Delete(ctx context.Context, args gopayd.InvoiceArgs) error {
	if err := args.Validate(); err != nil {
		return err
	}
	return errors.WithMessagef(i.store.InvoiceDelete(ctx, args),
		"failed to delete invoice with ID %s", args.InvoiceID)
}
