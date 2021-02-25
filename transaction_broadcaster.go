package gopayd

import (
	"context"
)

type BroadcastTransaction struct {
	TXHex string
}

// TransactionBroadcaster can be implemented to broadcast a raw transaction to a network.
type TransactionBroadcaster interface {
	Broadcast(ctx context.Context, req BroadcastTransaction) error
}
