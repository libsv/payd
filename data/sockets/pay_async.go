package sockets

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	validator "github.com/theflyingcodr/govalidator"
	"github.com/theflyingcodr/sockets"
	"github.com/theflyingcodr/sockets/client"

	"github.com/libsv/payd"
	"github.com/libsv/payd/config"
)

type paymentChannel struct {
	cli *client.Client
	cfg config.Socket
}

// NewPaymentChannel will setup a connection service.
func NewPaymentChannel(cfg config.Socket, cli *client.Client) *paymentChannel {
	return &paymentChannel{
		cli: cli,
		cfg: cfg,
	}
}

// Pay will join payd with a socket server and kick off the payment process.
func (c *paymentChannel) Pay(ctx context.Context, req payd.PayRequest) error {
	if err := validator.New().
		Validate("PayToUrl", validator.MatchString(req.PayToURL, reURL)).
		Err(); err != nil {
		return err
	}
	// parse url to get host connection and invoiceID
	parts := reURL.FindStringSubmatch(req.PayToURL)
	invoiceID := parts[2]
	if err := c.cli.JoinChannel(parts[1], invoiceID, nil, nil); err != nil {
		return errors.Wrapf(err, "failed to connect to channel %s", invoiceID)
	}
	// kick off the process - we will receive the messages via the socket transport listeners.
	h := http.Header{}
	h.Add("X-Origin-ID", c.cfg.ClientIdentifier)
	return errors.Wrap(c.cli.Publish(sockets.Request{
		ChannelID:  invoiceID,
		MessageKey: "paymentrequest.create",
		Body:       nil,
		Headers:    h,
	}), "failed to publish payment request socket message")
}
