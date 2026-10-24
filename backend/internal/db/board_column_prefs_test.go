package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSQLiteProjectsBoardColumnPrefsColumn(t *testing.T) {
	dir := t.TempDir()
	dsn := filepath.Join(dir, "prefs.db")
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
	err = b.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='board_column_prefs'`).Scan(&n)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatal("standalone board_column_prefs table must not exist")
	}

	rows, err := b.Query(`PRAGMA table_info(projects)`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	found := false
	for rows.Next() {
		var cid, notnull, pk int
		var cname, ctype string
		var dflt *string
		if err := rows.Scan(&cid, &cname, &ctype, &notnull, &dflt, &pk); err != nil {
			t.Fatal(err)
		}
		if cname == "board_column_prefs" {
			found = true
		}
	}
	if !found {
		t.Fatal("projects.board_column_prefs column missing")
	}
}

func TestPostgresProjectsBoardColumnPrefsColumnIfDSN(t *testing.T) {
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
	err = db.QueryRow(`SELECT COUNT(*) FROM information_schema.tables WHERE table_name='board_column_prefs'`).Scan(&n)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatal("standalone board_column_prefs table must not exist")
	}
	var colN int
	err = db.QueryRow(`SELECT COUNT(*) FROM information_schema.columns WHERE table_name='projects' AND column_name='board_column_prefs'`).Scan(&colN)
	if err != nil {
		t.Fatal(err)
	}
	if colN != 1 {
		t.Fatalf("want projects.board_column_prefs column, count=%d", colN)
	}
}
