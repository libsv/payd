package payd

import (
	"context"

	"github.com/libsv/go-bc/spv"
	"github.com/libsv/go-bt/v2"
	"github.com/pkg/errors"
	validator "github.com/theflyingcodr/govalidator"
	"gopkg.in/guregu/null.v3"
)

// PaymentCreate is submitted to validate and add a payment to the wallet.
type PaymentCreate struct {
	InvoiceID   string        `json:"-" param:"invoiceID"`
	SPVEnvelope *spv.Envelope `json:"spvEnvelope"`
	RawTX       null.String   `json:"rawTx"`
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
func (p PaymentCreate) Validate(spvRequired bool) error {
	v := validator.New()

	if spvRequired {
		v = v.Validate("spvEnvelope", func() error {
			if validator.NotEmpty(p.SPVEnvelope)() != nil {
				return errors.New("spvEnvelope is required by this payment")
			}
			return nil
		})
	}
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

// Payment is a payment.
type Payment struct {
	Transaction  string        `json:"transaction"`
	SPVEnvelope  *spv.Envelope `json:"spvEnvelope"`
	MerchantData User          `json:"merchantData"`
	Memo         string        `json:"memo"`
}

// PaymentSend is a send request to p4.
type PaymentSend struct {
	SPVEnvelope    *spv.Envelope            `json:"spvEnvelope"`
	ProofCallbacks map[string]ProofCallback `json:"proofCallbacks"`
	MerchantData   User                     `json:"merchantData"`
}
