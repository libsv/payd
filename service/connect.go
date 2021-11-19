package service

import (
	"context"

	"github.com/pkg/errors"

	"github.com/libsv/payd"
)

type connect struct {
	wtr    payd.ConnectWriter
	invRdr payd.InvoiceReader
}

// NewConnect will setup a new connect service used to connect this wallet to a p4 socket server.
func NewConnect(wtr payd.ConnectWriter, invRdr payd.InvoiceReader) *connect {
	return &connect{wtr: wtr, invRdr: invRdr}
}

// Connect will connect this server to a third party service using either a socket or peer channel protocol.
func (c *connect) Connect(ctx context.Context, args payd.ConnectArgs) error {
	if err := args.Validate(); err != nil {
		return err
	}
	// get the invoice if an error then it isn't here.
	if _, err := c.invRdr.Invoice(ctx, payd.InvoiceArgs{InvoiceID: args.InvoiceID}); err != nil {
		return errors.Wrapf(err, "failed to validate invoice %s when attempting to create connection", args.InvoiceID)
	}
	return c.wtr.Connect(ctx, args)
}
