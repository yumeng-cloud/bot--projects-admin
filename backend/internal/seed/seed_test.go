package seed_test

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/config"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/seed"
)

func testCFG(sample bool) config.Config {
	return config.Config{
		JWTSecret:      "test-secret",
		JWTTTL:         time.Hour,
		AppEnv:         "test",
		SeedAdminEmail: "admin@yumeng.tech",
		SeedAdminPass:  "Admin123!",
		SeedWSName:     "育蒙科技",
		SeedSampleData: sample,
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := filepath.Join(t.TempDir(), "seed.db")
	sqlDB, err := db.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return sqlDB
}

func countProjects(t *testing.T, sqlDB *sql.DB) int {
	t.Helper()
	var n int
	if err := sqlDB.QueryRow(`SELECT COUNT(1) FROM projects`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func countIssues(t *testing.T, sqlDB *sql.DB) int {
	t.Helper()
	var n int
	if err := sqlDB.QueryRow(`SELECT COUNT(1) FROM issues`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func countLabels(t *testing.T, sqlDB *sql.DB) int {
	t.Helper()
	var n int
	if err := sqlDB.QueryRow(`SELECT COUNT(1) FROM labels`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func countBotRoles(t *testing.T, sqlDB *sql.DB) int {
	t.Helper()
	var n int
	if err := sqlDB.QueryRow(`SELECT COUNT(1) FROM bot_roles`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestSeedDefaultNoSampleData(t *testing.T) {
	sqlDB := openTestDB(t)
	if err := seed.Run(sqlDB, testCFG(false)); err != nil {
		t.Fatal(err)
	}
	if n := countProjects(t, sqlDB); n != 0 {
		t.Fatalf("default seed must not create sample project, got %d", n)
	}
	if n := countIssues(t, sqlDB); n != 0 {
		t.Fatalf("default seed must not create sample issues, got %d", n)
	}
	if n := countLabels(t, sqlDB); n != 0 {
		t.Fatalf("default seed must not create demo labels, got %d", n)
	}
	if n := countBotRoles(t, sqlDB); n < 1 {
		t.Fatalf("structural bot roles should still seed, got %d", n)
	}
	var users int
	_ = sqlDB.QueryRow(`SELECT COUNT(1) FROM users`).Scan(&users)
	if users < 1 {
		t.Fatal("admin user should still seed")
	}
}

func TestSeedSampleDataCreatesDemoOnce(t *testing.T) {
	sqlDB := openTestDB(t)
	if err := seed.Run(sqlDB, testCFG(true)); err != nil {
		t.Fatal(err)
	}
	if n := countProjects(t, sqlDB); n != 1 {
		t.Fatalf("want 1 sample project, got %d", n)
	}
	if n := countIssues(t, sqlDB); n != 8 {
		t.Fatalf("want 8 sample issues, got %d", n)
	}
	if n := countLabels(t, sqlDB); n != 3 {
		t.Fatalf("want 3 demo labels, got %d", n)
	}

	if _, err := sqlDB.Exec(`DELETE FROM issue_labels`); err != nil {
		t.Fatal(err)
	}
	if _, err := sqlDB.Exec(`DELETE FROM issues`); err != nil {
		t.Fatal(err)
	}
	if n := countIssues(t, sqlDB); n != 0 {
		t.Fatal("precondition: issues cleared")
	}
	if err := seed.Run(sqlDB, testCFG(true)); err != nil {
		t.Fatal(err)
	}
	if n := countIssues(t, sqlDB); n != 0 {
		t.Fatalf("must not re-seed sample issues after clear when already marked, got %d", n)
	}
}

func TestSeedSampleDataOffAfterPriorSeed(t *testing.T) {
	sqlDB := openTestDB(t)
	if err := seed.Run(sqlDB, testCFG(true)); err != nil {
		t.Fatal(err)
	}
	if _, err := sqlDB.Exec(`DELETE FROM issue_labels`); err != nil {
		t.Fatal(err)
	}
	if _, err := sqlDB.Exec(`DELETE FROM issues`); err != nil {
		t.Fatal(err)
	}
	if err := seed.Run(sqlDB, testCFG(false)); err != nil {
		t.Fatal(err)
	}
	if n := countIssues(t, sqlDB); n != 0 {
		t.Fatalf("SEED_SAMPLE_DATA off must not re-seed issues, got %d", n)
	}
}

func TestConfigSeedSampleDataEnv(t *testing.T) {
	t.Setenv("SEED_SAMPLE_DATA", "")
	cfg := config.Load()
	if cfg.SeedSampleData {
		t.Fatal("unset SEED_SAMPLE_DATA must default false")
	}
	t.Setenv("SEED_SAMPLE_DATA", "0")
	cfg = config.Load()
	if cfg.SeedSampleData {
		t.Fatal("SEED_SAMPLE_DATA=0 must be false")
	}
	t.Setenv("SEED_SAMPLE_DATA", "1")
	cfg = config.Load()
	if !cfg.SeedSampleData {
		t.Fatal("SEED_SAMPLE_DATA=1 must be true")
	}
	t.Setenv("SEED_SAMPLE_DATA", "true")
	cfg = config.Load()
	if !cfg.SeedSampleData {
		t.Fatal("SEED_SAMPLE_DATA=true must be true")
	}
}
