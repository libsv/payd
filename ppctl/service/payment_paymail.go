package service

import (
	"context"

	"github.com/labstack/gommon/log"
	validator "github.com/theflyingcodr/govalidator"

	"github.com/libsv/go-payd/ipaymail"
	"github.com/libsv/go-payd/ppctl"
)

type paymentPaymailService struct {
	pmSvc ipaymail.TransactionSubmitter
}

func NewPaymailPaymentService(pmSvc ipaymail.TransactionSubmitter) *paymentPaymailService {
	return &paymentPaymailService{pmSvc: pmSvc}
}

// Create will setup a new payment and return the result.
func (p *paymentPaymailService) CreatePayment(ctx context.Context, args ppctl.CreatePaymentArgs, req ppctl.CreatePayment) (*ppctl.PaymentACK, error) {
	if err := validator.New().Validate("paymentID", validator.NotEmpty(args.PaymentID)); err.Err() != nil {
		return nil, err
	}
	pa := &ppctl.PaymentACK{
		Payment: &req,
	}
	ref := ipaymail.ReferencesMap[args.PaymentID]                                    // TODO - change to a redis call
	txID, note, err := p.pmSvc.SubmitTx("jad@moneybutton.com", req.Transaction, ref) // TODO - dont pay jad, he has enough!
	log.Debug(txID)
	if err != nil {
		pa.Error = 1
		pa.Memo = err.Error()
		return nil, err
	}
	pa.Error = 0
	pa.Memo = note
	return pa, nil
}
