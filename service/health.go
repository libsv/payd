package service

import (
	"context"
	"net/url"

	"github.com/libsv/payd"
	"github.com/libsv/payd/config"
	"github.com/theflyingcodr/sockets"
	"github.com/theflyingcodr/sockets/client"
)

type healthSvc struct {
	c   *client.Client
	cfg *config.P4
}

func NewHealthService(c *client.Client, cfg *config.P4) payd.HealthService {
	return &healthSvc{
		c:   c,
		cfg: cfg,
	}
}

func (h *healthSvc) Health(ctx context.Context) error {
	u, err := url.Parse(h.cfg.ServerHost)
	if err != nil {
		return err
	}
	switch u.Scheme {
	case "ws", "wss":
		if err := h.c.JoinChannel(h.cfg.ServerHost, "health", nil); err != nil {
			return err
		}
		if err := h.c.Publish(sockets.Request{
			ChannelID:  "health",
			MessageKey: "my-p4",
			Body:       "ping",
		}); err != nil {
			return err
		}
		h.c.LeaveChannel("health", nil)
	}

	return nil
}
