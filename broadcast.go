package gopayd

import (
	"context"

	"github.com/libsv/go-bt/v2"
)

// BroadcastWriter is used to submit a transaction for public broadcast to nodes.
type BroadcastWriter interface {
	// Broadcast will submit a tx to a blockchain network.
	Broadcast(ctx context.Context, tx *bt.Tx) error
}
