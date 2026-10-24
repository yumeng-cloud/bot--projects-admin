package botclient

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
)

func TestDefaultStdoutNoUUIDAndGetHasVersion(t *testing.T) {
	hex := regexp.MustCompile(`(?i)\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b|\b[0-9a-f]{32}\b`)
	c := &Client{}

	items, err := c.ProjectIssues(json.RawMessage(`{"issue":{"id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","identifier":"YM-9","title":"修好登录态","status":"in_progress","priority":"high","version":3,"assignee":{"display_name":"前端Bot"},"updated_at":"2026-10-24T01:02:03Z","description":"long desc"}}`), ViewOpts{Detail: true})
	if err != nil {
		t.Fatal(err)
	}
	getTxt := FormatIssueGetText(items[0])
	t.Logf("issue get sample:\n%s", getTxt)

	listItems, err := c.ProjectIssues(json.RawMessage(`{"issues":[{"id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","identifier":"YM-1","title":"A","status":"todo","priority":"low","version":1,"updated_at":"t"}]}`), ViewOpts{})
	if err != nil {
		t.Fatal(err)
	}
	listTxt := FormatIssueListText(listItems)

	whoTxt := FormatWhoAmIText(ProjectWhoAmI(json.RawMessage(`{"bot":{"id":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","name":"前端Bot","role_key":"fe"}}`), ViewOpts{}))

	mem, err := ProjectMemories(json.RawMessage(`{"memories":[{"id":"cccccccccccccccccccccccccccccccc","key_name":"k1","value_text":"v1-long-value-here","updated_at":"t"}]}`), ViewOpts{})
	if err != nil {
		t.Fatal(err)
	}
	memTxt := FormatMemoryListText(mem)

	com, err := c.ProjectComments(json.RawMessage(`{"comments":[{"id":"dddddddddddddddddddddddddddddddd","body":"hi","created_at":"t","author":{"display_name":"Bot"}}]}`), "YM-9", ViewOpts{})
	if err != nil {
		t.Fatal(err)
	}
	comTxt := FormatCommentListText(com)

	ho, err := c.ProjectHandoff(json.RawMessage(`{"role_key":"fe","issues":[{"identifier":"YM-9","title":"T","status":"todo"}],"memories":[{"key_name":"k"}],"todos":[]}`), ViewOpts{})
	if err != nil {
		t.Fatal(err)
	}
	hoTxt := FormatHandoffText(ho)

	all := strings.Join([]string{getTxt, listTxt, whoTxt, memTxt, comTxt, hoTxt}, "\n")
	if m := hex.FindAllString(all, -1); len(m) != 0 {
		t.Fatalf("UUID matches=%d: %v\nall:\n%s", len(m), m, all)
	}
	if !strings.Contains(getTxt, "assignee=前端Bot") {
		t.Fatalf("get missing nested assignee name: %s", getTxt)
	}
	if !strings.Contains(getTxt, "version=3") {
		t.Fatalf("get missing version: %s", getTxt)
	}
	if strings.Contains(listTxt, "version=") {
		t.Fatalf("list should not have version: %s", listTxt)
	}
}
