package service

import (
	"context"
	"encoding/json"

	"github.com/libsv/go-bk/envelope"
	"github.com/pkg/errors"
	validator "github.com/theflyingcodr/govalidator"

	"github.com/libsv/payd"
)

type proofs struct {
	wtr payd.ProofsWriter
}

// NewProofsService will setup and return a new merkle proof service.
func NewProofsService(rdrWtr payd.ProofsWriter) *proofs {
	return &proofs{
		wtr: rdrWtr,
	}
}

// Create will add a merkle proof to a data store for persistent storage once it has
// been validated.
func (p *proofs) Create(ctx context.Context, args payd.ProofCreateArgs, req envelope.JSONEnvelope) error {
	var proof *payd.ProofWrapper
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
	if err := p.wtr.ProofCreate(ctx, *proof); err != nil {
		return errors.Wrap(err, "failed to save proof")
	}
	return nil
}
