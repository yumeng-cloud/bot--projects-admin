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
	// V2：可发身份，但不得把 stopped 席自动补上
	if !res.TokenIssued {
		t.Fatalf("want token: %+v", res)
	}
	got, err := svc.GetRoster(ws, tm.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range got {
		if s.ID == seat.ID {
			if s.Status != "stopped" {
				t.Fatalf("V2 violated: status=%s bound=%v", s.Status, s.BoundMemberID)
			}
			if s.BoundMemberID != nil && *s.BoundMemberID != "" {
				t.Fatalf("V2 auto-filled stopped seat: %+v", s)
			}
		}
	}

	// left 同理
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
	if !res2.TokenIssued {
		t.Fatalf("want token left: %+v", res2)
	}
	got, _ = svc.GetRoster(ws, tm.ID)
	for _, s := range got {
		if s.ID == seat2.ID {
			if s.Status != "left" {
				t.Fatalf("V2 left status=%s", s.Status)
			}
			if s.BoundMemberID != nil && *s.BoundMemberID != "" {
				t.Fatalf("V2 auto-filled left: %+v", s)
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
