package attachments_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/attachments"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
)

func setupAttach(t *testing.T) (*attachments.Service, string, string, string, string, string) {
	t.Helper()
	root := t.TempDir()
	dsn := filepath.Join(t.TempDir(), "t.db")
	sqlDB, err := db.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })

	wsID := idgen.New()
	projID := idgen.New()
	issueID := idgen.New()
	uploaderID := idgen.New()
	otherID := idgen.New()
	adminID := idgen.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)

	mustExec := func(q string, args ...any) {
		t.Helper()
		if _, err := sqlDB.Exec(q, args...); err != nil {
			t.Fatalf("exec: %v\n%s", err, q)
		}
	}
	mustExec(`INSERT INTO workspaces (id, name, slug) VALUES (?,?,?)`, wsID, "T", "t-"+wsID[:8])
	mustExec(`INSERT INTO members (id, workspace_id, user_id, role, display_name, is_bot, bot_enabled)
		VALUES (?,?,NULL,'member','Uploader',0,1)`, uploaderID, wsID)
	mustExec(`INSERT INTO members (id, workspace_id, user_id, role, display_name, is_bot, bot_enabled)
		VALUES (?,?,NULL,'member','Other',0,1)`, otherID, wsID)
	mustExec(`INSERT INTO members (id, workspace_id, user_id, role, display_name, is_bot, bot_enabled)
		VALUES (?,?,NULL,'admin','Admin',0,1)`, adminID, wsID)
	mustExec(`INSERT INTO projects (id, workspace_id, name, key_prefix, created_at, updated_at)
		VALUES (?,?,?,?,?,?)`, projID, wsID, "P", "TST", now, now)
	mustExec(`INSERT INTO issues (id, workspace_id, project_id, number, identifier, title, status, priority, version, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		issueID, wsID, projID, 1, "TST-1", "Hello", "todo", "none", 1, now, now)

	svc := attachments.New(sqlDB, root)
	return svc, wsID, issueID, uploaderID, otherID, adminID
}

func insertAttachment(t *testing.T, svc *attachments.Service, wsID, issueID, uploaderID, commentID, filename string) *attachments.Attachment {
	t.Helper()
	id := idgen.New()
	key := filepath.ToSlash(filepath.Join(wsID, id))
	full := filepath.Join(svc.Store.Root, key)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte("fake-bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	var commentArg any
	if commentID != "" {
		commentArg = commentID
	}
	_, err := svc.DB.Exec(`
		INSERT INTO attachments
			(id, workspace_id, issue_id, comment_id, uploader_id, filename, content_type, size_bytes, storage_key, created_at)
		VALUES (?,?,?,?,?,?,?,?,?,?)`,
		id, wsID, issueID, commentArg, uploaderID, filename, "text/plain", int64(10), key, now)
	if err != nil {
		t.Fatal(err)
	}
	att, err := svc.Get(context.Background(), wsID, id)
	if err != nil {
		t.Fatal(err)
	}
	return att
}

func TestListByIssueIncludesCommentLevel(t *testing.T) {
	svc, ws, issueID, uploader, _, _ := setupAttach(t)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	commentID := idgen.New()
	if _, err := svc.DB.Exec(`INSERT INTO comments (id, issue_id, author_id, body, created_at, updated_at)
		VALUES (?,?,?,?,?,?)`, commentID, issueID, uploader, "c", now, now); err != nil {
		t.Fatal(err)
	}

	a1 := insertAttachment(t, svc, ws, issueID, uploader, "", "a.txt")
	a2 := insertAttachment(t, svc, ws, issueID, uploader, commentID, "b.txt")

	list, err := svc.ListByIssue(context.Background(), ws, issueID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("want 2, got %d", len(list))
	}
	ids := map[string]bool{list[0].ID: true, list[1].ID: true}
	if !ids[a1.ID] || !ids[a2.ID] {
		t.Fatalf("missing ids: %+v", list)
	}
	for _, a := range list {
		wantURL := "/api/v1/workspaces/" + ws + "/attachments/" + a.ID + "/content"
		if a.URL != wantURL {
			t.Fatalf("url=%q want %q", a.URL, wantURL)
		}
	}

	clist, err := svc.ListByComment(context.Background(), ws, issueID, commentID)
	if err != nil {
		t.Fatal(err)
	}
	if len(clist) != 1 || clist[0].ID != a2.ID {
		t.Fatalf("comment list=%+v", clist)
	}
}

func TestDeleteUploaderAndAdmin(t *testing.T) {
	svc, ws, issueID, uploader, other, admin := setupAttach(t)
	att := insertAttachment(t, svc, ws, issueID, uploader, "", "x.txt")
	full := filepath.Join(svc.Store.Root, att.StorageKey)
	if _, err := os.Stat(full); err != nil {
		t.Fatalf("file missing before delete: %v", err)
	}

	// other non-admin cannot
	if attachments.CanDelete(other, "member", att) {
		t.Fatal("other member should not CanDelete")
	}

	// uploader can
	if !attachments.CanDelete(uploader, "member", att) {
		t.Fatal("uploader should CanDelete")
	}
	if err := svc.Delete(context.Background(), ws, att.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(full); !os.IsNotExist(err) {
		t.Fatalf("file should be gone, err=%v", err)
	}
	_, err := svc.Get(context.Background(), ws, att.ID)
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.StatusCode != 404 {
		t.Fatalf("want 404 after delete, got %v", err)
	}

	// admin can delete another's
	att2 := insertAttachment(t, svc, ws, issueID, uploader, "", "y.txt")
	if !attachments.CanDelete(admin, "admin", att2) {
		t.Fatal("admin should CanDelete")
	}
	if attachments.CanDelete(other, "member", att2) {
		t.Fatal("other still should not")
	}
	if err := svc.Delete(context.Background(), ws, att2.ID); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteMissingFileOK(t *testing.T) {
	svc, ws, issueID, uploader, _, _ := setupAttach(t)
	att := insertAttachment(t, svc, ws, issueID, uploader, "", "z.txt")
	_ = os.Remove(filepath.Join(svc.Store.Root, att.StorageKey))
	if err := svc.Delete(context.Background(), ws, att.ID); err != nil {
		t.Fatalf("missing file should not block delete: %v", err)
	}
}

func TestCanDeleteNil(t *testing.T) {
	if attachments.CanDelete("x", "admin", nil) {
		t.Fatal("nil att")
	}
	if attachments.CanDelete("", "admin", &attachments.Attachment{UploaderID: "u"}) {
		t.Fatal("empty actor")
	}
}
