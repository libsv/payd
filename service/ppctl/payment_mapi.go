package ppctl

import (
	"context"

	gopayd "github.com/libsv/payd"
	"github.com/pkg/errors"
)

type paymentMapiService struct {
	broadcaster gopayd.TransactionBroadcaster
}

func NewPaymentWalletService(store gopayd.PaymentReaderWriter, broadcaster gopayd.TransactionBroadcaster) *paymentMapiService {
	return &paymentMapiService{broadcaster: broadcaster}
}

// CreatePayment will inform the merchant of a new payment being made,
// this payment will then be transmitted to the network and and acknowledgement sent to the user.
func (p *paymentMapiService) Send(ctx context.Context, args gopayd.CreatePaymentArgs, req gopayd.CreatePayment) error {
	// Broadcast the transaction.
	return errors.WithStack(p.broadcaster.Broadcast(ctx, gopayd.BroadcastTransaction{TXHex: req.Transaction}))
}
