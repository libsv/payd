package schema

import (
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/jmoiron/sqlx"
	// sqlite gets wired up automatically.
	_ "github.com/mattn/go-sqlite3"

	"github.com/libsv/payd/config"
)

// MustSetup will setup the database and panic if it fails.
func MustSetup(cfg *config.Db) {
	log.Println("setting up sqlite database")
	if cfg.Type == "" {
		log.Fatal("no database connection string provided")
	}
	db, err := sqlx.Connect("sqlite3", cfg.Dsn)
	if err != nil {
		log.Fatalf("failed to setup database: %s", err)
	}
	// nolint:errcheck // dont care about this error
	defer db.Close()

	var schemaCount int
	if err := db.Get(&schemaCount, `SELECT COUNT(name) FROM sqlite_master
	WHERE type='table'`); err != nil {
		log.Fatalf("failed to read schema count %s", err)
	}
	if schemaCount > 0 {
		log.Println("db already created, exiting setup")
		return
	}
	f, err := os.Open(path.Join(cfg.SchemaPath, "schema_create_v1.sql"))
	if err != nil {
		log.Fatalf("failed to read create schema %s", err)
	}
	bb, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalf("failed to read create schema %s", err)
	}
	if _, err := db.Exec(string(bb)); err != nil {
		log.Fatalf("failed to execute db setup %s", err)
	}
	log.Println("finished setting up sqlite database")
}
