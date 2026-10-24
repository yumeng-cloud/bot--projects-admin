package botclient

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestProjectTeamTemplatesSlim(t *testing.T) {
	raw := json.RawMessage(`{"templates":[{
		"key":"default","name":"默认班底","description":"五席最小班底",
		"seats":[
			{"slug":"frontend","display_name":"前端","playbook_ref":"r@main:p"},
			{"slug":"pm","display_name":"产品经理","playbook_ref":"r@main:p"},
			{"slug":"qa","display_name":"QA","playbook_ref":"r@main:p"},
			{"slug":"backend","display_name":"后端","playbook_ref":"r@main:p"},
			{"slug":"wiki","display_name":"知识库","playbook_ref":"r@main:p"}
		]
	}]}`)
	items, err := ProjectTeamTemplates(raw, ViewOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("len=%d", len(items))
	}
	if items[0]["key"] != "default" || items[0]["seats_n"] != 5 {
		t.Fatalf("%#v", items[0])
	}
	txt := FormatTeamTemplatesText(items)
	if ContainsHexUUID(txt) {
		t.Fatalf("UUID leak: %s", txt)
	}
	if !strings.Contains(txt, "default") || !strings.Contains(txt, "frontend") {
		t.Fatalf("text=%s", txt)
	}
}

func TestProjectCreateTeamFromTemplateSlim(t *testing.T) {
	raw := json.RawMessage(`{
		"team":{"id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","name":"TmplSmoke","slug":"tmpl-smoke","template_key":"default"},
		"roster":[
			{"id":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","role_slug":"frontend","display_name":"前端","claim":null},
			{"id":"cccccccccccccccccccccccccccccccc","role_slug":"pm","display_name":"产品经理","claim":null}
		]
	}`)
	v, err := ProjectCreateTeamFromTemplate(raw, ViewOpts{})
	if err != nil {
		t.Fatal(err)
	}
	team, _ := v["team"].(map[string]any)
	if team["name"] != "TmplSmoke" || team["template"] != "default" {
		t.Fatalf("team %#v", team)
	}
	if _, ok := team["id"]; ok {
		t.Fatalf("id leaked without --id: %#v", team)
	}
	if v["seats_n"] != 2 {
		t.Fatalf("seats_n=%v", v["seats_n"])
	}
	txt := FormatCreateTeamFromTemplateText(v)
	if ContainsHexUUID(txt) {
		t.Fatalf("UUID in text: %s", txt)
	}
	if !strings.Contains(txt, "created name=TmplSmoke") || !strings.Contains(txt, "frontend") {
		t.Fatalf("text=%s", txt)
	}

	v2, err := ProjectCreateTeamFromTemplate(raw, ViewOpts{IDs: true})
	if err != nil {
		t.Fatal(err)
	}
	team2, _ := v2["team"].(map[string]any)
	if team2["id"] == "" {
		t.Fatalf("expected id with --id: %#v", team2)
	}
}
