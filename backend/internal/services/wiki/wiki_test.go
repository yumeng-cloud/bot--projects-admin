package wiki_test

import (
	"path/filepath"
	"testing"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/wiki"
)

func setupDB(t *testing.T) (*wiki.Service, string, string) {
	t.Helper()
	dsn := filepath.Join(t.TempDir(), "test.db")
	sqlDB, err := db.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	wsID := idgen.New()
	userID := idgen.New()
	memID := idgen.New()
	_, err = sqlDB.Exec(`INSERT INTO users (id, email, password_hash, name) VALUES (?,?,?,?)`,
		userID, "wiki@test.local", "x", "Wiki Tester")
	if err != nil {
		t.Fatal(err)
	}
	_, err = sqlDB.Exec(`INSERT INTO workspaces (id, name, slug) VALUES (?,?,?)`, wsID, "WS", "ws-wiki")
	if err != nil {
		t.Fatal(err)
	}
	_, err = sqlDB.Exec(`INSERT INTO members (id, workspace_id, user_id, role, display_name) VALUES (?,?,?,?,?)`,
		memID, wsID, userID, "admin", "Tester")
	if err != nil {
		t.Fatal(err)
	}
	return wiki.New(sqlDB), wsID, memID
}

func TestCreatePageAndSearch(t *testing.T) {
	svc, wsID, memID := setupDB(t)

	sp, err := svc.CreateSpace(wsID, wiki.CreateSpaceInput{Name: "手册", Slug: "handbook"})
	if err != nil {
		t.Fatalf("create space: %v", err)
	}

	page, err := svc.CreatePage(wsID, sp.ID, wiki.CreatePageInput{
		Title: "产品约定", BodyMD: "# 约定\nBot 先读 Wiki 再动手。", MemberID: memID,
	})
	if err != nil {
		t.Fatalf("create page: %v", err)
	}
	if page.Title != "产品约定" || page.BodyMD == "" {
		t.Fatalf("unexpected page: %+v", page)
	}

	child, err := svc.CreatePage(wsID, sp.ID, wiki.CreatePageInput{
		Title: "子页", BodyMD: "child", ParentID: &page.ID, MemberID: memID,
	})
	if err != nil {
		t.Fatalf("create child: %v", err)
	}

	tree, err := svc.Tree(wsID, sp.ID)
	if err != nil {
		t.Fatalf("tree: %v", err)
	}
	if len(tree) != 1 || len(tree[0].Children) != 1 {
		t.Fatalf("tree shape: %+v", tree)
	}
	if tree[0].Children[0].ID != child.ID {
		t.Fatalf("child mismatch")
	}

	found, err := svc.Search(wsID, "Bot 先读", 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(found) != 1 || found[0].ID != page.ID {
		t.Fatalf("search result: %+v", found)
	}

	found2, err := svc.Search(wsID, "产品", 10)
	if err != nil || len(found2) == 0 {
		t.Fatalf("title search: %v %+v", err, found2)
	}
}

func TestSpaceCRUD(t *testing.T) {
	svc, wsID, _ := setupDB(t)
	sp, err := svc.CreateSpace(wsID, wiki.CreateSpaceInput{Name: "A", Description: "d"})
	if err != nil {
		t.Fatal(err)
	}
	name := "B"
	sp2, err := svc.UpdateSpace(wsID, sp.ID, wiki.UpdateSpaceInput{Name: &name})
	if err != nil || sp2.Name != "B" {
		t.Fatalf("update: %v %+v", err, sp2)
	}
	list, err := svc.ListSpaces(wsID)
	if err != nil || len(list) != 1 {
		t.Fatalf("list: %v %d", err, len(list))
	}
	if err := svc.DeleteSpace(wsID, sp.ID); err != nil {
		t.Fatal(err)
	}
}
