package issueboard_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/issueboard"
)

// TestConcurrentBoardOrderOptimisticLock：同 Issue 并发改 board_order/status/version，
// 至多一人成功，失败者必须 409，禁止静默覆盖。
func TestConcurrentBoardOrderOptimisticLock(t *testing.T) {
	svc, ws, pid, tid := setup(t)
	iss, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{TeamID: tid, ProjectID: pid, Title: "board", ActorKind: "human"})
	if err != nil {
		t.Fatal(err)
	}

	const n = 8
	var wg sync.WaitGroup
	var wins, conflicts atomic.Int64
	errs := make(chan error, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			order := float64(100 + i)
			status := "todo"
			if i%2 == 0 {
				status = "in_progress"
			}
			_, err := svc.UpdateIssue(ws, iss.ID, issueboard.UpdateIssueInput{
				Version:    iss.Version,
				BoardOrder: &order,
				Status:     &status,
				ActorKind:  "human",
			})
			if err == nil {
				wins.Add(1)
				return
			}
			ae, ok := err.(*apperr.AppError)
			if ok && ae.StatusCode == 409 {
				conflicts.Add(1)
				return
			}
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("unexpected: %v", err)
	}
	if wins.Load() != 1 {
		t.Fatalf("want exactly 1 winner, got wins=%d conflicts=%d", wins.Load(), conflicts.Load())
	}
	if conflicts.Load() != int64(n-1) {
		t.Fatalf("want %d conflicts, got %d", n-1, conflicts.Load())
	}
	got, err := svc.GetIssue(ws, iss.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != iss.Version+1 {
		t.Fatalf("version want %d got %d", iss.Version+1, got.Version)
	}
}

// TestStaleVersionBoardMove：跨列移动用旧 version 不得覆盖。
func TestStaleVersionBoardMove(t *testing.T) {
	svc, ws, pid, tid := setup(t)
	iss, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{TeamID: tid, ProjectID: pid, Title: "move", ActorKind: "human"})
	if err != nil {
		t.Fatal(err)
	}
	st := "todo"
	ord := 1.0
	upd, err := svc.UpdateIssue(ws, iss.ID, issueboard.UpdateIssueInput{
		Version: iss.Version, Status: &st, BoardOrder: &ord, ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	st2 := "in_progress"
	ord2 := 2.0
	_, err = svc.UpdateIssue(ws, iss.ID, issueboard.UpdateIssueInput{
		Version: iss.Version, Status: &st2, BoardOrder: &ord2, ActorKind: "human",
	})
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.StatusCode != 409 {
		t.Fatalf("want 409, got %v", err)
	}
	cur, _ := svc.GetIssue(ws, iss.ID)
	if cur.Status != "todo" || cur.BoardOrder != 1.0 || cur.Version != upd.Version {
		t.Fatalf("stale write must not apply: %+v", cur)
	}
}
