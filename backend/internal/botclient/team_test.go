package botclient

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSeatMemoryKey(t *testing.T) {
	if got := SeatMemoryKey("frontend", "last_focus"); got != "frontend:last_focus" {
		t.Fatalf("got %q", got)
	}
	if got := SeatMemoryKey("", "last_focus"); got != "last_focus" {
		t.Fatalf("empty seat: %q", got)
	}
}

func TestProjectTeamRosterSlimNoUUID(t *testing.T) {
	raw := json.RawMessage(`{"roster":[
		{"id":"aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee","role_slug":"frontend","display_name":"前端","playbook_ref":"playbook/frontend.md","claim":null},
		{"id":"11111111-2222-3333-4444-555555555555","role_slug":"pm","display_name":"产品经理","bound_member_id":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","bound_display_name":"BotB","claim":{"member_id":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","claimed_at":"2026-10-24T01:00:00Z"}}
	]}`)
	items, err := ProjectTeamRoster(raw, ViewOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("len=%d", len(items))
	}
	if items[0]["slug"] != "frontend" || items[0]["claim"] != "empty" {
		t.Fatalf("seat0 %#v", items[0])
	}
	if items[1]["slug"] != "pm" || items[1]["claim"] != "claimed" {
		t.Fatalf("seat1 %#v", items[1])
	}
	if items[1]["holder"] != "BotB" {
		t.Fatalf("holder %#v", items[1])
	}
	for _, it := range items {
		if _, ok := it["id"]; ok {
			t.Fatalf("id leaked: %#v", it)
		}
	}
	txt := FormatTeamRosterText(items)
	if ContainsHexUUID(txt) {
		t.Fatalf("UUID in roster text: %s", txt)
	}
	if !strings.Contains(txt, "frontend") || !strings.Contains(txt, "seats=2") {
		t.Fatalf("text=%s", txt)
	}
}

func TestProjectTeamRosterThreeStateOccupancy(t *testing.T) {
	raw := json.RawMessage(`{"roster":[
		{"role_slug":"frontend","display_name":"前端","status":"active","bound_member_id":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","bound_display_name":"BotFE","claim":{"member_id":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","claimed_at":"2026-10-24T01:00:00Z"}},
		{"role_slug":"backend","display_name":"后端","status":"open","claim":null},
		{"role_slug":"pm","display_name":"产品","status":"stopped","claim":null},
		{"role_slug":"qa","display_name":"QA","status":"left","claim":null},
		{"role_slug":"wiki","display_name":"知识库","status":"departed","claim":null}
	]}`)
	items, err := ProjectTeamRoster(raw, ViewOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 5 {
		t.Fatalf("len=%d", len(items))
	}
	if items[0]["claim"] != "claimed" || items[0]["holder"] != "BotFE" {
		t.Fatalf("active %#v", items[0])
	}
	if items[1]["claim"] != "empty" {
		t.Fatalf("open never-bound %#v", items[1])
	}
	for i, want := range []string{"refillable", "refillable", "refillable"} {
		if items[i+2]["claim"] != want {
			t.Fatalf("seat%d claim=%v want %s", i+2, items[i+2]["claim"], want)
		}
	}
	txt := FormatTeamRosterText(items)
	if !strings.Contains(txt, "占用中·BotFE") {
		t.Fatalf("missing active label: %s", txt)
	}
	if !strings.Contains(txt, "空闲/未进席") {
		t.Fatalf("missing empty label: %s", txt)
	}
	if !strings.Contains(txt, "停工可补") {
		t.Fatalf("missing refillable label: %s", txt)
	}
	// 禁：停工席被说成 empty/空闲 字面机态
	if strings.Contains(txt, "\tempty\t") || strings.Contains(txt, "\tclaimed\t") {
		t.Fatalf("raw machine claim leaked into human text: %s", txt)
	}
	hint := FormatUnboundBotsHint([]string{"产品体验", "QA"})
	if !strings.Contains(hint, "2 个 Bot 尚未占本队席") || !strings.Contains(hint, "产品体验") {
		t.Fatalf("hint=%q", hint)
	}
	if FormatUnboundBotsHint(nil) != "" {
		t.Fatal("empty hint")
	}
}

func TestProjectTeamRosterIdempotentShape(t *testing.T) {
	// 两次 ensure 同形：同 slug 不双建（投影侧用 slug 去重检查辅助）
	raw := json.RawMessage(`{"seats":[
		{"role_slug":"frontend","display_name":"前端"},
		{"role_slug":"pm","display_name":"PM"},
		{"role_slug":"qa","display_name":"QA"},
		{"role_slug":"backend","display_name":"后端"},
		{"role_slug":"wiki","display_name":"知识库"}
	]}`)
	a, _ := ProjectTeamRoster(raw, ViewOpts{})
	b, _ := ProjectTeamRoster(raw, ViewOpts{})
	if len(a) != len(b) || len(a) != 5 {
		t.Fatalf("len a=%d b=%d", len(a), len(b))
	}
	seen := map[string]int{}
	for _, it := range a {
		seen[strField(it, "slug")]++
	}
	for s, n := range seen {
		if n != 1 {
			t.Fatalf("duplicate slug %s n=%d", s, n)
		}
	}
}

func TestGuardExpandNoStealInProgress(t *testing.T) {
	if err := GuardExpandNoStealInProgress("todo", "other", "me"); err != nil {
		t.Fatal(err)
	}
	if err := GuardExpandNoStealInProgress("in_progress", "", "me"); err != nil {
		t.Fatal(err)
	}
	if err := GuardExpandNoStealInProgress("doing", "me", "me"); err != nil {
		t.Fatal(err)
	}
	err := GuardExpandNoStealInProgress("in_progress", "other-member", "me")
	if err == nil {
		t.Fatal("expected conflict")
	}
	ae, ok := err.(*APIError)
	if !ok || ae.Status != 409 {
		t.Fatalf("want 409 got %v", err)
	}
	if !strings.Contains(ae.Error(), "进行中") {
		t.Fatalf("msg=%s", ae.Error())
	}
}

func TestProjectTeamExpandDryRun(t *testing.T) {
	raw := json.RawMessage(`{"dry_run":true,"note":"expand API 未就绪；dry-run 仅列建议","suggestions":[{"action":"find_idle_seat","rule":"仅空闲"}]}`)
	v, err := ProjectTeamExpand(raw, ViewOpts{})
	if err != nil {
		t.Fatal(err)
	}
	txt := FormatTeamExpandText(v)
	if ContainsHexUUID(txt) {
		t.Fatalf("UUID: %s", txt)
	}
	if !strings.Contains(txt, "dry-run") {
		t.Fatalf("text=%s", txt)
	}
}

func TestProjectTeamExpandAssignedSkipped(t *testing.T) {
	raw := json.RawMessage(`{"assigned":[{"issue_id":"aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee","identifier":"YM-12","seat_id":"s1","member_id":"m1"}],"skipped":[{"issue_id":"bbbbbbbb-cccc-dddd-eeee-ffffffffffff","reason":"已有负责人"}]}`)
	v, err := ProjectTeamExpand(raw, ViewOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if v["assigned_n"] != 1 || v["skipped_n"] != 1 {
		t.Fatalf("%#v", v)
	}
	txt := FormatTeamExpandText(v)
	if ContainsHexUUID(txt) {
		t.Fatalf("UUID: %s", txt)
	}
	if !strings.Contains(txt, "YM-12") || !strings.Contains(txt, "已有负责人") {
		t.Fatalf("text=%s", txt)
	}
}

func TestFormatSeatEnterTextNoUUID(t *testing.T) {
	v := map[string]any{
		"role":            "frontend",
		"step":            "whoami→handoff get→seat bind",
		"whoami":          map[string]any{"display_name": "BotA", "actor_kind": "bot"},
		"handoff":         map[string]any{"claimed_role": "frontend", "memory_count": 1, "todo_count": 0, "issue_count": 0},
		"claim":           map[string]any{"ok": true, "claimed_at": "2026-10-24T01:02:03Z"},
		"memory_key_hint": "frontend:last_focus",
	}
	txt := FormatSeatEnterText(v)
	if ContainsHexUUID(txt) {
		t.Fatalf("UUID: %s", txt)
	}
	if !strings.Contains(txt, "frontend:last_focus") {
		t.Fatalf("missing seat key: %s", txt)
	}
}

func TestDefaultTeamSeatSlugs(t *testing.T) {
	if len(DefaultTeamSeatSlugs) != 5 {
		t.Fatalf("want 5 got %d", len(DefaultTeamSeatSlugs))
	}
}
