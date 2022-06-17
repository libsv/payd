package service

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/libsv/go-bk/envelope"
	"github.com/libsv/go-dpp"
	"github.com/libsv/go-spvchannels"
	"github.com/libsv/payd/log"
	"github.com/pkg/errors"
	validator "github.com/theflyingcodr/govalidator"

	"github.com/libsv/payd"
)

type proofs struct {
	wtr payd.ProofsWriter
	l   log.Logger
}

// NewProofsService will setup and return a new merkle proof service.
func NewProofsService(wtr payd.ProofsWriter, l log.Logger) *proofs {
	return &proofs{
		wtr: wtr,
		l:   l,
	}
}

// Create will add a merkle proof to a data store for persistent storage once it has
// been validated.
func (p *proofs) Create(ctx context.Context, args dpp.ProofCreateArgs, req envelope.JSONEnvelope) error {
	var proof dpp.ProofWrapper
	if err := json.Unmarshal([]byte(req.Payload), &proof); err != nil {
		return errors.Wrap(err, "failed to unmarshall JSONEnvelope")
	}
	if err := validator.New().Validate("jsonEnvelope", func() error {
		if ok, err := req.IsValid(); !ok || err != nil {
			return errors.Wrap(err, "invalid merkleproof envelope")
		}
		return nil
	}).Err(); err != nil {
		return err
	}
	if err := proof.Validate(args); err != nil {
		return err
	}
	if err := p.wtr.ProofCreate(ctx, proof); err != nil {
		return errors.Wrap(err, "failed to save proof")
	}
	return nil
}

func (p *proofs) HandlePeerChannelsMessage(ctx context.Context, msgs spvchannels.MessagesReply) (bool, error) {
	p.l.Debugf("handling peer channel messages %d", len(msgs))
	for _, msg := range msgs {
		p.l.Debugf("handling peer channel message %+v", msg.Received)
		payload, err := base64.StdEncoding.DecodeString(msg.Payload)
		if err != nil {
			return false, errors.Wrap(err, "error decoding payload")
		}
		var env envelope.JSONEnvelope
		if err := json.Unmarshal(payload, &env); err != nil {
			return false, errors.Wrap(err, "error unmarshalling json envelope")
		}
		p.l.Debugf("handling peer channel message - decoded envelope")
		mm := make(map[string]interface{})
		if err := json.Unmarshal([]byte(env.Payload), &mm); err != nil {
			return false, errors.Wrap(err, "error unmarshalling payload")
		}

		if mm["callbackReason"].(string) != "merkleProof" {
			p.l.Debugf("skipping msg %#v", msg)
			continue
		}
		p.l.Debugf("handling peer channel message - proof received")
		txID := mm["callbackTxId"].(string)
		if err := p.Create(ctx, dpp.ProofCreateArgs{
			TxID: txID,
		}, env); err != nil {
			return false, errors.Wrap(err, "failed to store proof msg")
		}
		p.l.Debugf("handling peer channel message - stored")
	}
	return false, nil
}
