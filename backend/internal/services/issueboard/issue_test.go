package issueboard_test

import (
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/models"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/issueboard"
)

func setup(t *testing.T) (*issueboard.Service, string, string) {
	t.Helper()
	dsn := filepath.Join(t.TempDir(), "t.db")
	sqlDB, err := db.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	svc := issueboard.New(sqlDB)
	wsID := idgen.New()
	_, err = sqlDB.Exec(`INSERT INTO workspaces (id, name, slug) VALUES (?,?,?)`, wsID, "T", "t-"+wsID[:8])
	if err != nil {
		t.Fatal(err)
	}
	proj, err := svc.CreateProject(wsID, issueboard.CreateProjectInput{Name: "P", KeyPrefix: "TST"})
	if err != nil {
		t.Fatal(err)
	}
	return svc, wsID, proj.ID
}

func TestIdentifierAllocationUnique(t *testing.T) {
	svc, ws, pid := setup(t)
	var wg sync.WaitGroup
	n := 20
	ids := make(chan string, n)
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			iss, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{ProjectID: pid, Title: "x", ActorKind: "human"})
			if err != nil {
				errs <- err
				return
			}
			ids <- iss.Identifier
		}()
	}
	wg.Wait()
	close(ids)
	close(errs)
	for err := range errs {
		t.Fatalf("create: %v", err)
	}
	seen := map[string]bool{}
	for id := range ids {
		if seen[id] {
			t.Fatalf("duplicate identifier %s", id)
		}
		seen[id] = true
	}
	if len(seen) != n {
		t.Fatalf("got %d unique, want %d", len(seen), n)
	}
}

func TestOptimisticLockConflict(t *testing.T) {
	svc, ws, pid := setup(t)
	iss, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{ProjectID: pid, Title: "lock", ActorKind: "human"})
	if err != nil {
		t.Fatal(err)
	}
	title := "ok"
	_, err = svc.UpdateIssue(ws, iss.ID, issueboard.UpdateIssueInput{Version: iss.Version, Title: &title, ActorKind: "human"})
	if err != nil {
		t.Fatal(err)
	}
	bad := "stale"
	_, err = svc.UpdateIssue(ws, iss.ID, issueboard.UpdateIssueInput{Version: iss.Version, Title: &bad, ActorKind: "human"})
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.StatusCode != 409 {
		t.Fatalf("want 409 conflict, got %v", err)
	}
}

func TestBotCannotSetDone(t *testing.T) {
	svc, ws, pid := setup(t)
	// Bot cannot create orphan issues (A7); create as human then bot tries done.
	iss, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{ProjectID: pid, Title: "bot", ActorKind: "human"})
	if err != nil {
		t.Fatal(err)
	}
	done := "done"
	_, err = svc.UpdateIssue(ws, iss.ID, issueboard.UpdateIssueInput{Version: iss.Version, Status: &done, ActorKind: "bot"})
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.StatusCode != 403 {
		t.Fatalf("want 403, got %v", err)
	}
}

func insertMember(t *testing.T, svc *issueboard.Service, wsID string) string {
	return insertMemberNamed(t, svc, wsID, "Tester")
}

func insertMemberNamed(t *testing.T, svc *issueboard.Service, wsID, displayName string) string {
	t.Helper()
	mid := idgen.New()
	_, err := svc.DB.Exec(`INSERT INTO members (id, workspace_id, user_id, role, display_name, is_bot, bot_enabled)
		VALUES (?,?,NULL,'member',?,0,1)`, mid, wsID, displayName)
	if err != nil {
		t.Fatal(err)
	}
	return mid
}

func TestRelationCycleDetection(t *testing.T) {
	svc, ws, pid := setup(t)
	a, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{ProjectID: pid, Title: "A", ActorKind: "human"})
	if err != nil {
		t.Fatal(err)
	}
	b, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{ProjectID: pid, Title: "B", ActorKind: "human"})
	if err != nil {
		t.Fatal(err)
	}
	c, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{ProjectID: pid, Title: "C", ActorKind: "human"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.CreateRelation(ws, a.ID, issueboard.CreateRelationInput{
		ToIssueID: b.ID, RelationType: "blocks", ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.CreateRelation(ws, b.ID, issueboard.CreateRelationInput{
		ToIssueID: c.ID, RelationType: "blocks", ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.CreateRelation(ws, c.ID, issueboard.CreateRelationInput{
		ToIssueID: a.ID, RelationType: "blocks", ActorKind: "human",
	})
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.StatusCode != 409 {
		t.Fatalf("want 409 cycle, got %v", err)
	}
	if !strings.Contains(ae.Message, "依赖成环") {
		t.Fatalf("message should contain 依赖成环, got %q", ae.Message)
	}

	// blocked_by is reverse of blocks: A blocks B ≡ B blocked_by A
	_, err = svc.CreateRelation(ws, a.ID, issueboard.CreateRelationInput{
		ToIssueID: c.ID, RelationType: "blocked_by", ActorKind: "human",
	})
	ae, ok = err.(*apperr.AppError)
	if !ok || ae.StatusCode != 409 {
		t.Fatalf("want 409 via blocked_by reverse, got %v", err)
	}

	_, err = svc.CreateRelation(ws, a.ID, issueboard.CreateRelationInput{
		ToIssueID: a.ID, RelationType: "related", ActorKind: "human",
	})
	if err == nil {
		t.Fatal("expected self-loop rejection")
	}
}

func TestRelationDuplicateConflict(t *testing.T) {
	svc, ws, pid := setup(t)
	a, _ := svc.CreateIssue(ws, issueboard.CreateIssueInput{ProjectID: pid, Title: "A", ActorKind: "human"})
	b, _ := svc.CreateIssue(ws, issueboard.CreateIssueInput{ProjectID: pid, Title: "B", ActorKind: "human"})
	_, err := svc.CreateRelation(ws, a.ID, issueboard.CreateRelationInput{
		ToIssueID: b.ID, RelationType: "related", ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.CreateRelation(ws, a.ID, issueboard.CreateRelationInput{
		ToIssueID: b.ID, RelationType: "related", ActorKind: "human",
	})
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.StatusCode != 409 {
		t.Fatalf("want 409 duplicate, got %v", err)
	}
}

func TestCommentCreate(t *testing.T) {
	svc, ws, pid := setup(t)
	mid := insertMemberNamed(t, svc, ws, "Alice")
	iss, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{ProjectID: pid, Title: "cmt", ActorKind: "human", ActorID: mid})
	if err != nil {
		t.Fatal(err)
	}
	body := "hello force audit readable"
	c, err := svc.CreateComment(ws, iss.ID, issueboard.CreateCommentInput{
		Body: body, AuthorID: mid, ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	if c.Body != body || c.AuthorID != mid || c.IssueID != iss.ID {
		t.Fatalf("unexpected comment: %+v", c)
	}
	if c.Author == nil || c.Author.ID != mid || c.Author.DisplayName != "Alice" {
		t.Fatalf("create author want Alice, got %+v", c.Author)
	}
	list, err := svc.ListComments(ws, iss.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("want 1 comment, got %d", len(list))
	}
	if list[0].Author == nil || list[0].Author.DisplayName != "Alice" {
		t.Fatalf("ListComments author.display_name want Alice, got %+v", list[0].Author)
	}
	acts, err := svc.ListIssueActivities(ws, iss.ID)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, a := range acts {
		if a.Action == "comment.created" {
			found = true
			if !strings.Contains(a.Payload, body) {
				t.Fatalf("comment.created payload missing body text: %s", a.Payload)
			}
			for _, key := range []string{`"body"`, `"summary"`, `"comment"`} {
				if !strings.Contains(a.Payload, key) {
					t.Fatalf("comment.created payload missing %s: %s", key, a.Payload)
				}
			}
			break
		}
	}
	if !found {
		t.Fatal("expected comment.created activity")
	}
}

func TestOptimisticLockConflictCode(t *testing.T) {
	svc, ws, pid := setup(t)
	iss, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{ProjectID: pid, Title: "lock2", ActorKind: "human"})
	if err != nil {
		t.Fatal(err)
	}
	title := "ok"
	_, err = svc.UpdateIssue(ws, iss.ID, issueboard.UpdateIssueInput{Version: iss.Version, Title: &title, ActorKind: "human"})
	if err != nil {
		t.Fatal(err)
	}
	bad := "stale"
	_, err = svc.UpdateIssue(ws, iss.ID, issueboard.UpdateIssueInput{Version: iss.Version, Title: &bad, ActorKind: "human"})
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.StatusCode != 409 {
		t.Fatalf("want 409, got %v", err)
	}
	if ae.Code != "version_conflict" {
		t.Fatalf("want code version_conflict, got %q", ae.Code)
	}
}

func TestConcurrentUpdateIssueSameVersion(t *testing.T) {
	svc, ws, pid := setup(t)
	iss, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{ProjectID: pid, Title: "race", ActorKind: "human"})
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	titles := []string{"one", "two"}
	for _, title := range titles {
		wg.Add(1)
		go func(title string) {
			defer wg.Done()
			tcopy := title
			_, err := svc.UpdateIssue(ws, iss.ID, issueboard.UpdateIssueInput{
				Version: iss.Version, Title: &tcopy, ActorKind: "human",
			})
			errs <- err
		}(title)
	}
	wg.Wait()
	close(errs)
	var okN, conflictN int
	for err := range errs {
		if err == nil {
			okN++
			continue
		}
		ae, ok := err.(*apperr.AppError)
		if !ok || ae.StatusCode != 409 || ae.Code != "version_conflict" {
			t.Fatalf("unexpected err: %v", err)
		}
		conflictN++
	}
	if okN != 1 || conflictN != 1 {
		t.Fatalf("want exactly 1 success and 1 version_conflict, got ok=%d conflict=%d", okN, conflictN)
	}
}

func TestStaleBoardOrderAfterStatusChange(t *testing.T) {
	svc, ws, pid := setup(t)
	iss, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{ProjectID: pid, Title: "board", ActorKind: "human"})
	if err != nil {
		t.Fatal(err)
	}
	status := "todo"
	updated, err := svc.UpdateIssue(ws, iss.ID, issueboard.UpdateIssueInput{
		Version: iss.Version, Status: &status, ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Version != iss.Version+1 {
		t.Fatalf("version: got %d want %d", updated.Version, iss.Version+1)
	}
	// Stale client still holds old version and tries to reorder.
	order := 42.0
	_, err = svc.UpdateIssue(ws, iss.ID, issueboard.UpdateIssueInput{
		Version: iss.Version, BoardOrder: &order, ActorKind: "human",
	})
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.StatusCode != 409 || ae.Code != "version_conflict" {
		t.Fatalf("want version_conflict, got %v", err)
	}
}

func TestConcurrentCreateRelationCycle(t *testing.T) {
	svc, ws, pid := setup(t)
	a, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{ProjectID: pid, Title: "A", ActorKind: "human"})
	if err != nil {
		t.Fatal(err)
	}
	b, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{ProjectID: pid, Title: "B", ActorKind: "human"})
	if err != nil {
		t.Fatal(err)
	}
	// Concurrent A→B and B→A (blocks). With tx-internal cycle check, at most one cycle-forming
	// pair can commit; under SQLite MaxOpenConns(1) the second sees the first.
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, err := svc.CreateRelation(ws, a.ID, issueboard.CreateRelationInput{
			ToIssueID: b.ID, RelationType: "blocks", ActorKind: "human",
		})
		errs <- err
	}()
	go func() {
		defer wg.Done()
		_, err := svc.CreateRelation(ws, b.ID, issueboard.CreateRelationInput{
			ToIssueID: a.ID, RelationType: "blocks", ActorKind: "human",
		})
		errs <- err
	}()
	wg.Wait()
	close(errs)
	var okN, cycleN int
	for err := range errs {
		if err == nil {
			okN++
			continue
		}
		ae, ok := err.(*apperr.AppError)
		if !ok || ae.StatusCode != 409 {
			t.Fatalf("unexpected err: %v", err)
		}
		if !strings.Contains(ae.Message, "依赖成环") && ae.Code != "conflict" {
			t.Fatalf("want cycle conflict, got %v", err)
		}
		cycleN++
	}
	if okN != 1 || cycleN != 1 {
		t.Fatalf("want 1 success + 1 cycle conflict, got ok=%d cycle=%d", okN, cycleN)
	}
	// Graph must not contain a cycle: exactly one of the two edges.
	relsA, err := svc.ListRelations(ws, a.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(relsA) != 1 {
		t.Fatalf("want exactly 1 relation involving A, got %d", len(relsA))
	}
}

func TestCommentOptimisticLock(t *testing.T) {
	svc, ws, pid := setup(t)
	mid := insertMember(t, svc, ws)
	iss, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{ProjectID: pid, Title: "cmt2", ActorKind: "human", ActorID: mid})
	if err != nil {
		t.Fatal(err)
	}
	c, err := svc.CreateComment(ws, iss.ID, issueboard.CreateCommentInput{
		Body: "v1", AuthorID: mid, ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	// First update with expected updated_at succeeds.
	u1, err := svc.UpdateComment(ws, iss.ID, c.ID, issueboard.UpdateCommentInput{
		Body: "v2", ExpectedUpdatedAt: c.UpdatedAt, ActorID: mid, ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	// Stale updated_at → version_conflict.
	_, err = svc.UpdateComment(ws, iss.ID, c.ID, issueboard.UpdateCommentInput{
		Body: "v3", ExpectedUpdatedAt: c.UpdatedAt, ActorID: mid, ActorKind: "human",
	})
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.StatusCode != 409 || ae.Code != "version_conflict" {
		t.Fatalf("want version_conflict, got %v", err)
	}
	if !strings.Contains(ae.Message, "评论已被他人更新") {
		t.Fatalf("message: %q", ae.Message)
	}
	// Omitted ExpectedUpdatedAt keeps last-write-wins.
	u2, err := svc.UpdateComment(ws, iss.ID, c.ID, issueboard.UpdateCommentInput{
		Body: "v4", ActorID: mid, ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	if u2.Body != "v4" || u1.Body != "v2" {
		t.Fatalf("unexpected bodies u1=%q u2=%q", u1.Body, u2.Body)
	}
}

func TestCommentConcurrentEdit(t *testing.T) {
	svc, ws, pid := setup(t)
	mid := insertMember(t, svc, ws)
	iss, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{ProjectID: pid, Title: "cmt3", ActorKind: "human", ActorID: mid})
	if err != nil {
		t.Fatal(err)
	}
	c, err := svc.CreateComment(ws, iss.ID, issueboard.CreateCommentInput{
		Body: "base", AuthorID: mid, ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	bodies := []string{"a", "b"}
	for _, body := range bodies {
		wg.Add(1)
		go func(body string) {
			defer wg.Done()
			_, err := svc.UpdateComment(ws, iss.ID, c.ID, issueboard.UpdateCommentInput{
				Body: body, ExpectedUpdatedAt: c.UpdatedAt, ActorID: mid, ActorKind: "human",
			})
			errs <- err
		}(body)
	}
	wg.Wait()
	close(errs)
	var okN, conflictN int
	for err := range errs {
		if err == nil {
			okN++
			continue
		}
		ae, ok := err.(*apperr.AppError)
		if !ok || ae.Code != "version_conflict" {
			t.Fatalf("unexpected err: %v", err)
		}
		conflictN++
	}
	if okN != 1 || conflictN != 1 {
		t.Fatalf("want 1 success + 1 version_conflict, got ok=%d conflict=%d", okN, conflictN)
	}
}

func TestIssueDatesCreateAndPatch(t *testing.T) {
	svc, ws, pid := setup(t)
	start := "2026-10-01"
	due := "2026-10-15"
	iss, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, Title: "dated", ActorKind: "human",
		StartAt: &start, DueAt: &due,
	})
	if err != nil {
		t.Fatal(err)
	}
	if iss.StartAt == nil || *iss.StartAt != start {
		t.Fatalf("start_at: %+v", iss.StartAt)
	}
	if iss.DueAt == nil || *iss.DueAt != due {
		t.Fatalf("due_at: %+v", iss.DueAt)
	}

	newDue := "2026-10-20"
	upd, err := svc.UpdateIssue(ws, iss.ID, issueboard.UpdateIssueInput{
		Version: iss.Version, DueAt: &newDue, ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	if upd.DueAt == nil || *upd.DueAt != newDue {
		t.Fatalf("patched due: %+v", upd.DueAt)
	}

	// null clears
	cleared, err := svc.UpdateIssue(ws, iss.ID, issueboard.UpdateIssueInput{
		Version: upd.Version, ClearStartAt: true, ClearDueAt: true, ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cleared.StartAt != nil || cleared.DueAt != nil {
		t.Fatalf("expected cleared dates, got start=%v due=%v", cleared.StartAt, cleared.DueAt)
	}
}

func TestIssueDatesRejectBadFormatAndRange(t *testing.T) {
	svc, ws, pid := setup(t)
	bad := "2026/10/01"
	_, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, Title: "bad", ActorKind: "human", StartAt: &bad,
	})
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.StatusCode != 400 {
		t.Fatalf("want 400 bad format, got %v", err)
	}

	start := "2026-10-15"
	due := "2026-10-01"
	_, err = svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, Title: "range", ActorKind: "human", StartAt: &start, DueAt: &due,
	})
	ae, ok = err.(*apperr.AppError)
	if !ok || ae.StatusCode != 400 {
		t.Fatalf("want 400 due<start, got %v", err)
	}
	if !strings.Contains(ae.Message, "due_at") && !strings.Contains(ae.Message, "开始") {
		t.Fatalf("message: %q", ae.Message)
	}

	iss, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, Title: "ok", ActorKind: "human", StartAt: &start,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.UpdateIssue(ws, iss.ID, issueboard.UpdateIssueInput{
		Version: iss.Version, DueAt: &due, ActorKind: "human",
	})
	ae, ok = err.(*apperr.AppError)
	if !ok || ae.StatusCode != 400 {
		t.Fatalf("want 400 patch due<start, got %v", err)
	}
}

func TestIsLeafChildCountTwoLevel(t *testing.T) {
	svc, ws, pid := setup(t)
	parent, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{ProjectID: pid, Title: "P", ActorKind: "human"})
	if err != nil {
		t.Fatal(err)
	}
	if !parent.IsLeaf || parent.ChildCount != 0 {
		t.Fatalf("new parent should be leaf: is_leaf=%v child_count=%d", parent.IsLeaf, parent.ChildCount)
	}
	pidPtr := parent.ID
	c1, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, Title: "C1", ParentID: &pidPtr, ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	c2, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, Title: "C2", ParentID: &pidPtr, ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !c1.IsLeaf || !c2.IsLeaf {
		t.Fatal("children should be leaves")
	}
	got, err := svc.GetIssue(ws, parent.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.IsLeaf || got.ChildCount != 2 {
		t.Fatalf("parent: is_leaf=%v child_count=%d want false/2", got.IsLeaf, got.ChildCount)
	}
	list, err := svc.ListIssues(ws, issueboard.ListIssuesFilter{ProjectID: pid})
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]struct {
		leaf bool
		n    int
	}{}
	for _, iss := range list {
		found[iss.ID] = struct {
			leaf bool
			n    int
		}{iss.IsLeaf, iss.ChildCount}
	}
	if p := found[parent.ID]; p.leaf || p.n != 2 {
		t.Fatalf("list parent: %+v", p)
	}
	if c := found[c1.ID]; !c.leaf || c.n != 0 {
		t.Fatalf("list c1: %+v", c)
	}
}

func TestClaimGateNonLeaf409(t *testing.T) {
	svc, ws, pid := setup(t)
	mid := insertMember(t, svc, ws)
	parent, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{ProjectID: pid, Title: "P", ActorKind: "human"})
	if err != nil {
		t.Fatal(err)
	}
	pp := parent.ID
	leaf, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, Title: "L", ParentID: &pp, ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	parent, err = svc.GetIssue(ws, parent.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.UpdateIssue(ws, parent.ID, issueboard.UpdateIssueInput{
		Version: parent.Version, AssigneeID: &mid, ActorKind: "human",
	})
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.StatusCode != 409 {
		t.Fatalf("want 409 on parent claim, got %v", err)
	}
	if !strings.Contains(ae.Message, "仅叶子任务可领取") {
		t.Fatalf("message: %q", ae.Message)
	}
	// leaf assign ok
	got, err := svc.UpdateIssue(ws, leaf.ID, issueboard.UpdateIssueInput{
		Version: leaf.Version, AssigneeID: &mid, ActorKind: "bot",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.AssigneeID == nil || *got.AssigneeID != mid {
		t.Fatalf("leaf assignee: %+v", got.AssigneeID)
	}
}

func TestBotOrphanCreateRejected(t *testing.T) {
	svc, ws, pid := setup(t)
	_, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, Title: "orphan", ActorKind: "bot",
	})
	ae, ok := err.(*apperr.AppError)
	if !ok || (ae.StatusCode != 403 && ae.StatusCode != 400) {
		t.Fatalf("want 403/400, got %v", err)
	}
	if !strings.Contains(ae.Message, "Bot 不能创建无父节点的游离 Issue") {
		t.Fatalf("message: %q", ae.Message)
	}
	// with parent ok
	parent, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, Title: "P", ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	pp := parent.ID
	child, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, Title: "leaf", ParentID: &pp, ActorKind: "bot",
	})
	if err != nil {
		t.Fatal(err)
	}
	if child.ParentID == nil || *child.ParentID != parent.ID {
		t.Fatalf("parent: %+v", child.ParentID)
	}
}

func TestLeafOnlyFilter(t *testing.T) {
	svc, ws, pid := setup(t)
	parent, _ := svc.CreateIssue(ws, issueboard.CreateIssueInput{ProjectID: pid, Title: "P", ActorKind: "human"})
	pp := parent.ID
	c1, _ := svc.CreateIssue(ws, issueboard.CreateIssueInput{ProjectID: pid, Title: "C1", ParentID: &pp, ActorKind: "human"})
	_, _ = svc.CreateIssue(ws, issueboard.CreateIssueInput{ProjectID: pid, Title: "solo", ActorKind: "human"})

	all, err := svc.ListIssues(ws, issueboard.ListIssuesFilter{ProjectID: pid})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 {
		t.Fatalf("want 3, got %d", len(all))
	}
	leaves, err := svc.ListIssues(ws, issueboard.ListIssuesFilter{ProjectID: pid, LeafOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, iss := range leaves {
		ids[iss.ID] = true
		if !iss.IsLeaf {
			t.Fatalf("leaf_only returned non-leaf %s", iss.Identifier)
		}
	}
	if ids[parent.ID] {
		t.Fatal("parent should be excluded by leaf_only")
	}
	if !ids[c1.ID] {
		t.Fatal("child should be included")
	}
	if len(leaves) != 2 {
		t.Fatalf("want 2 leaves, got %d", len(leaves))
	}
}

func TestBlockedAndOverdueFilters(t *testing.T) {
	svc, ws, pid := setup(t)
	blocker, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, Title: "blocker", Status: "todo", ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	blocked, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, Title: "blocked", ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	free, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, Title: "free", ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.CreateRelation(ws, blocker.ID, issueboard.CreateRelationInput{
		ToIssueID: blocked.ID, RelationType: "blocks", ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	list, err := svc.ListIssues(ws, issueboard.ListIssuesFilter{ProjectID: pid, Blocked: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ID != blocked.ID {
		t.Fatalf("blocked filter: %+v", list)
	}
	_ = free

	past := "2020-01-01"
	future := "2099-12-31"
	over, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, Title: "over", DueAt: &past, ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, Title: "future", DueAt: &future, ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	olist, err := svc.ListIssues(ws, issueboard.ListIssuesFilter{ProjectID: pid, Overdue: true})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, iss := range olist {
		if iss.ID == over.ID {
			found = true
		}
		if iss.DueAt == nil || *iss.DueAt >= "2026-10-24" && *iss.DueAt > "2090-01-01" {
			// future due must not appear; soft check via ID
		}
	}
	if !found {
		t.Fatal("expected overdue issue in filter")
	}
	for _, iss := range olist {
		if iss.Title == "future" {
			t.Fatal("future due should not be overdue")
		}
	}
}

func TestForceAssigneeGate(t *testing.T) {
	svc, ws, pid := setup(t)
	m1 := insertMember(t, svc, ws)
	m2 := insertMember(t, svc, ws)
	leaf, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{ProjectID: pid, Title: "L", ActorKind: "human"})
	if err != nil {
		t.Fatal(err)
	}
	got, err := svc.UpdateIssue(ws, leaf.ID, issueboard.UpdateIssueInput{
		Version: leaf.Version, AssigneeID: &m1, ActorKind: "bot",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.UpdateIssue(ws, leaf.ID, issueboard.UpdateIssueInput{
		Version: got.Version, AssigneeID: &m2, ActorKind: "bot",
	})
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.StatusCode != 409 {
		t.Fatalf("want 409 steal, got %v", err)
	}
	if !strings.Contains(ae.Message, "禁止抢占") {
		t.Fatalf("message: %q", ae.Message)
	}
	got2, err := svc.UpdateIssue(ws, leaf.ID, issueboard.UpdateIssueInput{
		Version: got.Version, AssigneeID: &m2, ActorKind: "bot", ForceAssignee: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got2.AssigneeID == nil || *got2.AssigneeID != m2 {
		t.Fatalf("force assignee: %+v", got2.AssigneeID)
	}
}

func TestIssueDispatchEvents(t *testing.T) {
	svc, ws, pid := setup(t)
	var mu sync.Mutex
	var events []string
	svc.DispatchEvent = func(workspaceID, event string, payload any) {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, event)
		if workspaceID != ws {
			t.Errorf("ws=%s want %s", workspaceID, ws)
		}
		if payload == nil {
			t.Errorf("nil payload for %s", event)
		}
	}
	m1 := insertMember(t, svc, ws)
	iss, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, Title: "hook", ActorKind: "human", AssigneeID: &m1,
	})
	if err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	got := append([]string(nil), events...)
	mu.Unlock()
	if len(got) < 2 || got[0] != "issue.created" || got[1] != "issue.assigned" {
		t.Fatalf("create events=%v want [issue.created issue.assigned …]", got)
	}

	events = nil
	m2 := insertMember(t, svc, ws)
	_, err = svc.UpdateIssue(ws, iss.ID, issueboard.UpdateIssueInput{
		Version: iss.Version, AssigneeID: &m2, ActorKind: "human", ForceAssignee: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	got = append([]string(nil), events...)
	mu.Unlock()
	hasUpdated, hasAssigned := false, false
	for _, e := range got {
		if e == "issue.updated" {
			hasUpdated = true
		}
		if e == "issue.assigned" {
			hasAssigned = true
		}
	}
	if !hasUpdated || !hasAssigned {
		t.Fatalf("update events=%v want issue.updated + issue.assigned", got)
	}
}

func TestIssueNestedAssigneeReporterDisplayName(t *testing.T) {
	svc, ws, pid := setup(t)
	assigneeName := "Alice Chen"
	reporterName := "Bob Reporter"
	assigneeID := insertMemberNamed(t, svc, ws, assigneeName)
	reporterID := insertMemberNamed(t, svc, ws, reporterName)

	// Assigned + reporter via ActorID
	iss, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, Title: "with assignee", ActorKind: "human",
		ActorID: reporterID, AssigneeID: &assigneeID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if iss.Assignee == nil || iss.Assignee.ID != assigneeID || iss.Assignee.DisplayName != assigneeName {
		t.Fatalf("Create assignee nested: %+v", iss.Assignee)
	}
	if iss.Reporter == nil || iss.Reporter.ID != reporterID || iss.Reporter.DisplayName != reporterName {
		t.Fatalf("Create reporter nested: %+v", iss.Reporter)
	}
	if iss.AssigneeID == nil || *iss.AssigneeID != assigneeID {
		t.Fatalf("assignee_id not kept: %v", iss.AssigneeID)
	}

	got, err := svc.GetIssue(ws, iss.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Assignee == nil || got.Assignee.DisplayName != assigneeName {
		t.Fatalf("Get assignee.display_name want %q got %+v", assigneeName, got.Assignee)
	}
	if got.Reporter == nil || got.Reporter.DisplayName != reporterName {
		t.Fatalf("Get reporter.display_name want %q got %+v", reporterName, got.Reporter)
	}

	list, err := svc.ListIssues(ws, issueboard.ListIssuesFilter{ProjectID: pid})
	if err != nil {
		t.Fatal(err)
	}
	var listed *models.Issue
	for i := range list {
		if list[i].ID == iss.ID {
			listed = &list[i]
			break
		}
	}
	if listed == nil {
		t.Fatal("issue missing from list")
	}
	if listed.Assignee == nil || listed.Assignee.DisplayName != assigneeName {
		t.Fatalf("List assignee.display_name want %q got %+v", assigneeName, listed.Assignee)
	}

	// Unassigned → assignee null
	unassigned, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, Title: "no assignee", ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	if unassigned.Assignee != nil {
		t.Fatalf("unassigned Create want assignee null, got %+v", unassigned.Assignee)
	}
	if unassigned.AssigneeID != nil {
		t.Fatalf("unassigned Create want assignee_id null, got %v", unassigned.AssigneeID)
	}
	ugot, err := svc.GetIssue(ws, unassigned.ID)
	if err != nil {
		t.Fatal(err)
	}
	if ugot.Assignee != nil {
		t.Fatalf("unassigned Get want assignee null, got %+v", ugot.Assignee)
	}
	ulist, err := svc.ListIssues(ws, issueboard.ListIssuesFilter{ProjectID: pid, Q: "no assignee"})
	if err != nil {
		t.Fatal(err)
	}
	if len(ulist) != 1 || ulist[0].Assignee != nil {
		t.Fatalf("unassigned List want assignee null, got %+v", ulist)
	}

}
