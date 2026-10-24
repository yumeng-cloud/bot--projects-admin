package botclient

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestProjectPlaybookRefOK(t *testing.T) {
	ref, warn := ProjectPlaybookRef("bot-team-template@main:roles/前端研发/")
	if warn != "" || !strings.Contains(ref, "roles/前端研发") {
		t.Fatalf("ref=%q warn=%q", ref, warn)
	}
	ref, warn = ProjectPlaybookRef(`{"repo":"bot-team-template","ref":"main","path":"playbook/班底与补位.md"}`)
	if warn != "" || !strings.Contains(ref, "bot-team-template@main:") {
		t.Fatalf("json ref=%q warn=%q", ref, warn)
	}
	ref, warn = ProjectPlaybookRef("playbook/frontend.md")
	if warn != "" || ref == "" {
		t.Fatalf("path ref=%q warn=%q", ref, warn)
	}
}

func TestProjectPlaybookRefEmptyNoFallback(t *testing.T) {
	ref, warn := ProjectPlaybookRef("")
	if ref != "" || warn != PlaybookRefEmptyWarn {
		t.Fatalf("ref=%q warn=%q", ref, warn)
	}
	if strings.Contains(warn, "playbook/frontend") || strings.Contains(ref, "playbook/") {
		t.Fatal("must not fall back to hardcoded playbook")
	}
}

func TestProjectPlaybookRefIllegalNoFallback(t *testing.T) {
	cases := []string{
		"no-slash-no-md",
		"repo@main:", // empty path
		`{"repo":"r","ref":"main","path":"../etc"}`,
		"bot-team-template@main:../secret",
	}
	for _, c := range cases {
		ref, warn := ProjectPlaybookRef(c)
		if warn != PlaybookRefBadWarn {
			t.Fatalf("case %q: warn=%q ref=%q", c, warn, ref)
		}
		if strings.Contains(ref, "playbook/frontend.md") {
			t.Fatalf("fallback in %q", c)
		}
	}
}

func TestAttachPlaybookRefNoProcessField(t *testing.T) {
	dst := map[string]any{}
	AttachPlaybookRef(dst, "bot-team-template@main:roles/QA/")
	if _, ok := dst["流程说明"]; ok {
		t.Fatal("must not add 流程说明 field")
	}
	if dst["playbook_ref"] == nil {
		t.Fatalf("%#v", dst)
	}
	dst2 := map[string]any{}
	AttachPlaybookRef(dst2, "")
	if dst2["playbook_warn"] != PlaybookRefEmptyWarn {
		t.Fatalf("%#v", dst2)
	}
	if _, ok := dst2["playbook_ref"]; ok {
		t.Fatal("empty must not invent playbook_ref")
	}
}

func TestProjectTeamRosterPlaybookRefField(t *testing.T) {
	raw := json.RawMessage(`{"roster":[
		{"id":"aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee","role_slug":"frontend","display_name":"前端","playbook_ref":"bot-team-template@main:roles/前端研发/","claim":null},
		{"id":"11111111-2222-3333-4444-555555555555","role_slug":"ops","display_name":"运维","playbook_ref":"","claim":null}
	]}`)
	items, err := ProjectTeamRoster(raw, ViewOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if items[0]["playbook_ref"] == nil || strings.Contains(strField(items[0], "playbook_ref"), "流程") {
		t.Fatalf("seat0 %#v", items[0])
	}
	if _, ok := items[0]["playbook"]; ok {
		t.Fatal("legacy playbook key should not compete; use playbook_ref")
	}
	if items[1]["playbook_warn"] != PlaybookRefEmptyWarn {
		t.Fatalf("empty seat warn %#v", items[1])
	}
	txt := FormatTeamRosterText(items)
	if !strings.Contains(txt, "roles/前端研发") {
		t.Fatalf("text missing handbook path: %s", txt)
	}
	if !strings.Contains(txt, "playbook_ref 为空") {
		t.Fatalf("text missing empty warn: %s", txt)
	}
}

func TestFormatSeatEnterTextPlaybookRef(t *testing.T) {
	v := map[string]any{
		"role":            "frontend",
		"step":            "whoami→handoff get→seat bind",
		"whoami":          map[string]any{"display_name": "BotA", "actor_kind": "bot"},
		"handoff":         map[string]any{"claimed_role": "frontend", "memory_count": 0, "todo_count": 0, "issue_count": 0},
		"claim":           map[string]any{"ok": true, "via": "bind"},
		"playbook_ref":    "bot-team-template@main:roles/前端研发/",
		"memory_key_hint": "frontend:last_focus",
	}
	txt := FormatSeatEnterText(v)
	if !strings.Contains(txt, "手册 playbook_ref=") || !strings.Contains(txt, "roles/前端研发") {
		t.Fatalf("missing handbook: %s", txt)
	}
	if strings.Contains(txt, "流程说明") {
		t.Fatalf("second process field: %s", txt)
	}
	v2 := map[string]any{"role": "ops", "step": "x", "playbook_warn": PlaybookRefEmptyWarn}
	txt2 := FormatSeatEnterText(v2)
	if !strings.Contains(txt2, "playbook_ref 为空") {
		t.Fatalf("warn missing: %s", txt2)
	}
}
