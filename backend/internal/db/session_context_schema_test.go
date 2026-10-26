package db

import (
	"path/filepath"
	"testing"
)

func TestSQLiteUserSessionContextsExist(t *testing.T) {
	dsn := filepath.Join(t.TempDir(), "sess.db")
	a, err := Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("second open: %v", err)
	}
	defer b.Close()

	var n int
	if err := b.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, "user_session_contexts").Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("table user_session_contexts missing")
	}
}
