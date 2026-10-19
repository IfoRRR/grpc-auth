package store

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func TestStorePerUserIsolation(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s, err := Open(db)
	if err != nil {
		t.Fatal(err)
	}
	a := s.Create("alice", "task-a")
	s.Create("bob", "task-b")

	if _, err := s.Get("bob", a.ID); err != ErrNotFound {
		t.Fatalf("expected cross-user get to fail")
	}
	if len(s.List("alice")) != 1 {
		t.Fatal("alice should see one task")
	}
	if len(s.List("bob")) != 1 {
		t.Fatal("bob should see one task")
	}
}
