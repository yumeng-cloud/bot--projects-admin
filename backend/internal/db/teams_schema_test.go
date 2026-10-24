package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSQLiteTeamsAndTeamSeatsExist(t *testing.T) {
	dsn := filepath.Join(t.TempDir(), "teams.db")
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

	for _, table := range []string{"teams", "team_seats"} {
		var n int
		if err := b.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&n); err != nil {
			t.Fatal(err)
		}
		if n != 1 {
			t.Fatalf("table %s missing", table)
		}
	}
}

func TestPostgresTeamsIfDSN(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("TEST_DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("no DATABASE_URL")
	}
	_ = os.Setenv("AUTO_MIGRATE", "1")
	sqlDB, err := Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	for _, table := range []string{"teams", "team_seats"} {
		var n int
		err := sqlDB.QueryRow(`SELECT COUNT(*) FROM information_schema.tables WHERE table_name=$1`, table).Scan(&n)
		if err != nil {
			t.Fatal(err)
		}
		if n < 1 {
			t.Fatalf("postgres table %s missing", table)
		}
	}
}
