package sqlite

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"

	"github.com/libsv/payd"
)

const (
	sqlBalance = `
	SELECT SUM(d.satoshis) as satoshis
	FROM txos as t INNER JOIN destinations as d on t.destination_id = d.destination_id
	WHERE t.spent_at IS NULL 
	`
)

// Balance will return the current account balance.
func (s *sqliteStore) Balance(ctx context.Context) (*payd.Balance, error) {
	var resp payd.Balance
	if err := s.db.GetContext(ctx, &resp, sqlBalance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &payd.Balance{Satoshis: 0}, nil
		}
		return nil, errors.Wrap(err, "failed to get balance")
	}
	return &resp, nil
}
