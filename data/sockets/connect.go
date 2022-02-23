package sockets

import (
	"context"

	"github.com/pkg/errors"
	"github.com/theflyingcodr/sockets/client"

	"github.com/libsv/payd"
	"github.com/libsv/payd/config"
)

type connect struct {
	cli *client.Client
	cfg *config.P4
}

// NewConnect will setup a connection service.
func NewConnect(cfg *config.P4, cli *client.Client) *connect {
	return &connect{cli: cli, cfg: cfg}
}

// Connect will join payd with a socket server and kick off the payment process.
func (c *connect) Connect(ctx context.Context, args payd.ConnectArgs) error {
	if err := c.cli.JoinChannel(c.cfg.ServerHost, args.InvoiceID, nil, map[string]string{
		"internal": "true",
	}); err != nil {
		return errors.Wrapf(err, "failed to connect to channel")
	}
	return nil
}
