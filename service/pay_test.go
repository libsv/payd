package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/libsv/go-bc/spv"
	"github.com/libsv/go-bt/v2"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/libsv/payd"
	"github.com/libsv/payd/config"
	"github.com/libsv/payd/internal"
	"github.com/libsv/payd/mocks"
	"github.com/libsv/payd/service"
)

func TestPayService_Pay(t *testing.T) {
	ts := time.Now()
	fq := bt.NewFeeQuote()
	tests := map[string]struct {
		req                    payd.PayRequest
		envelopeFunc           func(ctx context.Context, args payd.EnvelopeArgs, req payd.PaymentRequestResponse) (*payd.Envelope, error)
		paymentRequestFunc     func(context.Context, payd.PayRequest) (*payd.PaymentRequestResponse, error)
		utxoReserveFunc        func(context.Context, payd.UTXOReserve) ([]payd.UTXO, error)
		paymentSendFunc        func(context.Context, payd.PayRequest, payd.PaymentSend) (*payd.PaymentACK, error)
		destinationsCreateFunc func(context.Context, payd.DestinationsCreateArgs, []payd.DestinationCreate) ([]payd.Output, error)
		txCreateFunc           func(context.Context, payd.TransactionCreate) error
		utxoSpendFunc          func(context.Context, payd.UTXOSpend) error

		expKeyName       string
		expDeficits      []uint64
		expUTXOUnreserve bool
		expCallbackURL   string
		expTx            string
		expChangeCreate  bool
		expChange        payd.DestinationCreate
		expErr           error
	}{
		"successful payment": {
			req: payd.PayRequest{
				PayToURL: "http://p4-merchant/api/v1/payment/abc123",
			},
			paymentRequestFunc: func(ctx context.Context, req payd.PayRequest) (*payd.PaymentRequestResponse, error) {
				return &payd.PaymentRequestResponse{
					Network: "testnet",
					Destinations: payd.P4Destination{
						Outputs: []payd.P4Output{{
							Amount: 1000,
							Script: "76a9146e912a2a1c28448522c1eba7d73ce0719b0636b388ac",
						}, {
							Amount: 2000,
							Script: "76a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac",
						}},
					},
					CreationTimestamp:   ts,
					ExpirationTimestamp: ts.Add(24 * time.Hour),
					Fee:                 fq,
					Memo:                "payment abc123",
					MerchantData: payd.User{
						ExtendedData: map[string]interface{}{
							"paymentReference": "abc123",
						},
						Name: "Merchant Person",
					},
					PaymentURL: "http://p4-merchant/api/v1/payment/abc123",
				}, nil
			},
			utxoReserveFunc: func() func(ctx context.Context, req payd.UTXOReserve) ([]payd.UTXO, error) {
				idx := 0
				utxos := []payd.UTXO{{
					Satoshis:       2000,
					LockingScript:  "76a9141a4cc80bc3ee6567cb37f9c5121841a5f8e0b87d88ac",
					DerivationPath: "2147483648/2147483648/2147483650",
					TxID:           "77505879a3ab90f2082614389dc971e79e5d75c553e5c4bf28d4f845f38f2b40",
					Vout:           1,
				}, {
					Satoshis:       2000,
					LockingScript:  "76a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac",
					DerivationPath: "2147483648/2147483648/2147483648",
					TxID:           "2e944402554bdfba1bac2fb94abf9b23985b258a3996aea6fcfe9a51c82500ac",
					Vout:           0,
				}}
				return func(ctx context.Context, req payd.UTXOReserve) ([]payd.UTXO, error) {
					defer func() { idx++ }()
					return utxos[idx : idx+1], nil
				}
			}(),
			envelopeFunc: func(ctx context.Context, args payd.EnvelopeArgs, req payd.PaymentRequestResponse) (*payd.Envelope, error) {
				return &payd.Envelope{
					SPVEnvelope: &spv.Envelope{
						RawTx: "0100000002402b8ff345f8d428bfc4e553c5755d9ee771c99d38142608f290aba379585077010000006a47304402206a9c351ba35f43b3b3c4eac4cbaade8ce3e405fa9e8c9cf7bd74df084ea4396d02202073e344b6e7c21c97a5e0b2beb7c667784208f50d306f87b37bac92658e1db1412102f46acbd7a9825d5220464b761b6477a600a50664a6b8765a77ec7e1b19e8f36bffffffffac0025c8519afefca6ae96398a255b98239bbf4ab92fac1bbadf4b550244942e000000006a47304402207b80e3da87295641ffb9dfd5823cfb9a74634174774cdd0fc88f1dcd541f3558022001c88cdaaf1b4c5b8dfaf80de7fd9de45b1df3a4ebafc3dd13e3b1ca4820bf4041210251c7b5806db15e127a986611aa23f71a84879acb2ceef610f9eabbf355790a29ffffffff03e8030000000000001976a9146e912a2a1c28448522c1eba7d73ce0719b0636b388acd0070000000000001976a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac1c030000000000001976a9148b1ca598db87cfe283229bf724ad39cc4f1a665788ac00000000",
					},
					Change: payd.Output{
						LockingScript:  internal.StringToScript("76a9148b1ca598db87cfe283229bf724ad39cc4f1a665788ac"),
						Satoshis:       796,
						DerivationPath: "2147483648/2147483723/2147483648",
					},
				}, nil
			},
			paymentSendFunc: func(context.Context, payd.PayRequest, payd.PaymentSend) (*payd.PaymentACK, error) {
				return &payd.PaymentACK{}, nil
			},
			destinationsCreateFunc: func(context.Context, payd.DestinationsCreateArgs, []payd.DestinationCreate) ([]payd.Output, error) {
				return []payd.Output{{}}, nil
			},
			txCreateFunc: func(context.Context, payd.TransactionCreate) error {
				return nil
			},
			utxoSpendFunc: func(context.Context, payd.UTXOSpend) error {
				return nil
			},
			expDeficits:     []uint64{3039, 1113},
			expCallbackURL:  "https://myserver/api/v1/proofs/",
			expChangeCreate: true,
			expChange: payd.DestinationCreate{
				Satoshis:       796,
				DerivationPath: "2147483648/2147483723/2147483648",
				Script:         "76a9148b1ca598db87cfe283229bf724ad39cc4f1a665788ac",
				Keyname:        "masterkey",
			},
			expKeyName: "masterkey",
		},
		"successful payment no change": {
			req: payd.PayRequest{
				PayToURL: "http://p4-merchant/api/v1/payment/abc123",
			},
			paymentRequestFunc: func(ctx context.Context, req payd.PayRequest) (*payd.PaymentRequestResponse, error) {
				return &payd.PaymentRequestResponse{
					Network: "testnet",
					Destinations: payd.P4Destination{
						Outputs: []payd.P4Output{{
							Amount: 1000,
							Script: "76a9146e912a2a1c28448522c1eba7d73ce0719b0636b388ac",
						}, {
							Amount: 2000,
							Script: "76a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac",
						}},
					},
					CreationTimestamp:   ts,
					ExpirationTimestamp: ts.Add(24 * time.Hour),
					Fee:                 fq,
					Memo:                "payment abc123",
					MerchantData: payd.User{
						ExtendedData: map[string]interface{}{
							"paymentReference": "abc123",
						},
						Name: "Merchant Person",
					},
					PaymentURL: "http://p4-merchant/api/v1/payment/abc123",
				}, nil
			},
			utxoReserveFunc: func() func(ctx context.Context, req payd.UTXOReserve) ([]payd.UTXO, error) {
				idx := 0
				utxos := []payd.UTXO{{
					Satoshis:       2000,
					LockingScript:  "76a9141a4cc80bc3ee6567cb37f9c5121841a5f8e0b87d88ac",
					DerivationPath: "2147483648/2147483648/2147483650",
					TxID:           "77505879a3ab90f2082614389dc971e79e5d75c553e5c4bf28d4f845f38f2b40",
					Vout:           1,
				}, {
					Satoshis:       1187,
					LockingScript:  "76a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac",
					DerivationPath: "2147483648/2147483648/2147483648",
					TxID:           "2e944402554bdfba1bac2fb94abf9b23985b258a3996aea6fcfe9a51c82500ac",
					Vout:           0,
				}}
				return func(ctx context.Context, req payd.UTXOReserve) ([]payd.UTXO, error) {
					defer func() { idx++ }()
					return utxos[idx : idx+1], nil
				}
			}(),
			envelopeFunc: func(ctx context.Context, args payd.EnvelopeArgs, req payd.PaymentRequestResponse) (*payd.Envelope, error) {
				return &payd.Envelope{
					SPVEnvelope: &spv.Envelope{
						RawTx: "0100000002402b8ff345f8d428bfc4e553c5755d9ee771c99d38142608f290aba379585077010000006b483045022100b5364c0e25f6edb9a4d8bfeed6cab278e874f22e9ceddec8fb4ff3ad647731eb0220729480538ffc64de8954a7e5b608bc3d5748055eb4fd970d4d625b6ffbb60bfe412102f46acbd7a9825d5220464b761b6477a600a50664a6b8765a77ec7e1b19e8f36bffffffffac0025c8519afefca6ae96398a255b98239bbf4ab92fac1bbadf4b550244942e000000006a47304402207be0f7b8bbb629184fdb3180a73f4f571644f04da996c3b9ae845dbc567b506202204e7d9378b1f6d6c14d7ebce8a2376090c1651747641b84a98134f168a4ff6d8941210251c7b5806db15e127a986611aa23f71a84879acb2ceef610f9eabbf355790a29ffffffff02e8030000000000001976a9146e912a2a1c28448522c1eba7d73ce0719b0636b388acd0070000000000001976a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac00000000",
					},
				}, nil
			},
			paymentSendFunc: func(context.Context, payd.PayRequest, payd.PaymentSend) (*payd.PaymentACK, error) {
				return &payd.PaymentACK{}, nil
			},
			destinationsCreateFunc: func(context.Context, payd.DestinationsCreateArgs, []payd.DestinationCreate) ([]payd.Output, error) {
				return []payd.Output{{}}, nil
			},
			txCreateFunc: func(context.Context, payd.TransactionCreate) error {
				return nil
			},
			utxoSpendFunc: func(context.Context, payd.UTXOSpend) error {
				return nil
			},
			expDeficits:    []uint64{3039, 1113},
			expTx:          "0100000002402b8ff345f8d428bfc4e553c5755d9ee771c99d38142608f290aba379585077010000006b483045022100b5364c0e25f6edb9a4d8bfeed6cab278e874f22e9ceddec8fb4ff3ad647731eb0220729480538ffc64de8954a7e5b608bc3d5748055eb4fd970d4d625b6ffbb60bfe412102f46acbd7a9825d5220464b761b6477a600a50664a6b8765a77ec7e1b19e8f36bffffffffac0025c8519afefca6ae96398a255b98239bbf4ab92fac1bbadf4b550244942e000000006a47304402207be0f7b8bbb629184fdb3180a73f4f571644f04da996c3b9ae845dbc567b506202204e7d9378b1f6d6c14d7ebce8a2376090c1651747641b84a98134f168a4ff6d8941210251c7b5806db15e127a986611aa23f71a84879acb2ceef610f9eabbf355790a29ffffffff02e8030000000000001976a9146e912a2a1c28448522c1eba7d73ce0719b0636b388acd0070000000000001976a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac00000000",
			expCallbackURL: "https://myserver/api/v1/proofs/",
			expKeyName:     "masterkey",
		},
		"invalid url in request is rejected": {
			req: payd.PayRequest{
				PayToURL: ":p4-merchant/api/v1/payment/abc123",
			},
			expErr: errors.New(`[payToURL: parse ":p4-merchant/api/v1/payment/abc123": missing protocol scheme]`),
		},
		"error fetching payment request is reported": {
			req: payd.PayRequest{
				PayToURL: "http://p4-merchant/api/v1/payment/abc123",
			},
			paymentRequestFunc: func(ctx context.Context, req payd.PayRequest) (*payd.PaymentRequestResponse, error) {
				return nil, errors.New("no payment request for you")
			},
			expKeyName: "masterkey",
			expErr:     errors.New("failed to request payment for url http://p4-merchant/api/v1/payment/abc123: no payment request for you"),
		},
		"insufficient utxos errors, reserved funds are freed": {
			req: payd.PayRequest{
				PayToURL: "http://p4-merchant/api/v1/payment/abc123",
			},
			paymentRequestFunc: func(ctx context.Context, req payd.PayRequest) (*payd.PaymentRequestResponse, error) {
				return &payd.PaymentRequestResponse{
					Network: "testnet",
					Destinations: payd.P4Destination{
						Outputs: []payd.P4Output{{
							Amount: 1000,
							Script: "76a9146e912a2a1c28448522c1eba7d73ce0719b0636b388ac",
						}, {
							Amount: 2000,
							Script: "76a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac",
						}},
					},
					CreationTimestamp:   ts,
					ExpirationTimestamp: ts.Add(24 * time.Hour),
					Fee:                 fq,
					Memo:                "payment abc123",
					MerchantData: payd.User{
						ExtendedData: map[string]interface{}{
							"paymentReference": "abc123",
						},
						Name: "Merchant Person",
					},
					PaymentURL: "http://p4-merchant/api/v1/payment/abc123",
				}, nil
			},
			utxoReserveFunc: func(ctx context.Context, req payd.UTXOReserve) ([]payd.UTXO, error) {
				return []payd.UTXO{}, nil
			},
			envelopeFunc: func(ctx context.Context, args payd.EnvelopeArgs, req payd.PaymentRequestResponse) (*payd.Envelope, error) {
				return nil, errors.New("Unprocessable: insufficient funds provided")
			},
			expDeficits:      []uint64{3039, 1113},
			expKeyName:       "masterkey",
			expUTXOUnreserve: true,
			expErr:           errors.New("envelope creation failed for 'http://p4-merchant/api/v1/payment/abc123': Unprocessable: insufficient funds provided"),
		},
		"error on envelope create is reported": {
			req: payd.PayRequest{
				PayToURL: "http://p4-merchant/api/v1/payment/abc123",
			},
			paymentRequestFunc: func(ctx context.Context, req payd.PayRequest) (*payd.PaymentRequestResponse, error) {
				return &payd.PaymentRequestResponse{
					Network: "testnet",
					Destinations: payd.P4Destination{
						Outputs: []payd.P4Output{{
							Amount: 1000,
							Script: "76a9146e912a2a1c28448522c1eba7d73ce0719b0636b388ac",
						}, {
							Amount: 2000,
							Script: "76a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac",
						}},
					},
					CreationTimestamp:   ts,
					ExpirationTimestamp: ts.Add(24 * time.Hour),
					Fee:                 fq,
					Memo:                "payment abc123",
					MerchantData: payd.User{
						ExtendedData: map[string]interface{}{
							"paymentReference": "abc123",
						},
						Name: "Merchant Person",
					},
					PaymentURL: "http://p4-merchant/api/v1/payment/abc123",
				}, nil
			},
			utxoReserveFunc: func() func(ctx context.Context, req payd.UTXOReserve) ([]payd.UTXO, error) {
				idx := 0
				utxos := []payd.UTXO{{
					Satoshis:       2000,
					LockingScript:  "76a9141a4cc80bc3ee6567cb37f9c5121841a5f8e0b87d88ac",
					DerivationPath: "2147483648/2147483648/2147483650",
					TxID:           "77505879a3ab90f2082614389dc971e79e5d75c553e5c4bf28d4f845f38f2b40",
					Vout:           1,
				}, {
					Satoshis:       2000,
					LockingScript:  "76a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac",
					DerivationPath: "2147483648/2147483648/2147483648",
					TxID:           "2e944402554bdfba1bac2fb94abf9b23985b258a3996aea6fcfe9a51c82500ac",
					Vout:           0,
				}}
				return func(ctx context.Context, req payd.UTXOReserve) ([]payd.UTXO, error) {
					defer func() { idx++ }()
					return utxos[idx : idx+1], nil
				}
			}(),
			envelopeFunc: func(ctx context.Context, args payd.EnvelopeArgs, req payd.PaymentRequestResponse) (*payd.Envelope, error) {
				return nil, errors.New("no envelope for you")
			},
			expDeficits:      []uint64{3039, 1113},
			expTx:            "0100000002402b8ff345f8d428bfc4e553c5755d9ee771c99d38142608f290aba379585077010000006a47304402206a9c351ba35f43b3b3c4eac4cbaade8ce3e405fa9e8c9cf7bd74df084ea4396d02202073e344b6e7c21c97a5e0b2beb7c667784208f50d306f87b37bac92658e1db1412102f46acbd7a9825d5220464b761b6477a600a50664a6b8765a77ec7e1b19e8f36bffffffffac0025c8519afefca6ae96398a255b98239bbf4ab92fac1bbadf4b550244942e000000006a47304402207b80e3da87295641ffb9dfd5823cfb9a74634174774cdd0fc88f1dcd541f3558022001c88cdaaf1b4c5b8dfaf80de7fd9de45b1df3a4ebafc3dd13e3b1ca4820bf4041210251c7b5806db15e127a986611aa23f71a84879acb2ceef610f9eabbf355790a29ffffffff03e8030000000000001976a9146e912a2a1c28448522c1eba7d73ce0719b0636b388acd0070000000000001976a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac1c030000000000001976a9148b1ca598db87cfe283229bf724ad39cc4f1a665788ac00000000",
			expKeyName:       "masterkey",
			expUTXOUnreserve: true,
			expErr:           errors.New("envelope creation failed for 'http://p4-merchant/api/v1/payment/abc123': no envelope for you"),
		},
		"error on payment send is reported": {
			req: payd.PayRequest{
				PayToURL: "http://p4-merchant/api/v1/payment/abc123",
			},
			paymentRequestFunc: func(ctx context.Context, req payd.PayRequest) (*payd.PaymentRequestResponse, error) {
				return &payd.PaymentRequestResponse{
					Network: "testnet",
					Destinations: payd.P4Destination{
						Outputs: []payd.P4Output{{
							Amount: 1000,
							Script: "76a9146e912a2a1c28448522c1eba7d73ce0719b0636b388ac",
						}, {
							Amount: 2000,
							Script: "76a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac",
						}},
					},
					CreationTimestamp:   ts,
					ExpirationTimestamp: ts.Add(24 * time.Hour),
					Fee:                 fq,
					Memo:                "payment abc123",
					MerchantData: payd.User{
						ExtendedData: map[string]interface{}{
							"paymentReference": "abc123",
						},
						Name: "Merchant Person",
					},
					PaymentURL: "http://p4-merchant/api/v1/payment/abc123",
				}, nil
			},
			utxoReserveFunc: func() func(ctx context.Context, req payd.UTXOReserve) ([]payd.UTXO, error) {
				idx := 0
				utxos := []payd.UTXO{{
					Satoshis:       2000,
					LockingScript:  "76a9141a4cc80bc3ee6567cb37f9c5121841a5f8e0b87d88ac",
					DerivationPath: "2147483648/2147483648/2147483650",
					TxID:           "77505879a3ab90f2082614389dc971e79e5d75c553e5c4bf28d4f845f38f2b40",
					Vout:           1,
				}, {
					Satoshis:       2000,
					LockingScript:  "76a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac",
					DerivationPath: "2147483648/2147483648/2147483648",
					TxID:           "2e944402554bdfba1bac2fb94abf9b23985b258a3996aea6fcfe9a51c82500ac",
					Vout:           0,
				}}
				return func(ctx context.Context, req payd.UTXOReserve) ([]payd.UTXO, error) {
					defer func() { idx++ }()
					return utxos[idx : idx+1], nil
				}
			}(),
			envelopeFunc: func(ctx context.Context, args payd.EnvelopeArgs, req payd.PaymentRequestResponse) (*payd.Envelope, error) {
				return &payd.Envelope{
					SPVEnvelope: &spv.Envelope{RawTx: "0100000002402b8ff345f8d428bfc4e553c5755d9ee771c99d38142608f290aba379585077010000006a47304402206a9c351ba35f43b3b3c4eac4cbaade8ce3e405fa9e8c9cf7bd74df084ea4396d02202073e344b6e7c21c97a5e0b2beb7c667784208f50d306f87b37bac92658e1db1412102f46acbd7a9825d5220464b761b6477a600a50664a6b8765a77ec7e1b19e8f36bffffffffac0025c8519afefca6ae96398a255b98239bbf4ab92fac1bbadf4b550244942e000000006a47304402207b80e3da87295641ffb9dfd5823cfb9a74634174774cdd0fc88f1dcd541f3558022001c88cdaaf1b4c5b8dfaf80de7fd9de45b1df3a4ebafc3dd13e3b1ca4820bf4041210251c7b5806db15e127a986611aa23f71a84879acb2ceef610f9eabbf355790a29ffffffff03e8030000000000001976a9146e912a2a1c28448522c1eba7d73ce0719b0636b388acd0070000000000001976a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac1c030000000000001976a9148b1ca598db87cfe283229bf724ad39cc4f1a665788ac00000000"},
				}, nil
			},
			paymentSendFunc: func(context.Context, payd.PayRequest, payd.PaymentSend) (*payd.PaymentACK, error) {
				return nil, errors.New("no send for you")
			},
			expDeficits:      []uint64{3039, 1113},
			expTx:            "0100000002402b8ff345f8d428bfc4e553c5755d9ee771c99d38142608f290aba379585077010000006a47304402206a9c351ba35f43b3b3c4eac4cbaade8ce3e405fa9e8c9cf7bd74df084ea4396d02202073e344b6e7c21c97a5e0b2beb7c667784208f50d306f87b37bac92658e1db1412102f46acbd7a9825d5220464b761b6477a600a50664a6b8765a77ec7e1b19e8f36bffffffffac0025c8519afefca6ae96398a255b98239bbf4ab92fac1bbadf4b550244942e000000006a47304402207b80e3da87295641ffb9dfd5823cfb9a74634174774cdd0fc88f1dcd541f3558022001c88cdaaf1b4c5b8dfaf80de7fd9de45b1df3a4ebafc3dd13e3b1ca4820bf4041210251c7b5806db15e127a986611aa23f71a84879acb2ceef610f9eabbf355790a29ffffffff03e8030000000000001976a9146e912a2a1c28448522c1eba7d73ce0719b0636b388acd0070000000000001976a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac1c030000000000001976a9148b1ca598db87cfe283229bf724ad39cc4f1a665788ac00000000",
			expUTXOUnreserve: true,
			expCallbackURL:   "https://myserver/api/v1/proofs/",
			expKeyName:       "masterkey",
			expErr:           errors.New("failed to send payment http://p4-merchant/api/v1/payment/abc123: no send for you"),
		},
		"error on change destination create is reported": {
			req: payd.PayRequest{
				PayToURL: "http://p4-merchant/api/v1/payment/abc123",
			},
			paymentRequestFunc: func(ctx context.Context, req payd.PayRequest) (*payd.PaymentRequestResponse, error) {
				return &payd.PaymentRequestResponse{
					Network: "testnet",
					Destinations: payd.P4Destination{
						Outputs: []payd.P4Output{{
							Amount: 1000,
							Script: "76a9146e912a2a1c28448522c1eba7d73ce0719b0636b388ac",
						}, {
							Amount: 2000,
							Script: "76a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac",
						}},
					},
					CreationTimestamp:   ts,
					ExpirationTimestamp: ts.Add(24 * time.Hour),
					Fee:                 fq,
					Memo:                "payment abc123",
					MerchantData: payd.User{
						ExtendedData: map[string]interface{}{
							"paymentReference": "abc123",
						},
						Name: "Merchant Person",
					},
					PaymentURL: "http://p4-merchant/api/v1/payment/abc123",
				}, nil
			},
			utxoReserveFunc: func() func(ctx context.Context, req payd.UTXOReserve) ([]payd.UTXO, error) {
				idx := 0
				utxos := []payd.UTXO{{
					Satoshis:       2000,
					LockingScript:  "76a9141a4cc80bc3ee6567cb37f9c5121841a5f8e0b87d88ac",
					DerivationPath: "2147483648/2147483648/2147483650",
					TxID:           "77505879a3ab90f2082614389dc971e79e5d75c553e5c4bf28d4f845f38f2b40",
					Vout:           1,
				}, {
					Satoshis:       2000,
					LockingScript:  "76a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac",
					DerivationPath: "2147483648/2147483648/2147483648",
					TxID:           "2e944402554bdfba1bac2fb94abf9b23985b258a3996aea6fcfe9a51c82500ac",
					Vout:           0,
				}}
				return func(ctx context.Context, req payd.UTXOReserve) ([]payd.UTXO, error) {
					defer func() { idx++ }()
					return utxos[idx : idx+1], nil
				}
			}(),
			envelopeFunc: func(ctx context.Context, args payd.EnvelopeArgs, req payd.PaymentRequestResponse) (*payd.Envelope, error) {
				return &payd.Envelope{
					SPVEnvelope: &spv.Envelope{RawTx: "0100000002402b8ff345f8d428bfc4e553c5755d9ee771c99d38142608f290aba379585077010000006a47304402206a9c351ba35f43b3b3c4eac4cbaade8ce3e405fa9e8c9cf7bd74df084ea4396d02202073e344b6e7c21c97a5e0b2beb7c667784208f50d306f87b37bac92658e1db1412102f46acbd7a9825d5220464b761b6477a600a50664a6b8765a77ec7e1b19e8f36bffffffffac0025c8519afefca6ae96398a255b98239bbf4ab92fac1bbadf4b550244942e000000006a47304402207b80e3da87295641ffb9dfd5823cfb9a74634174774cdd0fc88f1dcd541f3558022001c88cdaaf1b4c5b8dfaf80de7fd9de45b1df3a4ebafc3dd13e3b1ca4820bf4041210251c7b5806db15e127a986611aa23f71a84879acb2ceef610f9eabbf355790a29ffffffff03e8030000000000001976a9146e912a2a1c28448522c1eba7d73ce0719b0636b388acd0070000000000001976a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac1c030000000000001976a9148b1ca598db87cfe283229bf724ad39cc4f1a665788ac00000000"},
					Change: payd.Output{
						Satoshis:       796,
						DerivationPath: "2147483648/2147483723/2147483648",
						LockingScript:  internal.StringToScript("76a9148b1ca598db87cfe283229bf724ad39cc4f1a665788ac"),
					},
				}, nil
			},
			paymentSendFunc: func(context.Context, payd.PayRequest, payd.PaymentSend) (*payd.PaymentACK, error) {
				return &payd.PaymentACK{}, nil
			},
			destinationsCreateFunc: func(context.Context, payd.DestinationsCreateArgs, []payd.DestinationCreate) ([]payd.Output, error) {
				return nil, errors.New("no change for you")
			},
			expDeficits:     []uint64{3039, 1113},
			expTx:           "0100000002402b8ff345f8d428bfc4e553c5755d9ee771c99d38142608f290aba379585077010000006a47304402206a9c351ba35f43b3b3c4eac4cbaade8ce3e405fa9e8c9cf7bd74df084ea4396d02202073e344b6e7c21c97a5e0b2beb7c667784208f50d306f87b37bac92658e1db1412102f46acbd7a9825d5220464b761b6477a600a50664a6b8765a77ec7e1b19e8f36bffffffffac0025c8519afefca6ae96398a255b98239bbf4ab92fac1bbadf4b550244942e000000006a47304402207b80e3da87295641ffb9dfd5823cfb9a74634174774cdd0fc88f1dcd541f3558022001c88cdaaf1b4c5b8dfaf80de7fd9de45b1df3a4ebafc3dd13e3b1ca4820bf4041210251c7b5806db15e127a986611aa23f71a84879acb2ceef610f9eabbf355790a29ffffffff03e8030000000000001976a9146e912a2a1c28448522c1eba7d73ce0719b0636b388acd0070000000000001976a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac1c030000000000001976a9148b1ca598db87cfe283229bf724ad39cc4f1a665788ac00000000",
			expChangeCreate: true,
			expChange: payd.DestinationCreate{
				Satoshis:       796,
				DerivationPath: "2147483648/2147483723/2147483648",
				Script:         "76a9148b1ca598db87cfe283229bf724ad39cc4f1a665788ac",
				Keyname:        "masterkey",
			},
			expKeyName:     "masterkey",
			expCallbackURL: "https://myserver/api/v1/proofs/",
			expErr:         errors.New("failed to create destination for change output: no change for you"),
		},
		"error on tx create is reported": {
			req: payd.PayRequest{
				PayToURL: "http://p4-merchant/api/v1/payment/abc123",
			},
			paymentRequestFunc: func(ctx context.Context, req payd.PayRequest) (*payd.PaymentRequestResponse, error) {
				return &payd.PaymentRequestResponse{
					Network: "testnet",
					Destinations: payd.P4Destination{
						Outputs: []payd.P4Output{{
							Amount: 1000,
							Script: "76a9146e912a2a1c28448522c1eba7d73ce0719b0636b388ac",
						}, {
							Amount: 2000,
							Script: "76a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac",
						}},
					},
					CreationTimestamp:   ts,
					ExpirationTimestamp: ts.Add(24 * time.Hour),
					Fee:                 fq,
					Memo:                "payment abc123",
					MerchantData: payd.User{
						ExtendedData: map[string]interface{}{
							"paymentReference": "abc123",
						},
						Name: "Merchant Person",
					},
					PaymentURL: "http://p4-merchant/api/v1/payment/abc123",
				}, nil
			},
			utxoReserveFunc: func() func(ctx context.Context, req payd.UTXOReserve) ([]payd.UTXO, error) {
				idx := 0
				utxos := []payd.UTXO{{
					Satoshis:       2000,
					LockingScript:  "76a9141a4cc80bc3ee6567cb37f9c5121841a5f8e0b87d88ac",
					DerivationPath: "2147483648/2147483648/2147483650",
					TxID:           "77505879a3ab90f2082614389dc971e79e5d75c553e5c4bf28d4f845f38f2b40",
					Vout:           1,
				}, {
					Satoshis:       2000,
					LockingScript:  "76a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac",
					DerivationPath: "2147483648/2147483648/2147483648",
					TxID:           "2e944402554bdfba1bac2fb94abf9b23985b258a3996aea6fcfe9a51c82500ac",
					Vout:           0,
				}}
				return func(ctx context.Context, req payd.UTXOReserve) ([]payd.UTXO, error) {
					defer func() { idx++ }()
					return utxos[idx : idx+1], nil
				}
			}(),
			envelopeFunc: func(ctx context.Context, args payd.EnvelopeArgs, req payd.PaymentRequestResponse) (*payd.Envelope, error) {
				return &payd.Envelope{
					SPVEnvelope: &spv.Envelope{RawTx: "0100000002402b8ff345f8d428bfc4e553c5755d9ee771c99d38142608f290aba379585077010000006a47304402206a9c351ba35f43b3b3c4eac4cbaade8ce3e405fa9e8c9cf7bd74df084ea4396d02202073e344b6e7c21c97a5e0b2beb7c667784208f50d306f87b37bac92658e1db1412102f46acbd7a9825d5220464b761b6477a600a50664a6b8765a77ec7e1b19e8f36bffffffffac0025c8519afefca6ae96398a255b98239bbf4ab92fac1bbadf4b550244942e000000006a47304402207b80e3da87295641ffb9dfd5823cfb9a74634174774cdd0fc88f1dcd541f3558022001c88cdaaf1b4c5b8dfaf80de7fd9de45b1df3a4ebafc3dd13e3b1ca4820bf4041210251c7b5806db15e127a986611aa23f71a84879acb2ceef610f9eabbf355790a29ffffffff03e8030000000000001976a9146e912a2a1c28448522c1eba7d73ce0719b0636b388acd0070000000000001976a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac1c030000000000001976a9148b1ca598db87cfe283229bf724ad39cc4f1a665788ac00000000"},
					Change: payd.Output{
						Satoshis:       796,
						DerivationPath: "2147483648/2147483723/2147483648",
						LockingScript:  internal.StringToScript("76a9148b1ca598db87cfe283229bf724ad39cc4f1a665788ac"),
					},
				}, nil
			},
			paymentSendFunc: func(context.Context, payd.PayRequest, payd.PaymentSend) (*payd.PaymentACK, error) {
				return &payd.PaymentACK{}, nil
			},
			destinationsCreateFunc: func(context.Context, payd.DestinationsCreateArgs, []payd.DestinationCreate) ([]payd.Output, error) {
				return []payd.Output{{}}, nil
			},
			txCreateFunc: func(context.Context, payd.TransactionCreate) error {
				return errors.New("no tx for you")
			},
			expDeficits:     []uint64{3039, 1113},
			expTx:           "0100000002402b8ff345f8d428bfc4e553c5755d9ee771c99d38142608f290aba379585077010000006a47304402206a9c351ba35f43b3b3c4eac4cbaade8ce3e405fa9e8c9cf7bd74df084ea4396d02202073e344b6e7c21c97a5e0b2beb7c667784208f50d306f87b37bac92658e1db1412102f46acbd7a9825d5220464b761b6477a600a50664a6b8765a77ec7e1b19e8f36bffffffffac0025c8519afefca6ae96398a255b98239bbf4ab92fac1bbadf4b550244942e000000006a47304402207b80e3da87295641ffb9dfd5823cfb9a74634174774cdd0fc88f1dcd541f3558022001c88cdaaf1b4c5b8dfaf80de7fd9de45b1df3a4ebafc3dd13e3b1ca4820bf4041210251c7b5806db15e127a986611aa23f71a84879acb2ceef610f9eabbf355790a29ffffffff03e8030000000000001976a9146e912a2a1c28448522c1eba7d73ce0719b0636b388acd0070000000000001976a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac1c030000000000001976a9148b1ca598db87cfe283229bf724ad39cc4f1a665788ac00000000",
			expCallbackURL:  "https://myserver/api/v1/proofs/",
			expKeyName:      "masterkey",
			expChangeCreate: true,
			expChange: payd.DestinationCreate{
				Satoshis:       796,
				DerivationPath: "2147483648/2147483723/2147483648",
				Script:         "76a9148b1ca598db87cfe283229bf724ad39cc4f1a665788ac",
				Keyname:        "masterkey",
			},
			expErr: errors.New("failed to create transaction for change output: no tx for you"),
		},
		"error on utxo spend is reported": {
			req: payd.PayRequest{
				PayToURL: "http://p4-merchant/api/v1/payment/abc123",
			},
			paymentRequestFunc: func(ctx context.Context, req payd.PayRequest) (*payd.PaymentRequestResponse, error) {
				return &payd.PaymentRequestResponse{
					Network: "testnet",
					Destinations: payd.P4Destination{
						Outputs: []payd.P4Output{{
							Amount: 1000,
							Script: "76a9146e912a2a1c28448522c1eba7d73ce0719b0636b388ac",
						}, {
							Amount: 2000,
							Script: "76a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac",
						}},
					},
					CreationTimestamp:   ts,
					ExpirationTimestamp: ts.Add(24 * time.Hour),
					Fee:                 fq,
					Memo:                "payment abc123",
					MerchantData: payd.User{
						ExtendedData: map[string]interface{}{
							"paymentReference": "abc123",
						},
						Name: "Merchant Person",
					},
					PaymentURL: "http://p4-merchant/api/v1/payment/abc123",
				}, nil
			},
			utxoReserveFunc: func() func(ctx context.Context, req payd.UTXOReserve) ([]payd.UTXO, error) {
				idx := 0
				utxos := []payd.UTXO{{
					Satoshis:       2000,
					LockingScript:  "76a9141a4cc80bc3ee6567cb37f9c5121841a5f8e0b87d88ac",
					DerivationPath: "2147483648/2147483648/2147483650",
					TxID:           "77505879a3ab90f2082614389dc971e79e5d75c553e5c4bf28d4f845f38f2b40",
					Vout:           1,
				}, {
					Satoshis:       2000,
					LockingScript:  "76a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac",
					DerivationPath: "2147483648/2147483648/2147483648",
					TxID:           "2e944402554bdfba1bac2fb94abf9b23985b258a3996aea6fcfe9a51c82500ac",
					Vout:           0,
				}}
				return func(ctx context.Context, req payd.UTXOReserve) ([]payd.UTXO, error) {
					defer func() { idx++ }()
					return utxos[idx : idx+1], nil
				}
			}(),
			envelopeFunc: func(ctx context.Context, args payd.EnvelopeArgs, req payd.PaymentRequestResponse) (*payd.Envelope, error) {
				return &payd.Envelope{
					SPVEnvelope: &spv.Envelope{RawTx: "0100000002402b8ff345f8d428bfc4e553c5755d9ee771c99d38142608f290aba379585077010000006a47304402206a9c351ba35f43b3b3c4eac4cbaade8ce3e405fa9e8c9cf7bd74df084ea4396d02202073e344b6e7c21c97a5e0b2beb7c667784208f50d306f87b37bac92658e1db1412102f46acbd7a9825d5220464b761b6477a600a50664a6b8765a77ec7e1b19e8f36bffffffffac0025c8519afefca6ae96398a255b98239bbf4ab92fac1bbadf4b550244942e000000006a47304402207b80e3da87295641ffb9dfd5823cfb9a74634174774cdd0fc88f1dcd541f3558022001c88cdaaf1b4c5b8dfaf80de7fd9de45b1df3a4ebafc3dd13e3b1ca4820bf4041210251c7b5806db15e127a986611aa23f71a84879acb2ceef610f9eabbf355790a29ffffffff03e8030000000000001976a9146e912a2a1c28448522c1eba7d73ce0719b0636b388acd0070000000000001976a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac1c030000000000001976a9148b1ca598db87cfe283229bf724ad39cc4f1a665788ac00000000"},
					Change: payd.Output{
						Satoshis:       796,
						DerivationPath: "2147483648/2147483723/2147483648",
						LockingScript:  internal.StringToScript("76a9148b1ca598db87cfe283229bf724ad39cc4f1a665788ac"),
					},
				}, nil
			},
			paymentSendFunc: func(context.Context, payd.PayRequest, payd.PaymentSend) (*payd.PaymentACK, error) {
				return &payd.PaymentACK{}, nil
			},
			destinationsCreateFunc: func(context.Context, payd.DestinationsCreateArgs, []payd.DestinationCreate) ([]payd.Output, error) {
				return []payd.Output{{}}, nil
			},
			txCreateFunc: func(context.Context, payd.TransactionCreate) error {
				return nil
			},
			utxoSpendFunc: func(context.Context, payd.UTXOSpend) error {
				return errors.New("no spend for you")
			},
			expDeficits:     []uint64{3039, 1113},
			expTx:           "0100000002402b8ff345f8d428bfc4e553c5755d9ee771c99d38142608f290aba379585077010000006a47304402206a9c351ba35f43b3b3c4eac4cbaade8ce3e405fa9e8c9cf7bd74df084ea4396d02202073e344b6e7c21c97a5e0b2beb7c667784208f50d306f87b37bac92658e1db1412102f46acbd7a9825d5220464b761b6477a600a50664a6b8765a77ec7e1b19e8f36bffffffffac0025c8519afefca6ae96398a255b98239bbf4ab92fac1bbadf4b550244942e000000006a47304402207b80e3da87295641ffb9dfd5823cfb9a74634174774cdd0fc88f1dcd541f3558022001c88cdaaf1b4c5b8dfaf80de7fd9de45b1df3a4ebafc3dd13e3b1ca4820bf4041210251c7b5806db15e127a986611aa23f71a84879acb2ceef610f9eabbf355790a29ffffffff03e8030000000000001976a9146e912a2a1c28448522c1eba7d73ce0719b0636b388acd0070000000000001976a914e6e4fa093b7146a4a36fca4b1305182fafa7a9a288ac1c030000000000001976a9148b1ca598db87cfe283229bf724ad39cc4f1a665788ac00000000",
			expCallbackURL:  "https://myserver/api/v1/proofs/",
			expChangeCreate: true,
			expChange: payd.DestinationCreate{
				Satoshis:       796,
				DerivationPath: "2147483648/2147483723/2147483648",
				Script:         "76a9148b1ca598db87cfe283229bf724ad39cc4f1a665788ac",
				Keyname:        "masterkey",
			},
			expKeyName: "masterkey",
			expErr:     errors.New("failed to mark utxos as spent: no spend for you"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var utxoUnreserveCalled, changeCreated bool
			svc := service.NewPayService(
				&mocks.TxoWriterMock{
					UTXOUnreserveFunc: func(context.Context, payd.UTXOUnreserve) error {
						utxoUnreserveCalled = true
						return nil
					},
					UTXOReserveFunc: func() func(ctx context.Context, req payd.UTXOReserve) ([]payd.UTXO, error) {
						idx := 0
						return func(ctx context.Context, req payd.UTXOReserve) ([]payd.UTXO, error) {
							defer func() { idx++ }()
							assert.Equal(t, test.expDeficits[idx], req.Satoshis)
							return test.utxoReserveFunc(ctx, req)
						}
					}(),
					UTXOSpendFunc: test.utxoSpendFunc,
				},
				&mocks.TransactionWriterMock{
					TransactionCreateFunc: test.txCreateFunc,
				},
				&mocks.DestinationsReaderWriterMock{
					DestinationsCreateFunc: func(ctx context.Context, args payd.DestinationsCreateArgs, req []payd.DestinationCreate) ([]payd.Output, error) {
						changeCreated = true
						assert.Equal(t, test.expChange, req[0])
						return test.destinationsCreateFunc(ctx, args, req)
					},
				},
				&mocks.P4Mock{
					PaymentRequestFunc: test.paymentRequestFunc,
					PaymentSendFunc: func(ctx context.Context, req payd.PayRequest, args payd.PaymentSend) (*payd.PaymentACK, error) {
						_, ok := args.ProofCallbacks[test.expCallbackURL]
						assert.True(t, ok, "%s not in %+v", test.expCallbackURL, args.ProofCallbacks)
						return test.paymentSendFunc(ctx, req, args)
					},
				},
				&mocks.EnvelopeServiceMock{EnvelopeFunc: test.envelopeFunc},
				&config.Server{Hostname: "myserver"},
			)

			_, err := svc.Pay(context.TODO(), test.req)
			if test.expErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, test.expErr.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, test.expUTXOUnreserve, utxoUnreserveCalled)
			assert.Equal(t, test.expChangeCreate, changeCreated)
		})
	}
}
