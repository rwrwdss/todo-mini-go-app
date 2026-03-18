package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

// ConnectPostgres opens a connection to PostgreSQL and pings it. Does not run migrations.
// Use this for the notifier or other services that must not run schema changes.
func ConnectPostgres(connStr string) (*sql.DB, error) {
	if connStr == "" {
		connStr = "postgres://localhost/todo?sslmode=disable"
	}
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "connect: connection refused") {
			return nil, errors.New("PostgreSQL: connection refused. Start PostgreSQL and ensure database exists, or set DATABASE_URL")
		}
		return nil, fmt.Errorf("postgres ping: %w", err)
	}
	return db, nil
}

// InitPostgres connects to PostgreSQL and runs all migrations. Use only in the main app (one process).
func InitPostgres(connStr string) (*sql.DB, error) {
	db, err := ConnectPostgres(connStr)
	if err != nil {
		return nil, err
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
		db.Close()
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
		db.Close()
		return nil, err
	}
	_, _ = db.Exec(`ALTER TABLE spaces ADD CONSTRAINT spaces_owner_type_uniq UNIQUE (owner_id, type)`)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS space_members (
			space_id INTEGER NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			role TEXT NOT NULL DEFAULT 'member',
			PRIMARY KEY (space_id, user_id)
		);
	`)
	if err != nil {
		db.Close()
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS space_invitations (
			id SERIAL PRIMARY KEY,
			space_id INTEGER NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
			invitee_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			inviter_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			role TEXT NOT NULL DEFAULT 'member',
			status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined')),
			created_at TIMESTAMPTZ DEFAULT NOW()
		);
	`)
	if err != nil {
		db.Close()
		return nil, err
	}
	_, _ = db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS space_invitations_pending_uniq ON space_invitations (space_id, invitee_id) WHERE status = 'pending'`)

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
		db.Close()
		return nil, err
	}

	_, _ = db.Exec(`ALTER TABLE todos ADD COLUMN IF NOT EXISTS space_id INTEGER REFERENCES spaces(id) ON DELETE CASCADE`)
	_, _ = db.Exec(`ALTER TABLE todos ADD COLUMN IF NOT EXISTS assignee_id INTEGER REFERENCES users(id) ON DELETE SET NULL`)
	_, _ = db.Exec(`ALTER TABLE todos ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW()`)
	_, _ = db.Exec(`ALTER TABLE todos ADD COLUMN IF NOT EXISTS due_date DATE`)
	_, _ = db.Exec(`ALTER TABLE todos ADD COLUMN IF NOT EXISTS due_at TIMESTAMPTZ`)
	_, _ = db.Exec(`UPDATE todos SET due_at = (due_date::text || ' 00:00:00')::timestamptz WHERE due_at IS NULL AND due_date IS NOT NULL`)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS notifications (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			todo_id INTEGER REFERENCES todos(id) ON DELETE CASCADE,
			type TEXT NOT NULL CHECK (type IN ('overdue', 'due_soon', 'space_invitation')),
			created_at TIMESTAMPTZ DEFAULT NOW(),
			read_at TIMESTAMPTZ,
			space_id INTEGER REFERENCES spaces(id) ON DELETE CASCADE,
			invitation_id INTEGER
		);
	`)
	if err != nil {
		db.Close()
		return nil, err
	}
	_, _ = db.Exec(`ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_type_check`)
	_, _ = db.Exec(`ALTER TABLE notifications ADD CONSTRAINT notifications_type_check CHECK (type IN ('overdue', 'due_soon', 'space_invitation'))`)
	_, _ = db.Exec(`ALTER TABLE notifications ALTER COLUMN todo_id DROP NOT NULL`)
	_, _ = db.Exec(`ALTER TABLE notifications ADD COLUMN IF NOT EXISTS space_id INTEGER REFERENCES spaces(id) ON DELETE CASCADE`)
	_, _ = db.Exec(`ALTER TABLE notifications ADD COLUMN IF NOT EXISTS invitation_id INTEGER`)

	_, err = db.Exec(`
		INSERT INTO spaces (name, type, owner_id)
		SELECT 'My Space', 'personal', id FROM users u
		WHERE NOT EXISTS (SELECT 1 FROM spaces s WHERE s.owner_id = u.id AND s.type = 'personal')
	`)
	if err != nil {
		db.Close()
		return nil, err
	}

	_, err = db.Exec(`
		INSERT INTO space_members (space_id, user_id, role)
		SELECT s.id, s.owner_id, 'admin' FROM spaces s
		WHERE s.type = 'personal'
		ON CONFLICT (space_id, user_id) DO NOTHING
	`)
	if err != nil {
		db.Close()
		return nil, err
	}

	_, err = db.Exec(`
		UPDATE todos t SET space_id = s.id, assignee_id = t.user_id
		FROM spaces s
		WHERE s.owner_id = t.user_id AND s.type = 'personal' AND (t.space_id IS NULL OR t.space_id = 0)
	`)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
