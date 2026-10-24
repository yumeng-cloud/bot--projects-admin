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
	if !strings.Contains(payload, "force 清脏绑") || !strings.Contains(payload, m1) || !strings.Contains(payload, m2) {
		t.Fatalf("payload=%s", payload)
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
	if err := svc.RevokeInvite(ws, team.ID, inv.ID); err != nil {
		// already at max; create fresh to revoke
		inv2, err2 := svc.CreateInvite(ws, team.ID, botmem.CreateInviteInput{CreatedBy: m1})
		if err2 != nil {
			t.Fatal(err2)
		}
		if err := svc.RevokeInvite(ws, team.ID, inv2.ID); err != nil {
			t.Fatal(err)
		}
		_, err = svc.RedeemInvite(inv2.Code)
		if err == nil {
			t.Fatal("expected revoked")
		}
	}
}
