package service

import (
	"context"

	"github.com/labstack/gommon/log"
	"github.com/pkg/errors"
	validator "github.com/theflyingcodr/govalidator"

	"github.com/libsv/go-bt"
	"github.com/libsv/go-payd/errs"
	"github.com/libsv/go-payd/ppctl"
)

type paymentWalletService struct {
	skStorer  ppctl.ScriptKeyStorer
	invStorer ppctl.InvoiceStorer
}

func NewPaymentWalletService(skStore ppctl.ScriptKeyStorer, invStore ppctl.InvoiceStorer) *paymentWalletService {
	return &paymentWalletService{skStorer: skStore, invStorer: invStore}
}

// Create will inform the merchant of a new payment being made,
// this payment will then be transmitted to the network and and acknowledgement sent to the user.
func (p *paymentWalletService) Create(ctx context.Context, args ppctl.CreatePaymentArgs, req ppctl.CreatePayment) (*ppctl.PaymentACK, error) {
	if err := validator.New().Validate("paymentID", validator.NotEmpty(args.PaymentID)); err.Err() != nil {
		return nil, err
	}
	pa := &ppctl.PaymentACK{
		Payment: &req,
		Success: "true",
	}
	// get and attempt to store transaction before processing payment.
	// TODO - is this logic correct?
	tx, err := bt.NewTxFromString(req.Transaction)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse transaction")
	}
	// check outputs
	// if we can sign it, then it's for us, pop into an array
	// increment the total
	// when done, check the value is at least invoice.satoshis
	outputTotal := uint64(0)
	for _, o := range tx.GetOutputs() {
		if _, err := p.skStorer.ScriptKey(ctx, ppctl.ScriptKeyArgs{LockingScript: o.LockingScript.ToString()}); err != nil {
			if errs.IsNotFound(err) {
				continue
			}
			// TODO will need to handle not found errors, if actual error we'll return?
			log.Error(err)
			continue
		}
		outputTotal += o.Satoshis
	}
	inv, err := p.invStorer.Invoice(ctx, ppctl.InvoiceArgs{PaymentID: args.PaymentID})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get invoice to validate output total.")
	}
	// if it doesn't fully pay the invoice, reject it
	if outputTotal < inv.Satoshis {
		pa.Error = 1
		pa.Success = "false"
		pa.Memo = "Outputs do not fully pay invoice for paymentID " + args.PaymentID
		return pa, nil
	}
	// TODO - store outputs

	// TODO - Transmit to network somehow

	return pa, nil
}
