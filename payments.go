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
	// MerchantData is copied from PaymentDetails.merchantData.
	// Payment hosts may use invoice numbers or any other data they require to match Payments to PaymentRequests.
	// Note that malicious clients may modify the merchantData, so should be authenticated
	// in some way (for example, signed with a payment host-only key).
	// Maximum length is 10000 characters.
	MerchantData User `json:"merchantData"`
	// RefundTo is a paymail to send a refund to should a refund be necessary.
	// Maximum length is 100 characters
	RefundTo null.String `json:"refundTo"  swaggertype:"primitive,string" example:"me@paymail.com"`
	// Memo is a plain-text note from the customer to the payment host.
	Memo string `json:"memo" example:"for invoice 123456"`
	// SPVEnvelope which contains the details of previous transaction and Merkle proof of each input UTXO.
	// Should be available if SPVRequired is set to true in the paymentRequest.
	// See https://tsc.bitcoinassociation.net/standards/spv-envelope/
	SPVEnvelope *spv.Envelope `json:"spvEnvelope"`
	// RawTX should be sent if SPVRequired is set to false in the payment request.
	RawTX null.String `json:"rawTx"`
	// ProofCallbacks are optional and can be supplied when the sender wants to receive
	// a merkleproof for the transaction they are submitting as part of the SPV Envelope.
	//
	// This is especially useful if they are receiving change and means when they use it
	// as an input, they can provide the merkle proof.
	ProofCallbacks map[string]ProofCallback `json:"proofCallbacks"`
}

// Validate will ensure the users request is correct.
func (p PaymentCreate) Validate(spvRequired bool) error {
	v := validator.New().
		Validate("spvEnvelope/rawTx", func() error {
			if p.RawTX.IsZero() && p.SPVEnvelope == nil {
				return errors.New("either an SPVEnvelope or a rawTX are required")
			}
			return nil
		}).
		Validate("merchantData.extendedData", validator.NotEmpty(p.MerchantData.ExtendedData))
	if p.MerchantData.ExtendedData != nil {
		v = v.Validate("merchantData.paymentReference", validator.NotEmpty(p.MerchantData.ExtendedData["paymentReference"]))
	}

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

// PaymentCreateArgs are used to identify a payment.
type PaymentCreateArgs struct {
	InvoiceID string
}

// ProofCallback contains information relating to a merkleproof callback.
type ProofCallback struct {
	// Token to use for authentication when sending the proof to the destination. Optional.
	Token string
}

// PaymentsService is used for handling payments.
type PaymentsService interface {
	// PaymentCreate will validate a new payment.
	PaymentCreate(ctx context.Context, args PaymentCreateArgs, req PaymentCreate) error
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
