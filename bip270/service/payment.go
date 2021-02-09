package service

import (
	"context"

	"github.com/labstack/gommon/log"
	"github.com/pkg/errors"
	validator "github.com/theflyingcodr/govalidator"

	"github.com/libsv/go-bt"
	"github.com/libsv/go-payd/bip270"
	"github.com/libsv/go-payd/ipaymail"
)

type paymentService struct {
	payMail ipaymail.TransactionSubmitter
	txStore bip270.TransactionStore
}

func NewPaymentService(payMail ipaymail.TransactionSubmitter, txStore bip270.TransactionStore) *paymentService {
	return &paymentService{payMail: payMail, txStore: txStore}
}

// Create will inform the merchant of a new payment being made,
// this payment will then be transmitted to the network and and acknowledgement sent to the user.
func (p *paymentService) Create(ctx context.Context, args bip270.CreatePaymentArgs, req bip270.CreatePayment) (*bip270.PaymentACK, error) {
	if err := validator.New().Validate("paymentID", validator.NotEmpty(args.PaymentID)); err.Err() != nil {
		return nil, err
	}
	pa := &bip270.PaymentACK{
		Payment: &req,
	}
	// get and attempt to store transaction before processing payment.
	// TODO - is this logic correct?
	tx, err := bt.NewTxFromString(req.Transaction)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse transaction")
	}
	if _, err := p.txStore.Create(ctx, tx); err != nil {
		pa.Error = 1
		pa.Memo = "failed to store transaction"
		log.Error(errors.Wrap(err, "failed to store transaction"))
		return pa, nil
	}
	// Transaction persisted safely, now try to transmit the payment.
	if args.UsePaymail {
		if err := p.execPayMail(args.PaymentID, req.Transaction, pa); err != nil {
			log.Error(errors.Wrap(err, "failed to submit payment transaction to paymail"))
			return pa, nil
		}
		return pa, nil
	}
	// TODO - Transmit to network somehow
	return pa, nil
}

func (p *paymentService) execPayMail(paymentID, transaction string, pa *bip270.PaymentACK) error {
	ref := ipaymail.ReferencesMap[paymentID]                                       // TODO - change to a redis call
	txID, note, err := p.payMail.SubmitTx("jad@moneybutton.com", transaction, ref) // TODO - dont pay jad, he has enough!
	log.Debug(txID)
	if err != nil {
		pa.Error = 1
		pa.Memo = err.Error()
		return err
	}
	pa.Error = 0
	pa.Memo = note
	return nil
}
