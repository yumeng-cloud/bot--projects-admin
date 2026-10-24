package botclient

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestProjectPinsLeanNoUUID(t *testing.T) {
	raw := json.RawMessage(`{"pins":[{"id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","workspace_id":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","team_id":"cccccccccccccccccccccccccccccccc","key":"decision","value":"采用方案 B 并且这是一段足够长用来截断的共享记忆墙内容用来验证瘦输出截断行为是否生效","updated_by":"dddddddddddddddddddddddddddddddd","updated_at":"2026-10-24T10:58:57.179316446Z"}]}`)
	items, err := ProjectPins(raw, ViewOpts{})
	if err != nil {
		t.Fatal(err)
	}
	txt := FormatPinListText(items)
	if ContainsHexUUID(txt) {
		t.Fatalf("UUID in pin list text: %s", txt)
	}
	if !strings.Contains(txt, "decision") || !strings.Contains(txt, "采用方案 B") {
		t.Fatalf("missing key/value: %s", txt)
	}
	if strings.Contains(txt, "workspace_id") || strings.Contains(txt, "team_id") {
		t.Fatalf("fat fields leaked: %s", txt)
	}
	withID, err := ProjectPins(raw, ViewOpts{IDs: true})
	if err != nil {
		t.Fatal(err)
	}
	idTxt := FormatPinListText(withID)
	if !strings.Contains(idTxt, "id=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa") {
		t.Fatalf("--id missing: %s", idTxt)
	}

	one := json.RawMessage(`{"pin":{"id":"eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","key":"k","value":"v","updated_at":"2026-10-24T11:00:00Z"}}`)
	got, err := ProjectPins(one, ViewOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || strField(got[0], "key") != "k" {
		t.Fatalf("single pin: %#v", got)
	}
	del := json.RawMessage(`{"ok":true}`)
	ditems, err := ProjectPins(del, ViewOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(FormatPinListText(ditems), "ok=") {
		t.Fatalf("delete ok: %s", FormatPinListText(ditems))
	}
	empty, err := ProjectPins(json.RawMessage(`{"pins":[]}`), ViewOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if FormatPinListText(empty) != "(无 pin)\n" {
		t.Fatalf("empty: %q", FormatPinListText(empty))
	}
}
