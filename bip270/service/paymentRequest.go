package service

import (
	"context"
	"fmt"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/libsv/go-bt"
	"github.com/pkg/errors"
	validator "github.com/theflyingcodr/govalidator"

	"github.com/libsv/go-payd/bip270"
	"github.com/libsv/go-payd/ipaymail"
	"github.com/libsv/go-payd/wallet"
)

const (
	keyname        = "keyname"
	derivationPath = "0"
)

type paymentRequestService struct {
	privKeySvc wallet.PrivateKeyService
	scStore    bip270.ScriptKeyStorer
}

// NewPaymentRequestService will create and return a new payment service.
func NewPaymentRequestService(privKeySvc wallet.PrivateKeyService, scStore bip270.ScriptKeyStorer) *paymentRequestService {
	return &paymentRequestService{privKeySvc: privKeySvc, scStore: scStore}
}

// CreatePaymentRequest handles setting up a new PaymentRequest response and can use and optional existing paymentID.
func (p *paymentRequestService) CreatePaymentRequest(ctx context.Context, args bip270.PaymentRequestArgs) (*bip270.PaymentRequest, error) {
	if err := validator.New().
		Validate("paymentID", validator.NotEmpty(args.PaymentID)).
		Validate("hostname", validator.NotEmpty(args.Hostname)); err.Err() != nil {
		return nil, err
	}
	// TODO: get amount from paymentID key (badger db) and get paymail p2p outputs when creating invoice not here

	// TODO - check for paymail - we'll not do it this version though
	// get the master key stored
	// TODO - later we will allow users to provide their own key for now we've hardcoded to keyname
	xprv, err := p.privKeySvc.PrivateKey(ctx, keyname)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	outs := make([]*bip270.Output, 0)
	// generate a new child for each output
	// TODO - figure out how many outputs we need?
	// TODO - what should derivation path be?
	key, err := p.privKeySvc.DeriveChildFromKey(xprv, derivationPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	pubKey, err := p.privKeySvc.PubFromXPrv(key)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	o, err := bt.NewP2PKHOutputFromPubKeyBytes(pubKey, 10000) // TODO: get amount from invoice
	if err != nil {
		return nil, errors.WithStack(err)
	}

	outs = append(outs, &bip270.Output{
		Amount: o.Satoshis,
		Script: o.GetLockingScriptHexString(),
	})
	// store outputs so we can get them later for validation
	if err := p.storeKeys(ctx, keyname, derivationPath, outs); err != nil {
		return nil, errors.WithStack(err)
	}

	return &bip270.PaymentRequest{
		Network:             "bitcoin-sv", // TODO: check if bitcoin or bitcoin-sv?
		Outputs:             outs,
		CreationTimestamp:   time.Now().UTC().Unix(),
		ExpirationTimestamp: time.Now().Add(24 * time.Hour).UTC().Unix(),
		PaymentURL:          fmt.Sprintf("http://%s/v1/payment/%s", args.Hostname, args.PaymentID),
		Memo:                fmt.Sprintf("Payment request for invoice %s", args.PaymentID),
		MerchantData: &bip270.MerchantData{ // TODO: get from settings
			AvatarURL:    "https://bit.ly/3c4iaup",
			MerchantName: "go-payd",
		},
	}, nil
}

func (p *paymentRequestService) createPaymailOutputs(paymentID string, outs []*bip270.Output) ([]*bip270.Output, error) {
	ref, os, err := ipaymail.GetP2POutputs("jad@moneybutton.com", 10000)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get paymail outputs")
	}
	log.Debugf("reference: %s", ref)

	ipaymail.ReferencesMap[paymentID] = ref

	// change returned hexString output script into bytes TODO: understand what i wrote
	for _, o := range os {
		out := &bip270.Output{
			Amount: o.Satoshis,
			Script: o.Script,
		}
		outs = append(outs, out)
	}
	return outs, nil
}

// storeKeys will store each key along with keyname and derivation path
// to allow us to validate the outputs sent in the users payment.
func (p *paymentRequestService) storeKeys(ctx context.Context, keyName, derivPath string, outs []*bip270.Output) error {
	keys := make([]bip270.CreateScriptKey, len(outs), len(outs))
	for _, o := range outs {
		keys = append(keys, bip270.CreateScriptKey{
			LockingScript:  o.Script,
			KeyName:        keyName,
			DerivationPath: derivPath,
		})
	}
	return errors.Wrap(p.scStore.Create(ctx, keys), "failed to create payment request when storing key map")
}
