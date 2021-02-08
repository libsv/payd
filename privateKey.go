package gopayd

import (
	"context"
	"time"
)

// Key describes a named private key.
type Key struct {
	Name      string    `db:"name"`
	Xpriv     string    `db:"xpriv"`
	CreatedAt time.Time `db:"createdAt"`
}

// KeyArgs defines all arguments required to get a key.
type KeyArgs struct {
	Name string
}

type KeyStorer interface {
	Key(ctx context.Context, args KeyArgs) (*Key, error)
	Create(ctx context.Context, req Key) (*Key, error)
}
