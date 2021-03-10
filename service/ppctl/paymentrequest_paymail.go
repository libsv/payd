package ppctl

import (
	"context"

	gopayd "github.com/libsv/payd"
)

type paymentReqPaymailService struct {
	rdrwtr gopayd.PaymailReaderWriter
}

func NewPaymentRequestPaymailService(rdrwtr gopayd.PaymailReaderWriter) *paymentReqPaymailService {
	return &paymentReqPaymailService{rdrwtr: rdrwtr}
}

// CreatePaymentRequest handles setting up a new PaymentRequest response and can use and optional existing paymentID.
func (p *paymentReqPaymailService) CreatePaymentRequest(ctx context.Context, args gopayd.PaymentRequestArgs) (*gopayd.PaymentRequest, error) {
	args.PaymentID

}
