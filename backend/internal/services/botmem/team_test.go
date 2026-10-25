package botmem_test

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/botmem"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/issueboard"
)

func TestCreateTeamIdempotentAndEnsureRoster(t *testing.T) {
	svc, ws, m1, _ := setup(t)
	teamID := idgen.New()
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{ID: teamID, Name: "班底A"})
	if err != nil {
		t.Fatal(err)
	}
	again, err := svc.CreateTeam(ws, botmem.CreateTeamInput{ID: teamID, Name: "班底A-改名忽略"})
	if err != nil {
		t.Fatal(err)
	}
	if again.ID != team.ID || again.Name != team.Name {
		t.Fatalf("idempotent: %+v vs %+v", again, team)
	}

	r1, err := svc.EnsureRoster(ws, teamID, botmem.EnsureRosterInput{})
	if err != nil {
		t.Fatal(err)
	}
	if len(r1) != 5 {
		t.Fatalf("default roster len=%d want 5", len(r1))
	}
	// bind one seat then ensure again — bound must survive
	seat := r1[0]
	bound, err := svc.BindSeat(ws, teamID, seat.ID, m1, false)
	if err != nil {
		t.Fatal(err)
	}
	if bound.BoundMemberID == nil || *bound.BoundMemberID != m1 {
		t.Fatalf("bind: %+v", bound)
	}
	if bound.Status != "active" {
		t.Fatalf("after bind status=%s want active", bound.Status)
	}
	r2, err := svc.EnsureRoster(ws, teamID, botmem.EnsureRosterInput{})
	if err != nil {
		t.Fatal(err)
	}
	if len(r2) != len(r1) {
		t.Fatalf("roster grew: %d -> %d", len(r1), len(r2))
	}
	found := false
	for _, s := range r2 {
		if s.ID == seat.ID {
			found = true
			if s.BoundMemberID == nil || *s.BoundMemberID != m1 {
				t.Fatalf("bound cleared on ensure: %+v", s)
			}
		}
	}
	if !found {
		t.Fatal("seat disappeared")
	}
}

func TestBindSeatSingleActiveForce(t *testing.T) {
	svc, ws, m1, m2 := setup(t)
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "T"})
	if err != nil {
		t.Fatal(err)
	}
	roster, err := svc.EnsureRoster(ws, team.ID, botmem.EnsureRosterInput{})
	if err != nil {
		t.Fatal(err)
	}
	seatID := roster[0].ID
	if _, err := svc.BindSeat(ws, team.ID, seatID, m1, false); err != nil {
		t.Fatal(err)
	}
	_, err = svc.BindSeat(ws, team.ID, seatID, m2, false)
	if err == nil {
		t.Fatal("expected 409")
	}
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.Code != "conflict" {
		t.Fatalf("want conflict got %#v", err)
	}
	// 活跃 claim + force 仍 409，禁止顶替
	_, err = svc.BindSeat(ws, team.ID, seatID, m2, true)
	if err == nil {
		t.Fatal("expected 409 even with force on active seat")
	}
	ae, ok = err.(*apperr.AppError)
	if !ok || ae.Code != "conflict" {
		t.Fatalf("want conflict got %#v", err)
	}
	if ae.Message != "该席位占用中，禁止顶替" {
		t.Fatalf("msg=%q", ae.Message)
	}
}

func TestClaimForce(t *testing.T) {
	svc, ws, m1, m2 := setup(t)
	role, err := svc.CreateRole(ws, botmem.CreateRoleInput{Name: "前端", Slug: "fe2"})
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
		t.Fatalf("want conflict %#v", err)
	}
	// 活跃 claim + force 仍 409
	_, err = svc.Claim(ws, role.ID, m2, true)
	if err == nil {
		t.Fatal("expected 409 even with force on active claim")
	}
	ae, ok = err.(*apperr.AppError)
	if !ok || ae.Code != "conflict" {
		t.Fatalf("want conflict %#v", err)
	}
	if ae.Message != "该席位占用中，禁止顶替" {
		t.Fatalf("msg=%q", ae.Message)
	}
}

func TestBindSeatForceDirtyBindWritesAudit(t *testing.T) {
	svc, ws, m1, m2 := setup(t)
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "Dirty"})
	if err != nil {
		t.Fatal(err)
	}
	roster, err := svc.EnsureRoster(ws, team.ID, botmem.EnsureRosterInput{})
	if err != nil {
		t.Fatal(err)
	}
	seat := roster[0]
	if _, err := svc.BindSeat(ws, team.ID, seat.ID, m1, false); err != nil {
		t.Fatal(err)
	}
	// 释放 claim，保留 bound → 脏绑
	if seat.BotRoleID == nil || *seat.BotRoleID == "" {
		t.Fatal("expected bot_role_id on roster seat")
	}
	if err := svc.Release(ws, *seat.BotRoleID); err != nil {
		t.Fatal(err)
	}
	forced, err := svc.BindSeat(ws, team.ID, seat.ID, m2, true)
	if err != nil {
		t.Fatal(err)
	}
	if forced.BoundMemberID == nil || *forced.BoundMemberID != m2 {
		t.Fatalf("force dirty bind: %+v", forced)
	}
	if forced.Status != "active" {
		t.Fatalf("force dirty status=%s want active", forced.Status)
	}
	var action, payload string
	err = svc.DB.QueryRow(
		`SELECT action, payload FROM activities WHERE workspace_id=? AND action=? ORDER BY created_at DESC LIMIT 1`,
		ws, "seat.force_bind",
	).Scan(&action, &payload)
	if err != nil {
		t.Fatalf("audit activity: %v", err)
	}
	if action != "seat.force_bind" {
		t.Fatalf("action=%s", action)
	}
	if !strings.Contains(payload, "强制补位") || !strings.Contains(payload, "force") {
		t.Fatalf("payload missing 人话 force: %s", payload)
	}
	if !strings.Contains(payload, m1) || !strings.Contains(payload, m2) {
		t.Fatalf("payload missing members: %s", payload)
	}
	if !strings.Contains(payload, `"force":true`) && !strings.Contains(payload, `"force": true`) {
		t.Fatalf("payload missing force flag: %s", payload)
	}
}

func TestExpandSeatsAssignsUnassignedLeaves(t *testing.T) {
	svc, ws, m1, m2 := setup(t)
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "Expand"})
	if err != nil {
		t.Fatal(err)
	}
	roster, err := svc.EnsureRoster(ws, team.ID, botmem.EnsureRosterInput{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.BindSeat(ws, team.ID, roster[0].ID, m1, false); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.BindSeat(ws, team.ID, roster[1].ID, m2, false); err != nil {
		t.Fatal(err)
	}

	ib := issueboard.New(svc.DB)
	proj, err := ib.CreateProject(ws, issueboard.CreateProjectInput{Name: "P", KeyPrefix: "EX"})
	if err != nil {
		t.Fatal(err)
	}
	i1, err := ib.CreateIssue(ws, issueboard.CreateIssueInput{ProjectID: proj.ID, TeamID: team.ID, Title: "a", Status: "todo", ActorKind: "human"})
	if err != nil {
		t.Fatal(err)
	}
	i2, err := ib.CreateIssue(ws, issueboard.CreateIssueInput{ProjectID: proj.ID, TeamID: team.ID, Title: "b", Status: "backlog", ActorKind: "human"})
	if err != nil {
		t.Fatal(err)
	}
	// already assigned — must skip
	i3, err := ib.CreateIssue(ws, issueboard.CreateIssueInput{ProjectID: proj.ID, TeamID: team.ID, Title: "c", Status: "todo", AssigneeID: &m1, ActorKind: "human"})
	if err != nil {
		t.Fatal(err)
	}
	_ = i3

	res, err := svc.ExpandSeats(ws, team.ID, botmem.ExpandInput{ProjectID: proj.ID, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Assigned) != 2 {
		t.Fatalf("assigned=%d skipped=%+v want 2", len(res.Assigned), res.Skipped)
	}
	got1, _ := ib.GetIssue(ws, i1.ID)
	got2, _ := ib.GetIssue(ws, i2.ID)
	if got1.AssigneeID == nil || got2.AssigneeID == nil {
		t.Fatalf("assignees missing: %+v %+v", got1.AssigneeID, got2.AssigneeID)
	}
}

func TestCreateSeatIdempotentAndStop(t *testing.T) {
	svc, ws, m1, m2 := setup(t)
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "G2"})
	if err != nil {
		t.Fatal(err)
	}
	s1, err := svc.CreateSeat(ws, team.ID, botmem.EnsureSeatInput{Slug: "ops", Name: "运维", PlaybookRef: "bot-team-template@main:playbook/ops.md"})
	if err != nil {
		t.Fatal(err)
	}
	if s1.Status != "open" {
		t.Fatalf("status=%s want open", s1.Status)
	}
	s2, err := svc.CreateSeat(ws, team.ID, botmem.EnsureSeatInput{Slug: "ops", Name: "运维改名忽略", PlaybookRef: "bot-team-template@main:playbook/ops.md"})
	if err != nil {
		t.Fatal(err)
	}
	if s2.ID != s1.ID {
		t.Fatalf("idempotent seat id %s vs %s", s2.ID, s1.ID)
	}
	bound, err := svc.BindSeat(ws, team.ID, s1.ID, m1, false)
	if err != nil {
		t.Fatal(err)
	}
	if bound.Status != "active" {
		t.Fatalf("after bind status=%s", bound.Status)
	}
	// active force by other → 409
	_, err = svc.BindSeat(ws, team.ID, s1.ID, m2, true)
	if err == nil {
		t.Fatal("expected 409")
	}
	stopped, err := svc.StopSeat(ws, team.ID, s1.ID, m1)
	if err != nil {
		t.Fatal(err)
	}
	if stopped.Status != "stopped" {
		t.Fatalf("status=%s", stopped.Status)
	}
	if stopped.BoundMemberID != nil {
		t.Fatalf("bound should clear: %+v", stopped.BoundMemberID)
	}
	// stopped 可补位（无需 force）
	re, err := svc.BindSeat(ws, team.ID, s1.ID, m2, false)
	if err != nil {
		t.Fatal(err)
	}
	if re.BoundMemberID == nil || *re.BoundMemberID != m2 || re.Status != "active" {
		t.Fatalf("rebind: %+v", re)
	}
}

func TestTeamInviteRedeem(t *testing.T) {
	svc, ws, m1, _ := setup(t)
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "Inv"})
	if err != nil {
		t.Fatal(err)
	}
	max := 2
	inv, err := svc.CreateInvite(ws, team.ID, botmem.CreateInviteInput{
		AllowExpandSeat: true,
		RoleHint:        "frontend",
		MaxUses:         &max,
		CreatedBy:       m1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if inv.Code == "" || !strings.HasPrefix(inv.Code, "ymt_") {
		t.Fatalf("code=%q", inv.Code)
	}
	code := inv.Code
	res, err := svc.RedeemInvite(code)
	if err != nil {
		t.Fatal(err)
	}
	if res.WorkspaceID != ws || res.TeamID != team.ID || !res.AllowExpandSeat {
		t.Fatalf("redeem: %+v", res)
	}
	if res.RoleHint == nil || *res.RoleHint != "frontend" {
		t.Fatalf("role_hint=%v", res.RoleHint)
	}
	// second use ok
	if _, err := svc.RedeemInvite(code); err != nil {
		t.Fatal(err)
	}
	// third hits max
	_, err = svc.RedeemInvite(code)
	if err == nil {
		t.Fatal("expected max uses")
	}
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.Code != "forbidden" {
		t.Fatalf("want forbidden %#v", err)
	}
	if _, err := svc.RevokeInvite(ws, team.ID, inv.ID); err != nil {
		// already at max; create fresh to revoke
		inv2, err2 := svc.CreateInvite(ws, team.ID, botmem.CreateInviteInput{CreatedBy: m1})
		if err2 != nil {
			t.Fatal(err2)
		}
		if _, err := svc.RevokeInvite(ws, team.ID, inv2.ID); err != nil {
			t.Fatal(err)
		}
		_, err = svc.RedeemInvite(inv2.Code)
		if err == nil {
			t.Fatal("expected revoked")
		}
	}
}

func TestListInvitesNoPlainCode(t *testing.T) {
	svc, ws, m1, _ := setup(t)
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "ListInv"})
	if err != nil {
		t.Fatal(err)
	}
	max := 1
	inv, err := svc.CreateInvite(ws, team.ID, botmem.CreateInviteInput{
		AllowExpandSeat: true,
		RoleHint:        "frontend",
		MaxUses:         &max,
		CreatedBy:       m1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if inv.Code == "" || inv.Status != "active" {
		t.Fatalf("create: code=%q status=%q", inv.Code, inv.Status)
	}
	plain := inv.Code

	list, err := svc.ListInvites(ws, team.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("len=%d", len(list))
	}
	got := list[0]
	if got.Code != "" {
		t.Fatalf("list must not return plain code: %q", got.Code)
	}
	if got.ID != inv.ID || got.Status != "active" || got.UseCount != 0 {
		t.Fatalf("list item: %+v", got)
	}
	if got.CodeTail == "" || len(got.CodeTail) > 4 {
		t.Fatalf("code_tail=%q", got.CodeTail)
	}
	if !got.AllowExpandSeat {
		t.Fatal("allow_expand_seat lost")
	}
	// redeem once → exhausted
	if _, err := svc.RedeemInvite(plain); err != nil {
		t.Fatal(err)
	}
	list, err = svc.ListInvites(ws, team.ID)
	if err != nil {
		t.Fatal(err)
	}
	if list[0].Status != "exhausted" || list[0].UseCount != 1 {
		t.Fatalf("after redeem: %+v", list[0])
	}
	if _, err := svc.RevokeInvite(ws, team.ID, inv.ID); err != nil {
		t.Fatal(err)
	}
	list, err = svc.ListInvites(ws, team.ID)
	if err != nil {
		t.Fatal(err)
	}
	if list[0].Status != "revoked" {
		t.Fatalf("want revoked got %q", list[0].Status)
	}
}

func TestBindSeatLeftForceActive(t *testing.T) {
	svc, ws, m1, m2 := setup(t)
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "LeftForce"})
	if err != nil {
		t.Fatal(err)
	}
	s1, err := svc.CreateSeat(ws, team.ID, botmem.EnsureSeatInput{Slug: "fe", Name: "前端", PlaybookRef: "bot-team-template@main:playbook/fe.md"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.BindSeat(ws, team.ID, s1.ID, m1, false); err != nil {
		t.Fatal(err)
	}
	// 活跃 force → 409
	_, err = svc.BindSeat(ws, team.ID, s1.ID, m2, true)
	if err == nil {
		t.Fatal("expected 409 on active force")
	}
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.Code != "conflict" {
		t.Fatalf("want conflict %#v", err)
	}
	left, _, err := svc.UnbindSeat(ws, team.ID, s1.ID, m1, false)
	if err != nil {
		t.Fatal(err)
	}
	if left.Status != "left" {
		t.Fatalf("status=%s want left", left.Status)
	}
	// left 可 force（或无需 force）成功且 active
	re, err := svc.BindSeat(ws, team.ID, s1.ID, m2, true)
	if err != nil {
		t.Fatal(err)
	}
	if re.BoundMemberID == nil || *re.BoundMemberID != m2 || re.Status != "active" {
		t.Fatalf("left force rebind: %+v", re)
	}
}

func countForceBindActivities(t *testing.T, svc *botmem.Service, ws string) int {
	t.Helper()
	var n int
	if err := svc.DB.QueryRow(`SELECT COUNT(*) FROM activities WHERE workspace_id=? AND action=?`, ws, "seat.force_bind").Scan(&n); err != nil {
		t.Fatalf("count force_bind: %v", err)
	}
	return n
}

func countAuditComments(t *testing.T, svc *botmem.Service, issueID string) int {
	t.Helper()
	var n int
	if err := svc.DB.QueryRow(`SELECT COUNT(*) FROM comments WHERE issue_id=? AND body LIKE ?`, issueID, "%强制补位%").Scan(&n); err != nil {
		t.Fatalf("count comments: %v", err)
	}
	return n
}

// TestBindSeatStoppedForceWritesVisibleAudit N3：停工席 force 成功 → 人话 activities + Issue comment。
func TestBindSeatStoppedForceWritesVisibleAudit(t *testing.T) {
	svc, ws, m1, m2 := setup(t)
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "StopForceAudit"})
	if err != nil {
		t.Fatal(err)
	}
	s1, err := svc.CreateSeat(ws, team.ID, botmem.EnsureSeatInput{Slug: "ops", Name: "运维", PlaybookRef: "bot-team-template@main:playbook/ops.md"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.BindSeat(ws, team.ID, s1.ID, m1, false); err != nil {
		t.Fatal(err)
	}
	ib := issueboard.New(svc.DB)
	proj, err := ib.CreateProject(ws, issueboard.CreateProjectInput{Name: "P", KeyPrefix: "SF"})
	if err != nil {
		t.Fatal(err)
	}
	iss, err := ib.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: proj.ID, TeamID: team.ID, Title: "停工补位审计", Status: "todo", ActorKind: "human", AssigneeID: &m1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.StopSeat(ws, team.ID, s1.ID, m1); err != nil {
		t.Fatal(err)
	}
	beforeAct := countForceBindActivities(t, svc, ws)
	beforeCmt := countAuditComments(t, svc, iss.ID)

	forced, err := svc.BindSeat(ws, team.ID, s1.ID, m2, true)
	if err != nil {
		t.Fatal(err)
	}
	if forced.BoundMemberID == nil || *forced.BoundMemberID != m2 || forced.Status != "active" {
		t.Fatalf("stopped force: %+v", forced)
	}
	if countForceBindActivities(t, svc, ws) != beforeAct+1 {
		t.Fatalf("want +1 seat.force_bind activity")
	}
	var payload string
	if err := svc.DB.QueryRow(
		`SELECT payload FROM activities WHERE workspace_id=? AND action=? ORDER BY created_at DESC LIMIT 1`,
		ws, "seat.force_bind",
	).Scan(&payload); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"强制补位", "运维", "停工", "force", "BotB"} {
		if !strings.Contains(payload, want) {
			t.Fatalf("payload missing %q: %s", want, payload)
		}
	}
	if countAuditComments(t, svc, iss.ID) != beforeCmt+1 {
		t.Fatalf("want Issue comment 人话审计 on %s", iss.ID)
	}
	var body string
	if err := svc.DB.QueryRow(
		`SELECT body FROM comments WHERE issue_id=? AND body LIKE ? ORDER BY created_at DESC LIMIT 1`,
		iss.ID, "%强制补位%",
	).Scan(&body); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, "force") || !strings.Contains(body, "停工") {
		t.Fatalf("comment body=%q", body)
	}
}

// TestBindSeatLeftForceWritesVisibleAudit N3：离职席 force 成功 → 人话审计。
func TestBindSeatLeftForceWritesVisibleAudit(t *testing.T) {
	svc, ws, m1, m2 := setup(t)
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "LeftForceAudit"})
	if err != nil {
		t.Fatal(err)
	}
	s1, err := svc.CreateSeat(ws, team.ID, botmem.EnsureSeatInput{Slug: "qa", Name: "QA", PlaybookRef: "bot-team-template@main:playbook/qa.md"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.BindSeat(ws, team.ID, s1.ID, m1, false); err != nil {
		t.Fatal(err)
	}
	ib := issueboard.New(svc.DB)
	proj, err := ib.CreateProject(ws, issueboard.CreateProjectInput{Name: "P", KeyPrefix: "LF"})
	if err != nil {
		t.Fatal(err)
	}
	iss, err := ib.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: proj.ID, TeamID: team.ID, Title: "离职补位审计", Status: "in_progress", ActorKind: "human", AssigneeID: &m1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.UnbindSeat(ws, team.ID, s1.ID, m1, false); err != nil {
		t.Fatal(err)
	}
	beforeAct := countForceBindActivities(t, svc, ws)
	beforeCmt := countAuditComments(t, svc, iss.ID)

	if _, err := svc.BindSeat(ws, team.ID, s1.ID, m2, true); err != nil {
		t.Fatal(err)
	}
	if countForceBindActivities(t, svc, ws) != beforeAct+1 {
		t.Fatal("want +1 seat.force_bind")
	}
	var payload string
	_ = svc.DB.QueryRow(
		`SELECT payload FROM activities WHERE workspace_id=? AND action=? ORDER BY created_at DESC LIMIT 1`,
		ws, "seat.force_bind",
	).Scan(&payload)
	if !strings.Contains(payload, "强制补位") || !strings.Contains(payload, "离职") {
		t.Fatalf("payload=%s", payload)
	}
	if countAuditComments(t, svc, iss.ID) != beforeCmt+1 {
		t.Fatal("want left-force Issue comment")
	}
}

// TestBindSeatActiveForceNoVisibleAudit 活跃 force 仍 409，且不写人话审计。
func TestBindSeatActiveForceNoVisibleAudit(t *testing.T) {
	svc, ws, m1, m2 := setup(t)
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "ActiveNoAudit"})
	if err != nil {
		t.Fatal(err)
	}
	s1, err := svc.CreateSeat(ws, team.ID, botmem.EnsureSeatInput{Slug: "fe2", Name: "前端2", PlaybookRef: "bot-team-template@main:playbook/fe2.md"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.BindSeat(ws, team.ID, s1.ID, m1, false); err != nil {
		t.Fatal(err)
	}
	ib := issueboard.New(svc.DB)
	proj, err := ib.CreateProject(ws, issueboard.CreateProjectInput{Name: "P", KeyPrefix: "AN"})
	if err != nil {
		t.Fatal(err)
	}
	iss, err := ib.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: proj.ID, TeamID: team.ID, Title: "活跃不可 force", Status: "todo", ActorKind: "human", AssigneeID: &m1,
	})
	if err != nil {
		t.Fatal(err)
	}
	beforeAct := countForceBindActivities(t, svc, ws)
	beforeCmt := countAuditComments(t, svc, iss.ID)

	_, err = svc.BindSeat(ws, team.ID, s1.ID, m2, true)
	if err == nil {
		t.Fatal("expected 409")
	}
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.Code != "conflict" {
		t.Fatalf("want conflict %#v", err)
	}
	if countForceBindActivities(t, svc, ws) != beforeAct {
		t.Fatal("active force must not write seat.force_bind")
	}
	if countAuditComments(t, svc, iss.ID) != beforeCmt {
		t.Fatal("active force must not write audit comment")
	}
}

func TestClaimSyncsLinkedSeatActive(t *testing.T) {
	svc, ws, m1, _ := setup(t)
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "ClaimSync"})
	if err != nil {
		t.Fatal(err)
	}
	s1, err := svc.CreateSeat(ws, team.ID, botmem.EnsureSeatInput{Slug: "ops2", Name: "运维2", PlaybookRef: "bot-team-template@main:playbook/ops2.md"})
	if err != nil {
		t.Fatal(err)
	}
	if s1.Status != "open" || s1.BotRoleID == nil {
		t.Fatalf("seat: %+v", s1)
	}
	if _, err := svc.Claim(ws, *s1.BotRoleID, m1, false); err != nil {
		t.Fatal(err)
	}
	got, err := svc.GetRoster(ws, team.ID)
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, s := range got {
		if s.ID == s1.ID {
			found = true
			if s.BoundMemberID == nil || *s.BoundMemberID != m1 {
				t.Fatalf("claim should bind seat: %+v", s)
			}
			if s.Status != "active" {
				t.Fatalf("after claim status=%s want active", s.Status)
			}
		}
	}
	if !found {
		t.Fatal("seat missing from roster")
	}
}

// TestTwoTeamsRosterClaimIsolation T6-R：同 WS 两队同 slug 席位 id 不交、claim/bound 不串。
func TestTwoTeamsRosterClaimIsolation(t *testing.T) {
	svc, ws, m1, m2 := setup(t)
	teamA, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "TeamA", Slug: "team-a"})
	if err != nil {
		t.Fatal(err)
	}
	teamB, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "TeamB", Slug: "team-b"})
	if err != nil {
		t.Fatal(err)
	}
	if teamA.ID == teamB.ID {
		t.Fatal("expected distinct team ids")
	}

	ra, err := svc.EnsureRoster(ws, teamA.ID, botmem.EnsureRosterInput{})
	if err != nil {
		t.Fatal(err)
	}
	rb, err := svc.EnsureRoster(ws, teamB.ID, botmem.EnsureRosterInput{})
	if err != nil {
		t.Fatal(err)
	}
	if len(ra) == 0 || len(rb) == 0 {
		t.Fatalf("empty roster a=%d b=%d", len(ra), len(rb))
	}

	idsA := map[string]struct{}{}
	slugSeatA := map[string]string{}
	for _, s := range ra {
		idsA[s.ID] = struct{}{}
		slugSeatA[s.RoleSlug] = s.ID
	}
	for _, s := range rb {
		if _, ok := idsA[s.ID]; ok {
			t.Fatalf("seat id %s appears in both teams", s.ID)
		}
	}

	feA := slugSeatA["frontend"]
	if feA == "" {
		t.Fatal("team A missing frontend")
	}
	var feB string
	for _, s := range rb {
		if s.RoleSlug == "frontend" {
			feB = s.ID
			break
		}
	}
	if feB == "" {
		t.Fatal("team B missing frontend")
	}

	boundA, err := svc.BindSeat(ws, teamA.ID, feA, m1, false)
	if err != nil {
		t.Fatal(err)
	}
	if boundA.BoundMemberID == nil || *boundA.BoundMemberID != m1 {
		t.Fatalf("A bind: %+v", boundA)
	}
	if boundA.Claim == nil || boundA.Claim.MemberID != m1 {
		t.Fatalf("A claim missing: %+v", boundA.Claim)
	}

	// 队 B 同 slug 须可独立 bind（不因 A 的 claim 409）
	boundB, err := svc.BindSeat(ws, teamB.ID, feB, m2, false)
	if err != nil {
		t.Fatalf("B bind blocked by A claim (isolation hole): %v", err)
	}
	if boundB.BoundMemberID == nil || *boundB.BoundMemberID != m2 {
		t.Fatalf("B bind: %+v", boundB)
	}

	// 再拉 roster：A 的 claim/bound 不得出现在 B
	rb2, err := svc.GetRoster(ws, teamB.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range rb2 {
		if s.RoleSlug != "frontend" {
			continue
		}
		if s.BoundMemberID != nil && *s.BoundMemberID == m1 {
			t.Fatalf("A member leaked onto B seat: %+v", s)
		}
		if s.Claim != nil && s.Claim.MemberID == m1 {
			t.Fatalf("A claim leaked onto B roster: %+v", s.Claim)
		}
		if s.BoundMemberID == nil || *s.BoundMemberID != m2 {
			t.Fatalf("B frontend bound want m2: %+v", s)
		}
	}

	ra2, err := svc.GetRoster(ws, teamA.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range ra2 {
		if s.RoleSlug != "frontend" {
			continue
		}
		if s.BoundMemberID == nil || *s.BoundMemberID != m1 {
			t.Fatalf("A frontend disturbed: %+v", s)
		}
		if s.Claim != nil && s.Claim.MemberID == m2 {
			t.Fatalf("B claim leaked onto A: %+v", s.Claim)
		}
	}

	// bot_role_id 须按队隔离
	if boundA.BotRoleID == nil || boundB.BotRoleID == nil || *boundA.BotRoleID == *boundB.BotRoleID {
		t.Fatalf("expected distinct bot_role_id per team seat; A=%v B=%v", boundA.BotRoleID, boundB.BotRoleID)
	}

	// ensure 幂等且不串队改绑
	ra3, err := svc.EnsureRoster(ws, teamA.ID, botmem.EnsureRosterInput{})
	if err != nil {
		t.Fatal(err)
	}
	if len(ra3) != len(ra) {
		t.Fatalf("A roster grew on re-ensure: %d -> %d", len(ra), len(ra3))
	}
}

func countActionActivities(t *testing.T, svc *botmem.Service, ws, action string) int {
	t.Helper()
	var n int
	if err := svc.DB.QueryRow(`SELECT COUNT(*) FROM activities WHERE workspace_id=? AND action=?`, ws, action).Scan(&n); err != nil {
		t.Fatalf("count %s: %v", action, err)
	}
	return n
}

func countBodyComments(t *testing.T, svc *botmem.Service, issueID, like string) int {
	t.Helper()
	var n int
	if err := svc.DB.QueryRow(`SELECT COUNT(*) FROM comments WHERE issue_id=? AND body LIKE ?`, issueID, like).Scan(&n); err != nil {
		t.Fatalf("count comments: %v", err)
	}
	return n
}

// TestStopSeatActiveAdminPath L1：Admin 对 active 席 stop → stopped；幂等不重复审计。
func TestStopSeatActiveAdminPath(t *testing.T) {
	svc, ws, m1, m2 := setup(t)
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "StopL1"})
	if err != nil {
		t.Fatal(err)
	}
	s1, err := svc.CreateSeat(ws, team.ID, botmem.EnsureSeatInput{Slug: "ops", Name: "运维", PlaybookRef: "bot-team-template@main:playbook/ops.md"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.BindSeat(ws, team.ID, s1.ID, m1, false); err != nil {
		t.Fatal(err)
	}
	before := countActionActivities(t, svc, ws, "seat.stop")
	stopped, err := svc.StopSeat(ws, team.ID, s1.ID, m2) // m2 = Admin 操作者
	if err != nil {
		t.Fatal(err)
	}
	if stopped.Status != "stopped" {
		t.Fatalf("status=%s want stopped", stopped.Status)
	}
	if stopped.BoundMemberID != nil {
		t.Fatalf("bound should clear: %+v", stopped.BoundMemberID)
	}
	if countActionActivities(t, svc, ws, "seat.stop") != before+1 {
		t.Fatal("want +1 seat.stop activity")
	}
	// 幂等：再 stop 不报错、不追加审计
	again, err := svc.StopSeat(ws, team.ID, s1.ID, m2)
	if err != nil {
		t.Fatal(err)
	}
	if again.Status != "stopped" {
		t.Fatalf("idempotent status=%s", again.Status)
	}
	if countActionActivities(t, svc, ws, "seat.stop") != before+1 {
		t.Fatal("idempotent stop must not duplicate audit")
	}
	// 活跃顶替硬闸仍在
	_, err = svc.BindSeat(ws, team.ID, s1.ID, m1, false)
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.BindSeat(ws, team.ID, s1.ID, m2, true)
	if err == nil {
		t.Fatal("expected 409 on active force")
	}
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.Code != "conflict" || ae.Message != "该席位占用中，禁止顶替" {
		t.Fatalf("want 中文 conflict %#v", err)
	}
}

// TestUnbindSeatActiveAdminPath L1：Admin asAdmin 对 active 席 unbind → left；错态中文；幂等。
func TestUnbindSeatActiveAdminPath(t *testing.T) {
	svc, ws, m1, m2 := setup(t)
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "UnbindL1"})
	if err != nil {
		t.Fatal(err)
	}
	s1, err := svc.CreateSeat(ws, team.ID, botmem.EnsureSeatInput{Slug: "fe", Name: "前端", PlaybookRef: "bot-team-template@main:playbook/fe.md"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.BindSeat(ws, team.ID, s1.ID, m1, false); err != nil {
		t.Fatal(err)
	}
	// 非 Admin 解他人存活占用 → U6 中文 Forbidden（含占用者显示名）
	_, _, err = svc.UnbindSeat(ws, team.ID, s1.ID, m2, false)
	if err == nil {
		t.Fatal("expected forbidden")
	}
	ae, ok := err.(*apperr.AppError)
	wantMsg := "席位由「BotA」占用；请对方离席，或让管理员停工/离职解绑"
	if !ok || ae.Code != "forbidden" || ae.Message != wantMsg {
		t.Fatalf("want U6 forbidden %q, got %#v", wantMsg, err)
	}
	before := countActionActivities(t, svc, ws, "seat.unbind")
	left, _, err := svc.UnbindSeat(ws, team.ID, s1.ID, m2, true) // Admin 代解
	if err != nil {
		t.Fatal(err)
	}
	if left.Status != "left" {
		t.Fatalf("status=%s want left", left.Status)
	}
	if left.BoundMemberID != nil {
		t.Fatalf("bound should clear: %+v", left.BoundMemberID)
	}
	if countActionActivities(t, svc, ws, "seat.unbind") != before+1 {
		t.Fatal("want +1 seat.unbind activity")
	}
	// 幂等：已解绑再解 → OK、不追加审计
	again, _, err := svc.UnbindSeat(ws, team.ID, s1.ID, m2, true)
	if err != nil {
		t.Fatal(err)
	}
	if again.Status != "left" {
		t.Fatalf("idempotent status=%s", again.Status)
	}
	if countActionActivities(t, svc, ws, "seat.unbind") != before+1 {
		t.Fatal("idempotent unbind must not duplicate audit")
	}
}

// TestStopSeatWritesVisibleAudit L2：stop 成功 → seat.stop 人话 activity + Issue comment。
func TestStopSeatWritesVisibleAudit(t *testing.T) {
	svc, ws, m1, m2 := setup(t)
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "StopAudit"})
	if err != nil {
		t.Fatal(err)
	}
	s1, err := svc.CreateSeat(ws, team.ID, botmem.EnsureSeatInput{Slug: "qa", Name: "QA", PlaybookRef: "bot-team-template@main:playbook/qa.md"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.BindSeat(ws, team.ID, s1.ID, m1, false); err != nil {
		t.Fatal(err)
	}
	ib := issueboard.New(svc.DB)
	proj, err := ib.CreateProject(ws, issueboard.CreateProjectInput{Name: "P", KeyPrefix: "SA"})
	if err != nil {
		t.Fatal(err)
	}
	iss, err := ib.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: proj.ID, TeamID: team.ID, Title: "停工审计", Status: "todo", ActorKind: "human", AssigneeID: &m1,
	})
	if err != nil {
		t.Fatal(err)
	}
	beforeAct := countActionActivities(t, svc, ws, "seat.stop")
	beforeCmt := countBodyComments(t, svc, iss.ID, "%标记停工%")

	if _, err := svc.StopSeat(ws, team.ID, s1.ID, m2); err != nil {
		t.Fatal(err)
	}
	if countActionActivities(t, svc, ws, "seat.stop") != beforeAct+1 {
		t.Fatal("want +1 seat.stop")
	}
	var payload string
	if err := svc.DB.QueryRow(
		`SELECT payload FROM activities WHERE workspace_id=? AND action=? ORDER BY created_at DESC LIMIT 1`,
		ws, "seat.stop",
	).Scan(&payload); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"标记停工", "QA", "停工", "BotA"} {
		if !strings.Contains(payload, want) {
			t.Fatalf("payload missing %q: %s", want, payload)
		}
	}
	if countBodyComments(t, svc, iss.ID, "%标记停工%") != beforeCmt+1 {
		t.Fatal("want Issue comment 人话审计")
	}
}

// TestUnbindSeatWritesVisibleAudit L2：unbind 成功 → seat.unbind 人话 activity + Issue comment。
func TestUnbindSeatWritesVisibleAudit(t *testing.T) {
	svc, ws, m1, m2 := setup(t)
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "UnbindAudit"})
	if err != nil {
		t.Fatal(err)
	}
	s1, err := svc.CreateSeat(ws, team.ID, botmem.EnsureSeatInput{Slug: "ops", Name: "运维", PlaybookRef: "bot-team-template@main:playbook/ops.md"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.BindSeat(ws, team.ID, s1.ID, m1, false); err != nil {
		t.Fatal(err)
	}
	ib := issueboard.New(svc.DB)
	proj, err := ib.CreateProject(ws, issueboard.CreateProjectInput{Name: "P", KeyPrefix: "UA"})
	if err != nil {
		t.Fatal(err)
	}
	iss, err := ib.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: proj.ID, TeamID: team.ID, Title: "离职审计", Status: "in_progress", ActorKind: "human", AssigneeID: &m1,
	})
	if err != nil {
		t.Fatal(err)
	}
	beforeAct := countActionActivities(t, svc, ws, "seat.unbind")
	beforeCmt := countBodyComments(t, svc, iss.ID, "%离职解绑%")

	if _, _, err := svc.UnbindSeat(ws, team.ID, s1.ID, m2, true); err != nil {
		t.Fatal(err)
	}
	if countActionActivities(t, svc, ws, "seat.unbind") != beforeAct+1 {
		t.Fatal("want +1 seat.unbind")
	}
	var payload string
	if err := svc.DB.QueryRow(
		`SELECT payload FROM activities WHERE workspace_id=? AND action=? ORDER BY created_at DESC LIMIT 1`,
		ws, "seat.unbind",
	).Scan(&payload); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"离职解绑", "运维", "离职", "BotA"} {
		if !strings.Contains(payload, want) {
			t.Fatalf("payload missing %q: %s", want, payload)
		}
	}
	if countBodyComments(t, svc, iss.ID, "%离职解绑%") != beforeCmt+1 {
		t.Fatal("want Issue comment 人话审计")
	}
}

// TestUG1CascadeUnbindOnMemberRemove U-G1：移除成员前级联清席→left + release claim + 审计。
func TestUG1CascadeUnbindOnMemberRemove(t *testing.T) {
	svc, ws, m1, m2 := setup(t)
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "UG1"})
	if err != nil {
		t.Fatal(err)
	}
	s1, err := svc.CreateSeat(ws, team.ID, botmem.EnsureSeatInput{Slug: "fe", Name: "前端", PlaybookRef: "bot-team-template@main:playbook/fe.md"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.BindSeat(ws, team.ID, s1.ID, m1, false); err != nil {
		t.Fatal(err)
	}
	if s1.BotRoleID == nil || *s1.BotRoleID == "" {
		t.Fatal("seat missing bot_role_id")
	}
	roleID := *s1.BotRoleID
	var claimMember string
	err = svc.DB.QueryRow(`SELECT member_id FROM bot_role_claims WHERE role_id=? AND released_at IS NULL`, roleID).Scan(&claimMember)
	if err != nil || claimMember != m1 {
		t.Fatalf("want active claim by m1: member=%q err=%v", claimMember, err)
	}
	before := countActionActivities(t, svc, ws, "seat.unbind")
	n, err := svc.CascadeUnbindSeatsForMember(ws, m1, m2)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("cascade n=%d want 1", n)
	}
	// 幂等 get：无 bound 的 UnbindSeat 返回当前席态
	seat, _, err := svc.UnbindSeat(ws, team.ID, s1.ID, m2, true)
	if err != nil {
		t.Fatal(err)
	}
	if seat.Status != "left" {
		t.Fatalf("status=%s want left", seat.Status)
	}
	if seat.BoundMemberID != nil {
		t.Fatalf("bound should clear: %+v", seat.BoundMemberID)
	}
	if countActionActivities(t, svc, ws, "seat.unbind") != before+1 {
		t.Fatal("want +1 seat.unbind audit")
	}
	var still int
	_ = svc.DB.QueryRow(`SELECT COUNT(*) FROM bot_role_claims WHERE role_id=? AND released_at IS NULL`, roleID).Scan(&still)
	if still != 0 {
		t.Fatalf("claim should be released, still=%d", still)
	}
	// 再删成员行：席应仍无脏绑（级联已清；FK SET NULL 兜底）
	if _, err := svc.DB.Exec(`DELETE FROM members WHERE id=? AND workspace_id=?`, m1, ws); err != nil {
		t.Fatal(err)
	}
	var bound sql.NullString
	if err := svc.DB.QueryRow(`SELECT bound_member_id FROM team_seats WHERE id=?`, s1.ID).Scan(&bound); err != nil {
		t.Fatal(err)
	}
	if bound.Valid && bound.String != "" {
		t.Fatalf("after DELETE members still dirty bound: %q", bound.String)
	}
}

// TestUG2StaleBoundUnbindClears U-G2：脏绑（bound→已删 member）unbind 成功+人话。
func TestUG2StaleBoundUnbindClears(t *testing.T) {
	svc, ws, m1, m2 := setup(t)
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "UG2"})
	if err != nil {
		t.Fatal(err)
	}
	s1, err := svc.CreateSeat(ws, team.ID, botmem.EnsureSeatInput{Slug: "be", Name: "后端", PlaybookRef: "bot-team-template@main:playbook/be.md"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.BindSeat(ws, team.ID, s1.ID, m1, false); err != nil {
		t.Fatal(err)
	}
	// 模拟历史脏绑：关 FK 后删成员，bound 残留指向已删 id
	if _, err := svc.DB.Exec(`PRAGMA foreign_keys=OFF`); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DB.Exec(`DELETE FROM members WHERE id=?`, m1); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DB.Exec(`PRAGMA foreign_keys=ON`); err != nil {
		t.Fatal(err)
	}
	var bound string
	if err := svc.DB.QueryRow(`SELECT bound_member_id FROM team_seats WHERE id=?`, s1.ID).Scan(&bound); err != nil {
		t.Fatal(err)
	}
	if bound != m1 {
		t.Fatalf("dirty bind not preserved: bound=%q want %s", bound, m1)
	}
	// 非 Admin、非原绑人（m2）也可清脏绑；禁再抛「只能解绑自己」
	left, msg, err := svc.UnbindSeat(ws, team.ID, s1.ID, m2, false)
	if err != nil {
		t.Fatalf("stale unbind: %v", err)
	}
	if msg != botmem.UnbindSeatMsgStaleBound {
		t.Fatalf("msg=%q want %q", msg, botmem.UnbindSeatMsgStaleBound)
	}
	if left.Status != "left" {
		t.Fatalf("status=%s want left", left.Status)
	}
	if left.BoundMemberID != nil {
		t.Fatalf("bound should clear: %+v", left.BoundMemberID)
	}
	if strings.Contains(msg, "只能解绑自己") {
		t.Fatal("must not reuse old forbidden copy")
	}
}

// TestUG3AliveOccupantForbiddenMessage U-G3：他人存活占用 → 403 新人话含显示名。
func TestUG3AliveOccupantForbiddenMessage(t *testing.T) {
	svc, ws, m1, m2 := setup(t)
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "UG3"})
	if err != nil {
		t.Fatal(err)
	}
	s1, err := svc.CreateSeat(ws, team.ID, botmem.EnsureSeatInput{Slug: "qa", Name: "QA", PlaybookRef: "bot-team-template@main:playbook/qa.md"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.BindSeat(ws, team.ID, s1.ID, m1, false); err != nil {
		t.Fatal(err)
	}
	_, msg, err := svc.UnbindSeat(ws, team.ID, s1.ID, m2, false)
	if err == nil {
		t.Fatal("expected forbidden")
	}
	if msg != "" {
		t.Fatalf("msg should be empty on err, got %q", msg)
	}
	ae, ok := err.(*apperr.AppError)
	want := "席位由「BotA」占用；请对方离席，或让管理员停工/离职解绑"
	if !ok || ae.Code != "forbidden" || ae.Message != want {
		t.Fatalf("want %q, got %#v", want, err)
	}
	if strings.Contains(ae.Message, "只能解绑自己") {
		t.Fatal("old copy must be gone")
	}
}

// TestUG4UnboundIdempotentUnbind U-G4：无 bound 幂等 unbind 回归。
func TestUG4UnboundIdempotentUnbind(t *testing.T) {
	svc, ws, m1, _ := setup(t)
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "UG4"})
	if err != nil {
		t.Fatal(err)
	}
	s1, err := svc.CreateSeat(ws, team.ID, botmem.EnsureSeatInput{Slug: "ops", Name: "运维", PlaybookRef: "bot-team-template@main:playbook/ops.md"})
	if err != nil {
		t.Fatal(err)
	}
	// 从未绑定
	before := countActionActivities(t, svc, ws, "seat.unbind")
	seat, msg, err := svc.UnbindSeat(ws, team.ID, s1.ID, m1, false)
	if err != nil {
		t.Fatal(err)
	}
	if msg != "" {
		t.Fatalf("idempotent msg should be empty, got %q", msg)
	}
	if seat.BoundMemberID != nil {
		t.Fatalf("unexpected bound: %+v", seat.BoundMemberID)
	}
	if countActionActivities(t, svc, ws, "seat.unbind") != before {
		t.Fatal("idempotent unbind must not write audit")
	}
	// bind → unbind → 再 unbind 幂等
	if _, err := svc.BindSeat(ws, team.ID, s1.ID, m1, false); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.UnbindSeat(ws, team.ID, s1.ID, m1, false); err != nil {
		t.Fatal(err)
	}
	after := countActionActivities(t, svc, ws, "seat.unbind")
	again, msg2, err := svc.UnbindSeat(ws, team.ID, s1.ID, m1, false)
	if err != nil {
		t.Fatal(err)
	}
	if msg2 != "" {
		t.Fatalf("second idempotent msg=%q", msg2)
	}
	if again.Status != "left" {
		t.Fatalf("status=%s", again.Status)
	}
	if countActionActivities(t, svc, ws, "seat.unbind") != after {
		t.Fatal("second unbind must not duplicate audit")
	}
}
