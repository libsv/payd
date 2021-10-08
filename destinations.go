package payd

import (
	"context"
	"time"

	"github.com/libsv/go-bt/v2"
	validator "github.com/theflyingcodr/govalidator"
	"gopkg.in/guregu/null.v3"
)

// DestinationsCreate will create new destinations.
type DestinationsCreate struct {
	InvoiceID     null.String
	Satoshis      uint64
	Denominations uint64
}

// Validate will ensure arguments for destinationsCreate are valid, otherwise an error is returned.
func (d DestinationsCreate) Validate() error {
	return validator.New().
		Validate("satoshis", validator.MinUInt64(d.Satoshis, 136)).
		Err()
}

// DestinationCreate can be used to create a single Output for storage.
type DestinationCreate struct {
	Script         string `db:"locking_script"`
	DerivationPath string `db:"derivation_path"`
	Satoshis       uint64 `db:"satoshis"`
	Keyname        string `db:"key_name"`
}

// Destination contains outputs and current fees
// required to construct a transaction.
type Destination struct {
	SPVRequired bool         `json:"spvRequired"`
	Network     string       `json:"network"`
	Outputs     []Output     `json:"outputs"`
	Fees        *bt.FeeQuote `json:"fees"`
	CreatedAt   time.Time    `json:"createdAt"`
	ExpiresAt   time.Time    `json:"expiresAt"`
}

// Output contains a single locking script
// and satoshi amount and can be used to construct
// transaction outputs.
type Output struct {
	ID uint64 `json:"-" db:"destination_id"`
	// LockingScript is the P2PKH locking script used.
	LockingScript string `json:"script" db:"locking_script"`
	Satoshis      uint64 `json:"satoshis" db:"satoshis"`
	// DerivationPath is the deterministic path for this destination.
	DerivationPath string `json:"-" db:"derivation_path"`
	// State will indicate if this destination is still waiting on a tx to fulfil it (pending)
	// has been paid to in a tx (received) or has been deleted.
	State string `json:"-" db:"state"  enums:"pending,received,deleted"`
}

// DestinationsArgs are used to get a set of Denominations
// for an existing InvoiceID.
type DestinationsArgs struct {
	InvoiceID string `param:"invoiceID"`
}

// Validate will ensure arguments to get destinations are as expected.
func (d DestinationsArgs) Validate() error {
	return validator.New().Validate("invoiceID", validator.NotEmpty(d.InvoiceID)).
		Err()
}

// DestinationsCreateArgs are used when adding destinations to a datastore
// and can be sued to group them.
type DestinationsCreateArgs struct {
	// InvoiceID, this is optional, if not supplied destinations
	// will not be associated with an invoice.
	InvoiceID null.String `db:"invoice_id"`
}

// DestinationsService enforces business rules and validation for Destinations.
type DestinationsService interface {
	// InvoiceCreate will split satoshis into multiple denominations and store
	// as denominations waiting to be fulfilled in a tx.
	DestinationsCreate(ctx context.Context, req DestinationsCreate) (*Destination, error)
	// Destinations given the args, will return a set of Destinations.
	Destinations(ctx context.Context, args DestinationsArgs) (*Destination, error)
}

// DestinationsWriter can be implemented to store new destinations in a data store.
type DestinationsWriter interface {
	// DestinationsCreate will add a set of destinations to a data store.
	DestinationsCreate(ctx context.Context, args DestinationsCreateArgs, req []DestinationCreate) ([]Output, error)
}

// DestinationsReader will return destination outputs.
type DestinationsReader interface {
	// Destinations will return destination outputs.
	Destinations(ctx context.Context, args DestinationsArgs) ([]Output, error)
}

// DestinationsReaderWriter combines the reader and writer interfaces for convenience.
type DestinationsReaderWriter interface {
	DestinationsReader
	DestinationsWriter
}
