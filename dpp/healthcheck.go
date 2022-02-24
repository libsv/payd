package dpp

import (
	"context"
	"net/url"
	"time"

	"github.com/InVisionApp/go-health/v2"
	"github.com/libsv/payd"
	"github.com/libsv/payd/config"
	"github.com/pkg/errors"
	"github.com/theflyingcodr/lathos"
	"github.com/theflyingcodr/sockets"
	"github.com/theflyingcodr/sockets/client"
)

type healthCheck struct {
	h      health.IHealth
	c      *client.Client
	cfg    *config.P4
	invSvc payd.InvoiceService
}

func NewHealthCheck(h health.IHealth, c *client.Client, invSvc payd.InvoiceService, cfg *config.P4) payd.HealthCheck {
	return &healthCheck{
		h:      h,
		c:      c,
		cfg:    cfg,
		invSvc: invSvc,
	}
}

func (h *healthCheck) Start() error {
	u, err := url.Parse(h.cfg.ServerHost)
	if err != nil {
		return err
	}
	if u.Scheme != "ws" && u.Scheme != "wss" {
		return nil
	}

	if err := h.commsCheck(); err != nil {
		return errors.Wrap(err, "failed to start comms health check")
	}
	if err := h.channelCheck(); err != nil {
		return errors.Wrap(err, "failed to start channel health check")
	}
	return nil
}

func (h *healthCheck) commsCheck() error {
	if err := h.h.AddCheck(&health.Config{
		Name: "p4-comms",
		Checker: &commsCheck{
			c:    h.c,
			host: h.cfg.ServerHost,
		},
		Interval: time.Duration(2) * time.Second,
	}); err != nil {
		return errors.Wrap(err, "failed to create p4-comms healthcheck")
	}
	if err := h.h.AddCheck(&health.Config{
		Name: "p4-channel-conn",
		Checker: &channelCheck{
			c:      h.c,
			host:   h.cfg.ServerHost,
			invSvc: h.invSvc,
		},
		Interval: time.Duration(10) * time.Second,
	}); err != nil {
		return errors.Wrap(err, "failed to create p4-channel-conn healthcheck")
	}
	return nil
}

func (h *healthCheck) channelCheck() error {
	return nil
}

type commsCheck struct {
	c    *client.Client
	host string
}

// Status of communication.
func (ch *commsCheck) Status() (interface{}, error) {
	if err := ch.c.JoinChannel(ch.host, "health", nil, map[string]string{
		"internal": "true",
	}); err != nil {
		return nil, errors.Wrap(err, "failed to join p4 health channel")
	}
	if err := ch.c.Publish(sockets.Request{
		ChannelID:  "health",
		MessageKey: "my-p4",
		Body:       "ping",
	}); err != nil {
		return nil, errors.Wrap(err, "failed to ping p4")
	}
	ch.c.LeaveChannel("health", nil)
	return nil, nil
}

type channelCheck struct {
	c      *client.Client
	host   string
	invSvc payd.InvoiceService
}

// Status of channels.
func (ch *channelCheck) Status() (interface{}, error) {
	invoices, err := ch.invSvc.InvoicesPending(context.Background())
	if err != nil {
		if lathos.IsNotFound(err) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to get invoices for channel check")
	}
	for _, invoice := range invoices {
		if invoice.ExpiresAt.Time.UTC().Before(time.Now().UTC()) {
			continue
		}
		if ch.c.HasChannel(invoice.ID) {
			continue
		}

		if err := ch.c.JoinChannel(ch.host, invoice.ID, nil, map[string]string{
			"internal": "true",
		}); err != nil {
			return nil, errors.Wrapf(err, "failed rejoining channel for invoice '%s'", invoice.ID)
		}
	}
	return nil, nil
}
