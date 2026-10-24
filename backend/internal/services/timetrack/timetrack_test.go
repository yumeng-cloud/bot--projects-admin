package timetrack_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/timetrack"
)

func setup(t *testing.T) (svc *timetrack.Service, wsID, memID, issueID string) {
	t.Helper()
	dsn := filepath.Join(t.TempDir(), "time.db")
	sqlDB, err := db.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	wsID = idgen.New()
	userID := idgen.New()
	memID = idgen.New()
	projID := idgen.New()
	issueID = idgen.New()

	must := func(q string, args ...any) {
		t.Helper()
		if _, err := sqlDB.Exec(q, args...); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	must(`INSERT INTO users (id, email, password_hash, name) VALUES (?,?,?,?)`, userID, "t@t.local", "x", "T")
	must(`INSERT INTO workspaces (id, name, slug) VALUES (?,?,?)`, wsID, "WS", "ws-time")
	must(`INSERT INTO members (id, workspace_id, user_id, role, display_name) VALUES (?,?,?,?,?)`,
		memID, wsID, userID, "member", "Dev")
	must(`INSERT INTO projects (id, workspace_id, name, key_prefix) VALUES (?,?,?,?)`,
		projID, wsID, "Core", "CORE")
	must(`INSERT INTO issues (id, project_id, workspace_id, number, identifier, title) VALUES (?,?,?,?,?,?)`,
		issueID, projID, wsID, 1, "CORE-1", "做工时")

	return timetrack.New(sqlDB), wsID, memID, issueID
}

func TestCreateEntryAndSummary(t *testing.T) {
	svc, wsID, memID, issueID := setup(t)

	_, err := svc.UpsertRate(wsID, timetrack.UpsertRateInput{
		MemberID: &memID, HourlyRate: 200, Currency: "CNY",
	})
	if err != nil {
		t.Fatalf("rate: %v", err)
	}

	billable := true
	rate := 250.0
	e, err := svc.CreateEntry(wsID, timetrack.CreateEntryInput{
		IssueID: issueID, MemberID: memID, Hours: 2.5,
		Note: "实现 wiki", Billable: &billable, Rate: &rate, WorkedAt: "2026-10-20",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if e.Hours != 2.5 || e.Rate == nil || *e.Rate != 250 {
		t.Fatalf("entry: %+v", e)
	}

	_, err = svc.CreateEntry(wsID, timetrack.CreateEntryInput{
		IssueID: issueID, MemberID: memID, Hours: 1,
		Note: "用默认费率", Billable: &billable, WorkedAt: "2026-10-21",
	})
	if err != nil {
		t.Fatal(err)
	}

	sum, err := svc.Summary(wsID, "2026-10-01", "2026-10-31", "member")
	if err != nil {
		t.Fatal(err)
	}
	if sum.TotalHours != 3.5 {
		t.Fatalf("total hours: %v", sum.TotalHours)
	}
	// 2.5*250 + 1*200 = 625+200 = 825
	if sum.TotalAmount != 825 {
		t.Fatalf("total amount: %v want 825", sum.TotalAmount)
	}
	if len(sum.Rows) != 1 || sum.Rows[0].GroupLabel != "Dev" {
		t.Fatalf("rows: %+v", sum.Rows)
	}

	sumP, err := svc.Summary(wsID, "2026-10-01", "2026-10-31", "project")
	if err != nil || len(sumP.Rows) != 1 || sumP.Rows[0].GroupLabel != "Core" {
		t.Fatalf("project summary: %v %+v", err, sumP)
	}

	var buf bytes.Buffer
	if err := svc.ExportCSV(&buf, wsID, "2026-10-01", "2026-10-31"); err != nil {
		t.Fatal(err)
	}
	csv := buf.String()
	if !strings.Contains(csv, "CORE-1") || !strings.Contains(csv, "Dev") {
		t.Fatalf("csv missing fields: %s", csv)
	}
}

func TestRejectNonPositiveHours(t *testing.T) {
	svc, wsID, memID, issueID := setup(t)
	_, err := svc.CreateEntry(wsID, timetrack.CreateEntryInput{
		IssueID: issueID, MemberID: memID, Hours: 0, WorkedAt: "2026-10-20",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
