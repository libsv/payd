package databases

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/libsv/payd/log"

	"github.com/libsv/payd/config"
)

type dbSetupFunc func(l log.Logger, c *config.Db) (*sqlx.DB, error)
type dbSetups map[config.DbType]dbSetupFunc

// NewDbSetup will load the db setup functions into a lookup map
// ready for being called in main.go.
func NewDbSetup() dbSetups {
	s := make(map[config.DbType]dbSetupFunc, 3)
	s[config.DBSqlite] = setupSqliteDB
	//s[config.DBMySQL] = setupMySqlDB
	//s[config.DBPostgres] = setupPostgresDB
	return s
}

func (d dbSetups) SetupDb(l log.Logger, cfg *config.Db) (*sqlx.DB, error) {
	fn, ok := d[cfg.Type]
	if !ok {
		return nil, fmt.Errorf("db type %s not supported", cfg.Type)
	}
	return fn(l, cfg)
}
