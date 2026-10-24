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
