package bip270

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/libsv/go-bt"
	"github.com/libsv/go-payd/paymail"
)

// SolicitPaymentRequestHandler is used to obtain a BIP270
// PaymentRequest from the merchant/receiver of the
// payment (similar to AnyPay's Pay Protocol).
func SolicitPaymentRequestHandler(c echo.Context) error {
	// Payment ID from path `r/:paymentID`
	paymentID := c.Param("paymentID")

	// TODO: get amount from paymentID key (badger db) and get paymail p2p outputs when creating invoice not here

	ref, os, err := paymail.GetP2POutputs("jad@moneybutton.com", 10000)
	if err != nil {
		return c.String(http.StatusInternalServerError, "error getting paymail outputs")
	}

	fmt.Println("reference: ", ref)

	paymail.ReferencesMap[paymentID] = ref

	// change returned hexString output script into bytes
	var outs []*Output

	for _, o := range os {
		out := &Output{
			Amount: o.Satoshis,
			Script: o.Script,
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
		MerchantData: &MerchantData{
			AvatarURL:    "https://bit.ly/3c4iaup",
			MerchantName: "Medium Coffee 12oz",
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

	ref := paymail.ReferencesMap[paymentID]
	pa := &PaymentACK{
		Payment: p,
	}

	txid, note, err := paymail.SubmitTx("jad@moneybutton.com", p.Transaction, ref)
	if err != nil {
		pa.Error = 1
		pa.Memo = err.Error()
	} else {
		log.Info(txid)
		pa.Error = 0
		pa.Memo = note
	}

	return c.JSON(http.StatusOK, pa)
}

// Validate BIP270 Payment message.
func (s *Payment) Validate() error {
	// check that the tx is valid by decoding the hex string
	_, err := bt.NewTxFromString(s.Transaction)

	return err
}
