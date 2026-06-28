package database

import (
	"database/sql"
	_ "modernc.org/sqlite"
)

func InitDB(filepath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		return nil, err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		completed BOOLEAN NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL
	);`

	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}

	return db, nil
}
