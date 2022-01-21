package sockets

import (
	"context"
	"fmt"

	"github.com/libsv/go-p4"
	"github.com/libsv/payd"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/theflyingcodr/sockets"
	"github.com/theflyingcodr/sockets/client"
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
	c.RegisterListener(RoutePayment, p.create)
	c.RegisterListener(RoutePaymentACK, p.ack)
}

func (p *payments) create(ctx context.Context, msg *sockets.Message) (*sockets.Message, error) {
	var req p4.Payment
	if err := msg.Bind(&req); err != nil {
		return nil, errors.Wrap(err, "failed to bind request")
	}
	resp := msg.NewFrom(RoutePaymentACK)
	_, err := p.svc.PaymentCreate(ctx, payd.PaymentCreateArgs{InvoiceID: msg.ChannelID()}, req)
	if err != nil {
		log.Err(err).Msg("failed to create payment, returning ack")
		_ = resp.WithBody(p4.PaymentACK{
			Payment: &req,
			Memo:    err.Error(),
			Error:   1,
		})
		return resp, nil
	}
	_ = resp.WithBody(p4.PaymentACK{
		Payment: &req,
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
		// ack the error
		log.Err(p.svc.Ack(ctx, payd.AckArgs{
			InvoiceID: msg.ChannelID(),
			TxID:      req.Payment.SPVEnvelope.TxID,
		}, payd.Ack{
			Failed: true,
			Reason: req.Memo,
		})).Msg("failed to updated tx state")
		return nil, fmt.Errorf("failed to send payment, code: %d reason: %s", req.Error, req.Memo)
	}

	// handle the success
	if err := p.svc.Ack(ctx, payd.AckArgs{
		InvoiceID: msg.ChannelID(),
		TxID:      req.Payment.SPVEnvelope.TxID,
	}, payd.Ack{
		Failed: false,
		Reason: "",
	}); err != nil {
		return nil, err
	}
	log.Info().Msgf("payment success for %s", msg.ChannelID())
	return msg.NoContent()
}
