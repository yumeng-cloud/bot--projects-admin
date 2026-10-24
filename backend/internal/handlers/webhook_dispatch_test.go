package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/issueboard"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestDispatchSyncWritesDelivery(t *testing.T) {
	dsn := filepath.Join(t.TempDir(), "wh.db")
	sqlDB, err := db.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	wsID := idgen.New()
	if _, err := sqlDB.Exec(`INSERT INTO workspaces (id, name, slug) VALUES (?,?,?)`, wsID, "W", "w-"+wsID[:8]); err != nil {
		t.Fatal(err)
	}
	hookID := idgen.New()
	eventsJSON, _ := json.Marshal([]string{"*"})
	if _, err := sqlDB.Exec(
		`INSERT INTO webhooks (id, workspace_id, url, secret, events, enabled) VALUES (?,?,?,?,?,1)`,
		hookID, wsID, "http://webhook.test/hook", "sec", string(eventsJSON),
	); err != nil {
		t.Fatal(err)
	}
	h := &WebhookHandler{
		DB: sqlDB,
		Client: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			_, _ = io.Copy(io.Discard, r.Body)
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader([]byte("ok"))),
				Header:     make(http.Header),
				Request:    r,
			}, nil
		})},
	}
	h.dispatchSync(wsID, "issue.assigned", map[string]any{"id": "iss1"})
	var n int
	if err := sqlDB.QueryRow(`SELECT COUNT(*) FROM webhook_deliveries WHERE webhook_id=? AND event=? AND success=1`,
		hookID, "issue.assigned").Scan(&n); err != nil || n != 1 {
		t.Fatalf("n=%d err=%v", n, err)
	}
}

func TestIssueUpdateDispatchesWebhookDelivery(t *testing.T) {
	dsn := filepath.Join(t.TempDir(), "wh2.db")
	sqlDB, err := db.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })

	wsID := idgen.New()
	if _, err := sqlDB.Exec(`INSERT INTO workspaces (id, name, slug) VALUES (?,?,?)`, wsID, "W", "w-"+wsID[:8]); err != nil {
		t.Fatal(err)
	}

	hookH := &WebhookHandler{
		DB: sqlDB,
		Client: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			_, _ = io.Copy(io.Discard, r.Body)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte("ok"))),
				Header:     make(http.Header),
				Request:    r,
			}, nil
		})},
	}

	hookID := idgen.New()
	eventsJSON, _ := json.Marshal([]string{"*"})
	if _, err := sqlDB.Exec(
		`INSERT INTO webhooks (id, workspace_id, url, secret, events, enabled) VALUES (?,?,?,?,?,1)`,
		hookID, wsID, "http://webhook.test/hook", "test-secret", string(eventsJSON),
	); err != nil {
		t.Fatal(err)
	}

	board := issueboard.New(sqlDB)
	board.DispatchEvent = func(workspaceID, event string, payload any) {
		hookH.dispatchSync(workspaceID, event, payload)
	}

	proj, err := board.CreateProject(wsID, issueboard.CreateProjectInput{Name: "P", KeyPrefix: "WH"})
	if err != nil {
		t.Fatal(err)
	}
	mid := idgen.New()
	if _, err := sqlDB.Exec(`INSERT INTO members (id, workspace_id, user_id, role, display_name, is_bot, bot_enabled)
		VALUES (?,?,NULL,'member','M',0,1)`, mid, wsID); err != nil {
		t.Fatal(err)
	}

	iss, err := board.CreateIssue(wsID, issueboard.CreateIssueInput{
		ProjectID: proj.ID, Title: "t", ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	var createdN int
	if err := sqlDB.QueryRow(`SELECT COUNT(*) FROM webhook_deliveries WHERE webhook_id=? AND event='issue.created' AND success=1`,
		hookID).Scan(&createdN); err != nil || createdN < 1 {
		t.Fatalf("created deliveries=%d err=%v", createdN, err)
	}

	_, err = board.UpdateIssue(wsID, iss.ID, issueboard.UpdateIssueInput{
		Version: iss.Version, AssigneeID: &mid, ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	var assignedN, updatedN int
	_ = sqlDB.QueryRow(`SELECT COUNT(*) FROM webhook_deliveries WHERE webhook_id=? AND event='issue.assigned' AND success=1`,
		hookID).Scan(&assignedN)
	_ = sqlDB.QueryRow(`SELECT COUNT(*) FROM webhook_deliveries WHERE webhook_id=? AND event='issue.updated' AND success=1`,
		hookID).Scan(&updatedN)
	if assignedN < 1 || updatedN < 1 {
		t.Fatalf("assigned=%d updated=%d", assignedN, updatedN)
	}
}
