package botclient

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFilterOpenIssues(t *testing.T) {
	in := []map[string]any{
		{"identifier": "YM-1", "status": "todo"},
		{"identifier": "YM-2", "status": "done"},
		{"identifier": "YM-3", "status": "in_progress"},
		{"identifier": "YM-4", "status": "backlog"},
		{"identifier": "YM-5", "status": "in_review"},
	}
	out := FilterOpenIssues(in)
	if len(out) != 3 {
		t.Fatalf("want 3 open, got %d %#v", len(out), out)
	}
}

func TestPatrolFingerprintStableAndSensitive(t *testing.T) {
	a := []map[string]any{
		{"identifier": "YM-2", "status": "todo", "updated_at": "t2"},
		{"identifier": "YM-1", "status": "todo", "updated_at": "t1"},
	}
	b := []map[string]any{
		{"identifier": "YM-1", "status": "todo", "updated_at": "t1"},
		{"identifier": "YM-2", "status": "todo", "updated_at": "t2"},
	}
	if PatrolFingerprint(a) != PatrolFingerprint(b) {
		t.Fatal("order should not matter")
	}
	c := []map[string]any{
		{"identifier": "YM-1", "status": "in_progress", "updated_at": "t1"},
		{"identifier": "YM-2", "status": "todo", "updated_at": "t2"},
	}
	if PatrolFingerprint(a) == PatrolFingerprint(c) {
		t.Fatal("status change should change fingerprint")
	}
	empty := PatrolFingerprint(nil)
	if empty != PatrolFingerprint([]map[string]any{}) {
		t.Fatal("nil and empty should match")
	}
	if empty == PatrolFingerprint(a) {
		t.Fatal("empty vs non-empty")
	}
}

func TestPatrolQuietDecision(t *testing.T) {
	emptyFP := PatrolFingerprint(nil)
	itemsFP := PatrolFingerprint([]map[string]any{
		{"identifier": "YM-9", "status": "todo", "updated_at": "u"},
	})
	if PatrolQuietDecision(false, "", itemsFP) {
		t.Fatal("non-quiet never suppress")
	}
	if !PatrolQuietDecision(true, "", emptyFP) {
		t.Fatal("first empty should suppress")
	}
	if PatrolQuietDecision(true, "", itemsFP) {
		t.Fatal("first with items should NOT suppress")
	}
	if !PatrolQuietDecision(true, itemsFP, itemsFP) {
		t.Fatal("unchanged should suppress")
	}
	if PatrolQuietDecision(true, emptyFP, itemsFP) {
		t.Fatal("new assign should NOT suppress")
	}
}

func TestPatrolStateFileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "state")
	fp := "abc123deadbeef"
	if err := WritePatrolStateFP(path, fp); err != nil {
		t.Fatal(err)
	}
	got, err := ReadPatrolStateFP(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != fp {
		t.Fatalf("got %q want %q", got, fp)
	}
	missing, err := ReadPatrolStateFP(filepath.Join(dir, "nope"))
	if err != nil || missing != "" {
		t.Fatalf("missing: %q %v", missing, err)
	}
	_ = os.WriteFile(path, []byte("# c\n\n"+fp+"\n"), 0o600)
	got, _ = ReadPatrolStateFP(path)
	if got != fp {
		t.Fatalf("comment skip: %q", got)
	}
}
