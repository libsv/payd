package sqlite

import (
	"context"

	"github.com/pkg/errors"

	"github.com/libsv/payd"
)

const (
	sqlDerivationPathExists = `
	SELECT EXISTS(
	    SELECT derivation_path FROM destinations WHERE derivation_path = $1 AND key_name = $2 AND user_id = $3 
	    )
	`
)

// DerivationPathExists will return true / false if the supplied derivation path exists or not.
func (s *sqliteStore) DerivationPathExists(ctx context.Context, args payd.DerivationExistsArgs) (bool, error) {
	var exists int
	if err := s.db.GetContext(ctx, &exists, sqlDerivationPathExists, args.Path, args.KeyName, args.UserID); err != nil {
		return false, errors.WithStack(err)
	}
	return exists == 1, nil
}
