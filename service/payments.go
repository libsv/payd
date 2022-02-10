package service

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/libsv/go-bc/spv"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-p4"
	"github.com/libsv/go-spvchannels"
	"github.com/libsv/payd/config"
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
	pcSvc         payd.PeerChannelsService
	pcStr         payd.PeerChannelsStore
	pcNotif       payd.PeerChannelsNotifyService
	feeRdr        payd.FeeQuoteReader
	pCfg          *config.PeerChannels
}

// NewPayments will setup and return a payments service.
func NewPayments(l log.Logger, paymentVerify spv.PaymentVerifier, txWtr payd.TransactionWriter, invRdr payd.InvoiceReaderWriter, destRdr payd.DestinationsReader, transacter payd.Transacter, broadcaster payd.BroadcastWriter, feeRdr payd.FeeQuoteReader, callbackWtr payd.ProofCallbackWriter, pcSvc payd.PeerChannelsService, pcNotif payd.PeerChannelsNotifyService, pCfg *config.PeerChannels) payd.PaymentsService {
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
		pcSvc:         pcSvc,
		pcNotif:       pcNotif,
		pCfg:          pCfg,
	}
	return svc
}

// PaymentCreate will validate and store the payment.
func (p *payments) PaymentCreate(ctx context.Context, args payd.PaymentCreateArgs, req p4.Payment) (*p4.PaymentACK, error) {
	if err := validator.New().
		Validate("invoiceID", validator.StrLength(args.InvoiceID, 1, 30)).Err(); err != nil {
		return nil, err
	}
	// Check tx pays enough to cover invoice and that invoice hasn't been paid already
	inv, err := p.invRdr.Invoice(ctx, payd.InvoiceArgs{InvoiceID: args.InvoiceID})
	if err != nil || inv.State == "" {
		return nil, errors.Wrapf(err, "failed to get invoice with ID '%s'", args.InvoiceID)
	}
	if inv.State != payd.StateInvoicePending {
		return nil, lathos.NewErrDuplicate("D001", fmt.Sprintf("payment already received for invoice ID '%s'", args.InvoiceID))
	}
	fq, err := p.feeRdr.FeeQuote(ctx, args.InvoiceID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read fees for payment with id %s", args.InvoiceID)
	}
	if fq.Expired() {
		return nil, lathos.NewErrUnprocessable("E001", "fee quote has expired, please make a new payment request")
	}

	tx, err := p.paymentVerify.VerifyPayment(ctx, req.SPVEnvelope, p.paymentVerifyOpts(inv.SPVRequired, fq)...)
	if err != nil {
		if errors.Is(err, spv.ErrFeePaidNotEnough) {
			return nil, validator.ErrValidation{
				"fees": {
					err.Error(),
				},
			}
		}
		// map error to a validation error
		return nil, validator.ErrValidation{
			"spvEnvelope": {
				err.Error(),
			},
		}
	}

	// get destinations
	oo, err := p.destRdr.Destinations(ctx, payd.DestinationsArgs{InvoiceID: args.InvoiceID})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get destinations with ID '%s'", args.InvoiceID)
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
				return nil, validator.ErrValidation{
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
		return nil, validator.ErrValidation{
			"transaction": {
				"tx does not pay enough to cover invoice, ensure all outputs are included, the correct destinations are used and try again",
			},
		}
	}
	if len(outputs) > 0 {
		return nil, validator.ErrValidation{
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
		return nil, errors.Wrapf(err, "failed to store transaction for invoiceID '%s'", args.InvoiceID)
	}

	// Create peer channel for merkle proof.
	ch, err := p.pcSvc.PeerChannelCreate(ctx, spvchannels.ChannelCreateRequest{
		AccountID:   1,
		PublicWrite: true,
		PublicRead:  true,
		Sequenced:   true,
		Retention: spvchannels.Retention{
			MaxAgeDays: 9999,
			MinAgeDays: 0,
			AutoPrune:  false,
		},
	})
	if err != nil {
		return nil, err
	}

	tokens, err := p.pcSvc.PeerChannelAPITokensCreate(ctx, &payd.PeerChannelAPITokenCreateArgs{
		Role:    "mapi",
		Persist: false,
		Request: spvchannels.TokenCreateRequest{
			AccountID:   1,
			CanRead:     false,
			CanWrite:    true,
			ChannelID:   ch.ID,
			Description: "publishing proofs for " + inv.ID,
		},
	}, &payd.PeerChannelAPITokenCreateArgs{
		Role:    "notification",
		Persist: true,
		Request: spvchannels.TokenCreateRequest{
			AccountID:   1,
			CanRead:     true,
			CanWrite:    false,
			ChannelID:   ch.ID,
			Description: "reading proofs for " + inv.ID,
		},
	}, &payd.PeerChannelAPITokenCreateArgs{
		Role: "notification",
		Request: spvchannels.TokenCreateRequest{
			AccountID:   1,
			CanRead:     true,
			CanWrite:    false,
			ChannelID:   ch.ID,
			Description: "reading proofs for " + inv.ID,
		},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "error creating token for channel %s", ch.ID)
	}

	// Store callbacks if we have any
	if len(req.ProofCallbacks) > 0 {
		if err := p.callbackWtr.ProofCallBacksCreate(ctx, payd.ProofCallbackArgs{InvoiceID: args.InvoiceID}, req.ProofCallbacks); err != nil {
			return nil, errors.Wrapf(err, "failed to store proof callbacks for invoiceID '%s'", args.InvoiceID)
		}
	}

	pc := &p4.PeerChannelData{
		Host:      p.pCfg.Host,
		ChannelID: ch.ID,
		Token:     tokens[2].Token,
	}

	// Broadcast the transaction
	if err := p.broadcaster.Broadcast(ctx, payd.BroadcastArgs{
		InvoiceID:   inv.ID,
		CallbackURL: fmt.Sprintf("http://%s%s", p.pCfg.Host, path.Join("/api/v1/channel/", ch.ID)),
		Token:       "Bearer " + tokens[0].Token,
	}, tx); err != nil {
		// set as failed
		if err := p.txWtr.TransactionUpdateState(ctx, payd.TransactionArgs{TxID: txID}, payd.TransactionStateUpdate{State: payd.StateTxFailed}); err != nil {
			p.l.Error(err, "failed to update tx after failed broadcast")
		}
		return nil, errors.Wrap(err, "failed to broadcast tx")
	}

	if err := p.pcNotif.Subscribe(ctx, &payd.PeerChannel{
		ID:        ch.ID,
		Token:     tokens[1].Token,
		CreatedAt: ch.CreatedAt,
		Type:      payd.PeerChannelHandlerTypeProof,
	}); err != nil {
		p.l.Error(err, "failed to subscribe to proof notifications")
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

	if err := p.transacter.Commit(ctx); err != nil {
		return nil, err
	}

	return &p4.PaymentACK{
		ID:          inv.ID,
		TxID:        tx.TxID(),
		PeerChannel: pc,
	}, nil
}

// Ack will handle an acknowledgement after a payment has been processed.
func (p *payments) Ack(ctx context.Context, args payd.AckArgs, req payd.Ack) error {
	if err := p.txWtr.TransactionUpdateState(ctx, payd.TransactionArgs{TxID: args.TxID}, payd.TransactionStateUpdate{
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
	}); err != nil {
		return errors.Wrap(err, "failed to update transaction state after payment ack")
	}
	if req.Failed {
		return nil
	}

	if err := p.pcStr.PeerChannelCreate(ctx, &payd.PeerChannelCreateArgs{
		PeerChannelAccountID: 0,
		ChannelHost:          args.PeerChannel.Host,
		ChannelID:            args.PeerChannel.ID,
		ChannelType:          args.PeerChannel.Type,
	}); err != nil {
		return errors.Wrapf(err, "failed to store channel '%s'", args.PeerChannel.Host)
	}

	if err := p.pcStr.PeerChannelAPITokenCreate(ctx, &payd.PeerChannelAPITokenStoreArgs{
		Token:                 args.PeerChannel.Token,
		CanRead:               true,
		CanWrite:              false,
		PeerChannelsChannelID: args.PeerChannel.ID,
		Role:                  "notifications",
	}); err != nil {
		return errors.Wrapf(err, "failed to store token '%s' for channel '%s'", args.PeerChannel.Token, args.PeerChannel.ID)
	}

	return errors.Wrapf(p.pcNotif.Subscribe(ctx, args.PeerChannel), "failed to subscribe to channel '%s'", args.PeerChannel.ID)
}

func (p *payments) paymentVerifyOpts(verifySPV bool, fq *bt.FeeQuote) []spv.VerifyOpt {
	if verifySPV {
		return []spv.VerifyOpt{spv.VerifyFees(fq), spv.VerifySPV()}
	}
	return []spv.VerifyOpt{spv.NoVerifySPV(), spv.NoVerifyFees()}
}
