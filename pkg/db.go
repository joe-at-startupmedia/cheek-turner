package cheek

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	_ "github.com/glebarez/go-sqlite"
)

func OpenDB(dbPath string) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := InitDB(db); err != nil {
		return nil, fmt.Errorf("init db: %w", err)
	}

	return db, nil
}

func InitDB(db *sqlx.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS log (
        id INTEGER PRIMARY KEY,
        job TEXT,
        triggered_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		triggered_by TEXT,
        duration INTEGER,
        status INTEGER,
        message TEXT
    )`)
	if err != nil {
		return fmt.Errorf("create log table: %w", err)
	}

	return nil
}
