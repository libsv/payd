package sockets

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/theflyingcodr/sockets"
	"gopkg.in/guregu/null.v3"

	"github.com/libsv/payd"
)

type paymentRequest struct {
	transacter payd.Transacter
	prSvc      payd.PaymentRequestService
	envSvc     payd.EnvelopeService
}

// NewPaymentRequest will setup and return a new PaymentRequest socket listener.
func NewPaymentRequest(transacter payd.Transacter, svc payd.PaymentRequestService, envSvc payd.EnvelopeService) *paymentRequest {
	return &paymentRequest{
		transacter: transacter,
		prSvc:      svc,
		envSvc:     envSvc,
	}
}

// RegisterListeners will setup a listener for payments.
func (p *paymentRequest) RegisterListeners(c sockets.Client) {
	c.RegisterListener(RoutePaymentRequestCreate, p.create).
		RegisterListener(RoutePaymentRequestResponse, p.response)
}

func (p *paymentRequest) create(ctx context.Context, msg *sockets.Message) (*sockets.Message, error) {
	log.Debug().Msg("socket: payment request create hit")
	pr, err := p.prSvc.PaymentRequest(ctx, payd.PaymentRequestArgs{InvoiceID: msg.ChannelID()})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	resp := msg.NewFrom(RoutePaymentRequestResponse)
	if err := resp.WithBody(pr); err != nil {
		fmt.Printf("body %+v\n", pr)
		return nil, err
	}
	resp.Expiration = &pr.ExpirationTimestamp
	return resp, nil
}

func (p *paymentRequest) response(ctx context.Context, msg *sockets.Message) (*sockets.Message, error) {
	var req payd.PaymentRequestResponse
	if err := msg.Bind(&req); err != nil {
		return nil, err
	}
	payment := payd.PaymentCreate{
		MerchantData: req.MerchantData,
		RefundTo:     null.String{}, // TODO - read users paymail
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
