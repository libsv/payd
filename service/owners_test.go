package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/libsv/payd"
	"github.com/libsv/payd/mocks"
	"github.com/libsv/payd/service"
	"github.com/stretchr/testify/assert"
)

func TestOwnerService_Owner(t *testing.T) {
	tests := map[string]struct {
		ownerFunc func(context.Context) (*payd.User, error)
		expErr    error
	}{
		"no error reported": {
			ownerFunc: func(context.Context) (*payd.User, error) {
				return nil, nil
			},
		},
		"error reported": {
			ownerFunc: func(context.Context) (*payd.User, error) {
				return nil, errors.New("no one here")
			},
			expErr: errors.New("no one here"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			svc := service.NewOwnerService(&mocks.OwnerStoreMock{
				OwnerFunc: test.ownerFunc,
			})

			_, err := svc.Owner(context.TODO())
			assert.Equal(t, test.expErr, err)
		})
	}
}
