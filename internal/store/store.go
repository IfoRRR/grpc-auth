package store

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

var ErrNotFound = errors.New("task not found")

type Task struct {
	ID    string `json:"id"`
	Owner string `json:"owner"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type Store struct {
	db *sql.DB
}

func Open(db *sql.DB) (*Store, error) {
	if db == nil {
		return nil, errors.New("db required")
	}
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    owner TEXT NOT NULL,
    title TEXT NOT NULL,
    done BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_tasks_owner ON tasks(owner);
`)
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Create(owner, title string) Task {
	t := Task{ID: uuid.NewString(), Owner: owner, Title: title}
	_, _ = s.db.Exec(
		`INSERT INTO tasks (id, owner, title, done) VALUES ($1, $2, $3, FALSE)`,
		t.ID, t.Owner, t.Title,
	)
	return t
}

func (s *Store) Get(owner, id string) (Task, error) {
	var t Task
	err := s.db.QueryRow(
		`SELECT id, owner, title, done FROM tasks WHERE id = $1 AND owner = $2`,
		id, owner,
	).Scan(&t.ID, &t.Owner, &t.Title, &t.Done)
	if errors.Is(err, sql.ErrNoRows) {
		return Task{}, ErrNotFound
	}
	if err != nil {
		return Task{}, err
	}
	return t, nil
}

func (s *Store) List(owner string) []Task {
	rows, err := s.db.Query(
		`SELECT id, owner, title, done FROM tasks WHERE owner = $1 ORDER BY created_at`,
		owner,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]Task, 0)
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Owner, &t.Title, &t.Done); err != nil {
			return out
		}
		out = append(out, t)
	}
	return out
}
