package wiki_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/embedder"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/wiki"
)

func TestPostgresHybridSearchSmoke(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" || os.Getenv("DB_DRIVER") != "postgres" && os.Getenv("DB_DRIVER") != "pgx" && os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("need Postgres TEST_DATABASE_URL / DB_DRIVER=postgres")
	}
	if os.Getenv("DB_DRIVER") == "" {
		os.Setenv("DB_DRIVER", "postgres")
	}
	os.Setenv("EMBED_PROVIDER", "hash")

	sqlDB, err := db.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer sqlDB.Close()

	ctx := context.Background()
	e := embedder.MustNew(embedder.LoadConfig())
	store := embedder.NewStore(sqlDB)

	wsID := "ws_smoke_" + idgen.New()[:8]
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := sqlDB.Exec(`INSERT INTO workspaces (id, name, slug, created_at, updated_at) VALUES (?,?,?,?,?)`,
		wsID, "Smoke", "smoke-"+wsID[len(wsID)-6:], now, now); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = sqlDB.Exec(`DELETE FROM workspaces WHERE id=?`, wsID) })

	svc := wiki.New(sqlDB).WithEmbedder(e, store)
	sp, err := svc.CreateSpace(wsID, wiki.CreateSpaceInput{Name: "KB", Slug: "kb"})
	if err != nil {
		t.Fatal(err)
	}
	p1, err := svc.CreatePage(wsID, sp.ID, wiki.CreatePageInput{
		Title: "pgvector hybrid search", BodyMD: "Wiki semantic recall uses Postgres pgvector cosine distance and RRF fusion.",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreatePage(wsID, sp.ID, wiki.CreatePageInput{
		Title: "unrelated cooking", BodyMD: "How to boil pasta with salt water.",
	}); err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func(n int) {
			chunks := []string{fmt.Sprintf("concurrent rewrite %d about pgvector RRF", n)}
			vecs, err := e.Embed(ctx, chunks)
			if err != nil {
				done <- err
				return
			}
			done <- store.ReplaceEntity(ctx, wsID, embedder.EntityWikiPage, p1.ID, chunks, vecs, e.ModelID(), e.Dims())
		}(i)
	}
	for i := 0; i < 2; i++ {
		if err := <-done; err != nil {
			t.Fatalf("concurrent replace: %v", err)
		}
	}

	res, err := svc.SearchEx(ctx, wsID, "semantic pgvector RRF", wiki.ModeHybrid, 10)
	if err != nil {
		t.Fatal(err)
	}
	if res.SearchBackend != "hybrid" {
		t.Fatalf("backend=%s", res.SearchBackend)
	}
	found := false
	for _, p := range res.Pages {
		if p.ID == p1.ID {
			found = true
		}
	}
	if !found {
		t.Fatalf("p1 not in results: %+v", res.Pages)
	}
	var n int
	if err := sqlDB.QueryRow(`SELECT COUNT(*) FROM embedding_chunks WHERE workspace_id=$1 AND entity_type=$2`,
		wsID, embedder.EntityWikiPage).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n < 1 {
		t.Fatalf("expected chunks, got %d", n)
	}
	t.Logf("OK mode=%s backend=%s pages=%d chunks=%d", res.Mode, res.SearchBackend, len(res.Pages), n)
}
