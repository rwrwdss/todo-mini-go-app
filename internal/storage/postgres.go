package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

func InitPostgres() (*sql.DB, error) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://localhost/todo?sslmode=disable"
	}
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "connect: connection refused") {
			return nil, errors.New("PostgreSQL connection refused. Backend uses only PostgreSQL. Start PostgreSQL (e.g. brew services start postgresql on macOS) and ensure database 'todo' exists, or set DATABASE_URL")
		}
		return nil, fmt.Errorf("postgres ping: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			name TEXT DEFAULT '',
			created_at TIMESTAMPTZ DEFAULT NOW()
		);
	`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS todos (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			title TEXT NOT NULL,
			done BOOLEAN DEFAULT FALSE,
			description TEXT DEFAULT '',
			priority TEXT DEFAULT 'none',
			tag TEXT DEFAULT '',
			parent_id INTEGER REFERENCES todos(id) ON DELETE SET NULL
		);
	`)
	if err != nil {
		return nil, err
	}

	return db, nil
}
