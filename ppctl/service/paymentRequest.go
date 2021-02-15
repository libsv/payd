package service

import (
	"context"
	"fmt"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/libsv/go-bt"
	"github.com/pkg/errors"
	validator "github.com/theflyingcodr/govalidator"

	"github.com/libsv/go-payd/config"
	"github.com/libsv/go-payd/ipaymail"
	"github.com/libsv/go-payd/ppctl"
	"github.com/libsv/go-payd/wallet"
)

const (
	keyname        = "keyname"
	derivationPath = "0"
)

type paymentRequestService struct {
	env        *config.Server
	privKeySvc wallet.PrivateKeyService
	scStore    ppctl.ScriptKeyStorer
	invStore   ppctl.InvoiceStorer
}

// NewPaymentRequestService will create and return a new payment service.
func NewPaymentRequestService(env *config.Server, privKeySvc wallet.PrivateKeyService, scStore ppctl.ScriptKeyStorer, invStore ppctl.InvoiceStorer) *paymentRequestService {
	if env == nil || env.Hostname == "" {
		log.Fatal("env hostname should be set")
	}
	return &paymentRequestService{env: env, privKeySvc: privKeySvc, scStore: scStore, invStore: invStore}
}

// CreatePaymentRequest handles setting up a new PaymentRequest response and can use and optional existing paymentID.
func (p *paymentRequestService) CreatePaymentRequest(ctx context.Context, args ppctl.PaymentRequestArgs) (*ppctl.PaymentRequest, error) {
	if err := validator.New().
		Validate("paymentID", validator.NotEmpty(args.PaymentID)).
		Validate("hostname", validator.NotEmpty(p.env)); err.Err() != nil {
		return nil, err
	}
	inv, err := p.invStore.Invoice(ctx, ppctl.InvoiceArgs{PaymentID: args.PaymentID})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get invoice when creating payment request")
	}

	// get the master key stored
	// TODO - later we will allow users to provide their own key for now we've hardcoded to keyname
	xprv, err := p.privKeySvc.PrivateKey(ctx, keyname)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	outs := make([]*ppctl.Output, 0)
	// generate a new child for each output
	// TODO - figure out how many outputs we need?
	// TODO - what should derivation path be, just hardcoded for now. Users could create their own paths which we lookup or something
	key, err := p.privKeySvc.DeriveChildFromKey(xprv, derivationPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	pubKey, err := p.privKeySvc.PubFromXPrv(key)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	o, err := bt.NewP2PKHOutputFromPubKeyBytes(pubKey, inv.Satoshis)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	outs = append(outs, &ppctl.Output{
		Amount: o.Satoshis,
		Script: o.GetLockingScriptHexString(),
	})
	// store outputs so we can get them later for validation
	if err := p.storeKeys(ctx, keyname, derivationPath, outs); err != nil {
		return nil, errors.WithStack(err)
	}
	return &ppctl.PaymentRequest{
		Network:             "bitcoin-sv", // TODO: check if bitcoin or bitcoin-sv?
		Outputs:             outs,
		CreationTimestamp:   time.Now().UTC().Unix(),
		ExpirationTimestamp: time.Now().Add(24 * time.Hour).UTC().Unix(),
		PaymentURL:          fmt.Sprintf("http://%s/v1/payment/%s", p.env.Hostname, args.PaymentID),
		Memo:                fmt.Sprintf("Payment request for invoice %s", args.PaymentID),
		MerchantData: &ppctl.MerchantData{ // TODO: get from settings
			AvatarURL:    "https://bit.ly/3c4iaup",
			MerchantName: "go-payd",
		},
	}, nil
}

// createPaymailOutputs is not currently used but will be when we incorporate this feature.
func (p *paymentRequestService) createPaymailOutputs(paymentID string, outs []*ppctl.Output) ([]*ppctl.Output, error) {
	ref, os, err := ipaymail.GetP2POutputs("jad@moneybutton.com", 10000)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get paymail outputs")
	}
	log.Debugf("reference: %s", ref)

	ipaymail.ReferencesMap[paymentID] = ref

	// change returned hexString output script into bytes TODO: understand what i wrote
	for _, o := range os {
		out := &ppctl.Output{
			Amount: o.Satoshis,
			Script: o.Script,
		}
		outs = append(outs, out)
	}
	return outs, nil
}

// storeKeys will store each key along with keyname and derivation path
// to allow us to validate the outputs sent in the users payment.
// If there is a failure all will be rolled back.
func (p *paymentRequestService) storeKeys(ctx context.Context, keyName, derivPath string, outs []*ppctl.Output) error {
	keys := make([]ppctl.CreateScriptKey, 0)
	for _, o := range outs {
		keys = append(keys, ppctl.CreateScriptKey{
			LockingScript:  o.Script,
			KeyName:        keyName,
			DerivationPath: derivPath,
		})
	}
	return errors.Wrap(p.scStore.Create(ctx, keys), "failed to create payment request when storing key map")
}
