package service_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/libsv/payd"
	"github.com/libsv/payd/config"
	"github.com/libsv/payd/mocks"
	"github.com/libsv/payd/service"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v3"
)

func TestInvoiceService_Invoice(t *testing.T) {
	tests := map[string]struct {
		invoiceFunc func(context.Context, payd.InvoiceArgs) (*payd.Invoice, error)
		args        payd.InvoiceArgs
		expErr      error
	}{
		"successful invoice get": {
			invoiceFunc: func(context.Context, payd.InvoiceArgs) (*payd.Invoice, error) {
				return nil, nil
			},
			args: payd.InvoiceArgs{
				InvoiceID: "abc123",
			},
		},
		"invalid invoice args rejected": {
			invoiceFunc: func(context.Context, payd.InvoiceArgs) (*payd.Invoice, error) {
				return nil, nil
			},
			expErr: errors.New("[invoiceID: value must be between 1 and 30 characters]"),
		},
		"store error is reported": {
			invoiceFunc: func(context.Context, payd.InvoiceArgs) (*payd.Invoice, error) {
				return nil, errors.New("whoopsie")
			},
			args: payd.InvoiceArgs{
				InvoiceID: "def123",
			},
			expErr: errors.New("failed to get invoice with id def123: whoopsie"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			svc := service.NewInvoice(nil, nil, &mocks.InvoiceReaderWriterMock{
				InvoiceFunc: test.invoiceFunc,
			}, nil, nil, nil)
			_, err := svc.Invoice(context.TODO(), test.args)
			if test.expErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, test.expErr.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestInvoiceService_Invoices(t *testing.T) {
	tests := map[string]struct {
		invoicesFunc func(context.Context) ([]payd.Invoice, error)
		expErr       error
	}{
		"successful invoices get": {
			invoicesFunc: func(context.Context) ([]payd.Invoice, error) {
				return nil, nil
			},
		},
		"store error is reported": {
			invoicesFunc: func(context.Context) ([]payd.Invoice, error) {
				return nil, errors.New("whoopsie")
			},
			expErr: errors.New("failed to get invoices: whoopsie"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			svc := service.NewInvoice(nil, nil, &mocks.InvoiceReaderWriterMock{
				InvoicesFunc: test.invoicesFunc,
			}, nil, nil, nil)
			_, err := svc.Invoices(context.TODO())
			if test.expErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, test.expErr.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestInvoiceService_Create(t *testing.T) {
	now := time.Now().UTC()
	tests := map[string]struct {
		nanosecondFunc         func() int
		nowUTCFunc             func() time.Time
		invoiceCreateFunc      func(context.Context, payd.InvoiceCreate) (*payd.Invoice, error)
		destinationsCreateFunc func(context.Context, payd.DestinationsCreate) (*payd.Destination, error)
		commitFunc             func(context.Context) error
		req                    payd.InvoiceCreate
		cfg                    *config.Server
		expReq                 payd.InvoiceCreate
		expErr                 error
	}{
		"successful invoice create": {
			nanosecondFunc: func() int {
				return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC).Nanosecond()
			},
			nowUTCFunc: func() time.Time {
				return now
			},
			commitFunc: func(context.Context) error {
				return nil
			},
			destinationsCreateFunc: func(context.Context, payd.DestinationsCreate) (*payd.Destination, error) {
				return nil, nil
			},
			invoiceCreateFunc: func(context.Context, payd.InvoiceCreate) (*payd.Invoice, error) {
				return nil, nil
			},
			req: payd.InvoiceCreate{
				Satoshis:    2000,
				Description: null.StringFrom("my cool invoice"),
				Reference:   null.StringFrom("sick"),
			},
			cfg: &config.Server{
				Hostname: "ohwow",
			},
			expReq: payd.InvoiceCreate{
				InvoiceID:   "pJ",
				SPVRequired: true,
				Satoshis:    2000,
				Description: null.StringFrom("my cool invoice"),
				Reference:   null.StringFrom("sick"),
				CreatedAt:   now,
				ExpiresAt:   null.TimeFrom(now.Add(time.Hour * 24)),
			},
		},
		"invoice below 1000 satsohis spv not required": {
			nanosecondFunc: func() int {
				return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC).Nanosecond()
			},
			nowUTCFunc: func() time.Time {
				return now
			},
			commitFunc: func(context.Context) error {
				return nil
			},
			destinationsCreateFunc: func(context.Context, payd.DestinationsCreate) (*payd.Destination, error) {
				return nil, nil
			},
			invoiceCreateFunc: func(context.Context, payd.InvoiceCreate) (*payd.Invoice, error) {
				return nil, nil
			},
			req: payd.InvoiceCreate{
				Satoshis:    999,
				Description: null.StringFrom("my cool invoice"),
				Reference:   null.StringFrom("sick"),
			},
			cfg: &config.Server{
				Hostname: "ohwow",
			},
			expReq: payd.InvoiceCreate{
				InvoiceID:   "Kr",
				SPVRequired: false,
				Satoshis:    999,
				Description: null.StringFrom("my cool invoice"),
				Reference:   null.StringFrom("sick"),
				CreatedAt:   now,
				ExpiresAt:   null.TimeFrom(now.Add(time.Hour * 24)),
			},
		},
		"custom expiry is applied": {
			nanosecondFunc: func() int {
				return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC).Nanosecond()
			},
			nowUTCFunc: func() time.Time {
				return now
			},
			commitFunc: func(context.Context) error {
				return nil
			},
			destinationsCreateFunc: func(context.Context, payd.DestinationsCreate) (*payd.Destination, error) {
				return nil, nil
			},
			invoiceCreateFunc: func(context.Context, payd.InvoiceCreate) (*payd.Invoice, error) {
				return nil, nil
			},
			req: payd.InvoiceCreate{
				Satoshis:    2000,
				Description: null.StringFrom("my cool invoice"),
				Reference:   null.StringFrom("sick"),
				ExpiresAt:   null.TimeFrom(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
			},
			cfg: &config.Server{
				Hostname: "ohwow",
			},
			expReq: payd.InvoiceCreate{
				InvoiceID:   "nY",
				SPVRequired: true,
				Satoshis:    2000,
				Description: null.StringFrom("my cool invoice"),
				Reference:   null.StringFrom("sick"),
				CreatedAt:   now,
				ExpiresAt:   null.TimeFrom(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
			},
		},
		"error on invoice create is reported": {
			nanosecondFunc: func() int {
				return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC).Nanosecond()
			},
			nowUTCFunc: func() time.Time {
				return now
			},
			commitFunc: func(context.Context) error {
				return nil
			},
			destinationsCreateFunc: func(context.Context, payd.DestinationsCreate) (*payd.Destination, error) {
				return nil, nil
			},
			invoiceCreateFunc: func(context.Context, payd.InvoiceCreate) (*payd.Invoice, error) {
				return nil, errors.New("nah")
			},
			req: payd.InvoiceCreate{
				Satoshis:    2000,
				Description: null.StringFrom("my cool invoice"),
				Reference:   null.StringFrom("sick"),
			},
			cfg: &config.Server{
				Hostname: "ohwow",
			},
			expReq: payd.InvoiceCreate{
				InvoiceID:   "pJ",
				SPVRequired: true,
				Satoshis:    2000,
				Description: null.StringFrom("my cool invoice"),
				Reference:   null.StringFrom("sick"),
				CreatedAt:   now,
				ExpiresAt:   null.TimeFrom(now.Add(time.Hour * 24)),
			},
			expErr: errors.New("nah"),
		},
		"error on destination create is reported": {
			nanosecondFunc: func() int {
				return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC).Nanosecond()
			},
			nowUTCFunc: func() time.Time {
				return now
			},
			commitFunc: func(context.Context) error {
				return nil
			},
			destinationsCreateFunc: func(context.Context, payd.DestinationsCreate) (*payd.Destination, error) {
				return nil, errors.New("nope")
			},
			invoiceCreateFunc: func(context.Context, payd.InvoiceCreate) (*payd.Invoice, error) {
				return nil, nil
			},
			req: payd.InvoiceCreate{
				Satoshis:    2000,
				Description: null.StringFrom("my cool invoice"),
				Reference:   null.StringFrom("sick"),
			},
			cfg: &config.Server{
				Hostname: "ohwow",
			},
			expReq: payd.InvoiceCreate{
				InvoiceID:   "pJ",
				SPVRequired: true,
				Satoshis:    2000,
				Description: null.StringFrom("my cool invoice"),
				Reference:   null.StringFrom("sick"),
				CreatedAt:   now,
				ExpiresAt:   null.TimeFrom(now.Add(time.Hour * 24)),
			},
			expErr: errors.New("failed to create payment destinations for invoice: nope"),
		},
		"error on commit is reported": {
			nanosecondFunc: func() int {
				return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC).Nanosecond()
			},
			nowUTCFunc: func() time.Time {
				return now
			},
			commitFunc: func(context.Context) error {
				return errors.New("afraid of commitment")
			},
			destinationsCreateFunc: func(context.Context, payd.DestinationsCreate) (*payd.Destination, error) {
				return nil, nil
			},
			invoiceCreateFunc: func(context.Context, payd.InvoiceCreate) (*payd.Invoice, error) {
				return nil, nil
			},
			req: payd.InvoiceCreate{
				Satoshis:    2000,
				Description: null.StringFrom("my cool invoice"),
				Reference:   null.StringFrom("sick"),
			},
			cfg: &config.Server{
				Hostname: "ohwow",
			},
			expReq: payd.InvoiceCreate{
				InvoiceID:   "pJ",
				SPVRequired: true,
				Satoshis:    2000,
				Description: null.StringFrom("my cool invoice"),
				Reference:   null.StringFrom("sick"),
				CreatedAt:   now,
				ExpiresAt:   null.TimeFrom(now.Add(time.Hour * 24)),
			},
			expErr: errors.New("afraid of commitment"),
		},
		"invalid request is rejected": {
			nanosecondFunc: func() int {
				return time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC).Nanosecond()
			},
			nowUTCFunc: func() time.Time {
				return now
			},
			commitFunc: func(context.Context) error {
				return nil
			},
			destinationsCreateFunc: func(context.Context, payd.DestinationsCreate) (*payd.Destination, error) {
				return nil, nil
			},
			invoiceCreateFunc: func(context.Context, payd.InvoiceCreate) (*payd.Invoice, error) {
				return nil, nil
			},
			req: payd.InvoiceCreate{
				Description: null.StringFrom(string(bytes.Repeat([]byte("invoice"), 1000))),
				Reference:   null.StringFrom(string(bytes.Repeat([]byte("sick"), 50))),
			},
			cfg: &config.Server{
				Hostname: "ohwow",
			},
			expReq: payd.InvoiceCreate{
				InvoiceID:   "pJ",
				SPVRequired: true,
				Satoshis:    2000,
				Description: null.StringFrom("my cool invoice"),
				Reference:   null.StringFrom("sick"),
				CreatedAt:   now,
				ExpiresAt:   null.TimeFrom(now.Add(time.Hour * 24)),
			},
			expErr: errors.New("[description: value must be between 0 and 1024 characters], [paymentReference: value must be between 0 and 32 characters], [satoshis: value 0 is smaller than minimum 136]"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			svc := service.NewInvoice(
				test.cfg,
				&config.Wallet{SPVRequired: true, PaymentExpiryHours: 24},
				&mocks.InvoiceReaderWriterMock{
					InvoiceCreateFunc: func(ctx context.Context, req payd.InvoiceCreate) (*payd.Invoice, error) {
						assert.Equal(t, test.expReq, req)
						return test.invoiceCreateFunc(ctx, req)
					},
				},
				&mocks.DestinationsServiceMock{
					DestinationsCreateFunc: test.destinationsCreateFunc,
				},
				&mocks.TransacterMock{
					CommitFunc: test.commitFunc,
					WithTxFunc: func(ctx context.Context) context.Context {
						return ctx
					},
					RollbackFunc: func(context.Context) error {
						return nil
					},
				},
				&mocks.TimestampServiceMock{
					NanosecondFunc: test.nanosecondFunc,
					NowUTCFunc:     test.nowUTCFunc,
				})

			_, err := svc.Create(context.TODO(), test.req)
			if test.expErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, test.expErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInvoiceService_Delete(t *testing.T) {
	tests := map[string]struct {
		invoiceDeleteFunc func(context.Context, payd.InvoiceArgs) error
		args              payd.InvoiceArgs
		expErr            error
	}{
		"successful invoice delete": {
			invoiceDeleteFunc: func(context.Context, payd.InvoiceArgs) error {
				return nil
			},
			args: payd.InvoiceArgs{
				InvoiceID: "cvb123",
			},
		},
		"invalid invoice delete args": {
			invoiceDeleteFunc: func(context.Context, payd.InvoiceArgs) error {
				return nil
			},
			expErr: errors.New("[invoiceID: value must be between 1 and 30 characters]"),
		},
		"store error is reported": {
			invoiceDeleteFunc: func(context.Context, payd.InvoiceArgs) error {
				return errors.New("whoopsie")
			},
			args: payd.InvoiceArgs{
				InvoiceID: "cvb123",
			},
			expErr: errors.New("failed to delete invoice with ID cvb123: whoopsie"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			svc := service.NewInvoice(nil, nil, &mocks.InvoiceReaderWriterMock{
				InvoiceDeleteFunc: test.invoiceDeleteFunc,
			}, nil, nil, nil)

			if test.expErr != nil {
				assert.EqualError(t, svc.Delete(context.TODO(), test.args), test.expErr.Error())
			} else {
				assert.NoError(t, svc.Delete(context.TODO(), test.args))
			}
		})
	}
}
