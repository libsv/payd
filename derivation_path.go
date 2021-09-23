package payd

import (
	"context"
)

// DerivationExistsArgs are used to check a derivation path exists for a specific
// master key and path.
type DerivationExistsArgs struct {
	KeyName string `db:"key_name"`
	Path    string `db:"derivation_path"`
}

// DerivationReader can be used to read derivation path data from a data store.
type DerivationReader interface {
	DerivationPathExists(ctx context.Context, args DerivationExistsArgs) (bool, error)
}
