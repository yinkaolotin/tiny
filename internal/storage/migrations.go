package storage

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func RunMigrations(databaseURL string) error {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS items (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_items_expires_at
		ON items (expires_at);

		CREATE INDEX IF NOT EXISTS idx_items_created_at
		ON items (created_at DESC);
	`)

	return err
}
