package payd

import (
	"context"

	"github.com/libsv/go-bt/v2"
)

// BroadcastArgs sends some meta identifying the invoice used when broadcasting.
type BroadcastArgs struct {
	InvoiceID string
}

// BroadcastWriter is used to submit a transaction for public broadcast to nodes.
type BroadcastWriter interface {
	// Broadcast will submit a tx to a blockchain network.
	Broadcast(ctx context.Context, args BroadcastArgs, tx *bt.Tx) error
}
