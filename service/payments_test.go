package service_test

import (
	"context"
	"testing"

	"github.com/libsv/go-bc/spv"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/payd"
	"github.com/libsv/payd/mocks"
	"github.com/libsv/payd/service"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v3"
)

func TestPaymentsService_PaymentCreate(t *testing.T) {
	fq := bt.NewFeeQuote()
	tests := map[string]struct {
		invoiceFunc             func(context.Context, payd.InvoiceArgs) (*payd.Invoice, error)
		feesFunc                func(context.Context) (*bt.FeeQuote, error)
		verifyPaymentFunc       func(context.Context, *spv.Envelope, ...spv.VerifyOpt) (*bt.Tx, error)
		destinationsFunc        func(context.Context, payd.DestinationsArgs) ([]payd.Output, error)
		txCreateFunc            func(context.Context, payd.TransactionCreate) error
		proofCallbackCreateFunc func(context.Context, payd.ProofCallbackArgs, map[string]payd.ProofCallback) error
		broadcastFunc           func(context.Context, *bt.Tx) error
		txUpdateStateFunc       func(context.Context, payd.TransactionArgs, payd.TransactionStateUpdate) error
		commitFunc              func(context.Context) error
		req                     payd.PaymentCreate
		expVerifyOpts           []spv.VerifyOpt
		expRawTx                string
		expTxState              payd.TxState
		expErr                  error
	}{
		"successful create": {
			invoiceFunc: func(ctx context.Context, args payd.InvoiceArgs) (*payd.Invoice, error) {
				return &payd.Invoice{ID: args.InvoiceID, State: payd.StateInvoicePending}, nil
			},
			feesFunc: func(ctx context.Context) (*bt.FeeQuote, error) {
				return fq, nil
			},
			verifyPaymentFunc: func(context.Context, *spv.Envelope, ...spv.VerifyOpt) (*bt.Tx, error) {
				return bt.NewTxFromString("010000000001e8030000000000001976a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac00000000")
			},
			destinationsFunc: func(context.Context, payd.DestinationsArgs) ([]payd.Output, error) {
				return []payd.Output{{
					LockingScript:  "76a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac",
					DerivationPath: "2147483648/2147483648/2147483648",
					Satoshis:       1000,
					State:          "pending",
				}}, nil
			},
			txCreateFunc: func(context.Context, payd.TransactionCreate) error {
				return nil
			},
			proofCallbackCreateFunc: func(context.Context, payd.ProofCallbackArgs, map[string]payd.ProofCallback) error {
				return nil
			},
			broadcastFunc: func(context.Context, *bt.Tx) error {
				return nil
			},
			txUpdateStateFunc: func(context.Context, payd.TransactionArgs, payd.TransactionStateUpdate) error {
				return nil
			},
			commitFunc: func(context.Context) error {
				return nil
			},
			req: payd.PaymentCreate{
				InvoiceID: "abc123",
				RawTX:     null.StringFrom("010000000001e8030000000000001976a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac00000000"),
			},
			expVerifyOpts: []spv.VerifyOpt{spv.VerifyFees(fq), spv.NoVerifySPV()},
			expRawTx:      "010000000001e8030000000000001976a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac00000000",
			expTxState:    payd.StateTxBroadcast,
		},
		"successful create with spv verification": {
			invoiceFunc: func(ctx context.Context, args payd.InvoiceArgs) (*payd.Invoice, error) {
				return &payd.Invoice{ID: args.InvoiceID, State: payd.StateInvoicePending, SPVRequired: true}, nil
			},
			feesFunc: func(ctx context.Context) (*bt.FeeQuote, error) {
				return fq, nil
			},
			verifyPaymentFunc: func(context.Context, *spv.Envelope, ...spv.VerifyOpt) (*bt.Tx, error) {
				return bt.NewTxFromString("010000000001e8030000000000001976a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac00000000")
			},
			destinationsFunc: func(context.Context, payd.DestinationsArgs) ([]payd.Output, error) {
				return []payd.Output{{
					LockingScript:  "76a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac",
					DerivationPath: "2147483648/2147483648/2147483648",
					Satoshis:       1000,
					State:          "pending",
				}}, nil
			},
			txCreateFunc: func(context.Context, payd.TransactionCreate) error {
				return nil
			},
			proofCallbackCreateFunc: func(context.Context, payd.ProofCallbackArgs, map[string]payd.ProofCallback) error {
				return nil
			},
			broadcastFunc: func(context.Context, *bt.Tx) error {
				return nil
			},
			txUpdateStateFunc: func(context.Context, payd.TransactionArgs, payd.TransactionStateUpdate) error {
				return nil
			},
			commitFunc: func(context.Context) error {
				return nil
			},
			req: payd.PaymentCreate{
				InvoiceID: "abc123",
				SPVEnvelope: &spv.Envelope{
					RawTx: "010000000001e8030000000000001976a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac00000000",
				},
			},
			expVerifyOpts: []spv.VerifyOpt{spv.VerifyFees(fq)},
			expRawTx:      "010000000001e8030000000000001976a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac00000000",
			expTxState:    payd.StateTxBroadcast,
		},
		"invalid request is rejected": {
			req:    payd.PaymentCreate{},
			expErr: errors.New("[invoiceID: value must be between 1 and 30 characters]"),
		},
		"invoice error is handled": {
			invoiceFunc: func(ctx context.Context, args payd.InvoiceArgs) (*payd.Invoice, error) {
				return nil, errors.New("no invoice 4 u")
			},
			req:    payd.PaymentCreate{InvoiceID: "abc123"},
			expErr: errors.New("failed to get invoice with ID 'abc123': no invoice 4 u"),
		},
		"invoice cannot be paid twice": {
			invoiceFunc: func(ctx context.Context, args payd.InvoiceArgs) (*payd.Invoice, error) {
				return &payd.Invoice{ID: args.InvoiceID, State: payd.StateInvoicePaid}, nil
			},
			req:    payd.PaymentCreate{InvoiceID: "abc123"},
			expErr: errors.New("Item already exists: payment already received for invoice ID 'abc123'"),
		},
		"error reading fees is reported": {
			invoiceFunc: func(ctx context.Context, args payd.InvoiceArgs) (*payd.Invoice, error) {
				return &payd.Invoice{ID: args.InvoiceID, State: payd.StateInvoicePending}, nil
			},
			feesFunc: func(ctx context.Context) (*bt.FeeQuote, error) {
				return nil, errors.New("fee error")
			},
			req:    payd.PaymentCreate{InvoiceID: "abc123"},
			expErr: errors.New("failed to read fees for payment with id abc123: fee error"),
		},
		"tx with insufficient fees is rejected": {
			invoiceFunc: func(ctx context.Context, args payd.InvoiceArgs) (*payd.Invoice, error) {
				return &payd.Invoice{ID: args.InvoiceID, State: payd.StateInvoicePending}, nil
			},
			feesFunc: func(ctx context.Context) (*bt.FeeQuote, error) {
				return fq, nil
			},
			verifyPaymentFunc: func(context.Context, *spv.Envelope, ...spv.VerifyOpt) (*bt.Tx, error) {
				return nil, spv.ErrFeePaidNotEnough
			},
			req:           payd.PaymentCreate{InvoiceID: "abc123"},
			expVerifyOpts: []spv.VerifyOpt{spv.VerifyFees(fq), spv.NoVerifySPV()},
			expErr:        errors.New("[fees: not enough fees paid]"),
		},
		"invalid spv envelope is rejected": {
			invoiceFunc: func(ctx context.Context, args payd.InvoiceArgs) (*payd.Invoice, error) {
				return &payd.Invoice{ID: args.InvoiceID, State: payd.StateInvoicePending}, nil
			},
			feesFunc: func(ctx context.Context) (*bt.FeeQuote, error) {
				return fq, nil
			},
			verifyPaymentFunc: func(context.Context, *spv.Envelope, ...spv.VerifyOpt) (*bt.Tx, error) {
				return nil, spv.ErrInvalidProof
			},
			req:           payd.PaymentCreate{InvoiceID: "abc123"},
			expVerifyOpts: []spv.VerifyOpt{spv.VerifyFees(fq), spv.NoVerifySPV()},
			expErr:        errors.New("[spvEnvelope: invalid merkle proof, payment invalid]"),
		},
		"error reading destinations is reported": {
			invoiceFunc: func(ctx context.Context, args payd.InvoiceArgs) (*payd.Invoice, error) {
				return &payd.Invoice{ID: args.InvoiceID, State: payd.StateInvoicePending}, nil
			},
			feesFunc: func(ctx context.Context) (*bt.FeeQuote, error) {
				return fq, nil
			},
			verifyPaymentFunc: func(context.Context, *spv.Envelope, ...spv.VerifyOpt) (*bt.Tx, error) {
				return bt.NewTxFromString("010000000001e8030000000000001976a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac00000000")
			},
			destinationsFunc: func(context.Context, payd.DestinationsArgs) ([]payd.Output, error) {
				return nil, errors.New("destinations unknown")
			},
			req:           payd.PaymentCreate{InvoiceID: "abc123"},
			expVerifyOpts: []spv.VerifyOpt{spv.VerifyFees(fq), spv.NoVerifySPV()},
			expErr:        errors.New("failed to get destinations with ID 'abc123': destinations unknown"),
		},
		"mismatch in satoshis tx output/destination satoshis is rejected": {
			invoiceFunc: func(ctx context.Context, args payd.InvoiceArgs) (*payd.Invoice, error) {
				return &payd.Invoice{ID: args.InvoiceID, State: payd.StateInvoicePending}, nil
			},
			feesFunc: func(ctx context.Context) (*bt.FeeQuote, error) {
				return fq, nil
			},
			verifyPaymentFunc: func(context.Context, *spv.Envelope, ...spv.VerifyOpt) (*bt.Tx, error) {
				return bt.NewTxFromString("010000000001e9030000000000001976a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac00000000")
			},
			destinationsFunc: func(context.Context, payd.DestinationsArgs) ([]payd.Output, error) {
				return []payd.Output{{
					LockingScript:  "76a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac",
					DerivationPath: "2147483648/2147483648/2147483648",
					Satoshis:       1000,
					State:          "pending",
				}}, nil
			},
			req:           payd.PaymentCreate{InvoiceID: "abc123"},
			expVerifyOpts: []spv.VerifyOpt{spv.VerifyFees(fq), spv.NoVerifySPV()},
			expErr:        errors.New("[tx.outputs: output satoshis do not match requested amount]"),
		},
		// TODO: fix
		//"same destination cannot be paid to twice": {
		//	invoiceFunc: func(ctx context.Context, args payd.InvoiceArgs) (*payd.Invoice, error) {
		//		return &payd.Invoice{ID: args.InvoiceID, State: payd.StateInvoicePending}, nil
		//	},
		//	feesFunc: func(ctx context.Context) (*bt.FeeQuote, error) {
		//		return fq, nil
		//	},
		//	verifyPaymentFunc: func(context.Context, *spv.Envelope, ...spv.VerifyOpt) (*bt.Tx, error) {
		//		return bt.NewTxFromString("010000000002e8030000000000001976a91474b0424726ca510399c1eb5c8374f974c68b2fa388ace8030000000000001976a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac00000000")
		//	},
		//	destinationsFunc: func(context.Context, payd.DestinationsArgs) ([]payd.Output, error) {
		//		return []payd.Output{{
		//			LockingScript:  "76a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac",
		//			DerivationPath: "2147483648/2147483648/2147483648",
		//			Satoshis:       1000,
		//			State:          "pending",
		//		}}, nil
		//	},
		//	req:    payd.PaymentCreate{InvoiceID: "abc123"},
		//	expErr: errors.New("[tx.outputs: ]"),
		//},
		"tx with insufficient outputs is rejected": {
			invoiceFunc: func(ctx context.Context, args payd.InvoiceArgs) (*payd.Invoice, error) {
				return &payd.Invoice{ID: args.InvoiceID, Satoshis: 1001, State: payd.StateInvoicePending}, nil
			},
			feesFunc: func(ctx context.Context) (*bt.FeeQuote, error) {
				return fq, nil
			},
			verifyPaymentFunc: func(context.Context, *spv.Envelope, ...spv.VerifyOpt) (*bt.Tx, error) {
				return bt.NewTxFromString("010000000001e8030000000000001976a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac00000000")
			},
			destinationsFunc: func(context.Context, payd.DestinationsArgs) ([]payd.Output, error) {
				return []payd.Output{{
					LockingScript:  "76a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac",
					DerivationPath: "2147483648/2147483648/2147483648",
					Satoshis:       1000,
					State:          "pending",
				}}, nil
			},
			req:           payd.PaymentCreate{InvoiceID: "abc123"},
			expVerifyOpts: []spv.VerifyOpt{spv.VerifyFees(fq), spv.NoVerifySPV()},
			expErr:        errors.New("[transaction: tx does not pay enough to cover invoice, ensure all outputs are included, the correct destinations are used and try again]"),
		},
		"tx that doesn't use all destinations is rejected": {
			invoiceFunc: func(ctx context.Context, args payd.InvoiceArgs) (*payd.Invoice, error) {
				return &payd.Invoice{ID: args.InvoiceID, Satoshis: 1000, State: payd.StateInvoicePending}, nil
			},
			feesFunc: func(ctx context.Context) (*bt.FeeQuote, error) {
				return fq, nil
			},
			verifyPaymentFunc: func(context.Context, *spv.Envelope, ...spv.VerifyOpt) (*bt.Tx, error) {
				return bt.NewTxFromString("010000000001e8030000000000001976a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac00000000")
			},
			destinationsFunc: func(context.Context, payd.DestinationsArgs) ([]payd.Output, error) {
				return []payd.Output{{
					LockingScript:  "76a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac",
					DerivationPath: "2147483648/2147483648/2147483648",
					Satoshis:       1000,
					State:          "pending",
				}, {
					LockingScript:  "76a9141a4cc80bc3ee6567cb37f9c5121841a5f8e0b87d88ac",
					DerivationPath: "2147483648/2147483648/2147483650",
					Satoshis:       1000,
					State:          "pending",
				}}, nil
			},
			req:           payd.PaymentCreate{InvoiceID: "abc123"},
			expVerifyOpts: []spv.VerifyOpt{spv.VerifyFees(fq), spv.NoVerifySPV()},
			expErr:        errors.New("[tx.outputs: expected '2' outputs, received '1', ensure all destinations are supplied]"),
		},
		"error on tx create is reported": {
			invoiceFunc: func(ctx context.Context, args payd.InvoiceArgs) (*payd.Invoice, error) {
				return &payd.Invoice{ID: args.InvoiceID, State: payd.StateInvoicePending}, nil
			},
			feesFunc: func(ctx context.Context) (*bt.FeeQuote, error) {
				return fq, nil
			},
			verifyPaymentFunc: func(context.Context, *spv.Envelope, ...spv.VerifyOpt) (*bt.Tx, error) {
				return bt.NewTxFromString("010000000001e8030000000000001976a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac00000000")
			},
			destinationsFunc: func(context.Context, payd.DestinationsArgs) ([]payd.Output, error) {
				return []payd.Output{{
					LockingScript:  "76a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac",
					DerivationPath: "2147483648/2147483648/2147483648",
					Satoshis:       1000,
					State:          "pending",
				}}, nil
			},
			txCreateFunc: func(context.Context, payd.TransactionCreate) error {
				return errors.New("tx not create")
			},
			req: payd.PaymentCreate{
				InvoiceID: "abc123",
				RawTX:     null.StringFrom("010000000001e8030000000000001976a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac00000000"),
			},
			expRawTx:      "010000000001e8030000000000001976a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac00000000",
			expVerifyOpts: []spv.VerifyOpt{spv.VerifyFees(fq), spv.NoVerifySPV()},
			expErr:        errors.New("failed to store transaction for invoiceID 'abc123': tx not create"),
		},
		"error on proof callback is reported": {
			invoiceFunc: func(ctx context.Context, args payd.InvoiceArgs) (*payd.Invoice, error) {
				return &payd.Invoice{
					ID:    args.InvoiceID,
					State: payd.StateInvoicePending,
				}, nil
			},
			feesFunc: func(ctx context.Context) (*bt.FeeQuote, error) {
				return fq, nil
			},
			verifyPaymentFunc: func(context.Context, *spv.Envelope, ...spv.VerifyOpt) (*bt.Tx, error) {
				return bt.NewTxFromString("010000000001e8030000000000001976a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac00000000")
			},
			destinationsFunc: func(context.Context, payd.DestinationsArgs) ([]payd.Output, error) {
				return []payd.Output{{
					LockingScript:  "76a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac",
					DerivationPath: "2147483648/2147483648/2147483648",
					Satoshis:       1000,
					State:          "pending",
				}}, nil
			},
			txCreateFunc: func(context.Context, payd.TransactionCreate) error {
				return nil
			},
			proofCallbackCreateFunc: func(context.Context, payd.ProofCallbackArgs, map[string]payd.ProofCallback) error {
				return errors.New("oh no")
			},
			req: payd.PaymentCreate{
				InvoiceID: "abc123",
				RawTX:     null.StringFrom("010000000001e8030000000000001976a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac00000000"),
				ProofCallbacks: map[string]payd.ProofCallback{
					"wow": {},
				},
			},
			expRawTx:      "010000000001e8030000000000001976a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac00000000",
			expVerifyOpts: []spv.VerifyOpt{spv.VerifyFees(fq), spv.NoVerifySPV()},
			expErr:        errors.New("failed to store proof callbacks for invoiceID 'abc123': oh no"),
		},
		"error on broadcast is reported": {
			invoiceFunc: func(ctx context.Context, args payd.InvoiceArgs) (*payd.Invoice, error) {
				return &payd.Invoice{ID: args.InvoiceID, State: payd.StateInvoicePending}, nil
			},
			feesFunc: func(ctx context.Context) (*bt.FeeQuote, error) {
				return fq, nil
			},
			verifyPaymentFunc: func(context.Context, *spv.Envelope, ...spv.VerifyOpt) (*bt.Tx, error) {
				return bt.NewTxFromString("010000000001e8030000000000001976a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac00000000")
			},
			destinationsFunc: func(context.Context, payd.DestinationsArgs) ([]payd.Output, error) {
				return []payd.Output{{
					LockingScript:  "76a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac",
					DerivationPath: "2147483648/2147483648/2147483648",
					Satoshis:       1000,
					State:          "pending",
				}}, nil
			},
			txCreateFunc: func(context.Context, payd.TransactionCreate) error {
				return nil
			},
			proofCallbackCreateFunc: func(context.Context, payd.ProofCallbackArgs, map[string]payd.ProofCallback) error {
				return nil
			},
			broadcastFunc: func(context.Context, *bt.Tx) error {
				return errors.New("broadcast error")
			},
			txUpdateStateFunc: func(context.Context, payd.TransactionArgs, payd.TransactionStateUpdate) error {
				return nil
			},
			req: payd.PaymentCreate{
				InvoiceID: "abc123",
				RawTX:     null.StringFrom("010000000001e8030000000000001976a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac00000000"),
			},
			expRawTx:      "010000000001e8030000000000001976a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac00000000",
			expTxState:    payd.StateTxFailed,
			expVerifyOpts: []spv.VerifyOpt{spv.VerifyFees(fq), spv.NoVerifySPV()},
			expErr:        errors.New("failed to broadcast tx: broadcast error"),
		},
		"error on commit is reported": {
			invoiceFunc: func(ctx context.Context, args payd.InvoiceArgs) (*payd.Invoice, error) {
				return &payd.Invoice{ID: args.InvoiceID, State: payd.StateInvoicePending}, nil
			},
			feesFunc: func(ctx context.Context) (*bt.FeeQuote, error) {
				return fq, nil
			},
			verifyPaymentFunc: func(context.Context, *spv.Envelope, ...spv.VerifyOpt) (*bt.Tx, error) {
				return bt.NewTxFromString("010000000001e8030000000000001976a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac00000000")
			},
			destinationsFunc: func(context.Context, payd.DestinationsArgs) ([]payd.Output, error) {
				return []payd.Output{{
					LockingScript:  "76a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac",
					DerivationPath: "2147483648/2147483648/2147483648",
					Satoshis:       1000,
					State:          "pending",
				}}, nil
			},
			txCreateFunc: func(context.Context, payd.TransactionCreate) error {
				return nil
			},
			proofCallbackCreateFunc: func(context.Context, payd.ProofCallbackArgs, map[string]payd.ProofCallback) error {
				return nil
			},
			broadcastFunc: func(context.Context, *bt.Tx) error {
				return nil
			},
			txUpdateStateFunc: func(context.Context, payd.TransactionArgs, payd.TransactionStateUpdate) error {
				return nil
			},
			commitFunc: func(context.Context) error {
				return errors.New("oh no")
			},
			req: payd.PaymentCreate{
				InvoiceID: "abc123",
				RawTX:     null.StringFrom("010000000001e8030000000000001976a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac00000000"),
			},
			expRawTx:      "010000000001e8030000000000001976a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac00000000",
			expTxState:    payd.StateTxBroadcast,
			expVerifyOpts: []spv.VerifyOpt{spv.VerifyFees(fq), spv.NoVerifySPV()},
			expErr:        errors.New("oh no"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			svc := service.NewPayments(
				&mocks.PaymentVerifierMock{
					VerifyPaymentFunc: func(ctx context.Context, envelope *spv.Envelope, opts ...spv.VerifyOpt) (*bt.Tx, error) {
						assert.Equal(t, len(test.expVerifyOpts), len(opts))
						return test.verifyPaymentFunc(ctx, envelope, opts...)
					},
				},
				&mocks.TransactionWriterMock{
					TransactionCreateFunc: func(ctx context.Context, req payd.TransactionCreate) error {
						assert.Equal(t, test.expRawTx, req.TxHex)
						return test.txCreateFunc(ctx, req)
					},
					TransactionUpdateStateFunc: func(ctx context.Context, args payd.TransactionArgs, req payd.TransactionStateUpdate) error {
						assert.Equal(t, test.expTxState, req.State)
						return test.txUpdateStateFunc(ctx, args, req)
					},
				},
				&mocks.InvoiceReaderWriterMock{
					InvoiceFunc: test.invoiceFunc,
				},
				&mocks.DestinationsReaderWriterMock{
					DestinationsFunc: test.destinationsFunc,
				},
				&mocks.TransacterMock{
					WithTxFunc: func(ctx context.Context) context.Context {
						return ctx
					},
					RollbackFunc: func(context.Context) error {
						return nil
					},
					CommitFunc: test.commitFunc,
				},
				&mocks.BroadcastWriterMock{
					BroadcastFunc: test.broadcastFunc,
				},
				&mocks.FeeReaderMock{
					FeesFunc: test.feesFunc,
				},
				&mocks.ProofCallbackWriterMock{
					ProofCallBacksCreateFunc: test.proofCallbackCreateFunc,
				},
			)

			err := svc.PaymentCreate(context.TODO(), test.req)
			if test.expErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, test.expErr, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
