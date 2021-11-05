package sockets

import (
	"github.com/libsv/go-bk/envelope"
	"github.com/pkg/errors"
	"github.com/theflyingcodr/sockets"

	"github.com/libsv/payd"

	"context"

	"github.com/theflyingcodr/sockets/client"
)

type proofs struct {
	svc payd.ProofsService
}

// NewProofs will setup proof handlers with their deps.
func NewProofs(svc payd.ProofsService) *proofs {
	return &proofs{svc: svc}
}

// RegisterListeners will setup handlers for dealing with proofs.
func (p *proofs) RegisterListeners(s *client.Client) {
	s = s.RegisterListener(RouteProofCreate, p.create)
}

// create will take a proof sent wrapped in a JSONEnvelope and store it, if valid.
func (p *proofs) create(ctx context.Context, msg *sockets.Message) (*sockets.Message, error) {
	var req envelope.JSONEnvelope
	if err := msg.Bind(&req); err != nil {
		return nil, errors.WithStack(err)
	}
	txID := msg.Headers.Get("x-tx-id")

	if err := p.svc.Create(ctx, payd.ProofCreateArgs{TxID: txID}, req); err != nil {
		return nil, errors.WithStack(err)
	}
	return msg.NoContent()
}
