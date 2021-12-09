package payd

import (
	"context"
)

// SpendingArgs are used to identify the payment we are spending.
type SpendingArgs struct {
	InvoiceID string
}

// SpendingCreate will wrap the request arguments.
type SpendingCreate struct {
	Tx string
}

// SpendingService is used to mark a transaction as spent on response
// from a payment ack.
// The sender wallet will insert their accepted transaction as broadcast
// and store the change txo if found.
type SpendingService interface {
	// SpendTx will mark the txos as spent and insert the change txo.
	SpendTx(ctx context.Context, args SpendingArgs, req SpendingCreate) error
}
