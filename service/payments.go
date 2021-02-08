package service

import (
	"context"
	"fmt"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/libsv/go-bt"
	"github.com/pkg/errors"

	gopayd "github.com/libsv/go-payd"
	"github.com/libsv/go-payd/ipaymail"
	"github.com/libsv/go-payd/wallet"
)

type paymentService struct {
}

func NewPaymentService() *paymentService {
	return &paymentService{}
}

func (p *paymentService) CreatePaymentRequest(ctx context.Context, args gopayd.PaymentRequestArgs) (*gopayd.PaymentRequest, error) {

	// TODO: get amount from paymentID key (badger db) and get paymail p2p outputs when creating invoice not here
	// TODO: if no paymentID, generate a random one
	var pID string
	if args.PaymentID != nil {
		pID = *args.PaymentID
	}
	var outs []*gopayd.Output

	if args.UsePaymail {
		ref, os, err := ipaymail.GetP2POutputs("jad@moneybutton.com", 10000)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get paymail outputs")
		}
		log.Debugf("reference: %s", ref)

		ipaymail.ReferencesMap[pID] = ref

		// change returned hexString output script into bytes TODO: understand what i wrote
		for _, o := range os {
			out := &gopayd.Output{
				Amount: o.Satoshis,
				Script: o.Script,
			}
			outs = append(outs, out)
		}
	} else {
		xprv, err := wallet.GetPrivateKey("keyname") // TODO: get from settings
		if err != nil {
			return nil, err
		}

		// TODO: derive new key for each payment!

		pubKey, err := wallet.PubFromXPrv(xprv)
		if err != nil {
			return nil, err
		}

		o, err := bt.NewP2PKHOutputFromPubKeyBytes(pubKey, 10000) // TODO: get amount from invoice
		if err != nil {
			return nil, err
		}

		out := &gopayd.Output{
			Amount: o.Satoshis,
			Script: o.GetLockingScriptHexString(),
		}

		outs = append(outs, out)
	}

	// endpoint := "localhost:1323" // TODO: get from settings
	endpoint := "178.62.87.120:1323" // TODO: get from settings

	return &gopayd.PaymentRequest{
		Network:             "bitcoin-sv", // TODO: check if bitcoin or bitcoin-sv?
		Outputs:             outs,
		CreationTimestamp:   time.Now().UTC().Unix(),
		ExpirationTimestamp: time.Now().AddDate(0, 0, 1).UTC().Unix(),
		PaymentURL:          fmt.Sprintf("http://%s/v1/payment/%s", endpoint, pID),
		Memo:                fmt.Sprintf("Payment request for invoice %s", pID),
		MerchantData: &gopayd.MerchantData{ // TODO: get from settings
			AvatarURL:    "https://bit.ly/3c4iaup",
			MerchantName: "go-payd",
		},
	}, nil
}
