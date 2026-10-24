package botmem_test

import (
	"strings"
	"testing"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/botmem"
)

// T7-3：roster / bind(enter) 回包带可读 playbook_ref；空/非法写入 → 4xx；同席两次出口同 ref。

func TestT73RosterAndBindIncludePlaybookRef(t *testing.T) {
	svc, ws, m1, _ := setup(t)
	res, err := svc.CreateTeamFromTemplate(ws, botmem.CreateTeamInput{
		Name: "T73", Slug: "t73-a", TemplateKey: "default",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Roster) == 0 {
		t.Fatal("empty roster")
	}
	for _, seat := range res.Roster {
		if seat.PlaybookRef == nil || strings.TrimSpace(*seat.PlaybookRef) == "" {
			t.Fatalf("roster seat %s missing playbook_ref: %+v", seat.RoleSlug, seat)
		}
		if _, err := botmem.ParsePlaybookRef(*seat.PlaybookRef); err != nil {
			t.Fatalf("roster playbook_ref %q: %v", *seat.PlaybookRef, err)
		}
	}
	seat := res.Roster[0]
	want := *seat.PlaybookRef

	roster2, err := svc.GetRoster(ws, res.Team.ID)
	if err != nil {
		t.Fatal(err)
	}
	var again string
	for _, s := range roster2 {
		if s.ID == seat.ID {
			if s.PlaybookRef == nil {
				t.Fatal("second roster missing playbook_ref")
			}
			again = *s.PlaybookRef
			break
		}
	}
	if again != want {
		t.Fatalf("roster twice: %q vs %q", again, want)
	}

	bound, err := svc.BindSeat(ws, res.Team.ID, seat.ID, m1, false)
	if err != nil {
		t.Fatal(err)
	}
	if bound.PlaybookRef == nil || *bound.PlaybookRef != want {
		t.Fatalf("bind/enter playbook=%v want %s", bound.PlaybookRef, want)
	}
	bound2, err := svc.BindSeat(ws, res.Team.ID, seat.ID, m1, false)
	if err != nil {
		t.Fatal(err)
	}
	if bound2.PlaybookRef == nil || *bound2.PlaybookRef != want {
		t.Fatalf("bind twice playbook=%v want %s", bound2.PlaybookRef, want)
	}
}

func TestT73EmptyPlaybookRefWrite4xx(t *testing.T) {
	svc, ws, _, _ := setup(t)
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "T73-empty"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.CreateSeat(ws, team.ID, botmem.EnsureSeatInput{Slug: "ops", Name: "运维", PlaybookRef: ""})
	if err == nil {
		t.Fatal("expected 400 empty playbook_ref")
	}
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.Code != "bad_request" {
		t.Fatalf("want bad_request got %#v", err)
	}
	if !strings.Contains(ae.Message, "playbook_ref") {
		t.Fatalf("msg=%q", ae.Message)
	}

	_, err = svc.EnsureRoster(ws, team.ID, botmem.EnsureRosterInput{
		Seats: []botmem.EnsureSeatInput{{Slug: "qa", Name: "QA", PlaybookRef: ""}},
	})
	if err == nil {
		t.Fatal("expected 400 ensure empty")
	}
	ae, ok = err.(*apperr.AppError)
	if !ok || ae.Code != "bad_request" || !strings.Contains(ae.Message, "playbook_ref") {
		t.Fatalf("got %#v", err)
	}
}

func TestT73IllegalPlaybookRefWrite4xx(t *testing.T) {
	svc, ws, _, _ := setup(t)
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "T73-bad"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.CreateSeat(ws, team.ID, botmem.EnsureSeatInput{
		Slug: "ops", Name: "运维", PlaybookRef: "bot-team-template@main:../etc/passwd",
	})
	if err == nil {
		t.Fatal("expected 400 illegal path")
	}
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.Code != "bad_request" {
		t.Fatalf("got %#v", err)
	}
	_, err = svc.EnsureRoster(ws, team.ID, botmem.EnsureRosterInput{
		Seats: []botmem.EnsureSeatInput{{Slug: "fe", Name: "前端", PlaybookRef: "{not-json"}},
	})
	if err == nil {
		t.Fatal("expected 400 illegal json")
	}
	ae, ok = err.(*apperr.AppError)
	if !ok || ae.Code != "bad_request" {
		t.Fatalf("got %#v", err)
	}
}

func TestT73CanonicalNoSilentFallback(t *testing.T) {
	_, err := botmem.CanonicalPlaybookRefString("")
	if err == nil {
		t.Fatal("empty must error")
	}
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.Code != "bad_request" {
		t.Fatalf("got %#v", err)
	}
	got, err := botmem.CanonicalPlaybookRefString("bot-team-template@main:playbook/x.md")
	if err != nil {
		t.Fatal(err)
	}
	if got != "bot-team-template@main:playbook/x.md" {
		t.Fatalf("got %q", got)
	}
}
