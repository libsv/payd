package ppctl

import (
	"context"
	"fmt"

	"github.com/labstack/gommon/log"
	gopayd "github.com/libsv/payd"
	"github.com/pkg/errors"
	validator "github.com/theflyingcodr/govalidator"
	"github.com/theflyingcodr/lathos"

	"github.com/libsv/go-bt"
)

type paymentSender interface {
}

type paymentWalletService struct {
	store       gopayd.PaymentReaderWriter
	broadcaster gopayd.TransactionBroadcaster
}

func NewPaymentWalletService(store gopayd.PaymentReaderWriter, broadcaster gopayd.TransactionBroadcaster) *paymentWalletService {
	return &paymentWalletService{store: store, broadcaster: broadcaster}
}

// CreatePayment will inform the merchant of a new payment being made,
// this payment will then be transmitted to the network and and acknowledgement sent to the user.
func (p *paymentWalletService) CreatePayment(ctx context.Context, args gopayd.CreatePaymentArgs, req gopayd.CreatePayment) (*gopayd.PaymentACK, error) {
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
		sk, err := p.store.ScriptKey(ctx, gopayd.ScriptKeyArgs{LockingScript: o.LockingScript.ToString()})
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
	inv, err := p.store.Invoice(ctx, gopayd.InvoiceArgs{PaymentID: args.PaymentID})
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
	// Broadcast the transaction.
	if err := p.broadcaster.Broadcast(ctx, gopayd.BroadcastTransaction{TXHex: req.Transaction}); err != nil {
		log.Error(err)
		pa.Error = 1
		pa.Success = "false"
		pa.Memo = err.Error()
		return pa, nil
	}
	// Store utxos and set invoice to paid.
	if _, err := p.store.CompletePayment(ctx, gopayd.CreateTransaction{
		PaymentID: inv.PaymentID,
		TxID:      tx.GetTxID(),
		TxHex:     req.Transaction,
		Outputs:   txos,
	}); err != nil {
		return nil, errors.Wrapf(err, "failed to complete payment for paymentID %s", args.PaymentID)
	}
	return pa, nil
}
