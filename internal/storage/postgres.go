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
		CREATE TABLE IF NOT EXISTS spaces (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL DEFAULT 'personal',
			owner_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);
	`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS space_members (
			space_id INTEGER NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			role TEXT NOT NULL DEFAULT 'member',
			PRIMARY KEY (space_id, user_id)
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
			parent_id INTEGER REFERENCES todos(id) ON DELETE SET NULL,
			space_id INTEGER REFERENCES spaces(id) ON DELETE CASCADE,
			assignee_id INTEGER REFERENCES users(id) ON DELETE SET NULL
		);
	`)
	if err != nil {
		return nil, err
	}

	// Migrate existing schema: add space_id/assignee_id if tables already existed without them
	_, _ = db.Exec(`ALTER TABLE todos ADD COLUMN IF NOT EXISTS space_id INTEGER REFERENCES spaces(id) ON DELETE CASCADE`)
	_, _ = db.Exec(`ALTER TABLE todos ADD COLUMN IF NOT EXISTS assignee_id INTEGER REFERENCES users(id) ON DELETE SET NULL`)
	_, _ = db.Exec(`ALTER TABLE todos ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW()`)
	_, _ = db.Exec(`ALTER TABLE todos ADD COLUMN IF NOT EXISTS due_date DATE`)

	// Notifications table (used by main app and notifier microservice)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS notifications (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			todo_id INTEGER NOT NULL REFERENCES todos(id) ON DELETE CASCADE,
			type TEXT NOT NULL CHECK (type IN ('overdue', 'due_soon')),
			created_at TIMESTAMPTZ DEFAULT NOW(),
			read_at TIMESTAMPTZ
		);
	`)
	if err != nil {
		return nil, err
	}

	// Ensure every user has a personal space
	_, err = db.Exec(`
		INSERT INTO spaces (name, type, owner_id)
		SELECT 'My Space', 'personal', id FROM users u
		WHERE NOT EXISTS (SELECT 1 FROM spaces s WHERE s.owner_id = u.id AND s.type = 'personal')
	`)
	if err != nil {
		return nil, err
	}

	// Add owner as admin member of their personal space
	_, err = db.Exec(`
		INSERT INTO space_members (space_id, user_id, role)
		SELECT s.id, s.owner_id, 'admin' FROM spaces s
		WHERE s.type = 'personal'
		ON CONFLICT (space_id, user_id) DO NOTHING
	`)
	if err != nil {
		return nil, err
	}

	// Migrate existing todos into personal space
	_, err = db.Exec(`
		UPDATE todos t SET space_id = s.id, assignee_id = t.user_id
		FROM spaces s
		WHERE s.owner_id = t.user_id AND s.type = 'personal' AND (t.space_id IS NULL OR t.space_id = 0)
	`)
	if err != nil {
		return nil, err
	}

	return db, nil
}
