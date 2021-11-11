package sockets

import (
	"context"
	"fmt"

	"github.com/libsv/go-bk/envelope"
	"github.com/pkg/errors"
	"github.com/theflyingcodr/sockets"

	"github.com/libsv/payd"

	"github.com/theflyingcodr/sockets/client"
)

type proofs struct {
	svc payd.ProofsService
	cli *client.Client
}

// NewProofs will setup proof handlers with their deps.
func NewProofs(svc payd.ProofsService, cli *client.Client) *proofs {
	return &proofs{svc: svc, cli: cli}
}

// RegisterListeners will setup handlers for dealing with proofs.
func (p *proofs) RegisterListeners(c sockets.Client) {
	c.RegisterListener(RouteProofCreate, p.create)
}

// create will take a proof sent wrapped in a JSONEnvelope and store it, if valid.
func (p *proofs) create(ctx context.Context, msg *sockets.Message) (*sockets.Message, error) {
	fmt.Println("adding new proof")
	var req envelope.JSONEnvelope
	if err := msg.Bind(&req); err != nil {
		return nil, errors.WithStack(err)
	}
	txID := msg.Headers.Get("x-tx-id")
	fmt.Println(txID)
	if err := p.svc.Create(ctx, payd.ProofCreateArgs{TxID: txID}, req); err != nil {
		return nil, errors.WithStack(err)
	}
	p.cli.LeaveChannel(msg.ChannelID(), nil)
	return msg.NoContent()
}
