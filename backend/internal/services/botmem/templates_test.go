package botmem_test

import (
	"strings"
	"testing"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/botmem"
)

func TestListTeamTemplatesHasDefault(t *testing.T) {
	list, err := botmem.ListTeamTemplateDTOs()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) == 0 {
		t.Fatal("expected at least 1 builtin template")
	}
	var def *botmem.TemplateDTO
	for i := range list {
		if list[i].Key == "default" {
			def = &list[i]
			break
		}
	}
	if def == nil {
		t.Fatal("missing default template")
	}
	if len(def.Seats) != 5 {
		t.Fatalf("default seats=%d want 5", len(def.Seats))
	}
	for _, s := range def.Seats {
		if s.Slug == "" || s.DisplayName == "" || s.PlaybookRef == "" {
			t.Fatalf("incomplete seat: %+v", s)
		}
		if _, err := botmem.ParsePlaybookRef(s.PlaybookRef); err != nil {
			t.Fatalf("playbook_ref %q: %v", s.PlaybookRef, err)
		}
	}
}

func TestCreateTeamFromTemplateR1(t *testing.T) {
	svc, ws, _, _ := setup(t)
	tmpl, err := botmem.GetTeamTemplate("default")
	if err != nil {
		t.Fatal(err)
	}
	res, err := svc.CreateTeamFromTemplate(ws, botmem.CreateTeamInput{
		Name: "队A", Slug: "team-a", TemplateKey: "default",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Team == nil || res.Team.TemplateKey != "default" {
		t.Fatalf("team: %+v", res.Team)
	}
	if len(res.Roster) != len(tmpl.Seats) {
		t.Fatalf("roster len=%d want %d", len(res.Roster), len(tmpl.Seats))
	}
	want := map[string]botmem.TemplateSeat{}
	for _, s := range tmpl.Seats {
		want[s.Slug] = s
	}
	for _, seat := range res.Roster {
		w, ok := want[seat.RoleSlug]
		if !ok {
			t.Fatalf("unexpected slug %s", seat.RoleSlug)
		}
		if seat.DisplayName != w.DisplayName {
			t.Fatalf("slug %s name=%q want %q", seat.RoleSlug, seat.DisplayName, w.DisplayName)
		}
		wantRef := botmem.FormatPlaybookRef(w.PlaybookRef)
		if seat.PlaybookRef == nil || *seat.PlaybookRef != wantRef {
			t.Fatalf("slug %s playbook=%v want %s", seat.RoleSlug, seat.PlaybookRef, wantRef)
		}
		if seat.BoundMemberID != nil {
			t.Fatalf("must not auto-bind Key: %+v", seat)
		}
		if seat.Status != "open" {
			t.Fatalf("status=%s want open", seat.Status)
		}
	}
}

func TestCreateTeamFromTemplateUnknownR2(t *testing.T) {
	svc, ws, _, _ := setup(t)
	_, err := svc.CreateTeamFromTemplate(ws, botmem.CreateTeamInput{
		Name: "X", TemplateKey: "no-such-template",
	})
	if err == nil {
		t.Fatal("expected 400")
	}
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.Code != "bad_request" {
		t.Fatalf("want bad_request got %#v", err)
	}
	if !strings.Contains(ae.Message, "未知模板") {
		t.Fatalf("msg=%q", ae.Message)
	}
}

func TestCreateTeamFromTemplateDualTeamsNoSeatLeakR1(t *testing.T) {
	svc, ws, _, _ := setup(t)
	a, err := svc.CreateTeamFromTemplate(ws, botmem.CreateTeamInput{
		Name: "队A", Slug: "dual-a", TemplateKey: "default",
	})
	if err != nil {
		t.Fatal(err)
	}
	b, err := svc.CreateTeamFromTemplate(ws, botmem.CreateTeamInput{
		Name: "队B", Slug: "dual-b", TemplateKey: "default",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Roster) != len(b.Roster) {
		t.Fatalf("shape len A=%d B=%d", len(a.Roster), len(b.Roster))
	}
	idsA := map[string]struct{}{}
	slugsA := map[string]string{}
	for _, s := range a.Roster {
		idsA[s.ID] = struct{}{}
		slugsA[s.RoleSlug] = deref(s.PlaybookRef)
	}
	for _, s := range b.Roster {
		if _, ok := idsA[s.ID]; ok {
			t.Fatalf("seat id leaked across teams: %s", s.ID)
		}
		if s.TeamID == a.Team.ID {
			t.Fatal("team B seat has team A id")
		}
		wantRef, ok := slugsA[s.RoleSlug]
		if !ok {
			t.Fatalf("B missing slug shape %s", s.RoleSlug)
		}
		if deref(s.PlaybookRef) != wantRef {
			t.Fatalf("slug %s playbook mismatch", s.RoleSlug)
		}
	}
}

func TestEnsureRosterUnknownTemplateR2(t *testing.T) {
	svc, ws, _, _ := setup(t)
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "T"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.EnsureRoster(ws, team.ID, botmem.EnsureRosterInput{TemplateKey: "ghost"})
	if err == nil {
		t.Fatal("expected 400")
	}
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.Code != "bad_request" || !strings.Contains(ae.Message, "未知模板") {
		t.Fatalf("got %#v", err)
	}
}

func TestParsePlaybookRefIllegalR2(t *testing.T) {
	_, err := botmem.ParsePlaybookRef("{not-json")
	if err == nil {
		t.Fatal("expected error")
	}
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.Code != "bad_request" {
		t.Fatalf("got %#v", err)
	}
	_, err = botmem.ParsePlaybookRef("bot-team-template@main:../etc/passwd")
	if err == nil {
		t.Fatal("expected .. reject")
	}
	p, err := botmem.ParsePlaybookRef("bot-team-template@main:playbook/x.md")
	if err != nil {
		t.Fatal(err)
	}
	if p.Repo != "bot-team-template" || p.Path != "playbook/x.md" {
		t.Fatalf("%+v", p)
	}
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
