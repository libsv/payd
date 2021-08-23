package sqlite

import (
	"context"

	"github.com/pkg/errors"

	gopayd "github.com/libsv/payd"
)

const (
	sqlDerivationCounter = `
	SELECT pathcounter FROM keys WHERE name = :name
	`
	sqlDerivationIncrement = `
	UPDATE keys set pathCounter = pathCounter + $1
	WHERE name = $2
	`

	sqlDerivationPathExists = `
	SELECT EXISTS(
	    SELECT derivationpath FROM txos WHERE derivationpath = :derivationPath AND keyname = :keyname 
	    )
	`
)

// DerivationPathExists will return true / false if the supplied derivation path exists or not.
func (s *sqliteStore) DerivationPathExists(ctx context.Context, args gopayd.DerivationExistsArgs) (bool, error) {
	var exists int
	if err := s.db.GetContext(ctx, &exists, sqlDerivationPathExists, args); err != nil {
		return false, errors.WithStack(err)
	}
	return exists == 1, nil
}
