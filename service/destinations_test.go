package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/payd"
	"github.com/libsv/payd/config"
	"github.com/libsv/payd/internal"
	"github.com/libsv/payd/mocks"
	"github.com/libsv/payd/service"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v3"
)

func TestDestinationService_DestinationsCreate(t *testing.T) {
	t.Parallel()

	fq := bt.NewFeeQuote()
	destinationsToOutputs := func(ctx context.Context, args payd.DestinationsCreateArgs, dests []payd.DestinationCreate) ([]payd.Output, error) {
		oo := make([]payd.Output, len(dests))
		for i, dest := range dests {
			oo[i] = payd.Output{
				LockingScript: func() *bscript.Script {
					s, _ := bscript.NewFromHexString(dest.Script)
					return s
				}(),
				Satoshis:       dest.Satoshis,
				DerivationPath: dest.DerivationPath,
				State:          "pending",
			}
		}
		return oo, nil
	}
	tests := map[string]struct {
		req                      payd.DestinationsCreate
		derivationPathExistsFunc func(context.Context, payd.DerivationExistsArgs) (bool, error)
		destinationsCreateFunc   func(context.Context, payd.DestinationsCreateArgs, []payd.DestinationCreate) ([]payd.Output, error)
		privateKeyFunc           func(context.Context, string) (*bip32.ExtendedKey, error)
		feesFunc                 func(context.Context, string) (*bt.FeeQuote, error)
		uint64Func               func() (uint64, error)
		expErr                   error
		expDests                 []payd.DestinationCreate
		expDestination           *payd.Destination
		expDerivationChecks      int
	}{
		"successful create": {
			req: payd.DestinationsCreate{
				InvoiceID: null.StringFrom("abc123"),
				Satoshis:  1000,
			},
			uint64Func: func() (uint64, error) {
				return 0, nil
			},
			derivationPathExistsFunc: func(context.Context, payd.DerivationExistsArgs) (bool, error) {
				return false, nil
			},
			destinationsCreateFunc: destinationsToOutputs,
			feesFunc: func(context.Context, string) (*bt.FeeQuote, error) {
				return fq, nil
			},
			expDests: []payd.DestinationCreate{{
				Satoshis:       1000,
				Script:         "76a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac",
				DerivationPath: "2147483648/2147483648/2147483648",
				Keyname:        "masterkey",
			}},
			expDestination: &payd.Destination{
				Outputs: []payd.Output{{
					LockingScript: func() *bscript.Script {
						s, _ := bscript.NewFromHexString("76a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac")
						return s
					}(),
					Satoshis:       1000,
					DerivationPath: "2147483648/2147483648/2147483648",
					State:          "pending",
				}},
			},
			expDerivationChecks: 1,
		},
		"success create after derivation path collision": {
			req: payd.DestinationsCreate{
				InvoiceID: null.StringFrom("abc123"),
				Satoshis:  1000,
			},
			uint64Func: func() func() (uint64, error) {
				itr := uint64(0)
				return func() (uint64, error) {
					defer func() { itr++ }()
					return itr, nil
				}
			}(),
			derivationPathExistsFunc: func(ctx context.Context, args payd.DerivationExistsArgs) (bool, error) {
				n, err := bip32.DeriveNumber(args.Path)
				return n < 2, err
			},
			destinationsCreateFunc: destinationsToOutputs,
			feesFunc: func(context.Context, string) (*bt.FeeQuote, error) {
				return fq, nil
			},
			expDests: []payd.DestinationCreate{{
				Satoshis:       1000,
				Script:         "76a9141a4cc80bc3ee6567cb37f9c5121841a5f8e0b87d88ac",
				DerivationPath: "2147483648/2147483648/2147483650",
				Keyname:        "masterkey",
			}},
			expDestination: &payd.Destination{
				Outputs: []payd.Output{{
					LockingScript:  internal.StringToScript("76a9141a4cc80bc3ee6567cb37f9c5121841a5f8e0b87d88ac"),
					Satoshis:       1000,
					DerivationPath: "2147483648/2147483648/2147483650",
					State:          "pending",
				}},
			},
			expDerivationChecks: 3,
		},
		"error on private key get is reported": {
			req: payd.DestinationsCreate{
				InvoiceID: null.StringFrom("abc123"),
				Satoshis:  1000,
			},
			uint64Func: func() (uint64, error) {
				return 0, nil
			},
			privateKeyFunc: func(context.Context, string) (*bip32.ExtendedKey, error) {
				return nil, errors.New("denied")
			},
			derivationPathExistsFunc: func(ctx context.Context, args payd.DerivationExistsArgs) (bool, error) {
				return false, nil
			},
			destinationsCreateFunc: destinationsToOutputs,
			feesFunc: func(context.Context, string) (*bt.FeeQuote, error) {
				return fq, nil
			},
			expErr: errors.New("denied"),
		},
		"error on seed is reported": {
			req: payd.DestinationsCreate{
				InvoiceID: null.StringFrom("abc123"),
				Satoshis:  1000,
			},
			uint64Func: func() (uint64, error) {
				return 0, errors.New("no seed 4 u")
			},
			derivationPathExistsFunc: func(ctx context.Context, args payd.DerivationExistsArgs) (bool, error) {
				return false, nil
			},
			destinationsCreateFunc: destinationsToOutputs,
			feesFunc: func(context.Context, string) (*bt.FeeQuote, error) {
				return fq, nil
			},
			expErr: errors.New("failed to create seed for derivation path: no seed 4 u"),
		},
		"error on derivation path existence check is reported": {
			req: payd.DestinationsCreate{
				InvoiceID: null.StringFrom("abc123"),
				Satoshis:  1000,
			},
			uint64Func: func() (uint64, error) {
				return 0, nil
			},
			derivationPathExistsFunc: func(ctx context.Context, args payd.DerivationExistsArgs) (bool, error) {
				return false, errors.New("flip a coin")
			},
			destinationsCreateFunc: destinationsToOutputs,
			feesFunc: func(context.Context, string) (*bt.FeeQuote, error) {
				return fq, nil
			},
			expErr:              errors.New("failed to check derivation path exists when creating new destination: flip a coin"),
			expDerivationChecks: 1,
		},
		"error created desintations is reported": {
			req: payd.DestinationsCreate{
				InvoiceID: null.StringFrom("abc123"),
				Satoshis:  1000,
			},
			uint64Func: func() (uint64, error) {
				return 0, nil
			},
			derivationPathExistsFunc: func(ctx context.Context, args payd.DerivationExistsArgs) (bool, error) {
				return false, nil
			},
			destinationsCreateFunc: func(ctx context.Context, args payd.DestinationsCreateArgs, dests []payd.DestinationCreate) ([]payd.Output, error) {
				return nil, errors.New("finaldestination")
			},
			feesFunc: func(context.Context, string) (*bt.FeeQuote, error) {
				return fq, nil
			},
			expDests: []payd.DestinationCreate{{
				Satoshis:       1000,
				Script:         "76a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac",
				DerivationPath: "2147483648/2147483648/2147483648",
				Keyname:        "masterkey",
			}},
			expErr:              errors.New("failed to store destinations: finaldestination"),
			expDerivationChecks: 1,
		},
		"satoshis below dust limit rejected": {
			req: payd.DestinationsCreate{
				Satoshis: 100,
			},
			destinationsCreateFunc: destinationsToOutputs,
			expErr:                 errors.New("[satoshis: value 100 is smaller than minimum 136]"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var derivationChecks int
			svc := service.NewDestinationsService(
				nil,
				&mocks.PrivateKeyServiceMock{
					PrivateKeyFunc: func(ctx context.Context, name string) (*bip32.ExtendedKey, error) {
						if test.privateKeyFunc != nil {
							return test.privateKeyFunc(ctx, name)
						}
						return bip32.NewKeyFromString("tprv8ZgxMBicQKsPcvvcLrg1PVzjNhVpU1ckb294dNKSYZ4YY4CwLfL9v3gzuW5WY96Cg7Wu58t7bukEezWFKzKapc4gJriYwgSYcHaN2VrTRKP")
					},
				},
				&mocks.DestinationsReaderWriterMock{
					DestinationsCreateFunc: func(ctx context.Context, args payd.DestinationsCreateArgs, dests []payd.DestinationCreate) ([]payd.Output, error) {
						assert.Equal(t, test.expDests, dests)
						return test.destinationsCreateFunc(ctx, args, dests)
					},
				},
				&mocks.DerivationReaderMock{
					DerivationPathExistsFunc: func(ctx context.Context, args payd.DerivationExistsArgs) (bool, error) {
						derivationChecks++
						return test.derivationPathExistsFunc(ctx, args)
					},
				},
				nil,
				&mocks.SeedServiceMock{
					Uint64Func: test.uint64Func,
				},
			)

			dests, err := svc.DestinationsCreate(context.TODO(), test.req)
			assert.Equal(t, test.expDerivationChecks, derivationChecks)
			if test.expErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, test.expErr.Error())
			} else {
				assert.NoError(t, err)
			}

			if test.expDestination != nil {
				assert.NotNil(t, dests)
				assert.Equal(t, test.expDestination, dests)
			} else {
				assert.Nil(t, dests)
			}
		})
	}
}

func TestDestinationService_Destinations(t *testing.T) {
	ts := time.Now().UTC()
	fq := bt.NewFeeQuote()
	tests := map[string]struct {
		args             payd.DestinationsArgs
		cfg              *config.Wallet
		invoiceFunc      func(context.Context, payd.InvoiceArgs) (*payd.Invoice, error)
		destinationsFunc func(context.Context, payd.DestinationsArgs) ([]payd.Output, error)
		feesFunc         func(context.Context, string) (*bt.FeeQuote, error)
		expErr           error
		expDestination   *payd.Destination
	}{
		"successful destinations network get": {
			cfg: &config.Wallet{
				Network: config.NetworkMainet,
			},
			args: payd.DestinationsArgs{
				InvoiceID: "abc123",
			},
			invoiceFunc: func(ctx context.Context, args payd.InvoiceArgs) (*payd.Invoice, error) {
				return &payd.Invoice{
					SPVRequired: true,
					Satoshis:    1000,
					ExpiresAt:   null.TimeFrom(ts.Add(time.Hour * 24)),
					MetaData: payd.MetaData{
						CreatedAt: ts,
					},
				}, nil
			},
			destinationsFunc: func(ctx context.Context, args payd.DestinationsArgs) ([]payd.Output, error) {
				return []payd.Output{{
					Satoshis: 1000,
					LockingScript: func() *bscript.Script {
						s, _ := bscript.NewFromHexString("76a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac")
						return s
					}(),
				}}, nil
			},
			feesFunc: func(context.Context, string) (*bt.FeeQuote, error) {
				return fq, nil
			},
			expDestination: &payd.Destination{
				Network:     string(config.NetworkMainet),
				SPVRequired: true,
				Outputs: []payd.Output{{
					Satoshis: 1000,
					LockingScript: func() *bscript.Script {
						s, _ := bscript.NewFromHexString("76a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac")
						return s
					}(),
				}},
				CreatedAt: ts,
				ExpiresAt: ts.Add(time.Hour * 24),
			},
		},
		"successful destinations network get on testnet": {
			cfg: &config.Wallet{
				Network: config.NetworkTestnet,
			},
			args: payd.DestinationsArgs{
				InvoiceID: "abc123",
			},
			invoiceFunc: func(ctx context.Context, args payd.InvoiceArgs) (*payd.Invoice, error) {
				return &payd.Invoice{
					SPVRequired: true,
					Satoshis:    1000,
					ExpiresAt:   null.TimeFrom(ts.Add(time.Hour * 24)),
					MetaData: payd.MetaData{
						CreatedAt: ts,
					},
				}, nil
			},
			destinationsFunc: func(ctx context.Context, args payd.DestinationsArgs) ([]payd.Output, error) {
				return []payd.Output{{
					Satoshis: 1000,
					LockingScript: func() *bscript.Script {
						s, _ := bscript.NewFromHexString("76a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac")
						return s
					}(),
				}}, nil
			},
			feesFunc: func(context.Context, string) (*bt.FeeQuote, error) {
				return fq, nil
			},
			expDestination: &payd.Destination{
				Network:     string(config.NetworkTestnet),
				SPVRequired: true,
				Outputs: []payd.Output{{
					Satoshis: 1000,
					LockingScript: func() *bscript.Script {
						s, _ := bscript.NewFromHexString("76a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac")
						return s
					}(),
				}},
				CreatedAt: ts,
				ExpiresAt: ts.Add(time.Hour * 24),
			},
		},
		"successful destinations network get spv not required": {
			cfg: &config.Wallet{
				Network: config.NetworkMainet,
			},
			args: payd.DestinationsArgs{
				InvoiceID: "abc123",
			},
			invoiceFunc: func(ctx context.Context, args payd.InvoiceArgs) (*payd.Invoice, error) {
				return &payd.Invoice{
					Satoshis:  1000,
					ExpiresAt: null.TimeFrom(ts.Add(time.Hour * 24)),
					MetaData: payd.MetaData{
						CreatedAt: ts,
					},
				}, nil
			},
			destinationsFunc: func(ctx context.Context, args payd.DestinationsArgs) ([]payd.Output, error) {
				return []payd.Output{{
					Satoshis: 1000,
					LockingScript: func() *bscript.Script {
						s, _ := bscript.NewFromHexString("76a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac")
						return s
					}(),
				}}, nil
			},
			feesFunc: func(context.Context, string) (*bt.FeeQuote, error) {
				return fq, nil
			},
			expDestination: &payd.Destination{
				Network: string(config.NetworkMainet),
				Outputs: []payd.Output{{
					Satoshis: 1000,
					LockingScript: func() *bscript.Script {
						s, _ := bscript.NewFromHexString("76a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac")
						return s
					}(),
				}},
				CreatedAt: ts,
				ExpiresAt: ts.Add(time.Hour * 24),
			},
		},
		"successful get with 2 hr expiry": {
			cfg: &config.Wallet{
				Network: config.NetworkMainet,
			},
			args: payd.DestinationsArgs{
				InvoiceID: "abc123",
			},
			invoiceFunc: func(ctx context.Context, args payd.InvoiceArgs) (*payd.Invoice, error) {
				return &payd.Invoice{
					Satoshis:  1000,
					ExpiresAt: null.TimeFrom(ts.Add(time.Hour * 2)),
					MetaData: payd.MetaData{
						CreatedAt: ts,
					},
				}, nil
			},
			destinationsFunc: func(ctx context.Context, args payd.DestinationsArgs) ([]payd.Output, error) {
				return []payd.Output{{
					Satoshis: 1000,
					LockingScript: func() *bscript.Script {
						s, _ := bscript.NewFromHexString("76a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac")
						return s
					}(),
				}}, nil
			},
			feesFunc: func(context.Context, string) (*bt.FeeQuote, error) {
				return fq, nil
			},
			expDestination: &payd.Destination{
				Network: string(config.NetworkMainet),
				Outputs: []payd.Output{{
					Satoshis: 1000,
					LockingScript: func() *bscript.Script {
						s, _ := bscript.NewFromHexString("76a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac")
						return s
					}(),
				}},
				CreatedAt: ts,
				ExpiresAt: ts.Add(time.Hour * 2),
			},
		},
		"error with invoice is reported": {
			cfg: &config.Wallet{
				Network: config.NetworkMainet,
			},
			args: payd.DestinationsArgs{
				InvoiceID: "abc123",
			},
			invoiceFunc: func(ctx context.Context, args payd.InvoiceArgs) (*payd.Invoice, error) {
				return nil, errors.New("outsilent")
			},
			destinationsFunc: func(ctx context.Context, args payd.DestinationsArgs) ([]payd.Output, error) {
				return []payd.Output{{
					Satoshis: 1000,
					LockingScript: func() *bscript.Script {
						s, _ := bscript.NewFromHexString("76a91474b0424726ca510399c1eb5c8374f974c68b2fa388ac")
						return s
					}(),
				}}, nil
			},
			feesFunc: func(context.Context, string) (*bt.FeeQuote, error) {
				return fq, nil
			},
			expErr: errors.New("failed to get invoice for invoiceID 'abc123' when getting destinations: outsilent"),
		},
		"error with destinations is reported": {
			cfg: &config.Wallet{
				Network: config.NetworkMainet,
			},
			args: payd.DestinationsArgs{
				InvoiceID: "abc123",
			},
			invoiceFunc: func(ctx context.Context, args payd.InvoiceArgs) (*payd.Invoice, error) {
				return &payd.Invoice{
					SPVRequired: true,
					Satoshis:    1000,
					ExpiresAt:   null.TimeFrom(ts.Add(time.Hour * 24)),
					MetaData: payd.MetaData{
						CreatedAt: ts,
					},
				}, nil
			},
			destinationsFunc: func(ctx context.Context, args payd.DestinationsArgs) ([]payd.Output, error) {
				return nil, errors.New("destination unknown")
			},
			feesFunc: func(context.Context, string) (*bt.FeeQuote, error) {
				return fq, nil
			},
			expErr: errors.New("failed to read destinations for invoiceID 'abc123': destination unknown"),
		},
		"invalid args are rejected": {
			args:   payd.DestinationsArgs{},
			expErr: errors.New("[invoiceID: value cannot be empty]"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			svc := service.NewDestinationsService(
				test.cfg,
				nil,
				&mocks.DestinationsReaderWriterMock{
					DestinationsFunc: test.destinationsFunc,
				},
				nil,
				&mocks.InvoiceReaderWriterMock{
					InvoiceFunc: test.invoiceFunc,
				},
				nil,
			)

			dests, err := svc.Destinations(context.Background(), test.args)
			if test.expErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, test.expErr.Error())
			} else {
				assert.NoError(t, err)
			}

			if test.expDestination != nil {
				assert.NotNil(t, dests)
				assert.Equal(t, *test.expDestination, *dests)
			} else {
				assert.Nil(t, dests)
			}
		})
	}
}
