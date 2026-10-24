package db_test

import (
	"os"
	"testing"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"
)

func TestLivePostgresMigrateAndRebind(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("no TEST_DATABASE_URL")
	}
	t.Setenv("AUTO_MIGRATE", "1")
	sqlDB, err := db.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer sqlDB.Close()

	_, err = sqlDB.Exec(`INSERT INTO workspaces (id, name, slug, created_at, updated_at) VALUES (?, ?, ?, NOW(), NOW())
		ON CONFLICT (id) DO NOTHING`, "ws_verify_1", "Verify WS", "verify-ws")
	if err != nil {
		t.Fatalf("rebind insert: %v", err)
	}
	var n int
	if err := sqlDB.QueryRow(`SELECT COUNT(*) FROM workspaces WHERE id = ?`, "ws_verify_1").Scan(&n); err != nil || n != 1 {
		t.Fatalf("rebind select: err=%v n=%d", err, n)
	}
	var ext string
	if err := sqlDB.QueryRow(`SELECT extversion FROM pg_extension WHERE extname='vector'`).Scan(&ext); err != nil {
		t.Fatalf("vector: %v", err)
	}
	var hasEmb int
	_ = sqlDB.QueryRow(`SELECT COUNT(*) FROM information_schema.tables WHERE table_name='embedding_chunks'`).Scan(&hasEmb)
	if hasEmb != 1 {
		t.Fatalf("embedding_chunks missing")
	}
	var tables int
	_ = sqlDB.QueryRow(`SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE'`).Scan(&tables)
	t.Logf("OK tables=%d vector=%s embedding_chunks=1 rebind=ok", tables, ext)
}
