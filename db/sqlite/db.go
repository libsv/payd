package sqlite

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// MustSetup will setup the database and panic if it fails.
func MustSetup(dsn string) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatalf("failed to setup database: %s", err)
	}

}
