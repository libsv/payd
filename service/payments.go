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

type payments struct {
	paymentVerify spv.PaymentVerifier
	txWtr         payd.TransactionWriter
	invRdr        payd.InvoiceReader
	destRdr       payd.DestinationsReader
	transacter    payd.Transacter
	callbackWtr   payd.ProofCallbackWriter
	broadcaster   payd.BroadcastWriter
	feeRdr        payd.FeeReader
}

// NewPayments will setup and return a payments service.
func NewPayments(paymentVerify spv.PaymentVerifier, txWtr payd.TransactionWriter, invRdr payd.InvoiceReader, destRdr payd.DestinationsReader, transacter payd.Transacter, broadcaster payd.BroadcastWriter, feeRdr payd.FeeReader, callbackWtr payd.ProofCallbackWriter) *payments {
	svc := &payments{
		paymentVerify: paymentVerify,
		invRdr:        invRdr,
		destRdr:       destRdr,
		transacter:    transacter,
		txWtr:         txWtr,
		broadcaster:   broadcaster,
		feeRdr:        feeRdr,
		callbackWtr:   callbackWtr,
	}
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
	fq, err := p.feeRdr.Fees(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to read fees for payment with id %s", req.InvoiceID)
	}

	tx, err := p.paymentVerify.VerifyPayment(ctx, req.SPVEnvelope, p.paymentVerifyOpts(inv.SPVRequired, fq)...)
	if err != nil {
		if errors.Is(err, spv.ErrFeePaidNotEnough) {
			return validator.ErrValidation{
				"fees": {
					err.Error(),
				},
			}
		}
		// map error to a validation error
		return validator.ErrValidation{
			"spvEnvelope": {
				err.Error(),
			},
		}
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
	// TODO: simple dust limit check
	for i, o := range tx.Outputs {
		if output, ok := outputs[o.LockingScript.String()]; ok {
			if o.Satoshis != output.Satoshis {
				return validator.ErrValidation{
					"tx.outputs": {
						"output satoshis do not match requested amount",
					},
				}
			}

			total += output.Satoshis
			txos = append(txos, &payd.TxoCreate{
				Outpoint:      fmt.Sprintf("%s%d", txID, i),
				DestinationID: output.ID,
				TxID:          txID,
				Vout:          uint64(i),
			})
			delete(outputs, o.LockingScript.String())
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
	if len(outputs) > 0 {
		return validator.ErrValidation{
			"tx.outputs": {
				fmt.Sprintf("expected '%d' outputs, received '%d', ensure all destinations are supplied", len(oo), len(tx.Outputs)),
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
	if err := p.broadcaster.Broadcast(ctx, payd.BroadcastArgs{InvoiceID: inv.ID}, tx); err != nil {
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

func (p *payments) paymentVerifyOpts(verifySPV bool, fq *bt.FeeQuote) []spv.VerifyOpt {
	opts := []spv.VerifyOpt{spv.VerifyFees(fq)}
	if !verifySPV {
		opts = append(opts, spv.NoVerifySPV())
	}
	return opts
}
