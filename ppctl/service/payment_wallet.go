package service

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	validator "github.com/theflyingcodr/govalidator"

	"github.com/libsv/go-bt"

	"github.com/libsv/go-payd/errs"
	"github.com/libsv/go-payd/ppctl"
)

type paymentWalletService struct {
	skStorer  ppctl.ScriptKeyStorer
	invStorer ppctl.InvoiceStorer
	txStore   ppctl.TransactionStorer
}

func NewPaymentWalletService(skStore ppctl.ScriptKeyStorer, invStore ppctl.InvoiceStorer, txStore ppctl.TransactionStorer) *paymentWalletService {
	return &paymentWalletService{skStorer: skStore, invStorer: invStore, txStore: txStore}
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
	tx, err := bt.NewTxFromString(req.Transaction)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse transaction for paymentID %s", args.PaymentID)
	}
	// TODO - validate the transaction inputs
	outputTotal := uint64(0)
	txos := make([]ppctl.CreateTxo, tx.OutputCount(), tx.OutputCount())
	// iterate outputs and gather the total satoshis for our known outputs
	for i, o := range tx.GetOutputs() {
		sk, err := p.skStorer.ScriptKey(ctx, ppctl.ScriptKeyArgs{LockingScript: o.LockingScript.ToString()})
		if err != nil {
			// script isn't known to us, could be a change utxo, skip and carry on
			if errs.IsNotFound(err) {
				continue
			}
			return nil, errors.Wrapf(err, "failed to get store output for paymentID %s", args.PaymentID)
		}
		// push new txo onto list for persistence later
		txos = append(txos, ppctl.CreateTxo{
			Outpoint:       fmt.Sprintf("%s%d", tx.GetTxID(), i),
			TxID:           tx.GetTxID(),
			Vout:           i,
			KeyName:        keyname,
			DerivationPath: sk.DerivationPath,
			LockingScript:  sk.LockingScript,
			Satoshis:       o.Satoshis,
		})
		outputTotal += o.Satoshis
	}
	// get the invoice for the paymentID to check total satoshis required.
	inv, err := p.invStorer.Invoice(ctx, ppctl.InvoiceArgs{PaymentID: args.PaymentID})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get invoice to validate output total for paymentID %s.", args.PaymentID)
	}
	// if it doesn't fully pay the invoice, reject it
	if outputTotal < inv.Satoshis {
		pa.Error = 1
		pa.Success = "false"
		pa.Memo = "Outputs do not fully pay invoice for paymentID " + args.PaymentID
		return pa, nil
	}
	// TODO - Transmit to network somehow

	if _, err := p.txStore.Create(ctx, ppctl.CreateTransaction{
		PaymentID: inv.PaymentID,
		TxID:      tx.GetTxID(),
		TxHex:     req.Transaction,
		Outputs:   txos,
	}); err != nil {
		return nil, errors.Wrapf(err, "failed to persist transaction outputs for paymentID %s", args.PaymentID)
	}
	return pa, nil
}
