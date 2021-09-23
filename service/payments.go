package service

import (
	"context"
	"fmt"

	"github.com/labstack/gommon/log"
	"github.com/libsv/go-bc/spv"
	"github.com/libsv/go-bt/v2"
	"github.com/pkg/errors"
	validator "github.com/theflyingcodr/govalidator"
	lathos "github.com/theflyingcodr/lathos/errs"

	gopayd "github.com/libsv/payd"
)

type payments struct {
	paymentVerify spv.PaymentVerifier
	txWtr         gopayd.TransactionWriter
	invRdr        gopayd.InvoiceReader
	destRdr       gopayd.DestinationsReader
	transacter    gopayd.Transacter
	callbackWtr   gopayd.ProofCallbackWriter
	broadcaster   gopayd.BroadcastWriter
}

// NewPayments will setup and return a payments service.
func NewPayments(paymentVerify spv.PaymentVerifier, txWtr gopayd.TransactionWriter, invRdr gopayd.InvoiceReader, destRdr gopayd.DestinationsReader, transacter gopayd.Transacter, broadcaster gopayd.BroadcastWriter, callbackWtr gopayd.ProofCallbackWriter) *payments {
	return &payments{
		paymentVerify: paymentVerify,
		invRdr:        invRdr,
		destRdr:       destRdr,
		transacter:    transacter,
		txWtr:         txWtr,
		broadcaster:   broadcaster,
		callbackWtr:   callbackWtr,
	}
}

// PaymentCreate will validate and store the payment.
func (p *payments) PaymentCreate(ctx context.Context, req gopayd.PaymentCreate) error {
	if err := req.Validate(); err != nil {
		return err
	}
	// Check tx pays enough to cover invoice and that invoice hasn't been paid already
	inv, err := p.invRdr.Invoice(ctx, gopayd.InvoiceArgs{InvoiceID: req.InvoiceID})
	if err != nil {
		return errors.Wrapf(err, "failed to get invoice with ID '%s'", req.InvoiceID)
	}
	if inv.State != gopayd.StateInvoicePending {
		return lathos.NewErrDuplicate("D001", fmt.Sprintf("payment already received for invoice ID '%s'", req.InvoiceID))
	}
	ok, err := p.paymentVerify.VerifyPayment(ctx, req.SPVEnvelope)
	if err != nil {
		// map error to a validation error
		return validator.ErrValidation{
			"spvEnvelope": {
				err.Error(),
			},
		}
	}
	if !ok {
		// map error to a validation error
		return validator.ErrValidation{
			"spvEnvelope": {
				"payment envelope is not valid",
			},
		}
	}
	// validate outputs match invoice
	// ensure tx pays enough fees.
	tx, err := bt.NewTxFromString(req.SPVEnvelope.RawTx)
	if err != nil {
		return errors.Wrap(err, "failed to read transaction")
	}
	// get destinations
	oo, err := p.destRdr.Destinations(ctx, gopayd.DestinationsArgs{InvoiceID: req.InvoiceID})
	if err != nil {
		return errors.Wrapf(err, "failed to get destinations with ID '%s'", req.InvoiceID)
	}
	// gather all outputs and add to a lookup map
	var total uint64
	outputs := map[string]gopayd.Output{}
	for _, o := range oo {
		outputs[o.LockingScript] = o
	}
	txos := make([]*gopayd.TxoCreate, 0)
	// get total of outputs that we know about
	txID := tx.TxID()
	for i, o := range tx.Outputs {
		if output, ok := outputs[o.LockingScript.String()]; ok {
			total += output.Satoshis
			txos = append(txos, &gopayd.TxoCreate{
				Outpoint:      fmt.Sprintf("%s%d", txID, i),
				DestinationID: output.ID,
				TxID:          txID,
				Vout:          uint64(i),
			})
		}
	}
	// fail if tx doesn't pay invoice in full
	if total < inv.Satoshis {
		return validator.ErrValidation{
			"transaction": {
				"tx does not pay enough to cover invoice, ensure all outputs are included, the correct destinations are used and try again",
			},
		}
	}
	ctx = p.transacter.WithTx(ctx)
	defer func() {
		_ = p.transacter.Rollback(ctx)
	}()
	// Store tx
	if err := p.txWtr.TransactionCreate(ctx, gopayd.TransactionCreate{
		InvoiceID: req.InvoiceID,
		TxID:      txID,
		TxHex:     req.SPVEnvelope.RawTx,
		Outputs:   txos,
	}); err != nil {
		return errors.Wrapf(err, "failed to store transaction for invoiceID '%s'", req.InvoiceID)
	}
	// Store callbacks if we have any
	if len(req.ProofCallbacks) > 0 {
		if err := p.callbackWtr.ProofCallBacksCreate(ctx, gopayd.ProofCallbackArgs{InvoiceID: req.InvoiceID}, req.ProofCallbacks); err != nil {
			return errors.Wrapf(err, "failed to store proof callbacks for invoiceID '%s'", req.InvoiceID)
		}
	}
	// Broadcast the transaction
	if err := p.broadcaster.Broadcast(ctx, tx); err != nil {
		// set as failed
		if err := p.txWtr.TransactionUpdateState(ctx, gopayd.TransactionArgs{TxID: txID}, gopayd.TransactionStateUpdate{State: gopayd.StateTxFailed}); err != nil {
			log.Error(err)
		}
		return errors.Wrap(err, "failed to broadcast tx")
	}

	// Update tx state to broadcast
	if err := p.txWtr.TransactionUpdateState(ctx, gopayd.TransactionArgs{TxID: txID}, gopayd.TransactionStateUpdate{State: gopayd.StateTxBroadcast}); err != nil {
		log.Error(err)
	}

	return p.transacter.Commit(ctx)
}