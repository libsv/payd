package sockets

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/theflyingcodr/sockets"
	"github.com/theflyingcodr/sockets/client"

	"github.com/libsv/payd"
)

type payments struct {
	svc payd.PaymentsService
}

// NewPayments will setup and return a new Payments socket listener.
func NewPayments(svc payd.PaymentsService) *payments {
	return &payments{svc: svc}
}

// RegisterListeners will setup a listener for payments.
func (p *payments) RegisterListeners(c *client.Client) {
	c = c.RegisterListener(RoutePayment, p.create).
		RegisterListener(RoutePaymentACK, p.ack)
}

func (p *payments) create(ctx context.Context, msg *sockets.Message) (*sockets.Message, error) {
	var req payd.PaymentCreate
	if err := msg.Bind(&req); err != nil {
		return nil, errors.Wrap(err, "failed to bind request")
	}
	resp := msg.NewFrom(RoutePaymentACK)
	if err := p.svc.PaymentCreate(ctx, payd.PaymentCreateArgs{InvoiceID: msg.ChannelID()}, req); err != nil {
		log.Err(err).Msg("failed to create payment, returning ack")
		_ = resp.WithBody(payd.PaymentACK{
			Memo:  err.Error(),
			Error: 1,
		})
		return resp, nil
	}
	_ = resp.WithBody(payd.PaymentACK{
		Payment: req,
		Memo:    req.Memo,
	})
	return resp, nil
}

// ack handles the ack from the payment.
// This isn't fully fleshed out yet, it could notify a front end
// via another message, for now it just logs an error or returns no content.
func (p *payments) ack(ctx context.Context, msg *sockets.Message) (*sockets.Message, error) {
	var req payd.PaymentACK
	if err := msg.Bind(&req); err != nil {
		return nil, errors.Wrap(err, "failed to bind request")
	}
	if req.Error > 0 {
		return nil, fmt.Errorf("failed to send payment, code: %d reason: %s", req.Error, req.Memo)
	}
	log.Info().Msgf("payment success for %s", msg.ChannelID())
	return msg.NoContent()
}
