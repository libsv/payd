package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/libsv/go-bc/spv"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-dpp"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/libsv/payd"
	"github.com/libsv/payd/config"
	"github.com/libsv/payd/mocks"
	"github.com/libsv/payd/service"
	lerrs "github.com/theflyingcodr/lathos/errs"
)

func TestPayService_Pay(t *testing.T) {
	ts := time.Now()
	fq := bt.NewFeeQuote()
	tests := map[string]struct {
		req                payd.PayRequest
		envelopeFunc       func(ctx context.Context, args payd.EnvelopeArgs, req dpp.PaymentRequest) (*spv.Envelope, error)
		paymentRequestFunc func(context.Context, payd.PayRequest) (*dpp.PaymentRequest, error)
		paymentSendFunc    func(context.Context, payd.PayRequest, dpp.Payment) (*dpp.PaymentACK, error)
		walletConfig       *config.Wallet
		expKeyName         string
		expDeficits        []uint64
		expUTXOUnreserve   bool
		expCallbackURL     string
		expTx              string
		expChangeCreate    bool
		expChange          payd.DestinationCreate
		expErr             error
	}{
		"successful payment": {
			req: payd.PayRequest{
				PayToURL: "http://dpp-merchant/api/v1/payment/abc123",
			},
			walletConfig: &config.Wallet{},
			paymentRequestFunc: func(ctx context.Context, req payd.PayRequest) (*dpp.PaymentRequest, error) {
				return &dpp.PaymentRequest{
					Network: "testnet",
					Destinations: dpp.PaymentDestinations{
						Outputs: []dpp.Output{{
							Amount: 1000,
							LockingScript: func() *bscript.Script {
								ls, err := bscript.NewFromHexString("76a9146e912a2a1c28448522c1eba7d73ce0719b0636b388ac")
								assert.NoError(t, err)
								return ls
							}(),
						}, {
							Amount: 2000,
							LockingScript: func() *bscript.Script {
								ls, err := bscript.NewFromHexString("76a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac")
								assert.NoError(t, err)
								return ls
							}(),
						}},
					},
					CreationTimestamp:   ts,
					ExpirationTimestamp: ts.Add(24 * time.Hour),
					FeeRate:             fq,
					Memo:                "payment abc123",
					MerchantData: &dpp.Merchant{
						ExtendedData: map[string]interface{}{
							"paymentReference": "abc123",
						},
						Name: "Merchant Person",
					},
					PaymentURL: "http://dpp-merchant/api/v1/payment/abc123",
				}, nil
			},
			envelopeFunc: func(ctx context.Context, args payd.EnvelopeArgs, req dpp.PaymentRequest) (*spv.Envelope, error) {
				return &spv.Envelope{
					RawTx: "0100000002402b8ff345f8d428bfc4e553c5755d9ee771c99d38142608f290aba379585077010000006a47304402206a9c351ba35f43b3b3c4eac4cbaade8ce3e405fa9e8c9cf7bd74df084ea4396d02202073e344b6e7c21c97a5e0b2beb7c667784208f50d306f87b37bac92658e1db1412102f46acbd7a9825d5220464b761b6477a600a50664a6b8765a77ec7e1b19e8f36bffffffffac0025c8519afefca6ae96398a255b98239bbf4ab92fac1bbadf4b550244942e000000006a47304402207b80e3da87295641ffb9dfd5823cfb9a74634174774cdd0fc88f1dcd541f3558022001c88cdaaf1b4c5b8dfaf80de7fd9de45b1df3a4ebafc3dd13e3b1ca4820bf4041210251c7b5806db15e127a986611aa23f71a84879acb2ceef610f9eabbf355790a29ffffffff03e8030000000000001976a9146e912a2a1c28448522c1eba7d73ce0719b0636b388acd0070000000000001976a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac1c030000000000001976a9148b1ca598db87cfe283229bf724ad39cc4f1a665788ac00000000"}, nil
			},
			paymentSendFunc: func(context.Context, payd.PayRequest, dpp.Payment) (*dpp.PaymentACK, error) {
				return &dpp.PaymentACK{
					PeerChannel: &dpp.PeerChannelData{
						Host:      "myhost",
						ChannelID: "abc123",
						Token:     "def456",
					},
				}, nil
			},
			expDeficits:     []uint64{3039, 1113},
			expCallbackURL:  "https://myserver/api/v1/proofs/",
			expChangeCreate: true,
			expChange: payd.DestinationCreate{
				Satoshis:       796,
				DerivationPath: "2147483648/2147483723/2147483648",
				Script:         "76a9148b1ca598db87cfe283229bf724ad39cc4f1a665788ac",
				UserID:         1,
				KeyName:        "masterkey",
			},
			expKeyName: "masterkey",
		},
		"successful payment no change": {
			req: payd.PayRequest{
				PayToURL: "http://dpp-merchant/api/v1/payment/abc123",
			},

			walletConfig: &config.Wallet{},
			paymentRequestFunc: func(ctx context.Context, req payd.PayRequest) (*dpp.PaymentRequest, error) {
				return &dpp.PaymentRequest{
					Network: "testnet",
					Destinations: dpp.PaymentDestinations{
						Outputs: []dpp.Output{{
							Amount: 1000,
							LockingScript: func() *bscript.Script {
								ls, err := bscript.NewFromHexString("76a9146e912a2a1c28448522c1eba7d73ce0719b0636b388ac")
								assert.NoError(t, err)
								return ls
							}(),
						}, {
							Amount: 2000,
							LockingScript: func() *bscript.Script {
								ls, err := bscript.NewFromHexString("76a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac")
								assert.NoError(t, err)
								return ls
							}(),
						}},
					},
					CreationTimestamp:   ts,
					ExpirationTimestamp: ts.Add(24 * time.Hour),
					FeeRate:             fq,
					Memo:                "payment abc123",
					MerchantData: &dpp.Merchant{
						ExtendedData: map[string]interface{}{
							"paymentReference": "abc123",
						},
						Name: "Merchant Person",
					},
					PaymentURL: "http://dpp-merchant/api/v1/payment/abc123",
				}, nil
			},
			envelopeFunc: func(ctx context.Context, args payd.EnvelopeArgs, req dpp.PaymentRequest) (*spv.Envelope, error) {
				return &spv.Envelope{
					RawTx: "0100000002402b8ff345f8d428bfc4e553c5755d9ee771c99d38142608f290aba379585077010000006b483045022100b5364c0e25f6edb9a4d8bfeed6cab278e874f22e9ceddec8fb4ff3ad647731eb0220729480538ffc64de8954a7e5b608bc3d5748055eb4fd970d4d625b6ffbb60bfe412102f46acbd7a9825d5220464b761b6477a600a50664a6b8765a77ec7e1b19e8f36bffffffffac0025c8519afefca6ae96398a255b98239bbf4ab92fac1bbadf4b550244942e000000006a47304402207be0f7b8bbb629184fdb3180a73f4f571644f04da996c3b9ae845dbc567b506202204e7d9378b1f6d6c14d7ebce8a2376090c1651747641b84a98134f168a4ff6d8941210251c7b5806db15e127a986611aa23f71a84879acb2ceef610f9eabbf355790a29ffffffff02e8030000000000001976a9146e912a2a1c28448522c1eba7d73ce0719b0636b388acd0070000000000001976a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac00000000"}, nil
			},
			paymentSendFunc: func(context.Context, payd.PayRequest, dpp.Payment) (*dpp.PaymentACK, error) {
				return &dpp.PaymentACK{}, nil
			},
			expDeficits:    []uint64{3039, 1113},
			expTx:          "0100000002402b8ff345f8d428bfc4e553c5755d9ee771c99d38142608f290aba379585077010000006b483045022100b5364c0e25f6edb9a4d8bfeed6cab278e874f22e9ceddec8fb4ff3ad647731eb0220729480538ffc64de8954a7e5b608bc3d5748055eb4fd970d4d625b6ffbb60bfe412102f46acbd7a9825d5220464b761b6477a600a50664a6b8765a77ec7e1b19e8f36bffffffffac0025c8519afefca6ae96398a255b98239bbf4ab92fac1bbadf4b550244942e000000006a47304402207be0f7b8bbb629184fdb3180a73f4f571644f04da996c3b9ae845dbc567b506202204e7d9378b1f6d6c14d7ebce8a2376090c1651747641b84a98134f168a4ff6d8941210251c7b5806db15e127a986611aa23f71a84879acb2ceef610f9eabbf355790a29ffffffff02e8030000000000001976a9146e912a2a1c28448522c1eba7d73ce0719b0636b388acd0070000000000001976a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac00000000",
			expCallbackURL: "https://myserver/api/v1/proofs/",
			expKeyName:     "masterkey",
		},
		"invalid url in request is rejected": {

			walletConfig: &config.Wallet{},
			req: payd.PayRequest{
				PayToURL: ":dpp-merchant/api/v1/payment/abc123",
			},
			expErr: errors.New(`[payToURL: parse ":dpp-merchant/api/v1/payment/abc123": missing protocol scheme]`),
		},
		"error for expired payment": {
			req: payd.PayRequest{
				PayToURL: "http://dpp-merchant/api/v1/payment/abc123",
			},
			paymentRequestFunc: func(ctx context.Context, req payd.PayRequest) (*dpp.PaymentRequest, error) {
				return nil, lerrs.NewErrUnprocessable("E001", "no payment request for you")
			},
			expKeyName: "masterkey",
			expErr:     errors.New("Unprocessable: failed to request payment for url http://dpp-merchant/api/v1/payment/abc123 : Unprocessable: no payment request for you"),
		},
		"error fetching payment request is reported": {

			walletConfig: &config.Wallet{},
			req: payd.PayRequest{
				PayToURL: "http://dpp-merchant/api/v1/payment/abc123",
			},
			paymentRequestFunc: func(ctx context.Context, req payd.PayRequest) (*dpp.PaymentRequest, error) {
				return nil, errors.New("no payment request for you")
			},
			expKeyName: "masterkey",
			expErr:     errors.New("failed to request payment for url http://dpp-merchant/api/v1/payment/abc123: no payment request for you"),
		},
		"insufficient utxos errors, reserved funds are freed": {

			walletConfig: &config.Wallet{},
			req: payd.PayRequest{
				PayToURL: "http://dpp-merchant/api/v1/payment/abc123",
			},
			paymentRequestFunc: func(ctx context.Context, req payd.PayRequest) (*dpp.PaymentRequest, error) {
				return &dpp.PaymentRequest{
					Network: "testnet",
					Destinations: dpp.PaymentDestinations{
						Outputs: []dpp.Output{{
							Amount: 1000,
							LockingScript: func() *bscript.Script {
								ls, err := bscript.NewFromHexString("76a9146e912a2a1c28448522c1eba7d73ce0719b0636b388ac")
								assert.NoError(t, err)
								return ls
							}(),
						}, {
							Amount: 2000,
							LockingScript: func() *bscript.Script {
								ls, err := bscript.NewFromHexString("76a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac")
								assert.NoError(t, err)
								return ls
							}(),
						}},
					},
					CreationTimestamp:   ts,
					ExpirationTimestamp: ts.Add(24 * time.Hour),
					FeeRate:             fq,
					Memo:                "payment abc123",
					MerchantData: &dpp.Merchant{
						ExtendedData: map[string]interface{}{
							"paymentReference": "abc123",
						},
						Name: "Merchant Person",
					},
					PaymentURL: "http://dpp-merchant/api/v1/payment/abc123",
				}, nil
			},
			envelopeFunc: func(ctx context.Context, args payd.EnvelopeArgs, req dpp.PaymentRequest) (*spv.Envelope, error) {
				return nil, errors.New("Unprocessable: insufficient funds provided")
			},
			expDeficits:      []uint64{3039, 1113},
			expKeyName:       "masterkey",
			expUTXOUnreserve: true,
			expErr:           errors.New("envelope creation failed for 'http://dpp-merchant/api/v1/payment/abc123': Unprocessable: insufficient funds provided"),
		},
		"error on envelope create is reported": {

			walletConfig: &config.Wallet{},
			req: payd.PayRequest{
				PayToURL: "http://dpp-merchant/api/v1/payment/abc123",
			},
			paymentRequestFunc: func(ctx context.Context, req payd.PayRequest) (*dpp.PaymentRequest, error) {
				return &dpp.PaymentRequest{
					Network: "testnet",
					Destinations: dpp.PaymentDestinations{
						Outputs: []dpp.Output{{
							Amount: 1000,
							LockingScript: func() *bscript.Script {
								ls, err := bscript.NewFromHexString("76a9146e912a2a1c28448522c1eba7d73ce0719b0636b388ac")
								assert.NoError(t, err)
								return ls
							}(),
						}, {
							Amount: 2000,
							LockingScript: func() *bscript.Script {
								ls, err := bscript.NewFromHexString("76a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac")
								assert.NoError(t, err)
								return ls
							}(),
						}},
					},
					CreationTimestamp:   ts,
					ExpirationTimestamp: ts.Add(24 * time.Hour),
					FeeRate:             fq,
					Memo:                "payment abc123",
					MerchantData: &dpp.Merchant{
						ExtendedData: map[string]interface{}{
							"paymentReference": "abc123",
						},
						Name: "Merchant Person",
					},
					PaymentURL: "http://dpp-merchant/api/v1/payment/abc123",
				}, nil
			},
			envelopeFunc: func(ctx context.Context, args payd.EnvelopeArgs, req dpp.PaymentRequest) (*spv.Envelope, error) {
				return nil, errors.New("no envelope for you")
			},
			expDeficits:      []uint64{3039, 1113},
			expTx:            "0100000002402b8ff345f8d428bfc4e553c5755d9ee771c99d38142608f290aba379585077010000006a47304402206a9c351ba35f43b3b3c4eac4cbaade8ce3e405fa9e8c9cf7bd74df084ea4396d02202073e344b6e7c21c97a5e0b2beb7c667784208f50d306f87b37bac92658e1db1412102f46acbd7a9825d5220464b761b6477a600a50664a6b8765a77ec7e1b19e8f36bffffffffac0025c8519afefca6ae96398a255b98239bbf4ab92fac1bbadf4b550244942e000000006a47304402207b80e3da87295641ffb9dfd5823cfb9a74634174774cdd0fc88f1dcd541f3558022001c88cdaaf1b4c5b8dfaf80de7fd9de45b1df3a4ebafc3dd13e3b1ca4820bf4041210251c7b5806db15e127a986611aa23f71a84879acb2ceef610f9eabbf355790a29ffffffff03e8030000000000001976a9146e912a2a1c28448522c1eba7d73ce0719b0636b388acd0070000000000001976a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac1c030000000000001976a9148b1ca598db87cfe283229bf724ad39cc4f1a665788ac00000000",
			expKeyName:       "masterkey",
			expUTXOUnreserve: true,
			expErr:           errors.New("envelope creation failed for 'http://dpp-merchant/api/v1/payment/abc123': no envelope for you"),
		},
		"error on payment send is reported": {
			walletConfig: &config.Wallet{},
			req: payd.PayRequest{
				PayToURL: "http://dpp-merchant/api/v1/payment/abc123",
			},
			paymentRequestFunc: func(ctx context.Context, req payd.PayRequest) (*dpp.PaymentRequest, error) {
				return &dpp.PaymentRequest{
					Network: "testnet",
					Destinations: dpp.PaymentDestinations{
						Outputs: []dpp.Output{{
							Amount: 1000,
							LockingScript: func() *bscript.Script {
								ls, err := bscript.NewFromHexString("76a9146e912a2a1c28448522c1eba7d73ce0719b0636b388ac")
								assert.NoError(t, err)
								return ls
							}(),
						}, {
							Amount: 2000,
							LockingScript: func() *bscript.Script {
								ls, err := bscript.NewFromHexString("76a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac")
								assert.NoError(t, err)
								return ls
							}(),
						}},
					},
					CreationTimestamp:   ts,
					ExpirationTimestamp: ts.Add(24 * time.Hour),
					FeeRate:             fq,
					Memo:                "payment abc123",
					MerchantData: &dpp.Merchant{
						ExtendedData: map[string]interface{}{
							"paymentReference": "abc123",
						},
						Name: "Merchant Person",
					},
					PaymentURL: "http://dpp-merchant/api/v1/payment/abc123",
				}, nil
			},
			envelopeFunc: func(ctx context.Context, args payd.EnvelopeArgs, req dpp.PaymentRequest) (*spv.Envelope, error) {
				return &spv.Envelope{RawTx: "0100000002402b8ff345f8d428bfc4e553c5755d9ee771c99d38142608f290aba379585077010000006a47304402206a9c351ba35f43b3b3c4eac4cbaade8ce3e405fa9e8c9cf7bd74df084ea4396d02202073e344b6e7c21c97a5e0b2beb7c667784208f50d306f87b37bac92658e1db1412102f46acbd7a9825d5220464b761b6477a600a50664a6b8765a77ec7e1b19e8f36bffffffffac0025c8519afefca6ae96398a255b98239bbf4ab92fac1bbadf4b550244942e000000006a47304402207b80e3da87295641ffb9dfd5823cfb9a74634174774cdd0fc88f1dcd541f3558022001c88cdaaf1b4c5b8dfaf80de7fd9de45b1df3a4ebafc3dd13e3b1ca4820bf4041210251c7b5806db15e127a986611aa23f71a84879acb2ceef610f9eabbf355790a29ffffffff03e8030000000000001976a9146e912a2a1c28448522c1eba7d73ce0719b0636b388acd0070000000000001976a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac1c030000000000001976a9148b1ca598db87cfe283229bf724ad39cc4f1a665788ac00000000"}, nil
			},
			paymentSendFunc: func(context.Context, payd.PayRequest, dpp.Payment) (*dpp.PaymentACK, error) {
				return nil, errors.New("no send for you")
			},
			expDeficits:      []uint64{3039, 1113},
			expTx:            "0100000002402b8ff345f8d428bfc4e553c5755d9ee771c99d38142608f290aba379585077010000006a47304402206a9c351ba35f43b3b3c4eac4cbaade8ce3e405fa9e8c9cf7bd74df084ea4396d02202073e344b6e7c21c97a5e0b2beb7c667784208f50d306f87b37bac92658e1db1412102f46acbd7a9825d5220464b761b6477a600a50664a6b8765a77ec7e1b19e8f36bffffffffac0025c8519afefca6ae96398a255b98239bbf4ab92fac1bbadf4b550244942e000000006a47304402207b80e3da87295641ffb9dfd5823cfb9a74634174774cdd0fc88f1dcd541f3558022001c88cdaaf1b4c5b8dfaf80de7fd9de45b1df3a4ebafc3dd13e3b1ca4820bf4041210251c7b5806db15e127a986611aa23f71a84879acb2ceef610f9eabbf355790a29ffffffff03e8030000000000001976a9146e912a2a1c28448522c1eba7d73ce0719b0636b388acd0070000000000001976a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac1c030000000000001976a9148b1ca598db87cfe283229bf724ad39cc4f1a665788ac00000000",
			expUTXOUnreserve: true,
			expCallbackURL:   "https://myserver/api/v1/proofs/",
			expKeyName:       "masterkey",
			expErr:           errors.New("failed to send payment http://dpp-merchant/api/v1/payment/abc123: no send for you"),
		}, "payment limit enabled with destinations exceeding limit": {
			req: payd.PayRequest{
				PayToURL: "http://dpp-merchant/api/v1/payment/abc123",
			},
			walletConfig: &config.Wallet{
				PayoutLimitSatoshis: 1000,
				PayoutLimitEnabled:  true,
			},
			paymentRequestFunc: func(ctx context.Context, req payd.PayRequest) (*dpp.PaymentRequest, error) {
				return &dpp.PaymentRequest{
					Network: "testnet",
					Destinations: dpp.PaymentDestinations{
						Outputs: []dpp.Output{{
							Amount: 1000,
							LockingScript: func() *bscript.Script {
								ls, err := bscript.NewFromHexString("76a9146e912a2a1c28448522c1eba7d73ce0719b0636b388ac")
								assert.NoError(t, err)
								return ls
							}(),
						}, {
							Amount: 2000,
							LockingScript: func() *bscript.Script {
								ls, err := bscript.NewFromHexString("76a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac")
								assert.NoError(t, err)
								return ls
							}(),
						}},
					},
					CreationTimestamp:   ts,
					ExpirationTimestamp: ts.Add(24 * time.Hour),
					FeeRate:             fq,
					Memo:                "payment abc123",
					MerchantData: &dpp.Merchant{
						ExtendedData: map[string]interface{}{
							"paymentReference": "abc123",
						},
						Name: "Merchant Person",
					},
					PaymentURL: "http://dpp-merchant/api/v1/payment/abc123",
				}, nil
			},
			expErr: errors.New("Unprocessable: amount requested 3000 satoshis is larger than our max payout of 1000 satoshis"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			svc := service.NewPayService(
				&mocks.TransacterMock{
					WithTxFunc: func(ctx context.Context) context.Context {
						return ctx
					},
					RollbackFunc: func(context.Context) error {
						return nil
					},
					CommitFunc: func(ctx context.Context) error {
						return nil
					},
				},
				&mocks.DPPMock{
					PaymentRequestFunc: test.paymentRequestFunc,
					PaymentSendFunc: func(ctx context.Context, req payd.PayRequest, args dpp.Payment) (*dpp.PaymentACK, error) {
						_, ok := args.ProofCallbacks[test.expCallbackURL]
						assert.True(t, ok, "%s not in %+v", test.expCallbackURL, args.ProofCallbacks)
						return test.paymentSendFunc(ctx, req, args)
					},
				},
				&mocks.EnvelopeServiceMock{EnvelopeFunc: test.envelopeFunc},
				&config.Server{Hostname: "myserver"},
				&mocks.PeerChannelsNotifyServiceMock{
					SubscribeFunc: func(ctx context.Context, args *payd.PeerChannel) error {
						return nil
					},
				},
				&mocks.PeerChannelsStoreMock{
					PeerChannelCreateFunc: func(ctx context.Context, args *payd.PeerChannelCreateArgs) error {
						return nil
					},
					PeerChannelAPITokenCreateFunc: func(ctx context.Context, args *payd.PeerChannelAPITokenStoreArgs) error {
						return nil
					},
				},
				&mocks.TransactionWriterMock{
					TransactionUpdateStateFunc: func(ctx context.Context, args payd.TransactionArgs, req payd.TransactionStateUpdate) error {
						return nil
					},
				},
				test.walletConfig,
			)

			_, err := svc.Pay(context.TODO(), test.req)
			if test.expErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, test.expErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
