package db_test

import (
	"os"
	"testing"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"
)

// TestOpenPostgres_SkipWithoutDSN：无 DATABASE_URL / 无服务时跳过；有服务则 Ping + 可选迁移。
func TestOpenPostgres_SkipWithoutDSN(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" || !(os.Getenv("DB_DRIVER") == "postgres" || os.Getenv("DB_DRIVER") == "pgx" || os.Getenv("TEST_DATABASE_URL") != "") {
		t.Skip("no Postgres TEST_DATABASE_URL; compile-time pgx link verified by build")
	}
	t.Setenv("AUTO_MIGRATE", "0") // ping only; migrate needs scripts/pg-bootstrap.sh
	sqlDB, err := db.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer sqlDB.Close()
	if err := sqlDB.Ping(); err != nil {
		t.Fatal(err)
	}
}

func TestOpenSQLiteStillWorks(t *testing.T) {
	sqlDB, err := db.Open("sqlite", t.TempDir()+"/x.db")
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	var n int
	if err := sqlDB.QueryRow(`SELECT COUNT(1) FROM users`).Scan(&n); err != nil {
		t.Fatal(err)
	}
}
