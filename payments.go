package gopayd

import (
	"context"

	"github.com/libsv/go-bc/spv"
	"github.com/libsv/go-bt/v2"
	"github.com/pkg/errors"
	validator "github.com/theflyingcodr/govalidator"
)

// PaymentCreate is submitted to validate and add a payment to the wallet.
type PaymentCreate struct {
	InvoiceID   string        `json:"-" param:"invoiceID"`
	SPVEnvelope *spv.Envelope `json:"spvEnvelope"`
	// ProofCallbacks allow support of multiple callbacks for merkle proofs
	// this will help support multisig and also transmitting proofs to the sender wallet.
	//    "proofCallbacks": {
	//        "http://domain.com/proofs": {
	//            "token": "abc123"
	//        }
	//    }
	ProofCallbacks map[string]ProofCallback `json:"proofCallbacks"`
}

// Validate will ensure the users request is correct.
func (p PaymentCreate) Validate() error {
	v := validator.New().
		Validate("spvEnvelope", validator.NotEmpty(p.SPVEnvelope))

	// perform a light validation of the envelope, make sure we have a valid root txID
	// the root rawTx is actually a tx and that the supplied root txhex and txid match
	if p.SPVEnvelope != nil {
		v = v.Validate("spvEnvelope.txId", validator.StrLengthExact(p.SPVEnvelope.TxID, 64)).
			Validate("spvEnvelope.rawTx", func() error {
				tx, err := bt.NewTxFromString(p.SPVEnvelope.RawTx)
				if err != nil {
					return errors.Wrap(err, "invalid rawTx hex supplied")
				}
				if tx.TxID() != p.SPVEnvelope.TxID {
					return errors.New("transaction mismatch, root txId does not match rawTx supplied")
				}

				return nil
			})
	}
	return v.Err()
}

// ProofCallback contains information relating to a merkleproof callback.
type ProofCallback struct {
	// Token to use for authentication when sending the proof to the destination. Optional.
	Token string
}

// PaymentsService is used for handling payments.
type PaymentsService interface {
	// PaymentCreate will validate a new payment.
	PaymentCreate(ctx context.Context, req PaymentCreate) error
}
