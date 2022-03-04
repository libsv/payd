package dpp

import (
	"context"

	"github.com/libsv/go-bc/spv"
	"github.com/libsv/go-bt/v2"
	"github.com/pkg/errors"
	validator "github.com/theflyingcodr/govalidator"
)

// Payment is a Payment message used in BIP270.
// See https://github.com/moneybutton/bips/blob/master/bip-0270.mediawiki#payment
type Payment struct {
	// MerchantData is copied from PaymentDetails.merchantData.
	// Payment hosts may use invoice numbers or any other data they require to match Payments to PaymentRequests.
	// Note that malicious clients may modify the merchantData, so should be authenticated
	// in some way (for example, signed with a payment host-only key).
	// Maximum length is 10000 characters.
	MerchantData Merchant `json:"merchantData"`
	// RefundTo is a paymail to send a refund to should a refund be necessary.
	// Maximum length is 100 characters
	RefundTo *string `json:"refundTo"  swaggertype:"primitive,string" example:"me@paymail.com"`
	// Memo is a plain-text note from the customer to the payment host.
	Memo string `json:"memo" example:"for invoice 123456"`
	// SPVEnvelope which contains the details of previous transaction and Merkle proof of each input UTXO.
	// Should be available if SPVRequired is set to true in the paymentRequest.
	// See https://tsc.bitcoinassociation.net/standards/spv-envelope/
	SPVEnvelope *spv.Envelope `json:"spvEnvelope"`
	// RawTX should be sent if SPVRequired is set to false in the payment request.
	RawTX *string `json:"rawTx"`
	// ProofCallbacks are optional and can be supplied when the sender wants to receive
	// a merkleproof for the transaction they are submitting as part of the SPV Envelope.
	//
	// This is especially useful if they are receiving change and means when they use it
	// as an input, they can provide the merkle proof.
	ProofCallbacks map[string]ProofCallback `json:"proofCallbacks"`
}

// Validate will ensure the users request is correct.
func (p Payment) Validate() error {
	v := validator.New().
		Validate("spvEnvelope/rawTx", func() error {
			if p.RawTX == nil && p.SPVEnvelope == nil {
				return errors.New("either an SPVEnvelope or a rawTX are required")
			}
			return nil
		}).
		Validate("merchantData.extendedData", validator.NotEmpty(p.MerchantData.ExtendedData))
	if p.MerchantData.ExtendedData != nil {
		v = v.Validate("merchantData.paymentReference", validator.NotEmpty(p.MerchantData.ExtendedData["paymentReference"]))
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
	if p.RawTX != nil {
		v = v.Validate("rawTx", func() error {
			if _, err := bt.NewTxFromString(*p.RawTX); err != nil {
				return errors.Wrap(err, "invalid rawTx supplied")
			}
			return nil
		})
	}
	if p.RefundTo != nil {
		v = v.Validate("refundTo", validator.StrLength(*p.RefundTo, 0, 100))
	}
	return v.Err()
}

// ProofCallback is used by a payee to request a merkle proof is sent to them
// as proof of acceptance of the tx they have provided in the spvEnvelope.
type ProofCallback struct {
	Token string `json:"token"`
}

// PaymentACK message used in BIP270.
// See https://github.com/moneybutton/bips/blob/master/bip-0270.mediawiki#paymentack
type PaymentACK struct {
	ID          string           `json:"id"`
	TxID        string           `json:"tx_id"`
	Memo        string           `json:"memo"`
	PeerChannel *PeerChannelData `json:"peer_channel"`
	// A number indicating why the transaction was not accepted. 0 or undefined indicates no error.
	// A 1 or any other positive integer indicates an error. The errors are left undefined for now;
	// it is recommended only to use “1” and to fill the memo with a textual explanation about why
	// the transaction was not accepted until further numbers are defined and standardised.
	Error int `json:"error,omitempty"`
}

// PeerChannelData holds peer channel information for subscribing to and reading from a peer channel.
type PeerChannelData struct {
	Host      string `json:"host"`
	Path      string `json:"path"`
	ChannelID string `json:"channel_id"`
	Token     string `json:"token"`
}

// PaymentCreateArgs identifies the paymentID used for the payment.
type PaymentCreateArgs struct {
	PaymentID string `param:"paymentID"`
}

// Validate will ensure that the PaymentCreateArgs are supplied and correct.
func (p PaymentCreateArgs) Validate() error {
	return validator.New().
		Validate("paymentID", validator.NotEmpty(p.PaymentID)).
		Err()
}

// PaymentService enforces business rules when creating payments.
type PaymentService interface {
	PaymentCreate(ctx context.Context, args PaymentCreateArgs, req Payment) (*PaymentACK, error)
}

// PaymentWriter will write a payment to a data store.
type PaymentWriter interface {
	PaymentCreate(ctx context.Context, args PaymentCreateArgs, req Payment) (*PaymentACK, error)
}
