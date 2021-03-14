package ppctl

import (
	"context"
	"fmt"

	"github.com/labstack/gommon/log"
	"github.com/libsv/go-bt"
	"github.com/pkg/errors"
	validator "github.com/theflyingcodr/govalidator"
	"github.com/theflyingcodr/lathos"

	gopayd "github.com/libsv/payd"
)

// payment is a layer on top of the payment services of which we currently support:
// * wallet payments, that are handled by the wallet and transmitted to the network
// * paymail payments, that use the paymail protocol for making the payments.
type payment struct {
	store    gopayd.PaymentWriter
	script   gopayd.ScriptKeyReader
	invStore gopayd.InvoiceReaderWriter
	sender   gopayd.PaymentSender
	txrunner gopayd.Transacter
}

// NewPaymentFacade will create and return a new facade to determine between payments to use.
func NewPayment(store gopayd.PaymentWriter, script gopayd.ScriptKeyReader, invStore gopayd.InvoiceReaderWriter, sender gopayd.PaymentSender, txrunner gopayd.Transacter) *payment {
	return &payment{
		store:    store,
		script:   script,
		invStore: invStore,
		sender:   sender,
		txrunner: txrunner,
	}
}

// Create will setup a new payment and return the result.
func (p *payment) CreatePayment(ctx context.Context, args gopayd.CreatePaymentArgs, req gopayd.CreatePayment) (*gopayd.PaymentACK, error) {
	if err := validator.New().Validate("paymentID", validator.NotEmpty(args.PaymentID)); err.Err() != nil {
		return nil, err
	}
	pa := &gopayd.PaymentACK{
		Payment: &req,
		Success: "true",
	}
	// get and attempt to store transaction before processing payment.
	tx, err := bt.NewTxFromString(req.Transaction)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse transaction for paymentID %s", args.PaymentID)
	}
	// TODO: validate the transaction inputs
	outputTotal := uint64(0)
	txos := make([]gopayd.CreateTxo, tx.OutputCount(), tx.OutputCount())
	// iterate outputs and gather the total satoshis for our known outputs
	for i, o := range tx.GetOutputs() {
		sk, err := p.script.ScriptKey(ctx, gopayd.ScriptKeyArgs{LockingScript: o.LockingScript.ToString()})
		if err != nil {
			// script isn't known to us, could be a change utxo, skip and carry on
			if lathos.IsNotFound(err) {
				continue
			}
			return nil, errors.Wrapf(err, "failed to get store output for paymentID %s", args.PaymentID)
		}
		// push new txo onto list for persistence later
		txos = append(txos, gopayd.CreateTxo{
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
	inv, err := p.invStore.Invoice(ctx, gopayd.InvoiceArgs{PaymentID: args.PaymentID})
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
	ctx = p.txrunner.WithTx(ctx)
	// Broadcast the transaction.
	if err := p.sender.Send(ctx, args, req); err != nil {
		log.Error(err)
		pa.Error = 1
		pa.Success = "false"
		pa.Memo = err.Error()
		return pa, nil
	}
	// Store utxos and set invoice to paid.
	if _, err := p.store.StoreUtxos(ctx, gopayd.CreateTransaction{
		PaymentID: inv.PaymentID,
		TxID:      tx.GetTxID(),
		TxHex:     req.Transaction,
		Outputs:   txos,
	}); err != nil {
		pa.Error = 1
		pa.Success = "false"
		pa.Memo = err.Error()
		return nil, errors.Wrapf(err, "failed to complete payment for paymentID %s", args.PaymentID)
	}
	inv, err = p.invStore.Update(ctx, gopayd.UpdateInvoiceArgs{PaymentID: args.PaymentID}, gopayd.UpdateInvoice{
		RefundTo: req.RefundTo,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return pa, errors.WithStack(p.txrunner.Commit(ctx))
}
