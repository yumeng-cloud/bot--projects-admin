package botmem_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/models"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/embedder"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/botmem"
)

func setup(t *testing.T) (*botmem.Service, string, string, string) {
	t.Helper()
	dsn := filepath.Join(t.TempDir(), "t.db")
	sqlDB, err := db.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })

	wsID := idgen.New()
	_, err = sqlDB.Exec(`INSERT INTO workspaces (id, name, slug) VALUES (?,?,?)`, wsID, "WS", "ws-"+wsID[:8])
	if err != nil {
		t.Fatal(err)
	}
	m1, m2 := idgen.New(), idgen.New()
	_, err = sqlDB.Exec(`INSERT INTO members (id, workspace_id, user_id, role, display_name, is_bot, bot_enabled) VALUES (?,?,NULL,'bot','BotA',1,1)`, m1, wsID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = sqlDB.Exec(`INSERT INTO members (id, workspace_id, user_id, role, display_name, is_bot, bot_enabled) VALUES (?,?,NULL,'bot','BotB',1,1)`, m2, wsID)
	if err != nil {
		t.Fatal(err)
	}
	return botmem.New(sqlDB), wsID, m1, m2
}

func TestClaimConflict(t *testing.T) {
	svc, ws, m1, m2 := setup(t)
	role, err := svc.CreateRole(ws, botmem.CreateRoleInput{Name: "前端", Slug: "frontend"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Claim(ws, role.ID, m1, false); err != nil {
		t.Fatal(err)
	}
	_, err = svc.Claim(ws, role.ID, m2, false)
	if err == nil {
		t.Fatal("expected conflict")
	}
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.Code != "conflict" {
		t.Fatalf("want conflict, got %#v", err)
	}
	// same member re-claim ok
	if _, err := svc.Claim(ws, role.ID, m1, false); err != nil {
		t.Fatal(err)
	}
	if err := svc.Release(ws, role.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Claim(ws, role.ID, m2, false); err != nil {
		t.Fatal(err)
	}
}

func TestMemoryUpsert(t *testing.T) {
	svc, ws, m1, _ := setup(t)
	a, err := svc.UpsertMemory(ws, m1, botmem.UpsertMemoryInput{KeyName: "ctx", ValueText: "v1", Tags: `["a"]`})
	if err != nil {
		t.Fatal(err)
	}
	b, err := svc.UpsertMemory(ws, m1, botmem.UpsertMemoryInput{KeyName: "ctx", ValueText: "v2", Tags: `["b"]`})
	if err != nil {
		t.Fatal(err)
	}
	if a.ID != b.ID {
		t.Fatalf("upsert should keep id: %s vs %s", a.ID, b.ID)
	}
	if b.ValueText != "v2" {
		t.Fatalf("value=%s", b.ValueText)
	}
	list, err := svc.ListMemories(ws, m1, "v2")
	if err != nil || len(list) != 1 {
		t.Fatalf("search: %v len=%d", err, len(list))
	}
}

func TestHandoffAssemble(t *testing.T) {
	svc, ws, m1, _ := setup(t)
	role, err := svc.CreateRole(ws, botmem.CreateRoleInput{Name: "后端", Slug: "backend"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Claim(ws, role.ID, m1, false); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UpsertMemory(ws, m1, botmem.UpsertMemoryInput{KeyName: "note", ValueText: "hello"}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateTodo(ws, m1, botmem.CreateTodoInput{Title: "open", Status: "todo"}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateTodo(ws, m1, botmem.CreateTodoInput{Title: "done", Status: "done"}); err != nil {
		t.Fatal(err)
	}

	proj := idgen.New()
	_, err = svc.DB.Exec(`INSERT INTO projects (id, workspace_id, name, key_prefix) VALUES (?,?,?,?)`, proj, ws, "P", "P")
	if err != nil {
		t.Fatal(err)
	}
	issOpen, issDone := idgen.New(), idgen.New()
	_, err = svc.DB.Exec(`INSERT INTO issues (id, project_id, workspace_id, number, identifier, title, status, priority, assignee_id)
		VALUES (?,?,?,1,'P-1','open iss','in_progress','medium',?)`, issOpen, proj, ws, m1)
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.DB.Exec(`INSERT INTO issues (id, project_id, workspace_id, number, identifier, title, status, priority, assignee_id)
		VALUES (?,?,?,2,'P-2','done iss','done','low',?)`, issDone, proj, ws, m1)
	if err != nil {
		t.Fatal(err)
	}

	h, err := svc.Handoff(ws, m1)
	if err != nil {
		t.Fatal(err)
	}
	if h.ClaimedRole == nil || h.ClaimedRole.Slug != "backend" {
		t.Fatalf("claimed role: %+v", h.ClaimedRole)
	}
	if len(h.Memories) != 1 || h.Memories[0].KeyName != "note" {
		t.Fatalf("memories: %+v", h.Memories)
	}
	if len(h.Todos) != 1 || h.Todos[0].Title != "open" {
		t.Fatalf("todos: %+v", h.Todos)
	}
	if len(h.Issues) != 1 || h.Issues[0].Identifier != "P-1" {
		t.Fatalf("issues: %+v", h.Issues)
	}
}

func TestImportHandoff(t *testing.T) {
	svc, ws, m1, m2 := setup(t)
	role, err := svc.CreateRole(ws, botmem.CreateRoleInput{Name: "前端", Slug: "frontend"})
	if err != nil {
		t.Fatal(err)
	}
	// m2 already claims → import claim for m1 should warn
	if _, err := svc.Claim(ws, role.ID, m2, false); err != nil {
		t.Fatal(err)
	}

	proj := idgen.New()
	if _, err := svc.DB.Exec(`INSERT INTO projects (id, workspace_id, name, key_prefix) VALUES (?,?,?,?)`, proj, ws, "P", "P"); err != nil {
		t.Fatal(err)
	}
	issueID := idgen.New()
	if _, err := svc.DB.Exec(`INSERT INTO issues (id, project_id, workspace_id, number, identifier, title, status, priority)
		VALUES (?,?,?,1,'P-1','open','todo','medium')`, issueID, proj, ws); err != nil {
		t.Fatal(err)
	}
	in := botmem.ImportHandoffInput{
		Memories: []models.BotMemory{
			{KeyName: "ctx", ValueText: "imported", Tags: `["x"]`},
			{ID: "foreign-id-ignored", KeyName: "note", ValueText: "n1"},
		},
		Todos: []models.BotTodo{
			{ID: "foreign-todo", Title: "work-a", Status: "todo", IssueID: &issueID},
			{Title: "done-skip", Status: "done"},
			{Title: "work-b", Status: "doing"},
		},
		ClaimedRoleSlug: strPtr("frontend"),
		Issues:          []models.HandoffIssue{{ID: issueID}, {ID: "iss-2"}},
	}
	res, err := svc.ImportHandoff(ws, m1, in)
	if err != nil {
		t.Fatal(err)
	}
	if res.ImportedMemories != 2 {
		t.Fatalf("imported_memories=%d", res.ImportedMemories)
	}
	if res.ImportedTodos != 2 {
		t.Fatalf("imported_todos=%d want 2", res.ImportedTodos)
	}
	if res.Claimed {
		t.Fatal("should not claim when conflict")
	}
	if len(res.Warnings) == 0 {
		t.Fatal("expected claim conflict warning")
	}
	if len(res.IssueIDs) != 2 {
		t.Fatalf("issue_ids=%v", res.IssueIDs)
	}

	mems, err := svc.ListMemories(ws, m1, "")
	if err != nil || len(mems) != 2 {
		t.Fatalf("ListMemories: %v len=%d", err, len(mems))
	}
	todos, err := svc.ListTodos(ws, m1, "")
	if err != nil || len(todos) != 2 {
		t.Fatalf("ListTodos: %v len=%d", err, len(todos))
	}

	// re-import same title+issue_id updates, no duplicate
	res2, err := svc.ImportHandoff(ws, m1, botmem.ImportHandoffInput{
		Handoff: &models.BotHandoff{
			Todos:    []models.BotTodo{{Title: "work-a", Status: "doing", IssueID: &issueID}},
			Memories: []models.BotMemory{{KeyName: "ctx", ValueText: "v2"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res2.ImportedTodos != 1 || res2.ImportedMemories != 1 {
		t.Fatalf("reimport counts mem=%d todo=%d", res2.ImportedMemories, res2.ImportedTodos)
	}
	todos, err = svc.ListTodos(ws, m1, "")
	if err != nil || len(todos) != 2 {
		t.Fatalf("after reimport ListTodos len=%d err=%v", len(todos), err)
	}
	found := false
	for _, td := range todos {
		if td.Title == "work-a" && td.Status == "doing" {
			found = true
		}
	}
	if !found {
		t.Fatalf("work-a not updated to doing: %+v", todos)
	}
	if res2.Handoff == nil || len(res2.Handoff.Memories) < 1 {
		t.Fatal("result handoff missing")
	}
}

func strPtr(s string) *string { return &s }

func TestSearchMemoriesKeywordFallback(t *testing.T) {
	svc, ws, m1, _ := setup(t)
	if _, err := svc.UpsertMemory(ws, m1, botmem.UpsertMemoryInput{KeyName: "pref", ValueText: "likes dark mode UI"}); err != nil {
		t.Fatal(err)
	}
	// 无 Embed/Store：semantic/hybrid → keyword_fallback
	res, err := svc.SearchMemories(context.Background(), ws, m1, "dark", botmem.ModeHybrid, 10)
	if err != nil {
		t.Fatal(err)
	}
	if res.SearchBackend != "keyword_fallback" {
		t.Fatalf("backend=%s", res.SearchBackend)
	}
	if len(res.Memories) != 1 || res.Memories[0].KeyName != "pref" {
		t.Fatalf("memories=%+v", res.Memories)
	}

	res2, err := svc.SearchMemories(context.Background(), ws, m1, "dark", botmem.ModeKeyword, 10)
	if err != nil {
		t.Fatal(err)
	}
	if res2.SearchBackend != "keyword" {
		t.Fatalf("backend=%s", res2.SearchBackend)
	}
}

func TestSearchMemoriesWithHashEmbedderNoStore(t *testing.T) {
	svc, ws, m1, _ := setup(t)
	// 仅 Embed 无 Store → 仍应 fallback（canVector 需两者）
	svc.WithEmbedder(embedder.NewHash("test", 32), nil)
	if _, err := svc.UpsertMemory(ws, m1, botmem.UpsertMemoryInput{KeyName: "a", ValueText: "hello vector world"}); err != nil {
		t.Fatal(err)
	}
	res, err := svc.SearchMemories(context.Background(), ws, m1, "vector", botmem.ModeSemantic, 5)
	if err != nil {
		t.Fatal(err)
	}
	if res.SearchBackend != "keyword_fallback" {
		t.Fatalf("want keyword_fallback got %s", res.SearchBackend)
	}
	if len(res.Memories) != 1 {
		t.Fatalf("len=%d", len(res.Memories))
	}
}

// 可选：DB_DRIVER=postgres + DATABASE_URL + 已装 pgvector / embedding_chunks 时跑通。
func TestSearchMemoriesPostgresOptional(t *testing.T) {
	if !db.IsPostgres(os.Getenv("DB_DRIVER")) {
		t.Skip("DB_DRIVER!=postgres")
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL empty")
	}
	// 避免无超权用户时 CREATE EXTENSION 失败拖垮可选测
	t.Setenv("AUTO_MIGRATE", "0")
	sqlDB, err := db.Open("postgres", dsn)
	if err != nil {
		t.Skipf("open postgres: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })

	var ext int
	if err := sqlDB.QueryRow(`SELECT COUNT(1) FROM pg_extension WHERE extname='vector'`).Scan(&ext); err != nil || ext == 0 {
		t.Skip("pgvector extension not installed (run scripts/pg-bootstrap.sh as superuser)")
	}
	var tbl int
	if err := sqlDB.QueryRow(`SELECT COUNT(1) FROM information_schema.tables WHERE table_name='embedding_chunks'`).Scan(&tbl); err != nil || tbl == 0 {
		t.Skip("embedding_chunks missing; enable AUTO_MIGRATE after bootstrap")
	}

	wsID := idgen.New()
	m1 := idgen.New()
	if _, err := sqlDB.Exec(`INSERT INTO workspaces (id, name, slug) VALUES ($1,$2,$3)`, wsID, "WS-emb", "ws-emb-"+wsID[:8]); err != nil {
		t.Skipf("insert workspace: %v", err)
	}
	t.Cleanup(func() {
		_, _ = sqlDB.Exec(`DELETE FROM embedding_chunks WHERE workspace_id=$1`, wsID)
		_, _ = sqlDB.Exec(`DELETE FROM bot_memories WHERE workspace_id=$1`, wsID)
		_, _ = sqlDB.Exec(`DELETE FROM members WHERE workspace_id=$1`, wsID)
		_, _ = sqlDB.Exec(`DELETE FROM workspaces WHERE id=$1`, wsID)
	})
	if _, err := sqlDB.Exec(`INSERT INTO members (id, workspace_id, user_id, role, display_name, is_bot, bot_enabled) VALUES ($1,$2,NULL,'bot','BotEmb',1,1)`, m1, wsID); err != nil {
		t.Fatal(err)
	}

	svc := botmem.New(sqlDB).WithEmbedder(embedder.NewHash("hash-test", 384), embedder.NewStore(sqlDB))
	if _, err := svc.UpsertMemory(wsID, m1, botmem.UpsertMemoryInput{KeyName: "theme", ValueText: "prefer dark theme for IDE"}); err != nil {
		t.Fatal(err)
	}
	res, err := svc.SearchMemories(context.Background(), wsID, m1, "dark theme", botmem.ModeHybrid, 10)
	if err != nil {
		t.Fatal(err)
	}
	if res.SearchBackend != "hybrid" {
		t.Fatalf("backend=%s", res.SearchBackend)
	}
	if len(res.Memories) < 1 {
		t.Fatalf("expected hit, got %+v", res)
	}
}

// X-G2：错仓 mid 读/写 memories+todos → 404「成员不存在」；同仓正常。
func TestWSIsolationMemoriesTodos(t *testing.T) {
	svc, ws1, m1, _ := setup(t)
	sqlDB := svc.DB

	ws2 := idgen.New()
	if _, err := sqlDB.Exec(`INSERT INTO workspaces (id, name, slug) VALUES (?,?,?)`, ws2, "WS2", "ws2-"+ws2[:8]); err != nil {
		t.Fatal(err)
	}
	m2 := idgen.New()
	if _, err := sqlDB.Exec(`INSERT INTO members (id, workspace_id, user_id, role, display_name, is_bot, bot_enabled) VALUES (?,?,NULL,'bot','BotWS2',1,1)`, m2, ws2); err != nil {
		t.Fatal(err)
	}

	// 同仓写/读 OK
	mem, err := svc.UpsertMemory(ws1, m1, botmem.UpsertMemoryInput{KeyName: "k", ValueText: "v"})
	if err != nil {
		t.Fatalf("same-ws UpsertMemory: %v", err)
	}
	list, err := svc.ListMemories(ws1, m1, "")
	if err != nil || len(list) != 1 {
		t.Fatalf("same-ws ListMemories: %v len=%d", err, len(list))
	}
	todo, err := svc.CreateTodo(ws1, m1, botmem.CreateTodoInput{Title: "t1"})
	if err != nil {
		t.Fatalf("same-ws CreateTodo: %v", err)
	}
	todos, err := svc.ListTodos(ws1, m1, "")
	if err != nil || len(todos) != 1 {
		t.Fatalf("same-ws ListTodos: %v len=%d", err, len(todos))
	}

	must404 := func(name string, err error) {
		t.Helper()
		if err == nil {
			t.Fatalf("%s: want 404 成员不存在, got nil", name)
		}
		ae, ok := err.(*apperr.AppError)
		if !ok || ae.Code != "not_found" || ae.Message != "成员不存在" {
			t.Fatalf("%s: want not_found 成员不存在, got %#v", name, err)
		}
	}

	// 错仓：ws2 path + ws1 mid（写 IDOR 回归）
	_, err = svc.UpsertMemory(ws2, m1, botmem.UpsertMemoryInput{KeyName: "evil", ValueText: "x"})
	must404("UpsertMemory cross-ws", err)
	_, err = svc.CreateTodo(ws2, m1, botmem.CreateTodoInput{Title: "evil"})
	must404("CreateTodo cross-ws", err)
	_, err = svc.ListMemories(ws2, m1, "")
	must404("ListMemories cross-ws", err)
	_, err = svc.ListTodos(ws2, m1, "")
	must404("ListTodos cross-ws", err)
	_, err = svc.PatchMemory(ws2, m1, mem.ID, botmem.PatchMemoryInput{})
	must404("PatchMemory cross-ws", err)
	err = svc.DeleteMemory(ws2, m1, mem.ID)
	must404("DeleteMemory cross-ws", err)
	_, err = svc.PatchTodo(ws2, m1, todo.ID, botmem.PatchTodoInput{})
	must404("PatchTodo cross-ws", err)

	// 同仓 patch/delete 仍可用
	if _, err := svc.PatchMemory(ws1, m1, mem.ID, botmem.PatchMemoryInput{ValueText: strPtr("v2")}); err != nil {
		t.Fatalf("same-ws PatchMemory: %v", err)
	}
	st := "done"
	if _, err := svc.PatchTodo(ws1, m1, todo.ID, botmem.PatchTodoInput{Status: &st}); err != nil {
		t.Fatalf("same-ws PatchTodo: %v", err)
	}
}
