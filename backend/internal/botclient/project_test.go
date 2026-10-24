package botclient

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestProjectIssuesSlimNoUUID(t *testing.T) {
	raw := json.RawMessage(`{"issues":[{"id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","identifier":"YM-9","title":"T","status":"todo","priority":"medium","assignee_id":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","description":"LONG","version":3,"updated_at":"2026-10-24T01:02:03Z"}]}`)
	c := &Client{}
	c.ensureCaches()
	c.memberNames = map[string]string{"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb": "BotA"}
	c.membersLoaded = true
	items, err := c.ProjectIssues(raw, ViewOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("len=%d", len(items))
	}
	it := items[0]
	if it["identifier"] != "YM-9" || it["assignee"] != "BotA" {
		t.Fatalf("got %#v", it)
	}
	for _, ban := range []string{"id", "description", "version", "assignee_id", "project_id"} {
		if _, ok := it[ban]; ok {
			t.Fatalf("banned field %s present", ban)
		}
	}
	txt := FormatIssueListText(items)
	if ContainsHexUUID(txt) {
		t.Fatalf("UUID in text: %s", txt)
	}
	if !strings.Contains(txt, "YM-9") {
		t.Fatalf("missing YM-9: %s", txt)
	}
}

func TestProjectIssueGetHasVersion(t *testing.T) {
	raw := json.RawMessage(`{"issue":{"id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","identifier":"YM-9","title":"T","status":"todo","priority":"medium","version":7,"updated_at":"2026-10-24T01:02:03Z"}}`)
	c := &Client{}
	items, err := c.ProjectIssues(raw, ViewOpts{Detail: true})
	if err != nil {
		t.Fatal(err)
	}
	it := items[0]
	if _, ok := it["id"]; ok {
		t.Fatalf("id leaked in detail slim: %#v", it)
	}
	if it["version"] != float64(7) && it["version"] != 7 {
		// json.Number vs float64 depending on decoder; ProjectIssues uses map via encoding/json → float64
		v := it["version"]
		switch n := v.(type) {
		case float64:
			if n != 7 {
				t.Fatalf("version=%v", v)
			}
		case int:
			if n != 7 {
				t.Fatalf("version=%v", v)
			}
		default:
			t.Fatalf("version type %#v", v)
		}
	}
	txt := FormatIssueGetText(it)
	if ContainsHexUUID(txt) {
		t.Fatalf("UUID in get text: %s", txt)
	}
	if !strings.Contains(txt, "version=7") {
		t.Fatalf("missing version in get text: %s", txt)
	}
	if !strings.Contains(txt, "YM-9") {
		t.Fatalf("missing YM-9: %s", txt)
	}
}

func TestProjectIssuesWithIDs(t *testing.T) {
	raw := json.RawMessage(`{"issue":{"id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","identifier":"YM-1","title":"X","status":"todo","priority":"low"}}`)
	c := &Client{}
	items, err := c.ProjectIssues(raw, ViewOpts{IDs: true})
	if err != nil {
		t.Fatal(err)
	}
	if items[0]["id"] != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
		t.Fatalf("expected id with IDs: %#v", items[0])
	}
}

func TestProjectWhoAmINoUUID(t *testing.T) {
	raw := json.RawMessage(`{"bot":{"id":"cccccccccccccccccccccccccccccccc","display_name":"MemBot","workspace_id":"dddddddddddddddddddddddddddddddd","member_id":"eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"}}`)
	v := ProjectWhoAmI(raw, ViewOpts{})
	txt := FormatWhoAmIText(v)
	if ContainsHexUUID(txt) {
		t.Fatalf("UUID: %s", txt)
	}
	if _, ok := v["id"]; ok {
		t.Fatal("id leaked")
	}
}

func TestProjectMemoryHandoffComment(t *testing.T) {
	memRaw := json.RawMessage(`{"memories":[{"id":"ffffffffffffffffffffffffffffffff","key_name":"focus","value_text":"hello world this is long enough to truncate somewhere past eighty characters maybe","tags":"a,b"}]}`)
	items, err := ProjectMemories(memRaw, ViewOpts{})
	if err != nil {
		t.Fatal(err)
	}
	txt := FormatMemoryListText(items)
	if ContainsHexUUID(txt) || strings.Contains(txt, "ffffffffffffffff") {
		t.Fatalf("UUID in memory: %s", txt)
	}
	if !strings.Contains(txt, "focus") {
		t.Fatal(txt)
	}

	c := &Client{}
	c.ensureCaches()
	c.memberNames = map[string]string{"11111111111111111111111111111111": "Alice"}
	c.membersLoaded = true
	comRaw := json.RawMessage(`{"comments":[{"id":"22222222222222222222222222222222","author_id":"11111111111111111111111111111111","body":"hi there","created_at":"2026-10-24T12:00:00Z","issue_id":"33333333333333333333333333333333"}]}`)
	coms, err := c.ProjectComments(comRaw, "YM-3", ViewOpts{})
	if err != nil {
		t.Fatal(err)
	}
	ctxt := FormatCommentListText(coms)
	if ContainsHexUUID(ctxt) {
		t.Fatalf("UUID in comment: %s", ctxt)
	}
	if !strings.Contains(ctxt, "Alice") || !strings.Contains(ctxt, "YM-3") {
		t.Fatal(ctxt)
	}

	hoRaw := json.RawMessage(`{"handoff":{"claimed_role":{"name":"记忆"},"memories":[{},{}],"todos":[{}],"issues":[{"id":"44444444444444444444444444444444","identifier":"YM-7","title":"Do it","status":"todo"}]}}`)
	hv, err := c.ProjectHandoff(hoRaw, ViewOpts{})
	if err != nil {
		t.Fatal(err)
	}
	ht := FormatHandoffText(hv)
	if ContainsHexUUID(ht) {
		t.Fatalf("UUID in handoff: %s", ht)
	}
	if !strings.Contains(ht, "YM-7") {
		t.Fatal(ht)
	}
}

func TestLooksLikeIdentifier(t *testing.T) {
	if !LooksLikeIdentifier("YM-12") || LooksLikeIdentifier("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa") {
		t.Fatal("identifier detect")
	}
}

func TestProjectIssuesNestedAssigneePreferred(t *testing.T) {
	raw := json.RawMessage(`{"issues":[{"id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","identifier":"YM-2","title":"N","status":"todo","priority":"low","assignee_id":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","assignee":{"id":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","display_name":"嵌套名"},"updated_at":"2026-10-24T01:02:03Z"}]}`)
	c := &Client{}
	c.ensureCaches()
	c.memberNames = map[string]string{"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb": "Members名"}
	c.membersLoaded = true
	items, err := c.ProjectIssues(raw, ViewOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if items[0]["assignee"] != "嵌套名" {
		t.Fatalf("want nested name, got %#v", items[0]["assignee"])
	}
}

func TestProjectIssuesAssigneeFallbackMembers(t *testing.T) {
	raw := json.RawMessage(`{"issues":[{"id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","identifier":"YM-3","title":"F","status":"todo","priority":"low","assignee_id":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","updated_at":"2026-10-24T01:02:03Z"}]}`)
	c := &Client{}
	c.ensureCaches()
	c.memberNames = map[string]string{"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb": "BotFromMembers"}
	c.membersLoaded = true
	items, err := c.ProjectIssues(raw, ViewOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if items[0]["assignee"] != "BotFromMembers" {
		t.Fatalf("got %#v", items[0]["assignee"])
	}
}

func TestProjectIssuesMembersFailShowsDash(t *testing.T) {
	raw := json.RawMessage(`{"issues":[{"id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","identifier":"YM-4","title":"D","status":"todo","priority":"low","assignee_id":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}]}`)
	c := &Client{} // no HTTP, MemberNames → empty; membersLoaded false then empty
	items, err := c.ProjectIssues(raw, ViewOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if items[0]["assignee"] != "-" {
		t.Fatalf("members miss should be -, got %#v", items[0]["assignee"])
	}
	txt := FormatIssueListText(items)
	if ContainsHexUUID(txt) {
		t.Fatalf("UUID leaked on members miss: %s", txt)
	}
}

func TestProjectIssuesDropsArchivedByDefault(t *testing.T) {
	raw := json.RawMessage(`{"issues":[
		{"id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","identifier":"YM-1","title":"Open","status":"todo","priority":"low","version":1,"updated_at":"t"},
		{"id":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","identifier":"YM-2","title":"Archived","status":"done","priority":"low","version":1,"updated_at":"t","archived_at":"2026-10-24T12:00:00Z"}
	]}`)
	c := &Client{}
	items, err := c.ProjectIssues(raw, ViewOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("want 1 after drop archived, got %d %#v", len(items), items)
	}
	if items[0]["identifier"] != "YM-1" {
		t.Fatalf("got %#v", items[0])
	}

	kept, err := c.ProjectIssues(raw, ViewOpts{IncludeArchived: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(kept) != 2 {
		t.Fatalf("include want 2, got %d", len(kept))
	}

	// Detail/get：单条已归档仍投影
	one := json.RawMessage(`{"issue":{"id":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","identifier":"YM-2","title":"Archived","status":"done","priority":"low","version":1,"archived_at":"2026-10-24T12:00:00Z"}}`)
	got, err := c.ProjectIssues(one, ViewOpts{Detail: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0]["identifier"] != "YM-2" {
		t.Fatalf("detail archived: %#v", got)
	}

	filtered := DropArchivedFromIssueListJSON(raw)
	items2, err := c.ProjectIssues(filtered, ViewOpts{IncludeArchived: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(items2) != 1 || items2[0]["identifier"] != "YM-1" {
		t.Fatalf("DropArchivedFromIssueListJSON: %#v", items2)
	}
}

func TestFormatHandoffFocusEmptyExplicit(t *testing.T) {
	c := &Client{}
	// 无 memories → 明示空
	hv, err := c.ProjectHandoff(json.RawMessage(`{"handoff":{"claimed_role":{"name":"SRE","slug":"sre"},"memories":[],"todos":[],"issues":[],"playbook_ref":"docs/sre.md"}}`), ViewOpts{})
	if err != nil {
		t.Fatal(err)
	}
	ht := FormatHandoffText(hv)
	if !strings.Contains(ht, "focus=空") {
		t.Fatalf("want focus=空, got:\n%s", ht)
	}
	if strings.Contains(ht, "假 focus") || strings.Contains(ht, "流程说明") {
		t.Fatalf("must not invent second流程文: %s", ht)
	}
	// 有 seat-scoped last_focus → 展示值、不写空
	hv2, err := c.ProjectHandoff(json.RawMessage(`{"handoff":{"claimed_role":{"name":"前端"},"memories":[{"key_name":"frontend:last_focus","value_text":"AEJ-focus-marker"}],"todos":[],"issues":[]}}`), ViewOpts{})
	if err != nil {
		t.Fatal(err)
	}
	ht2 := FormatHandoffText(hv2)
	if strings.Contains(ht2, "focus=空") {
		t.Fatalf("should not be empty when last_focus present:\n%s", ht2)
	}
	if !strings.Contains(ht2, "focus=AEJ-focus-marker") {
		t.Fatalf("want focus value, got:\n%s", ht2)
	}
}

func TestFocusFromMemories(t *testing.T) {
	v, empty := focusFromMemories(nil)
	if !empty || v != "" {
		t.Fatalf("nil mems: v=%q empty=%v", v, empty)
	}
	v, empty = focusFromMemories([]any{
		map[string]any{"key_name": "decision", "value_text": "x"},
		map[string]any{"key_name": "qa:last_focus", "value_text": "seat-f"},
	})
	if empty || v != "seat-f" {
		t.Fatalf("seat scoped: v=%q empty=%v", v, empty)
	}
	v, empty = focusFromMemories([]any{
		map[string]any{"key_name": "last_focus", "value_text": "bare"},
		map[string]any{"key_name": "qa:last_focus", "value_text": "seat-f"},
	})
	if empty || v != "bare" {
		t.Fatalf("bare preferred: v=%q empty=%v", v, empty)
	}
}
