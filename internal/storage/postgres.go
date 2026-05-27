package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(databaseURL string) (*PostgresStore, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Ready() bool {
	if err := s.db.Ping(); err != nil {
		return false
	}

	var exists bool
	err := s.db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = 'items'
		)
	`).Scan(&exists)

	return err == nil && exists
}

func (s *PostgresStore) Create(name string, ttl time.Duration) Item {
	item := Item{
		ID:        uuid.NewString(),
		Name:      name,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(ttl),
	}

	_, err := s.db.Exec(`
		INSERT INTO items (id, name, created_at, expires_at)
		VALUES ($1, $2, $3, $4)
	`, item.ID, item.Name, item.CreatedAt, item.ExpiresAt)
	if err != nil {
		return Item{}
	}

	return item
}

func (s *PostgresStore) Get(id string) (Item, error) {
	var item Item

	err := s.db.QueryRow(`
		SELECT id, name, created_at, expires_at
		FROM items
		WHERE id = $1
	`, id).Scan(&item.ID, &item.Name, &item.CreatedAt, &item.ExpiresAt)

	if err == sql.ErrNoRows {
		return Item{}, ErrNotFound
	}
	if err != nil {
		return Item{}, err
	}

	return item, nil
}

func (s *PostgresStore) List() []Item {
	rows, err := s.db.Query(`
		SELECT id, name, created_at, expires_at
		FROM items
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	items := []Item{}

	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name, &item.CreatedAt, &item.ExpiresAt); err != nil {
			return nil
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil
	}

	return items
}

func (s *PostgresStore) Delete(id string) error {
	result, err := s.db.Exec(`
		DELETE FROM items
		WHERE id = $1
	`, id)
	if err != nil {
		return err
	}

	deleted, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if deleted == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *PostgresStore) CleanupExpired() int {
	result, err := s.db.Exec(`
		DELETE FROM items
		WHERE expires_at <= now()
	`)
	if err != nil {
		return 0
	}

	deleted, err := result.RowsAffected()
	if err != nil {
		return 0
	}

	return int(deleted)
}

func (s *PostgresStore) Close() error {
	if s.db == nil {
		return nil
	}

	return s.db.Close()
}

func BuildPostgresURL(host, port, database, user, password string) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user,
		password,
		host,
		port,
		database,
	)
}
