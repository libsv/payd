package bip270

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/libsv/go-bt"
	"github.com/libsv/go-payd/ipaymail"
	"github.com/libsv/go-payd/wallet"
)

// SolicitPaymentRequestHandler is used to obtain a BIP270
// PaymentRequest from the merchant/receiver of the
// payment (similar to AnyPay's Pay Protocol).
func SolicitPaymentRequestHandler(c echo.Context) error {
	// Payment ID from path `r/:paymentID`
	paymentID := c.Param("paymentID")

	// TODO: get amount from paymentID key (badger db) and get paymail p2p outputs when creating invoice not here

	usePaymail := false // TODO: use setting or something like that

	var outs []*Output

	if usePaymail == true {
		ref, os, err := ipaymail.GetP2POutputs("jad@moneybutton.com", 10000)
		if err != nil {
			return c.String(http.StatusInternalServerError, "error getting paymail outputs")
		}

		fmt.Println("reference: ", ref)

		ipaymail.ReferencesMap[paymentID] = ref

		// change returned hexString output script into bytes TODO: understand what i wrote

		for _, o := range os {
			out := &Output{
				Amount: o.Satoshis,
				Script: o.Script,
			}

			outs = append(outs, out)
		}
	} else {
		xprv, err := wallet.GetPrivateKey("keyname") // TODO: get from settings
		if err != nil {
			return err
		}

		// TODO: derive new key for each payment!

		pubKey, err := wallet.PubFromXPrv(xprv)
		if err != nil {
			return err
		}

		o, err := bt.NewP2PKHOutputFromPubKeyBytes(pubKey, 10000) // TODO: get amount from invoice
		if err != nil {
			return err
		}

		out := &Output{
			Amount: o.Satoshis,
			Script: o.GetLockingScriptHexString(),
		}

		outs = append(outs, out)
	}

	// endpoint := "localhost:1323" // TODO: get from settings
	endpoint := "178.62.87.120:1323" // TODO: get from settings

	pr := &PaymentRequest{
		Network:             "bitcoin-sv", // TODO: check if bitcoin or bitcoin-sv?
		Outputs:             outs,
		CreationTimestamp:   time.Now().UTC().Unix(),
		ExpirationTimestamp: time.Now().AddDate(0, 0, 1).UTC().Unix(),
		PaymentURL:          fmt.Sprintf("http://%s/v1/payment/%s", endpoint, paymentID),
		Memo:                fmt.Sprintf("Payment request for invoice %s", paymentID),
		MerchantData: &MerchantData{ // TODO: get from settings
			AvatarURL:    "https://bit.ly/3c4iaup",
			MerchantName: "go-payd",
		},
	}

	return c.JSON(http.StatusOK, pr)
}

// PaymentHandler is used to submit a transaction
// in the form of a BIP270 Payment to the
// merchant/receiver.
func PaymentHandler(c echo.Context) error {
	// Payment ID from path `r/:paymentID`
	paymentID := c.Param("paymentID")

	p := new(Payment)
	if err := c.Bind(p); err != nil {
		return err
	}

	pa := &PaymentACK{
		Payment: p,
	}

	usePaymail := false // TODO: use setting or something like that

	if usePaymail == true {

		ref := ipaymail.ReferencesMap[paymentID]

		txid, note, err := ipaymail.SubmitTx("jad@moneybutton.com", p.Transaction, ref)
		if err != nil {
			pa.Error = 1
			pa.Memo = err.Error()
		} else {
			log.Info(txid)
			pa.Error = 0
			pa.Memo = note
		}
	} else {
		// TODO: insert tx into db
	}

	return c.JSON(http.StatusOK, pa)
}

// Validate BIP270 Payment message.
func (s *Payment) Validate() error {
	// check that the tx is valid by decoding the hex string
	_, err := bt.NewTxFromString(s.Transaction)

	return err
}
