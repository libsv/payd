package gopayd

import (
	"context"
	"time"
)

// DerivationPath defines a single derivation path, used when signing
// tx outputs.
type DerivationPath struct {
	ID        int       `db:"ID"`
	PaymentID string    `db:"paymentID"`
	Path      string    `db:"path"`
	Prefix    string    `db:"prefix"`
	Index     int       `db:"pathIndex"`
	CreatedAt time.Time `db:"createdAt"`
}

// DerivationPathCreate is used to create a new derivationPath.
type DerivationPathCreate struct {
	PaymentID string `db:"paymentID"`
	Prefix    string `db:"prefix"`
}

// DerivationPathArgs is used to return a single derivation path matching the args.
type DerivationPathArgs struct {
	ID int `db:"id"`
}

// DerivationPathExistsArgs are used to identify a derivPath by the paymentID .
type DerivationPathExistsArgs struct {
	PaymentID string `db:"paymentID"`
}

// DerivationPathWriter can be used to write derivation path data to a data store.
type DerivationPathWriter interface {
	// ReserveDerivationPath will create a derivation path for an invoice and
	// return with the index incremented ready for use.
	DerivationPathCreate(ctx context.Context, req DerivationPathCreate) (*DerivationPath, error)
}

// DerivationPathReader can be used to read derivation path data from a data store.
type DerivationPathReader interface {
	// DerivationPath will return a derivationPath that matches the supplied args.
	DerivationPath(ctx context.Context, args DerivationPathArgs) (*DerivationPath, error)
	// DerivationPathExists will return a true/false if key/s existing matching the args field.
	DerivationPathExists(ctx context.Context, args DerivationPathExistsArgs) (bool, error)
}

// DerivationPathReaderWriter allows derivation paths to be written and read from a data store.
type DerivationPathReaderWriter interface {
	DerivationPathReader
	DerivationPathWriter
}
