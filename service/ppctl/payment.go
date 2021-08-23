package ppctl

import (
	"context"
	"fmt"

	"github.com/labstack/gommon/log"
	"github.com/libsv/go-bt/v2"
	"github.com/pkg/errors"
	validator "github.com/theflyingcodr/govalidator"
	"github.com/theflyingcodr/lathos"
	"gopkg.in/guregu/null.v3"

	gopayd "github.com/libsv/payd"
)

const (
	paymentFailed  = "false"
	paymentSuccess = "true"
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

// NewPayment will create and return a new payment service.
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
	}
	// get and attempt to store transaction before processing payment.
	tx, err := bt.NewTxFromString(req.Transaction)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse transaction for paymentID %s", args.PaymentID)
	}
	// TODO: validate the transaction inputs
	outputTotal := uint64(0)
	txos := make([]gopayd.CreateTxo, 0)
	// iterate outputs and gather the total satoshis for our known outputs
	for i, o := range tx.Outputs {
		sk, err := p.script.ScriptKey(ctx, gopayd.ScriptKeyArgs{LockingScript: o.LockingScript.String()})
		if err != nil {
			// script isn't known to us, could be a change utxo, skip and carry on
			if lathos.IsNotFound(err) {
				continue
			}
			return nil, errors.Wrapf(err, "failed to get store output for paymentID %s", args.PaymentID)
		}
		// push new txo onto list for persistence later
		txos = append(txos, gopayd.CreateTxo{
			Outpoint:       fmt.Sprintf("%s%d", tx.TxID(), i),
			TxID:           tx.TxID(),
			Vout:           i,
			KeyName:        null.StringFrom(keyname),
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
		pa.Memo = "Outputs do not fully pay invoice for paymentID " + args.PaymentID
		return pa, nil
	}
	ctx = p.txrunner.WithTx(ctx)
	// Store utxos and set invoice to paid.
	if _, err = p.store.StoreUtxos(ctx, gopayd.CreateTransaction{
		PaymentID: inv.PaymentID,
		TxID:      tx.TxID(),
		TxHex:     req.Transaction,
		Outputs:   txos,
	}); err != nil {
		log.Error(err)
		pa.Error = 1
		pa.Memo = err.Error()
		return nil, errors.Wrapf(err, "failed to complete payment for paymentID %s", args.PaymentID)
	}
	if _, err := p.invStore.Update(ctx, gopayd.InvoiceUpdateArgs{PaymentID: args.PaymentID}, gopayd.InvoiceUpdate{
		RefundTo: req.RefundTo,
	}); err != nil {
		log.Error(err)
		pa.Error = 1
		pa.Memo = err.Error()
		return nil, errors.Wrapf(err, "failed to update invoice payment for paymentID %s", args.PaymentID)
	}
	// Broadcast the transaction.
	if err := p.sender.Send(ctx, gopayd.SendTransactionArgs{TxID: tx.TxID()}, req); err != nil {
		log.Error(err)
		pa.Error = 1
		pa.Memo = err.Error()
		return pa, errors.Wrapf(err, "failed to send payment for paymentID %s", args.PaymentID)
	}
	return pa, errors.WithStack(p.txrunner.Commit(ctx))
}
