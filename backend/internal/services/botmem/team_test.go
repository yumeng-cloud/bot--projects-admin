package botmem_test

import (
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
	i1, err := ib.CreateIssue(ws, issueboard.CreateIssueInput{ProjectID: proj.ID, Title: "a", Status: "todo", ActorKind: "human"})
	if err != nil {
		t.Fatal(err)
	}
	i2, err := ib.CreateIssue(ws, issueboard.CreateIssueInput{ProjectID: proj.ID, Title: "b", Status: "backlog", ActorKind: "human"})
	if err != nil {
		t.Fatal(err)
	}
	// already assigned — must skip
	i3, err := ib.CreateIssue(ws, issueboard.CreateIssueInput{ProjectID: proj.ID, Title: "c", Status: "todo", AssigneeID: &m1, ActorKind: "human"})
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
	s1, err := svc.CreateSeat(ws, team.ID, botmem.EnsureSeatInput{Slug: "ops", Name: "运维"})
	if err != nil {
		t.Fatal(err)
	}
	if s1.Status != "open" {
		t.Fatalf("status=%s want open", s1.Status)
	}
	s2, err := svc.CreateSeat(ws, team.ID, botmem.EnsureSeatInput{Slug: "ops", Name: "运维改名忽略"})
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
	stopped, err := svc.StopSeat(ws, team.ID, s1.ID)
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
	s1, err := svc.CreateSeat(ws, team.ID, botmem.EnsureSeatInput{Slug: "fe", Name: "前端"})
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
	left, err := svc.UnbindSeat(ws, team.ID, s1.ID, m1, false)
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
	s1, err := svc.CreateSeat(ws, team.ID, botmem.EnsureSeatInput{Slug: "ops", Name: "运维"})
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
		ProjectID: proj.ID, Title: "停工补位审计", Status: "todo", ActorKind: "human", AssigneeID: &m1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.StopSeat(ws, team.ID, s1.ID); err != nil {
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
	s1, err := svc.CreateSeat(ws, team.ID, botmem.EnsureSeatInput{Slug: "qa", Name: "QA"})
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
		ProjectID: proj.ID, Title: "离职补位审计", Status: "in_progress", ActorKind: "human", AssigneeID: &m1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UnbindSeat(ws, team.ID, s1.ID, m1, false); err != nil {
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
	s1, err := svc.CreateSeat(ws, team.ID, botmem.EnsureSeatInput{Slug: "fe2", Name: "前端2"})
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
		ProjectID: proj.ID, Title: "活跃不可 force", Status: "todo", ActorKind: "human", AssigneeID: &m1,
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
	s1, err := svc.CreateSeat(ws, team.ID, botmem.EnsureSeatInput{Slug: "ops2", Name: "运维2"})
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
