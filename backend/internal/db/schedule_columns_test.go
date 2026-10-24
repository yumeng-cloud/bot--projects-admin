package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSQLiteScheduleColumnsIdempotent(t *testing.T) {
	dir := t.TempDir()
	dsn := filepath.Join(dir, "sched.db")
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
	rows, err := b.Query(`PRAGMA table_info(issues)`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	found := map[string]bool{}
	for rows.Next() {
		var cid, notnull, pk int
		var name, ctype string
		var dflt *string
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			t.Fatal(err)
		}
		found[name] = true
	}
	if !found["start_at"] || !found["due_at"] || !found["parent_id"] {
		t.Fatalf("expected parent_id/start_at/due_at, got %v", found)
	}
}

func TestPostgresScheduleColumnsIfDSN(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("TEST_DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("no DATABASE_URL")
	}
	_ = os.Setenv("AUTO_MIGRATE", "1")
	db, err := Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var n int
	err = db.QueryRow(`SELECT COUNT(*) FROM information_schema.columns WHERE table_name='issues' AND column_name IN ('start_at','due_at')`).Scan(&n)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("want 2 schedule cols, got %d", n)
	}
}

func TestSQLiteIssuesTeamIDColumn(t *testing.T) {
	dsn := filepath.Join(t.TempDir(), "team.db")
	db, err := Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query(`PRAGMA table_info(issues)`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	found := false
	for rows.Next() {
		var cid, notnull, pk int
		var name, ctype string
		var dflt *string
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			t.Fatal(err)
		}
		if name == "team_id" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected issues.team_id column")
	}
}
