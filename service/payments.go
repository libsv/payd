package service

import (
	"context"
	"fmt"
	"time"

	"github.com/libsv/go-bc/spv"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-p4"
	"github.com/libsv/payd/log"
	"github.com/pkg/errors"
	validator "github.com/theflyingcodr/govalidator"
	lathos "github.com/theflyingcodr/lathos/errs"
	"gopkg.in/guregu/null.v3"

	"github.com/libsv/payd"
)

type payments struct {
	l             log.Logger
	paymentVerify spv.PaymentVerifier
	txWtr         payd.TransactionWriter
	invRdr        payd.InvoiceReaderWriter
	destRdr       payd.DestinationsReader
	transacter    payd.Transacter
	callbackWtr   payd.ProofCallbackWriter
	broadcaster   payd.BroadcastWriter
	feeRdr        payd.FeeReader
}

// NewPayments will setup and return a payments service.
func NewPayments(l log.Logger, paymentVerify spv.PaymentVerifier, txWtr payd.TransactionWriter, invRdr payd.InvoiceReaderWriter, destRdr payd.DestinationsReader, transacter payd.Transacter, broadcaster payd.BroadcastWriter, feeRdr payd.FeeReader, callbackWtr payd.ProofCallbackWriter) *payments {
	svc := &payments{
		l:             l,
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
func (p *payments) PaymentCreate(ctx context.Context, args payd.PaymentCreateArgs, req p4.Payment) error {
	if err := validator.New().
		Validate("invoiceID", validator.StrLength(args.InvoiceID, 1, 30)).Err(); err != nil {
		return err
	}
	// Check tx pays enough to cover invoice and that invoice hasn't been paid already
	inv, err := p.invRdr.Invoice(ctx, payd.InvoiceArgs{InvoiceID: args.InvoiceID})
	if err != nil {
		return errors.Wrapf(err, "failed to get invoice with ID '%s'", args.InvoiceID)
	}
	if inv.State != payd.StateInvoicePending {
		return lathos.NewErrDuplicate("D001", fmt.Sprintf("payment already received for invoice ID '%s'", args.InvoiceID))
	}
	fq, err := p.feeRdr.Fees(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to read fees for payment with id %s", args.InvoiceID)
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
	oo, err := p.destRdr.Destinations(ctx, payd.DestinationsArgs{InvoiceID: args.InvoiceID})
	if err != nil {
		return errors.Wrapf(err, "failed to get destinations with ID '%s'", args.InvoiceID)
	}
	// gather all outputs and add to a lookup map
	var total uint64
	outputs := map[string]payd.Output{}
	for _, o := range oo {
		outputs[o.LockingScript.String()] = o
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
		InvoiceID: args.InvoiceID,
		TxID:      txID,
		RefundTo:  null.StringFromPtr(req.RefundTo),
		TxHex:     req.SPVEnvelope.RawTx,
		Outputs:   txos,
	}); err != nil {
		return errors.Wrapf(err, "failed to store transaction for invoiceID '%s'", args.InvoiceID)
	}
	// Store callbacks if we have any
	if len(req.ProofCallbacks) > 0 {
		if err := p.callbackWtr.ProofCallBacksCreate(ctx, payd.ProofCallbackArgs{InvoiceID: args.InvoiceID}, req.ProofCallbacks); err != nil {
			return errors.Wrapf(err, "failed to store proof callbacks for invoiceID '%s'", args.InvoiceID)
		}
	}
	// Broadcast the transaction
	if err := p.broadcaster.Broadcast(ctx, payd.BroadcastArgs{InvoiceID: inv.ID}, tx); err != nil {
		// set as failed
		if err := p.txWtr.TransactionUpdateState(ctx, payd.TransactionArgs{TxID: txID}, payd.TransactionStateUpdate{State: payd.StateTxFailed}); err != nil {
			p.l.Error(err, "failed to update tx after failed broadcast")
		}
		return errors.Wrap(err, "failed to broadcast tx")
	}

	// Update tx state to broadcast
	// Just logging errors here as I don't want to roll back tx now tx is broadcast.
	if err := p.txWtr.TransactionUpdateState(ctx, payd.TransactionArgs{TxID: txID}, payd.TransactionStateUpdate{State: payd.StateTxBroadcast}); err != nil {
		p.l.Error(err, "failed to update tx to broadcast state")
	}
	// set invoice as paid
	if _, err := p.invRdr.InvoiceUpdate(ctx, payd.InvoiceUpdateArgs{InvoiceID: args.InvoiceID}, payd.InvoiceUpdatePaid{
		PaymentReceivedAt: time.Now().UTC(),
		RefundTo: func() string {
			if req.RefundTo == nil {
				return ""
			}
			return *req.RefundTo
		}(),
	}); err != nil {
		p.l.Error(err, "failed to update invoice to paid")
	}

	return p.transacter.Commit(ctx)
}

// Ack will handle an acknowledgement after a payment has been processed.
func (p *payments) Ack(ctx context.Context, args payd.AckArgs, req payd.Ack) error {
	err := p.txWtr.TransactionUpdateState(ctx, payd.TransactionArgs{TxID: args.TxID}, payd.TransactionStateUpdate{
		State: func() payd.TxState {
			if req.Failed {
				return payd.StateTxFailed
			}
			return payd.StateTxBroadcast
		}(),
		FailReason: func() null.String {
			if req.Reason == "" {
				return null.String{}
			}
			return null.StringFrom(req.Reason)
		}(),
	})
	return errors.Wrap(err, "failed to update transaction state after payment ack")
}

func (p *payments) paymentVerifyOpts(verifySPV bool, fq *bt.FeeQuote) []spv.VerifyOpt {
	if verifySPV {
		return []spv.VerifyOpt{spv.VerifyFees(fq), spv.VerifySPV()}
	}
	return []spv.VerifyOpt{spv.NoVerifySPV(), spv.NoVerifyFees()}
}
