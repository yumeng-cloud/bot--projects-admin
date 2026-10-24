package issueboard_test

import (
	"strings"
	"testing"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/issueboard"
)

// insertTeam seeds a second (or Nth) team in the same workspace via raw SQL —
// same style as insertMember; Issue-domain tests do not import botmem.
func insertTeam(t *testing.T, svc *issueboard.Service, wsID, name string) string {
	t.Helper()
	tid := idgen.New()
	_, err := svc.DB.Exec(
		`INSERT INTO teams (id, workspace_id, name, slug) VALUES (?,?,?,?)`,
		tid, wsID, name, "slug-"+tid[:8],
	)
	if err != nil {
		t.Fatal(err)
	}
	return tid
}

// TestTeamPool_ListHidesOtherTeam — Case: List 队 A 不见队 B.
func TestTeamPool_ListHidesOtherTeam(t *testing.T) {
	svc, ws, pid, teamA := setup(t)
	teamB := insertTeam(t, svc, ws, "B")

	a, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, TeamID: teamA, Title: "A-only", ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	b, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, TeamID: teamB, Title: "B-only", ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}

	listA, err := svc.ListIssues(ws, issueboard.ListIssuesFilter{ProjectID: pid, TeamID: teamA})
	if err != nil {
		t.Fatal(err)
	}
	listB, err := svc.ListIssues(ws, issueboard.ListIssuesFilter{ProjectID: pid, TeamID: teamB})
	if err != nil {
		t.Fatal(err)
	}

	seenA := map[string]bool{}
	for _, it := range listA {
		seenA[it.ID] = true
	}
	seenB := map[string]bool{}
	for _, it := range listB {
		seenB[it.ID] = true
	}
	if !seenA[a.ID] || seenA[b.ID] {
		t.Fatalf("team A list want A not B: %+v", seenA)
	}
	if !seenB[b.ID] || seenB[a.ID] {
		t.Fatalf("team B list want B not A: %+v", seenB)
	}
}

// TestTeamPool_DefaultListHidesNULL — Case: 默认 list（TeamID=A）不含 team_id IS NULL.
func TestTeamPool_DefaultListHidesNULL(t *testing.T) {
	svc, ws, pid, teamA := setup(t)

	owned, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, TeamID: teamA, Title: "owned", ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, TeamID: teamA, Title: "legacy-null", ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DB.Exec(`UPDATE issues SET team_id=NULL WHERE id=?`, legacy.ID); err != nil {
		t.Fatal(err)
	}

	list, err := svc.ListIssues(ws, issueboard.ListIssuesFilter{ProjectID: pid, TeamID: teamA})
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, it := range list {
		seen[it.ID] = true
	}
	if !seen[owned.ID] {
		t.Fatal("want owned issue in default team list")
	}
	if seen[legacy.ID] {
		t.Fatal("default team list must hide team_id IS NULL (use UnassignedOnly / ?unassigned=1)")
	}
}

// TestTeamPool_UnassignedFilterOnlyNULL — Case: UnassignedOnly 仅返回 NULL（handler: ?unassigned=1）.
func TestTeamPool_UnassignedFilterOnlyNULL(t *testing.T) {
	svc, ws, pid, teamA := setup(t)

	owned, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, TeamID: teamA, Title: "owned", ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, TeamID: teamA, Title: "legacy", ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DB.Exec(`UPDATE issues SET team_id=NULL WHERE id=?`, legacy.ID); err != nil {
		t.Fatal(err)
	}

	list, err := svc.ListIssues(ws, issueboard.ListIssuesFilter{
		ProjectID: pid, UnassignedOnly: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ID != legacy.ID {
		t.Fatalf("unassigned filter want only NULL legacy, got %+v", list)
	}
	for _, it := range list {
		if it.ID == owned.ID {
			t.Fatal("unassigned must not return assigned-team issues")
		}
		if it.TeamID != nil && *it.TeamID != "" {
			t.Fatalf("unassigned row must have nil team_id, got %+v", it.TeamID)
		}
	}
}

// TestTeamPool_CreateRequiresTeamID — Case: Create 无/错 team_id → 4xx 人话.
// Actual Go: empty → 400 "创建 Issue 必须指定 team_id（当前团队）";
// foreign/unknown → 400 "team_id 无效或不属于当前工作区".
func TestTeamPool_CreateRequiresTeamID(t *testing.T) {
	svc, ws, pid, _ := setup(t)
	_, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, Title: "no-team", ActorKind: "human",
	})
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.StatusCode < 400 || ae.StatusCode >= 500 {
		t.Fatalf("want 4xx, got %v", err)
	}
	if ae.StatusCode != 400 {
		t.Fatalf("empty team_id status want 400, got %d", ae.StatusCode)
	}
	if !strings.Contains(ae.Message, "team_id") && !strings.Contains(ae.Message, "团队") {
		t.Fatalf("want human message mentioning team_id/团队, got %q", ae.Message)
	}

	_, err = svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, TeamID: idgen.New(), Title: "bad-team", ActorKind: "human",
	})
	ae, ok = err.(*apperr.AppError)
	if !ok || ae.StatusCode < 400 || ae.StatusCode >= 500 {
		t.Fatalf("want 4xx for foreign team_id, got %v", err)
	}
	if ae.StatusCode != 400 {
		t.Fatalf("foreign team_id status want 400, got %d", ae.StatusCode)
	}
	if !strings.Contains(ae.Message, "team_id") {
		t.Fatalf("want human message mentioning team_id, got %q", ae.Message)
	}
}

// TestTeamPool_GetPatchWrongTeam404 — Case: GET/PATCH 错队 → 404（非 403）.
// Uses GetIssueInTeam + UpdateIssue.RequireTeamID (handler wires ?team_id=).
func TestTeamPool_GetPatchWrongTeam404(t *testing.T) {
	svc, ws, pid, teamA := setup(t)
	teamB := insertTeam(t, svc, ws, "B")

	iss, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, TeamID: teamA, Title: "A-iss", ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.GetIssueInTeam(ws, iss.ID, teamB)
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.StatusCode != 404 {
		t.Fatalf("wrong-team GET want 404 not 403, got %v", err)
	}
	if ae.StatusCode == 403 {
		t.Fatal("must not use 403 for wrong-team (existence leak)")
	}

	title := "hijack"
	_, err = svc.UpdateIssue(ws, iss.ID, issueboard.UpdateIssueInput{
		Version: iss.Version, Title: &title, ActorKind: "human",
		RequireTeamID: teamB,
	})
	ae, ok = err.(*apperr.AppError)
	if !ok || ae.StatusCode != 404 {
		t.Fatalf("wrong-team PATCH want 404, got %v", err)
	}

	_, err = svc.GetIssueInTeam(ws, iss.ID, teamA)
	if err != nil {
		t.Fatalf("same-team GET: %v", err)
	}
}

// TestTeamPool_SearchQScopedToTeam — Search handler wires ListIssuesFilter{Q, TeamID};
// Q+TeamID must not leak other-team title hits (side-path parity with List).
func TestTeamPool_SearchQScopedToTeam(t *testing.T) {
	svc, ws, pid, teamA := setup(t)
	teamB := insertTeam(t, svc, ws, "B")

	_, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, TeamID: teamA, Title: "shared-token-alpha", ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, TeamID: teamB, Title: "shared-token-beta", ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Unscoped Q (empty TeamID) still sees both — opt-in isolation.
	all, err := svc.ListIssues(ws, issueboard.ListIssuesFilter{Q: "shared-token"})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("unscoped search want 2, got %d", len(all))
	}

	scoped, err := svc.ListIssues(ws, issueboard.ListIssuesFilter{Q: "shared-token", TeamID: teamA})
	if err != nil {
		t.Fatal(err)
	}
	if len(scoped) != 1 {
		t.Fatalf("team-A search want 1, got %d", len(scoped))
	}
	if scoped[0].Title != "shared-token-alpha" {
		t.Fatalf("want alpha, got %q", scoped[0].Title)
	}
}
