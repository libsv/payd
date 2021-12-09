package databases

import (
	"fmt"

	"github.com/libsv/payd/log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"

	// this is needed for laoding file based migrations.
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"

	// used to import the sqlite drivers.
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"

	"github.com/libsv/payd/config"
)

func setupSqliteDB(l log.Logger, c *config.Db) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite3", c.Dsn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to setup database")
	}
	if !c.MigrateDb {
		l.Info("migrate database set to false, skipping migration")
		return db, nil
	}
	l.Info("migrating database")
	driver, err := sqlite3.WithInstance(db.DB, &sqlite3.Config{})
	if err != nil {
		l.Fatalf(err, "creating sqlite3 db driver failed")
	}
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", c.SchemaPath), "sqlite3",
		driver)
	if err != nil {
		l.Fatal(err, "failed to migrate file instance")
	}

	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			l.Fatal(err, "failed to exec migrations")
		}

	}
	l.Info("migrating database completed")
	return db, nil
}
