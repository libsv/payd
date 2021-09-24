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

	"github.com/libsv/payd"
)

type paymentValidatorFunc func(ctx context.Context, req payd.PaymentCreate) (*bt.Tx, error)

type payments struct {
	paymentVerify spv.PaymentVerifier
	txWtr         payd.TransactionWriter
	invRdr        payd.InvoiceReader
	destRdr       payd.DestinationsReader
	transacter    payd.Transacter
	callbackWtr   payd.ProofCallbackWriter
	broadcaster   payd.BroadcastWriter
	validator     map[bool]paymentValidatorFunc
}

// NewPayments will setup and return a payments service.
func NewPayments(paymentVerify spv.PaymentVerifier, txWtr payd.TransactionWriter, invRdr payd.InvoiceReader, destRdr payd.DestinationsReader, transacter payd.Transacter, broadcaster payd.BroadcastWriter, callbackWtr payd.ProofCallbackWriter) *payments {
	svc := &payments{
		paymentVerify: paymentVerify,
		invRdr:        invRdr,
		destRdr:       destRdr,
		transacter:    transacter,
		txWtr:         txWtr,
		broadcaster:   broadcaster,
		callbackWtr:   callbackWtr,
		validator:     map[bool]paymentValidatorFunc{},
	}
	// setup validators for spv and rawTX
	svc.validator[true] = svc.spvHandler
	svc.validator[false] = svc.rawTxHandler

	return svc
}

// PaymentCreate will validate and store the payment.
func (p *payments) PaymentCreate(ctx context.Context, req payd.PaymentCreate) error {
	if err := validator.New().
		Validate("invoiceID", validator.StrLength(req.InvoiceID, 1, 30)).Err(); err != nil {
		return err
	}
	// Check tx pays enough to cover invoice and that invoice hasn't been paid already
	inv, err := p.invRdr.Invoice(ctx, payd.InvoiceArgs{InvoiceID: req.InvoiceID})
	if err != nil {
		return errors.Wrapf(err, "failed to get invoice with ID '%s'", req.InvoiceID)
	}
	if inv.State != payd.StateInvoicePending {
		return lathos.NewErrDuplicate("D001", fmt.Sprintf("payment already received for invoice ID '%s'", req.InvoiceID))
	}
	// validate request tx or envelope and return tx.
	tx, err := p.validator[inv.SPVRequired](ctx, req)
	if err != nil {
		return errors.WithStack(err)
	}
	// get destinations
	oo, err := p.destRdr.Destinations(ctx, payd.DestinationsArgs{InvoiceID: req.InvoiceID})
	if err != nil {
		return errors.Wrapf(err, "failed to get destinations with ID '%s'", req.InvoiceID)
	}
	// gather all outputs and add to a lookup map
	var total uint64
	outputs := map[string]payd.Output{}
	for _, o := range oo {
		outputs[o.LockingScript] = o
	}
	txos := make([]*payd.TxoCreate, 0)
	// get total of outputs that we know about
	txID := tx.TxID()
	for i, o := range tx.Outputs {
		if output, ok := outputs[o.LockingScript.String()]; ok {
			total += output.Satoshis
			txos = append(txos, &payd.TxoCreate{
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
	if err := p.txWtr.TransactionCreate(ctx, payd.TransactionCreate{
		InvoiceID: req.InvoiceID,
		TxID:      txID,
		TxHex: func() string {
			if inv.SPVRequired {
				return req.SPVEnvelope.RawTx
			}
			return req.RawTX.ValueOrZero()
		}(),
		Outputs: txos,
	}); err != nil {
		return errors.Wrapf(err, "failed to store transaction for invoiceID '%s'", req.InvoiceID)
	}
	// Store callbacks if we have any
	if len(req.ProofCallbacks) > 0 {
		if err := p.callbackWtr.ProofCallBacksCreate(ctx, payd.ProofCallbackArgs{InvoiceID: req.InvoiceID}, req.ProofCallbacks); err != nil {
			return errors.Wrapf(err, "failed to store proof callbacks for invoiceID '%s'", req.InvoiceID)
		}
	}
	// Broadcast the transaction
	if err := p.broadcaster.Broadcast(ctx, tx); err != nil {
		// set as failed
		if err := p.txWtr.TransactionUpdateState(ctx, payd.TransactionArgs{TxID: txID}, payd.TransactionStateUpdate{State: payd.StateTxFailed}); err != nil {
			log.Error(err)
		}
		return errors.Wrap(err, "failed to broadcast tx")
	}

	// Update tx state to broadcast
	if err := p.txWtr.TransactionUpdateState(ctx, payd.TransactionArgs{TxID: txID}, payd.TransactionStateUpdate{State: payd.StateTxBroadcast}); err != nil {
		log.Error(err)
	}

	return p.transacter.Commit(ctx)
}

func (p *payments) spvHandler(ctx context.Context, req payd.PaymentCreate) (*bt.Tx, error) {
	if err := req.Validate(true); err != nil {
		return nil, err
	}
	ok, err := p.paymentVerify.VerifyPayment(ctx, req.SPVEnvelope)
	if err != nil {
		// map error to a validation error
		return nil, validator.ErrValidation{
			"spvEnvelope": {
				err.Error(),
			},
		}
	}
	if !ok {
		// map error to a validation error
		return nil, validator.ErrValidation{
			"spvEnvelope": {
				"payment envelope is not valid",
			},
		}
	}
	// validate outputs match invoice
	// ensure tx pays enough fees.
	tx, err := bt.NewTxFromString(req.SPVEnvelope.RawTx)
	if err != nil {
		// convert to validation error
		if err := validator.New().Validate("rawTx", func() error {
			return errors.Wrap(err, "invalid transaction received")
		}).Err(); err != nil {
			return nil, err
		}
	}
	return tx, nil
}

func (p *payments) rawTxHandler(ctx context.Context, req payd.PaymentCreate) (*bt.Tx, error) {
	if err := validator.New().Validate("rawTx", validator.NotEmpty(req.RawTX.ValueOrZero())).Err(); err != nil {
		return nil, err
	}
	tx, err := bt.NewTxFromString(req.RawTX.ValueOrZero())
	if err != nil {
		// convert to validation error
		if err := validator.New().Validate("rawTx", func() error {
			return errors.Wrap(err, "invalid transaction received")
		}).Err(); err != nil {
			return nil, err
		}
	}
	return tx, nil
}
