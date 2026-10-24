package attachments_test

import (
	"bytes"
	"context"
	"database/sql"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/attachments"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
)

type fixture struct {
	DB       *sql.DB
	Svc      *attachments.Service
	StoreDir string
	WS       string
	Issue    string
	Comment  string
	Admin    string
	Member   string
	Other    string
}

func setup(t *testing.T) *fixture {
	t.Helper()
	dir := t.TempDir()
	dsn := filepath.Join(dir, "t.db")
	sqlDB, err := db.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	storeDir := filepath.Join(dir, "attachments")
	svc := attachments.New(sqlDB, storeDir)

	ws := idgen.New()
	proj := idgen.New()
	issue := idgen.New()
	comment := idgen.New()
	admin := idgen.New()
	member := idgen.New()
	other := idgen.New()
	u1, u2, u3 := idgen.New(), idgen.New(), idgen.New()

	mustExec(t, sqlDB, `INSERT INTO workspaces (id, name, slug) VALUES (?,?,?)`, ws, "W", "w-"+ws[:8])
	mustExec(t, sqlDB, `INSERT INTO users (id, email, password_hash, name) VALUES (?,?,?,?)`, u1, "a@t.com", "x", "A")
	mustExec(t, sqlDB, `INSERT INTO users (id, email, password_hash, name) VALUES (?,?,?,?)`, u2, "m@t.com", "x", "M")
	mustExec(t, sqlDB, `INSERT INTO users (id, email, password_hash, name) VALUES (?,?,?,?)`, u3, "o@t.com", "x", "O")
	mustExec(t, sqlDB, `INSERT INTO members (id, workspace_id, user_id, role, display_name) VALUES (?,?,?,?,?)`, admin, ws, u1, "admin", "Admin")
	mustExec(t, sqlDB, `INSERT INTO members (id, workspace_id, user_id, role, display_name) VALUES (?,?,?,?,?)`, member, ws, u2, "member", "Member")
	mustExec(t, sqlDB, `INSERT INTO members (id, workspace_id, user_id, role, display_name) VALUES (?,?,?,?,?)`, other, ws, u3, "member", "Other")
	mustExec(t, sqlDB, `INSERT INTO projects (id, workspace_id, name, key_prefix) VALUES (?,?,?,?)`, proj, ws, "P", "P")
	mustExec(t, sqlDB, `INSERT INTO issue_counters (project_id, next_number) VALUES (?,?)`, proj, 2)
	mustExec(t, sqlDB, `INSERT INTO issues (id, project_id, workspace_id, number, identifier, title) VALUES (?,?,?,?,?,?)`,
		issue, proj, ws, 1, "P-1", "t")
	mustExec(t, sqlDB, `INSERT INTO comments (id, issue_id, author_id, body) VALUES (?,?,?,?)`, comment, issue, member, "hi")

	return &fixture{DB: sqlDB, Svc: svc, StoreDir: storeDir, WS: ws, Issue: issue, Comment: comment, Admin: admin, Member: member, Other: other}
}

func mustExec(t *testing.T, db *sql.DB, q string, args ...any) {
	t.Helper()
	if _, err := db.Exec(q, args...); err != nil {
		t.Fatalf("exec: %v\n%s", err, q)
	}
}

func tinyPNG() []byte {
	// 1x1 PNG
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xde, 0x00, 0x00, 0x00,
		0x0c, 0x49, 0x44, 0x41, 0x54, 0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0x00,
		0x00, 0x00, 0x03, 0x00, 0x01, 0x00, 0x05, 0xfe, 0xd4, 0xef, 0x00, 0x00,
		0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
	}
}

func asAppErr(t *testing.T, err error) *apperr.AppError {
	t.Helper()
	ae, ok := err.(*apperr.AppError)
	if !ok {
		t.Fatalf("want AppError, got %T %v", err, err)
	}
	return ae
}

func countAttachments(t *testing.T, db *sql.DB) int {
	t.Helper()
	var n int
	if err := db.QueryRow(`SELECT COUNT(1) FROM attachments`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func countFiles(t *testing.T, root string) int {
	t.Helper()
	n := 0
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err == nil && info != nil && !info.IsDir() {
			n++
		}
		return nil
	})
	return n
}

func TestUploadIssuePNG(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	png := tinyPNG()
	att, err := f.Svc.UploadIssue(ctx, f.WS, f.Issue, f.Member, "shot.png", "image/png", bytes.NewReader(png), int64(len(png)))
	if err != nil {
		t.Fatal(err)
	}
	if att.ContentType != "image/png" || att.SizeBytes != int64(len(png)) {
		t.Fatalf("meta %+v", att)
	}
	if !strings.Contains(att.URL, "/attachments/"+att.ID+"/content") {
		t.Fatalf("url %s", att.URL)
	}
	got, err := f.Svc.Get(ctx, f.WS, att.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Filename != "shot.png" {
		t.Fatalf("filename %s", got.Filename)
	}
	meta, rc, err := f.Svc.OpenContent(ctx, f.WS, att.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	body, _ := io.ReadAll(rc)
	if len(body) != len(png) || meta.ID != att.ID {
		t.Fatalf("content mismatch")
	}
}

func TestUploadComment(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	txt := []byte("hello")
	att, err := f.Svc.UploadComment(ctx, f.WS, f.Issue, f.Comment, f.Member, "note.txt", "text/plain", bytes.NewReader(txt), int64(len(txt)))
	if err != nil {
		t.Fatal(err)
	}
	if att.CommentID == nil || *att.CommentID != f.Comment {
		t.Fatalf("comment_id %+v", att.CommentID)
	}
	list, err := f.Svc.ListByIssue(ctx, f.WS, f.Issue)
	if err != nil || len(list) != 1 {
		t.Fatalf("list by issue: %v len=%d", err, len(list))
	}
	clist, err := f.Svc.ListByComment(ctx, f.WS, f.Issue, f.Comment)
	if err != nil || len(clist) != 1 {
		t.Fatalf("list by comment: %v", err)
	}
}

func TestRejectEXENoDirty(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	payload := []byte{0x4d, 0x5a, 0x00, 0x00} // MZ
	_, err := f.Svc.UploadIssue(ctx, f.WS, f.Issue, f.Member, "evil.exe", "application/octet-stream", bytes.NewReader(payload), int64(len(payload)))
	ae := asAppErr(t, err)
	if ae.Message != "不支持的文件类型" {
		t.Fatalf("msg=%q", ae.Message)
	}
	if countAttachments(t, f.DB) != 0 {
		t.Fatal("ghost DB row")
	}
	if countFiles(t, f.StoreDir) != 0 {
		t.Fatal("ghost file")
	}
}

func TestRejectOversizeNoDirty(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	old := attachments.MaxBytes
	attachments.MaxBytes = 32
	t.Cleanup(func() { attachments.MaxBytes = old })

	big := bytes.Repeat([]byte("a"), 64)
	_, err := f.Svc.UploadIssue(ctx, f.WS, f.Issue, f.Member, "big.txt", "text/plain", bytes.NewReader(big), int64(len(big)))
	ae := asAppErr(t, err)
	if ae.Message != "文件过大，上限 10MB" {
		t.Fatalf("msg=%q", ae.Message)
	}
	if countAttachments(t, f.DB) != 0 || countFiles(t, f.StoreDir) != 0 {
		t.Fatal("dirty after oversize")
	}
}

func TestCanDeleteAndDelete(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	png := tinyPNG()
	att, err := f.Svc.UploadIssue(ctx, f.WS, f.Issue, f.Member, "a.png", "image/png", bytes.NewReader(png), int64(len(png)))
	if err != nil {
		t.Fatal(err)
	}

	// uploader ok
	if !attachments.CanDelete(f.Member, "member", att) {
		t.Fatal("uploader should delete")
	}
	// other member no
	if attachments.CanDelete(f.Other, "member", att) {
		t.Fatal("other member must not delete")
	}
	// admin ok
	if !attachments.CanDelete(f.Admin, "admin", att) {
		t.Fatal("admin should delete")
	}
	// bot with uploader member id ok
	if !attachments.CanDelete(f.Member, "bot", att) {
		t.Fatal("bot bound to uploader should delete")
	}

	if err := f.Svc.Delete(ctx, f.WS, att.ID); err != nil {
		t.Fatal(err)
	}
	_, err = f.Svc.Get(ctx, f.WS, att.ID)
	if asAppErr(t, err).StatusCode != 404 {
		t.Fatalf("want 404 after delete, got %v", err)
	}
	_, _, err = f.Svc.OpenContent(ctx, f.WS, att.ID)
	if asAppErr(t, err).StatusCode != 404 {
		t.Fatalf("content after delete: %v", err)
	}
}

func TestListIncludesCommentAttachments(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	png := tinyPNG()
	_, _ = f.Svc.UploadIssue(ctx, f.WS, f.Issue, f.Member, "i.png", "image/png", bytes.NewReader(png), int64(len(png)))
	_, _ = f.Svc.UploadComment(ctx, f.WS, f.Issue, f.Comment, f.Member, "c.txt", "text/plain", bytes.NewReader([]byte("x")), 1)
	list, err := f.Svc.ListByIssue(ctx, f.WS, f.Issue)
	if err != nil || len(list) != 2 {
		t.Fatalf("want 2, got %d err=%v", len(list), err)
	}
}
