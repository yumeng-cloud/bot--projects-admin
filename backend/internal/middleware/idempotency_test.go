package middleware_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
)

func TestIdempotencyReplaySameKey(t *testing.T) {
	sqlDB, err := db.Open("sqlite", filepath.Join(t.TempDir(), "idemp.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	wsID := idgen.New()
	_, err = sqlDB.Exec(`INSERT INTO workspaces (id, name, slug) VALUES (?,?,?)`, wsID, "W", "w-"+wsID[:8])
	if err != nil {
		t.Fatal(err)
	}

	var hits atomic.Int64
	h := middleware.Idempotency(sqlDB)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))

	mk := func() *http.Request {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+wsID+"/issues", nil)
		req.Header.Set("Idempotency-Key", "k-same")
		return req.WithContext(context.WithValue(req.Context(), middleware.WorkspaceIDKey, wsID))
	}

	rr1 := httptest.NewRecorder()
	h.ServeHTTP(rr1, mk())
	if rr1.Code != 201 {
		t.Fatalf("first: %d", rr1.Code)
	}

	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, mk())
	if rr2.Code != 201 {
		t.Fatalf("replay: %d", rr2.Code)
	}
	if rr2.Header().Get("X-Idempotent-Replay") != "true" {
		t.Fatal("expected replay header")
	}
	body, _ := io.ReadAll(rr2.Body)
	if string(body) != `{"ok":true}` {
		t.Fatalf("body %s", body)
	}
	if hits.Load() != 1 {
		t.Fatalf("handler should run once, got %d", hits.Load())
	}
}

func TestIdempotencyDifferentKeysBothRun(t *testing.T) {
	sqlDB, err := db.Open("sqlite", filepath.Join(t.TempDir(), "idemp2.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	wsID := idgen.New()
	_, _ = sqlDB.Exec(`INSERT INTO workspaces (id, name, slug) VALUES (?,?,?)`, wsID, "W", "w-"+wsID[:8])

	var hits atomic.Int64
	h := middleware.Idempotency(sqlDB)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))

	for _, key := range []string{"a", "b"} {
		req := httptest.NewRequest(http.MethodPost, "/x", nil)
		req.Header.Set("Idempotency-Key", key)
		req = req.WithContext(context.WithValue(req.Context(), middleware.WorkspaceIDKey, wsID))
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != 200 {
			t.Fatalf("%s: %d", key, rr.Code)
		}
	}
	if hits.Load() != 2 {
		t.Fatalf("want 2 hits, got %d", hits.Load())
	}
}
