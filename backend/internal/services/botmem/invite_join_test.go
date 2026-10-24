package botmem_test

import (
	"strings"
	"testing"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/botmem"
)

func TestRedeemJoinBootstrapEmptySeat(t *testing.T) {
	svc, ws, _, _ := setup(t)
	tm, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "JoinAuto"})
	if err != nil {
		t.Fatal(err)
	}
	roster, err := svc.EnsureRoster(ws, tm.ID, botmem.EnsureRosterInput{})
	if err != nil {
		t.Fatal(err)
	}
	seat := roster[0]
	inv, err := svc.CreateInvite(ws, tm.ID, botmem.CreateInviteInput{
		SeatSlug: seat.RoleSlug, RoleHint: seat.RoleSlug,
	})
	if err != nil {
		t.Fatal(err)
	}

	res, err := svc.RedeemInviteAndJoin(botmem.RedeemJoinInput{
		Code: inv.Code, Provision: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.TokenIssued || !strings.HasPrefix(res.Token, "ymk_") {
		t.Fatalf("want ymk_ token: %+v", res)
	}
	if res.MemberID == "" || res.SeatID == "" {
		t.Fatalf("want member+seat: %+v", res)
	}
	if res.WorkspaceID != ws || res.TeamID != tm.ID {
		t.Fatalf("locate: %+v", res)
	}
	got, err := svc.GetRoster(ws, tm.ID)
	if err != nil {
		t.Fatal(err)
	}
	bound := false
	for _, s := range got {
		if s.ID == res.SeatID && s.BoundMemberID != nil && *s.BoundMemberID == res.MemberID {
			bound = true
			if s.Status != "active" {
				t.Fatalf("status=%s", s.Status)
			}
		}
	}
	if !bound {
		t.Fatal("seat not bound")
	}
}

func TestRedeemJoinIdempotentNoRetoken(t *testing.T) {
	svc, ws, _, _ := setup(t)
	tm, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "Idem"})
	if err != nil {
		t.Fatal(err)
	}
	roster, err := svc.EnsureRoster(ws, tm.ID, botmem.EnsureRosterInput{})
	if err != nil {
		t.Fatal(err)
	}
	seat := roster[0]
	inv, err := svc.CreateInvite(ws, tm.ID, botmem.CreateInviteInput{
		SeatSlug: seat.RoleSlug, RoleHint: seat.RoleSlug,
	})
	if err != nil {
		t.Fatal(err)
	}
	first, err := svc.RedeemInviteAndJoin(botmem.RedeemJoinInput{Code: inv.Code, Provision: true})
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.RedeemInviteAndJoin(botmem.RedeemJoinInput{
		Code:             inv.Code,
		ExistingMemberID: first.MemberID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if second.TokenIssued || second.Token != "" {
		t.Fatalf("must not reissue token: %+v", second)
	}
	if second.MemberID != first.MemberID {
		t.Fatalf("member changed: %s -> %s", first.MemberID, second.MemberID)
	}
}

func TestRedeemJoinV1SecondProvisionConflict(t *testing.T) {
	svc, ws, _, _ := setup(t)
	tm, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "V1"})
	if err != nil {
		t.Fatal(err)
	}
	roster, err := svc.EnsureRoster(ws, tm.ID, botmem.EnsureRosterInput{})
	if err != nil {
		t.Fatal(err)
	}
	// 两空席：若无 V1 会造第二人
	if len(roster) < 2 {
		t.Fatal("need >=2 seats")
	}
	inv, err := svc.CreateInvite(ws, tm.ID, botmem.CreateInviteInput{
		SeatSlug: roster[0].RoleSlug, RoleHint: roster[0].RoleSlug,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.RedeemInviteAndJoin(botmem.RedeemJoinInput{Code: inv.Code, Provision: true}); err != nil {
		t.Fatal(err)
	}
	_, err = svc.RedeemInviteAndJoin(botmem.RedeemJoinInput{Code: inv.Code, Provision: true})
	if err == nil {
		t.Fatal("expected V1 conflict")
	}
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.Code != "conflict" {
		t.Fatalf("want conflict got %#v", err)
	}
	if !strings.Contains(ae.Message, "请粘贴已发的身份钥匙续队") {
		t.Fatalf("msg=%q", ae.Message)
	}
}

// TestRedeemJoinActiveNoReplace B2：目标席 active 仍 409 禁顶替。
func TestRedeemJoinActiveNoReplace(t *testing.T) {
	svc, ws, m1, _ := setup(t)
	tm, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "Active"})
	if err != nil {
		t.Fatal(err)
	}
	roster, err := svc.EnsureRoster(ws, tm.ID, botmem.EnsureRosterInput{})
	if err != nil {
		t.Fatal(err)
	}
	seat := roster[0]
	if _, err := svc.BindSeat(ws, tm.ID, seat.ID, m1, false); err != nil {
		t.Fatal(err)
	}
	inv, err := svc.CreateInvite(ws, tm.ID, botmem.CreateInviteInput{
		SeatSlug: seat.RoleSlug, RoleHint: seat.RoleSlug,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.RedeemInviteAndJoin(botmem.RedeemJoinInput{Code: inv.Code, Provision: true})
	if err == nil {
		t.Fatal("expected conflict on active")
	}
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.Code != "conflict" {
		t.Fatalf("want conflict %#v", err)
	}
	got, _ := svc.GetRoster(ws, tm.ID)
	for _, s := range got {
		if s.ID == seat.ID {
			if s.BoundMemberID == nil || *s.BoundMemberID != m1 {
				t.Fatalf("active replaced: %+v", s)
			}
		}
	}
}

// TestRedeemJoinV2NoAutoFillLeftStopped B2：未锚定通码仍遵 HUMAN V2 —— 不静默补 stopped|left。
func TestRedeemJoinV2NoAutoFillLeftStopped(t *testing.T) {
	svc, ws, m1, _ := setup(t)
	tm, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "V2"})
	if err != nil {
		t.Fatal(err)
	}
	roster, err := svc.EnsureRoster(ws, tm.ID, botmem.EnsureRosterInput{})
	if err != nil {
		t.Fatal(err)
	}
	// 把所有席停工/离职，通码无可绑 open，仍不得静默补
	for i, seat := range roster {
		if _, err := svc.BindSeat(ws, tm.ID, seat.ID, m1, false); err != nil {
			t.Fatal(err)
		}
		if i%2 == 0 {
			if _, err := svc.StopSeat(ws, tm.ID, seat.ID, m1); err != nil {
				t.Fatal(err)
			}
		} else {
			if _, err := svc.UnbindSeat(ws, tm.ID, seat.ID, m1, false); err != nil {
				t.Fatal(err)
			}
		}
	}
	// 通码：无 seat_slug 锚定（仅 RoleHint 不算锚定）
	inv, err := svc.CreateInvite(ws, tm.ID, botmem.CreateInviteInput{
		RoleHint: "frontend",
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := svc.RedeemInviteAndJoin(botmem.RedeemJoinInput{Code: inv.Code, Provision: true})
	if err != nil {
		t.Fatal(err)
	}
	if !res.TokenIssued {
		t.Fatalf("want token: %+v", res)
	}
	if res.SeatID != "" {
		t.Fatalf("V2 通码不得绑停工/离职席: seat_id=%s", res.SeatID)
	}
	got, err := svc.GetRoster(ws, tm.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range got {
		if s.BoundMemberID != nil && *s.BoundMemberID != "" {
			t.Fatalf("V2 auto-filled: %+v", s)
		}
		st := s.Status
		if st != "stopped" && st != "left" {
			t.Fatalf("want stopped|left got %s for %s", st, s.RoleSlug)
		}
	}
}

// TestRedeemJoinAnchoredRefillBindStopped B2：锚定补位码对 stopped 席 redeem 直接绑（非 skipBind）。
func TestRedeemJoinAnchoredRefillBindStopped(t *testing.T) {
	svc, ws, m1, _ := setup(t)
	tm, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "RefillBind"})
	if err != nil {
		t.Fatal(err)
	}
	roster, err := svc.EnsureRoster(ws, tm.ID, botmem.EnsureRosterInput{})
	if err != nil {
		t.Fatal(err)
	}
	seat := roster[0]
	if _, err := svc.BindSeat(ws, tm.ID, seat.ID, m1, false); err != nil {
		t.Fatal(err)
	}
	stopped, err := svc.StopSeat(ws, tm.ID, seat.ID, m1)
	if err != nil {
		t.Fatal(err)
	}
	if stopped.Status != "stopped" {
		t.Fatalf("status=%s", stopped.Status)
	}
	inv, err := svc.CreateInvite(ws, tm.ID, botmem.CreateInviteInput{
		SeatSlug: seat.RoleSlug, RoleHint: seat.RoleSlug,
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := svc.RedeemInviteAndJoin(botmem.RedeemJoinInput{Code: inv.Code, Provision: true})
	if err != nil {
		t.Fatal(err)
	}
	if !res.TokenIssued || res.MemberID == "" {
		t.Fatalf("want bootstrap: %+v", res)
	}
	if res.SeatID != seat.ID {
		t.Fatalf("want bind stopped seat %s got %s", seat.ID, res.SeatID)
	}
	got, err := svc.GetRoster(ws, tm.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range got {
		if s.ID == seat.ID {
			if s.Status != "active" {
				t.Fatalf("status=%s want active", s.Status)
			}
			if s.BoundMemberID == nil || *s.BoundMemberID != res.MemberID {
				t.Fatalf("bound=%v want %s", s.BoundMemberID, res.MemberID)
			}
		}
	}

	// left 同理：锚定码直接绑
	seat2 := roster[1]
	if _, err := svc.BindSeat(ws, tm.ID, seat2.ID, m1, false); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UnbindSeat(ws, tm.ID, seat2.ID, m1, false); err != nil {
		t.Fatal(err)
	}
	inv2, err := svc.CreateInvite(ws, tm.ID, botmem.CreateInviteInput{
		SeatSlug: seat2.RoleSlug, RoleHint: seat2.RoleSlug,
	})
	if err != nil {
		t.Fatal(err)
	}
	res2, err := svc.RedeemInviteAndJoin(botmem.RedeemJoinInput{Code: inv2.Code, Provision: true})
	if err != nil {
		t.Fatal(err)
	}
	if res2.SeatID != seat2.ID {
		t.Fatalf("want bind left seat %s got %s", seat2.ID, res2.SeatID)
	}
	got, _ = svc.GetRoster(ws, tm.ID)
	for _, s := range got {
		if s.ID == seat2.ID {
			if s.Status != "active" || s.BoundMemberID == nil || *s.BoundMemberID != res2.MemberID {
				t.Fatalf("left refill bind: %+v", s)
			}
		}
	}
}

func TestRedeemInviteClassicStillWorks(t *testing.T) {
	svc, ws, m1, _ := setup(t)
	tm, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "Classic"})
	if err != nil {
		t.Fatal(err)
	}
	max := 2
	inv, err := svc.CreateInvite(ws, tm.ID, botmem.CreateInviteInput{
		RoleHint: "frontend", MaxUses: &max, CreatedBy: m1,
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := svc.RedeemInvite(inv.Code)
	if err != nil {
		t.Fatal(err)
	}
	if res.TokenIssued || res.Token != "" || res.MemberID != "" {
		t.Fatalf("classic must not mint: %+v", res)
	}
	if res.WorkspaceID != ws {
		t.Fatalf("%+v", res)
	}
}
