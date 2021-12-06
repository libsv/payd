package sockets

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/theflyingcodr/sockets"
	"github.com/theflyingcodr/sockets/client"

	"github.com/libsv/go-p4"
	"github.com/libsv/payd"
	"github.com/libsv/payd/config"
)

type paymentRequest struct {
	transacter payd.Transacter
	prSvc      payd.PaymentRequestService
	envSvc     payd.EnvelopeService
	p4Cfg      *config.P4
}

// NewPaymentRequest will setup and return a new PaymentRequest socket listener.
func NewPaymentRequest(transacter payd.Transacter, svc payd.PaymentRequestService, envSvc payd.EnvelopeService, p4Cfg *config.P4) *paymentRequest {
	return &paymentRequest{
		transacter: transacter,
		prSvc:      svc,
		envSvc:     envSvc,
		p4Cfg:      p4Cfg,
	}
}

// RegisterListeners will setup a listener for payments.
func (p *paymentRequest) RegisterListeners(c *client.Client) {
	c.RegisterListener(RoutePaymentRequestCreate, p.create)
	c.RegisterListener(RoutePaymentRequestResponse, p.response)
}

func (p *paymentRequest) create(ctx context.Context, msg *sockets.Message) (*sockets.Message, error) {
	log.Debug().Msg("socket: payment request create hit")
	invoiceID := msg.ChannelID()
	pr, err := p.prSvc.PaymentRequest(ctx, payd.PaymentRequestArgs{InvoiceID: invoiceID})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	pr.PaymentURL = fmt.Sprintf("%s/%s", p.p4Cfg.ServerHost, invoiceID)
	resp := msg.NewFrom(RoutePaymentRequestResponse)
	if err := resp.WithBody(pr); err != nil {
		fmt.Printf("body %+v\n", pr)
		return nil, err
	}
	resp.Expiration = &pr.ExpirationTimestamp
	return resp, nil
}

func (p *paymentRequest) response(ctx context.Context, msg *sockets.Message) (*sockets.Message, error) {
	var req p4.PaymentRequest
	if err := msg.Bind(&req); err != nil {
		return nil, err
	}
	payment := p4.Payment{
		MerchantData: *req.MerchantData,
		RefundTo:     nil, // TODO - read users paymail
		Memo:         req.Memo,
	}
	// TODO : fix this, shouldn't be in this layer
	ctx = p.transacter.WithTx(ctx)
	defer func() {
		_ = p.transacter.Rollback(ctx)
	}()
	env, err := p.envSvc.Envelope(ctx, payd.EnvelopeArgs{PayToURL: msg.ChannelID()}, req)
	if err != nil {
		return nil, err
	}
	payment.SPVEnvelope = env
	resp := msg.NewFrom(RoutePayment)
	if err := resp.WithBody(&payment); err != nil {
		return nil, err
	}
	if err := p.transacter.Commit(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction")
	}
	return resp, nil
}
