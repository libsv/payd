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
	"github.com/libsv/go-payd/wallet/service"
)

type paymentRequestService struct {
	privKeySvc wallet.PrivateKeyService
}

// NewPaymentRequestService
func NewPaymentRequestService(privKeySvc wallet.PrivateKeyService) *paymentRequestService {
	return &paymentRequestService{privKeySvc: privKeySvc}
}

// CreatePaymentRequest handles setting up a new PaymentRequest response and can use and optional existing paymentID.
func (p *paymentRequestService) CreatePaymentRequest(ctx context.Context, args bip270.PaymentRequestArgs) (*bip270.PaymentRequest, error) {
	if err := validator.New().
		Validate("paymentID", validator.NotEmpty(args.PaymentID)).
		Validate("hostname", validator.NotEmpty(args.Hostname)); err.Err() != nil {
		return nil, err
	}
	// TODO: get amount from paymentID key (badger db) and get paymail p2p outputs when creating invoice not here
	var outs []*bip270.Output
	if args.UsePaymail {
		ref, os, err := ipaymail.GetP2POutputs("jad@moneybutton.com", 10000)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get paymail outputs")
		}
		log.Debugf("reference: %s", ref)

		ipaymail.ReferencesMap[args.PaymentID] = ref

		// change returned hexString output script into bytes TODO: understand what i wrote
		for _, o := range os {
			out := &bip270.Output{
				Amount: o.Satoshis,
				Script: o.Script,
			}
			outs = append(outs, out)
		}
	} else {
		xprv, err := p.privKeySvc.PrivateKey(ctx, "keyname") // TODO: get from settings
		if err != nil {
			return nil, err
		}

		// TODO: derive new key for each payment!

		pubKey, err := service.PubFromXPrv(xprv)
		if err != nil {
			return nil, err
		}
		o, err := bt.NewP2PKHOutputFromPubKeyBytes(pubKey, 10000) // TODO: get amount from invoice
		if err != nil {
			return nil, err
		}
		outs = append(outs, &bip270.Output{
			Amount: o.Satoshis,
			Script: o.GetLockingScriptHexString(),
		})
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
