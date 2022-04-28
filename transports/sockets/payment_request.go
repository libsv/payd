package sockets

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/theflyingcodr/lathos"
	"github.com/theflyingcodr/sockets"
	"github.com/theflyingcodr/sockets/client"

	"github.com/libsv/go-dpp"
	"github.com/libsv/payd"
	"github.com/libsv/payd/config"
)

type paymentRequest struct {
	transacter payd.Transacter
	prSvc      payd.PaymentRequestService
	envSvc     payd.AncestryService
	dppCfg     *config.DPP
}

// NewPaymentRequest will setup and return a new PaymentRequest socket listener.
func NewPaymentRequest(transacter payd.Transacter, svc payd.PaymentRequestService, envSvc payd.AncestryService, dppCfg *config.DPP) *paymentRequest {
	return &paymentRequest{
		transacter: transacter,
		prSvc:      svc,
		envSvc:     envSvc,
		dppCfg:     dppCfg,
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
		resp := msg.NewFrom(RoutePaymentRequestError)

		var marshalErr error
		var clientErr lathos.ClientError
		if errors.As(err, &clientErr) {
			marshalErr = resp.WithBody(payd.ClientError{
				ID:      clientErr.ID(),
				Code:    clientErr.Code(),
				Title:   clientErr.Title(),
				Message: clientErr.Detail(),
			})
		} else {
			log.Error().Err(err)
			marshalErr = resp.WithBody(payd.ClientError{
				ID:      "",
				Code:    "500",
				Title:   "Internal Server Error",
				Message: "Internal server error",
			})
		}
		if marshalErr != nil {
			return nil, errors.WithStack(err)
		}

		return resp, nil
	}
	pr.PaymentURL = fmt.Sprintf("%s/%s", p.dppCfg.ServerHost, invoiceID)
	resp := msg.NewFrom(RoutePaymentRequestResponse)
	if err := resp.WithBody(pr); err != nil {
		fmt.Printf("body %+v\n", pr)
		return nil, err
	}
	resp.Expiration = &pr.ExpirationTimestamp
	return resp, nil
}

func (p *paymentRequest) response(ctx context.Context, msg *sockets.Message) (*sockets.Message, error) {
	var req dpp.PaymentRequest
	if err := msg.Bind(&req); err != nil {
		return nil, err
	}
	payment := dpp.Payment{
		MerchantData: *req.MerchantData,
		RefundTo:     nil, // TODO - read users paymail
		Memo:         req.Memo,
	}
	// TODO : fix this, shouldn't be in this layer
	ctx = p.transacter.WithTx(ctx)
	defer func() {
		_ = p.transacter.Rollback(ctx)
	}()
	env, err := p.envSvc.AncestryCreate(ctx, payd.EnvelopeArgs{PayToURL: msg.ChannelID()}, req)
	if err != nil {
		return nil, err
	}
	bb, err := env.Bytes()
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert ancestry to bytes")
	}
	ancestry := hex.EncodeToString(bb)
	payment.Ancestry = &ancestry
	payment.RawTx = &env.RawTx
	resp := msg.NewFrom(RoutePayment)
	if err := resp.WithBody(&payment); err != nil {
		return nil, err
	}
	if err := p.transacter.Commit(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction")
	}
	return resp, nil
}
