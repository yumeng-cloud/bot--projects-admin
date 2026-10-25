package botmem_test

import (
	"strings"
	"testing"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/models"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/botmem"
)

func TestInviteOnDemandIG1CreateRequireSeatOrRole(t *testing.T) {
	svc, ws, m1, _ := setup(t)
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "IG1-require"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.CreateInvite(ws, team.ID, botmem.CreateInviteInput{CreatedBy: m1})
	if err == nil {
		t.Fatal("expected 400")
	}
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.Code != "bad_request" {
		t.Fatalf("want bad_request %#v", err)
	}
	if !strings.Contains(ae.Message, "seat_slug") {
		t.Fatalf("msg=%q", ae.Message)
	}
}

func TestInviteOnDemandIG1ActiveSeatConflict(t *testing.T) {
	svc, ws, m1, _ := setup(t)
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "IG1-409"})
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
	_, err = svc.CreateInvite(ws, team.ID, botmem.CreateInviteInput{
		SeatSlug:  seat.RoleSlug,
		CreatedBy: m1,
	})
	if err == nil {
		t.Fatal("expected 409")
	}
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.Code != "conflict" {
		t.Fatalf("want conflict %#v", err)
	}
	if !strings.Contains(ae.Message, "席位已有人在席") {
		t.Fatalf("msg=%q", ae.Message)
	}
}

func TestInviteOnDemandIG1ReplaceRevokesThenMints(t *testing.T) {
	svc, ws, m1, _ := setup(t)
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "IG1-replace"})
	if err != nil {
		t.Fatal(err)
	}
	roster, err := svc.EnsureRoster(ws, team.ID, botmem.EnsureRosterInput{})
	if err != nil {
		t.Fatal(err)
	}
	slug := roster[0].RoleSlug

	old, err := svc.CreateInvite(ws, team.ID, botmem.CreateInviteInput{
		SeatSlug:  slug,
		CreatedBy: m1,
	})
	if err != nil {
		t.Fatal(err)
	}
	neu, err := svc.CreateInvite(ws, team.ID, botmem.CreateInviteInput{
		SeatSlug:  slug,
		Replace:   true,
		CreatedBy: m1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if neu.ID == old.ID {
		t.Fatal("expected new invite id")
	}
	active, err := svc.ListInvites(ws, team.ID, botmem.ListInvitesFilter{SeatSlug: slug, Status: models.InviteStatusActive})
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 1 || active[0].ID != neu.ID {
		t.Fatalf("active=%+v", active)
	}
	got, err := svc.RevokeInvite(ws, team.ID, old.ID) // already revoked; idempotent
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != models.InviteStatusRevoked {
		t.Fatalf("old status=%s", got.Status)
	}
}

func TestInviteOnDemandIG1ListFilterAndLabel(t *testing.T) {
	svc, ws, m1, _ := setup(t)
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "IG1-list"})
	if err != nil {
		t.Fatal(err)
	}
	roster, err := svc.EnsureRoster(ws, team.ID, botmem.EnsureRosterInput{})
	if err != nil {
		t.Fatal(err)
	}
	slugA := roster[0].RoleSlug
	slugB := roster[1].RoleSlug

	invA, err := svc.CreateInvite(ws, team.ID, botmem.CreateInviteInput{SeatSlug: slugA, CreatedBy: m1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateInvite(ws, team.ID, botmem.CreateInviteInput{SeatSlug: slugB, CreatedBy: m1}); err != nil {
		t.Fatal(err)
	}
	if invA.Label == "" {
		t.Fatal("create label empty")
	}
	if invA.Label == "未命名席" && roster[0].DisplayName != "" && !strings.HasPrefix(roster[0].DisplayName, "team:") {
		t.Fatalf("label=%q display=%q", invA.Label, roster[0].DisplayName)
	}

	filtered, err := svc.ListInvites(ws, team.ID, botmem.ListInvitesFilter{SeatSlug: slugA, Status: models.InviteStatusActive})
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 1 || filtered[0].ID != invA.ID {
		t.Fatalf("filter=%+v", filtered)
	}
	if filtered[0].Label == "" {
		t.Fatal("list label empty")
	}
	if filtered[0].Code != "" {
		t.Fatal("list leaked code")
	}

	// dirty / unknown slug → 未命名席
	invDirty, err := svc.CreateInvite(ws, team.ID, botmem.CreateInviteInput{
		RoleHint:  "no-such-role-" + idgen.New()[:8],
		CreatedBy: m1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if invDirty.Label != "未命名席" {
		t.Fatalf("dirty label=%q", invDirty.Label)
	}
}

func TestInviteOnDemandIG1LookupEntryYmtThen404(t *testing.T) {
	svc, ws, m1, _ := setup(t)
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "IG1-lookup"})
	if err != nil {
		t.Fatal(err)
	}
	roster, err := svc.EnsureRoster(ws, team.ID, botmem.EnsureRosterInput{})
	if err != nil {
		t.Fatal(err)
	}
	slug := roster[0].RoleSlug

	_, err = svc.LookupEntry(ws, team.ID, "", "")
	if err == nil {
		t.Fatal("expected 400")
	}
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.Code != "bad_request" {
		t.Fatalf("want bad_request %#v", err)
	}

	_, err = svc.LookupEntry(ws, team.ID, slug, "")
	if err == nil {
		t.Fatal("expected 404")
	}
	ae, ok = err.(*apperr.AppError)
	if !ok || ae.Code != "not_found" {
		t.Fatalf("want not_found %#v", err)
	}
	if !strings.Contains(ae.Message, "该席位暂无可用入场凭证") {
		t.Fatalf("msg=%q", ae.Message)
	}

	inv, err := svc.CreateInvite(ws, team.ID, botmem.CreateInviteInput{SeatSlug: slug, CreatedBy: m1})
	if err != nil {
		t.Fatal(err)
	}
	entry, err := svc.LookupEntry(ws, team.ID, slug, "")
	if err != nil {
		t.Fatal(err)
	}
	if entry.Kind != "ymt_" || entry.InviteID != inv.ID {
		t.Fatalf("entry=%+v", entry)
	}
	if entry.CodeTail == "" {
		t.Fatal("code_tail empty")
	}
}

func TestInviteOnDemandIG1LookupPrefersYmk(t *testing.T) {
	svc, ws, m1, _ := setup(t)
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "IG1-ymk"})
	if err != nil {
		t.Fatal(err)
	}
	roster, err := svc.EnsureRoster(ws, team.ID, botmem.EnsureRosterInput{})
	if err != nil {
		t.Fatal(err)
	}
	seat := roster[0]
	if _, err := svc.CreateInvite(ws, team.ID, botmem.CreateInviteInput{SeatSlug: seat.RoleSlug, CreatedBy: m1}); err != nil {
		t.Fatal(err)
	}
	// leave seat empty first then bind — but CreateInvite on active is 409, so bind after invite exists
	bound, err := svc.BindSeat(ws, team.ID, seat.ID, m1, false)
	if err != nil {
		t.Fatal(err)
	}
	if bound.BoundMemberID == nil {
		t.Fatal("not bound")
	}
	// insert unrevoked api key for bound member
	keyID := idgen.New()
	prefix := "ymk_abcd1234ef"
	_, err = svc.DB.Exec(`INSERT INTO api_keys (id, workspace_id, member_id, name, key_hash, key_prefix, created_at)
		VALUES (?,?,?,?,?,?,datetime('now'))`, keyID, ws, m1, "t", "hash-"+keyID, prefix)
	if err != nil {
		t.Fatal(err)
	}
	entry, err := svc.LookupEntry(ws, team.ID, seat.RoleSlug, "")
	if err != nil {
		t.Fatal(err)
	}
	if entry.Kind != "ymk_" || entry.KeyID != keyID {
		t.Fatalf("want ymk_ got %+v", entry)
	}
	if entry.MemberID != m1 {
		t.Fatalf("member=%s", entry.MemberID)
	}
}
