package service

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/libsv/go-dpp"
	"github.com/libsv/payd"
)

// payChannel is used to initiate payments down an async payment channel.
// This differs enough from the pay service to need it's own service.
type payChannel struct {
	wtr payd.PayWriter
}

// NewPayChannel will setup and return a new payment channel handler.
func NewPayChannel(wtr payd.PayWriter) *payChannel {
	return &payChannel{wtr: wtr}
}

// Pay will initiate an async payment flow.
func (p *payChannel) Pay(ctx context.Context, req payd.PayRequest) (*dpp.PaymentACK, error) {
	err := p.wtr.Pay(ctx, req)
	if err != nil {
		log.Err(err).Msg("failed to setup async channel")
		return &dpp.PaymentACK{
			Memo:  "failed to setup channel " + err.Error(),
			Error: 1,
		}, nil
	}
	// once pending in received, the receiver should listen on a channel for the async ack.
	return &dpp.PaymentACK{
		Memo:  "pending",
		Error: 0,
	}, nil
}
