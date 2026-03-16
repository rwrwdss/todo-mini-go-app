package storage

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB() (*sql.DB, error) {
	path := os.Getenv("DB_PATH")
	if path == "" {
		path = "todos.db"
	}
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	query := `
	CREATE TABLE IF NOT EXISTS todos(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT,
		done BOOLEAN,
		description TEXT,
		priority TEXT,
		tag TEXT,
		parent_id INTEGER REFERENCES todos(id)
	);`
	if _, err = db.Exec(query); err != nil {
		return nil, err
	}

	if err = migrateAddColumns(db); err != nil {
		return nil, err
	}
	return db, nil
}

func migrateAddColumns(db *sql.DB) error {
	columns := map[string]string{
		"description": "ALTER TABLE todos ADD COLUMN description TEXT DEFAULT ''",
		"priority":    "ALTER TABLE todos ADD COLUMN priority TEXT DEFAULT 'none'",
		"tag":          "ALTER TABLE todos ADD COLUMN tag TEXT DEFAULT ''",
		"parent_id":    "ALTER TABLE todos ADD COLUMN parent_id INTEGER REFERENCES todos(id)",
	}
	rows, err := db.Query("PRAGMA table_info(todos)")
	if err != nil {
		return err
	}
	defer rows.Close()
	existing := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull, pk int
		var dflt *string
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return err
		}
		existing[name] = true
	}
	for col, alterSQL := range columns {
		if existing[col] {
			continue
		}
		if _, err := db.Exec(alterSQL); err != nil {
			return err
		}
	}
	return nil
}
