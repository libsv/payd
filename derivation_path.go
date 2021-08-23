package gopayd

import (
	"context"
)

// DerivationIncrementArgs are used to increment a derivation counter.
type DerivationIncrementArgs struct {
	// Key is the private key name to increment.
	Key string `db:"key"`
	// Offset is the amount we are going to increment by.
	Offset uint64 `db:"offset"`
}

// DerivationCounterArgs are used to return the current derivation counter for a master key.
type DerivationCounterArgs struct {
	Key string `db:"key"`
}

// DerivationCounterWriter can be used to write derivation path data to a data store.
type DerivationCounterWriter interface {
	// IncrementKeyCounter will increment the key counter using the provided arguments.
	IncrementKeyCounter(ctx context.Context, args DerivationIncrementArgs) error
}

// DerivationCounterReader can be used to read derivation path data from a data store.
type DerivationCounterReader interface {
	// DerivationCounter will return the current counter for a private key.
	DerivationCounter(ctx context.Context, args DerivationCounterArgs) (uint64, error)
}

// DerivationCounterReaderWriter allows derivation paths to be written and read from a data store.
type DerivationCounterReaderWriter interface {
	DerivationCounterReader
	DerivationCounterWriter
}
