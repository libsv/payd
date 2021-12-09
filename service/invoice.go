package service

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/guregu/null.v3"

	"github.com/libsv/payd"
	"github.com/libsv/payd/config"

	"github.com/speps/go-hashids"
)

type destinationCreator interface {
	// DestinationsCreate will split satoshis into multiple denominations and store
	// as denominations waiting to be fulfilled in a tx.
	DestinationsCreate(ctx context.Context, req payd.DestinationsCreate) (*payd.Destination, error)
}

// invoice represents a purchase order system or other such system that a merchant would use
// to receive orders from customers.
// This could be a Pos system or online retailer etc.
// The invoice system would create an invoice / PO and then the protocol
// server would be sent this invoice for lookup.
// This invoicing system is separate to the protocol server itself but added here
// as a very basic example.
type invoice struct {
	timeSvc    payd.TimestampService
	store      payd.InvoiceReaderWriter
	destSvc    destinationCreator
	connSvc    payd.ConnectService
	cfg        *config.Server
	wallCfg    *config.Wallet
	transacter payd.Transacter
}

// NewInvoice will setup and return a new invoice service.
func NewInvoice(cfg *config.Server, wallCfg *config.Wallet, store payd.InvoiceReaderWriter, destSvc destinationCreator, transacter payd.Transacter, timeSvc payd.TimestampService) *invoice {
	return &invoice{
		cfg:        cfg,
		wallCfg:    wallCfg,
		store:      store,
		destSvc:    destSvc,
		transacter: transacter,
		timeSvc:    timeSvc,
	}
}

// Invoice will return an invoice by paymentID.
func (i *invoice) Invoice(ctx context.Context, args payd.InvoiceArgs) (*payd.Invoice, error) {
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
func (i *invoice) Invoices(ctx context.Context) ([]payd.Invoice, error) {
	ii, err := i.store.Invoices(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get invoices")
	}
	return ii, nil
}

// Create will add a new invoice to the system.
func (i *invoice) Create(ctx context.Context, req payd.InvoiceCreate) (*payd.Invoice, error) {
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
	id, err := h.Encode([]int{i.timeSvc.Nanosecond()})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	req.InvoiceID = id
	ctx = i.transacter.WithTx(ctx)
	defer func() {
		_ = i.transacter.Rollback(ctx)
	}()
	req.SPVRequired = i.wallCfg.SPVRequired // set default requirement
	// any payment 1000sat or below, we don't want spv
	// NOTE - this is just an example
	if req.Satoshis <= 1000 {
		req.SPVRequired = false
	}
	timestamp := i.timeSvc.NowUTC()
	req.CreatedAt = timestamp
	if req.ExpiresAt.IsZero() {
		// set to default expiry hours
		req.ExpiresAt = null.TimeFrom(timestamp.Add(time.Hour * time.Duration(i.wallCfg.PaymentExpiryHours)))
	}
	inv, err := i.store.InvoiceCreate(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if _, err := i.destSvc.DestinationsCreate(ctx, payd.DestinationsCreate{
		InvoiceID: null.StringFrom(req.InvoiceID),
		Satoshis:  req.Satoshis,
	}); err != nil {
		return nil, errors.Wrapf(err, "failed to create payment destinations for invoice")
	}
	if err := i.transacter.Commit(ctx); err != nil {
		return nil, errors.WithStack(err)
	}
	return inv, i.connectPaymentChannel(ctx, req.InvoiceID)
}

// Delete will permanently remove an invoice from the system.
func (i *invoice) Delete(ctx context.Context, args payd.InvoiceArgs) error {
	if err := args.Validate(); err != nil {
		return err
	}
	return errors.WithMessagef(i.store.InvoiceDelete(ctx, args),
		"failed to delete invoice with ID %s", args.InvoiceID)
}

// SetConnectionService set the connection service.
func (i *invoice) SetConnectionService(connSvc payd.ConnectService) {
	i.connSvc = connSvc
}

func (i *invoice) connectPaymentChannel(ctx context.Context, invoiceID string) error {
	if i.connSvc == nil {
		return nil
	}

	return i.connSvc.Connect(ctx, payd.ConnectArgs{InvoiceID: invoiceID})
}
