package service_test

import (
	"context"
	"testing"

	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/payd"
	"github.com/libsv/payd/mocks"
	"github.com/libsv/payd/service"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestPrivateKeyService_Create(t *testing.T) {
	tests := map[string]struct {
		privateKeyFunc            func(context.Context, payd.KeyArgs) (*payd.PrivateKey, error)
		privateKeyCreateFunc      func(context.Context, payd.PrivateKey) (*payd.PrivateKey, error)
		keyname                   string
		expPrivateKeyCreateCalled bool
		expErr                    error
	}{
		"successful create": {
			privateKeyFunc: func(ctx context.Context, args payd.KeyArgs) (*payd.PrivateKey, error) {
				return nil, nil
			},
			privateKeyCreateFunc: func(ctx context.Context, args payd.PrivateKey) (*payd.PrivateKey, error) {
				_, err := bip32.NewKeyFromString(args.Xprv)
				return nil, err
			},
			keyname:                   "test",
			expPrivateKeyCreateCalled: true,
		},
		"don't create on key if already exists": {
			privateKeyFunc: func(ctx context.Context, args payd.KeyArgs) (*payd.PrivateKey, error) {
				return &payd.PrivateKey{}, nil
			},
			privateKeyCreateFunc: func(ctx context.Context, args payd.PrivateKey) (*payd.PrivateKey, error) {
				_, err := bip32.NewKeyFromString(args.Xprv)
				return nil, err
			},
			keyname: "test",
		},
		"error on key check is reported": {
			privateKeyFunc: func(ctx context.Context, args payd.KeyArgs) (*payd.PrivateKey, error) {
				return nil, errors.New("key go bye bye")
			},
			privateKeyCreateFunc: func(ctx context.Context, args payd.PrivateKey) (*payd.PrivateKey, error) {
				_, err := bip32.NewKeyFromString(args.Xprv)
				return nil, err
			},
			keyname: "test",
			expErr:  errors.New("failed to get key test by name: key go bye bye"),
		},
		"error on key reate is reported": {
			privateKeyFunc: func(ctx context.Context, args payd.KeyArgs) (*payd.PrivateKey, error) {
				return nil, nil
			},
			privateKeyCreateFunc: func(ctx context.Context, args payd.PrivateKey) (*payd.PrivateKey, error) {
				return nil, errors.New("oh wow")
			},
			keyname:                   "test",
			expPrivateKeyCreateCalled: true,
			expErr:                    errors.New("failed to create private key: oh wow"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var createPrivateKeyCalled bool
			svc := service.NewPrivateKeys(&mocks.PrivateKeyReaderWriterMock{
				PrivateKeyFunc: test.privateKeyFunc,
				PrivateKeyCreateFunc: func(ctx context.Context, args payd.PrivateKey) (*payd.PrivateKey, error) {
					createPrivateKeyCalled = true
					assert.Equal(t, test.keyname, args.Name)
					return test.privateKeyCreateFunc(ctx, args)
				},
			}, false)

			err := svc.Create(context.Background(), test.keyname)
			if test.expErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, test.expErr.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, test.expPrivateKeyCreateCalled, createPrivateKeyCalled)
		})
	}
}

func TestPrivateKeyService_PrivateKey(t *testing.T) {
	tests := map[string]struct {
		privateKeyFunc func(context.Context, payd.KeyArgs) (*payd.PrivateKey, error)
		keyname        string
		expKey         string
		expErr         error
	}{
		"successful get": {
			privateKeyFunc: func(ctx context.Context, args payd.KeyArgs) (*payd.PrivateKey, error) {
				return &payd.PrivateKey{
					Xprv: "tprv8ZgxMBicQKsPcvvcLrg1PVzjNhVpU1ckb294dNKSYZ4YY4CwLfL9v3gzuW5WY96Cg7Wu58t7bukEezWFKzKapc4gJriYwgSYcHaN2VrTRKP",
				}, nil
			},
			keyname: "test",
			expKey:  "tprv8ZgxMBicQKsPcvvcLrg1PVzjNhVpU1ckb294dNKSYZ4YY4CwLfL9v3gzuW5WY96Cg7Wu58t7bukEezWFKzKapc4gJriYwgSYcHaN2VrTRKP",
		},
		"key not found errors": {
			privateKeyFunc: func(ctx context.Context, args payd.KeyArgs) (*payd.PrivateKey, error) {
				return nil, nil
			},
			keyname: "menoexist",
			expErr:  errors.New("key not found"),
		},
		"error on key get is reported": {
			privateKeyFunc: func(ctx context.Context, args payd.KeyArgs) (*payd.PrivateKey, error) {
				return nil, errors.New("key go bye bye")
			},
			keyname: "meerr",
			expErr:  errors.New("failed to get key meerr by name: key go bye bye"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			svc := service.NewPrivateKeys(&mocks.PrivateKeyReaderWriterMock{
				PrivateKeyFunc: test.privateKeyFunc,
			}, false)

			key, err := svc.PrivateKey(context.Background(), test.keyname)
			if test.expErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, test.expErr.Error())
			} else {
				assert.NoError(t, err)
			}

			if test.expKey != "" {
				assert.Equal(t, test.expKey, key.String())
			} else {
				assert.Nil(t, key)
			}
		})
	}
}
