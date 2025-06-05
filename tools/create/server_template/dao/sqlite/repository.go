package sqlite

import (
	"database/sql"
	"sync"
	// _ "github.com/mattn/go-sqlite3"
)

type Repository struct {
	db *sql.DB
	mu sync.RWMutex
}

func New(dbPath string) (*Repository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	repo := &Repository{
		db: db,
	}

	if err := repo.initSchema(); err != nil {
		db.Close()
		return nil, err
	}

	return repo, nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func (r *Repository) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		status TEXT NOT NULL,
		start_time DATETIME NOT NULL,
		parent_id TEXT,
		sub_tasks TEXT,  -- JSON array of subtasks
		FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE
	);
	`

	_, err := r.db.Exec(schema)
	return err
}
