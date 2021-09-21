package gopayd

import (
	"context"
	"time"

	"gopkg.in/guregu/null.v3"
)

const (
	// Dustlimit is the minimum amount a miner will accept from an output.
	DustLimit = 136
)

// Transacter can be implemented to provide context based transactions.
type Transacter interface {
	WithTx(ctx context.Context) context.Context
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// MetaData contains common meta info for objects.
type MetaData struct {
	// CreatedAt is the UTC time the object was created.
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	// UpdatedAt is the UTC time the object was updated.
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
	// DeletedAt is the date the object was removed.
	DeletedAt null.Time `json:"deletedAt,omitempty" db:"deleted_at"`
}
