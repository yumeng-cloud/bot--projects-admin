package botmem_test

import (
	"testing"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/botmem"
)

// TestPinsTeamIsolate P1：队 A 写入的 pin 仅 A 可见；B 列表无该 key。
// P2：队 B 对队 A 的 key upsert/delete → 鉴权 404。
func TestPinsTeamIsolate(t *testing.T) {
	svc, ws, m1, m2 := setup(t)
	teamA, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "PinA", Slug: "pin-a"})
	if err != nil {
		t.Fatal(err)
	}
	teamB, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "PinB", Slug: "pin-b"})
	if err != nil {
		t.Fatal(err)
	}
	ra, err := svc.EnsureRoster(ws, teamA.ID, botmem.EnsureRosterInput{})
	if err != nil {
		t.Fatal(err)
	}
	rb, err := svc.EnsureRoster(ws, teamB.ID, botmem.EnsureRosterInput{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.BindSeat(ws, teamA.ID, ra[0].ID, m1, false); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.BindSeat(ws, teamB.ID, rb[0].ID, m2, false); err != nil {
		t.Fatal(err)
	}

	// P1 access: m1→A ok；m2→A 404
	if err := svc.RequireTeamPinAccess(ws, teamA.ID, m1, false); err != nil {
		t.Fatalf("m1 should access A: %v", err)
	}
	err = svc.RequireTeamPinAccess(ws, teamA.ID, m2, false)
	if err == nil {
		t.Fatal("m2 must not access A")
	}
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.Code != "not_found" || ae.StatusCode != 404 {
		t.Fatalf("want 404 not_found, got %#v", err)
	}
	if ae.Message != "团队不存在或无权访问" {
		t.Fatalf("msg=%q", ae.Message)
	}

	pin, err := svc.UpsertPin(ws, teamA.ID, botmem.UpsertPinInput{Key: "wall.note", Value: "from-A", UpdatedBy: m1})
	if err != nil {
		t.Fatal(err)
	}
	if pin.Key != "wall.note" || pin.Value != "from-A" || pin.TeamID != teamA.ID {
		t.Fatalf("pin=%+v", pin)
	}

	listA, err := svc.ListPins(ws, teamA.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(listA) != 1 || listA[0].Key != "wall.note" {
		t.Fatalf("A list=%+v", listA)
	}
	listB, err := svc.ListPins(ws, teamB.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range listB {
		if p.Key == "wall.note" {
			t.Fatalf("A pin leaked to B: %+v", p)
		}
	}
	if len(listB) != 0 {
		t.Fatalf("B should be empty, got %+v", listB)
	}

	// P2: B 不得对 A 的 key 做写/删（鉴权层）
	if err := svc.RequireTeamPinAccess(ws, teamA.ID, m2, false); err == nil {
		t.Fatal("P2: B access A should 404")
	}
	// 即便绕过鉴权直接写，也不应改到 B 的池（不同 team_id）
	if _, err := svc.UpsertPin(ws, teamB.ID, botmem.UpsertPinInput{Key: "wall.note", Value: "from-B", UpdatedBy: m2}); err != nil {
		t.Fatal(err)
	}
	listA2, err := svc.ListPins(ws, teamA.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(listA2) != 1 || listA2[0].Value != "from-A" {
		t.Fatalf("A pin overwritten by same key on B: %+v", listA2)
	}
	listB2, err := svc.ListPins(ws, teamB.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(listB2) != 1 || listB2[0].Value != "from-B" {
		t.Fatalf("B pin=%+v", listB2)
	}

	// Admin 可跨队访问
	if err := svc.RequireTeamPinAccess(ws, teamA.ID, m2, true); err != nil {
		t.Fatalf("admin should access: %v", err)
	}

	// 无队
	err = svc.RequireTeamPinAccess(ws, "no-such-team", m1, false)
	if err == nil {
		t.Fatal("missing team should 404")
	}
	ae, ok = err.(*apperr.AppError)
	if !ok || ae.Code != "not_found" {
		t.Fatalf("want not_found %#v", err)
	}

	if err := svc.DeletePin(ws, teamA.ID, "wall.note"); err != nil {
		t.Fatal(err)
	}
	listA3, _ := svc.ListPins(ws, teamA.ID)
	if len(listA3) != 0 {
		t.Fatalf("after delete A=%+v", listA3)
	}
	// B 同名仍在
	listB3, _ := svc.ListPins(ws, teamB.ID)
	if len(listB3) != 1 || listB3[0].Key != "wall.note" {
		t.Fatalf("B pin deleted wrongly: %+v", listB3)
	}
}

// TestPinsDoNotClobberBotMemories P3：同名 key 与 bot_memories 互不影响。
func TestPinsDoNotClobberBotMemories(t *testing.T) {
	svc, ws, m1, _ := setup(t)
	team, err := svc.CreateTeam(ws, botmem.CreateTeamInput{Name: "PinMem"})
	if err != nil {
		t.Fatal(err)
	}
	roster, err := svc.EnsureRoster(ws, team.ID, botmem.EnsureRosterInput{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.BindSeat(ws, team.ID, roster[0].ID, m1, false); err != nil {
		t.Fatal(err)
	}

	const sharedKey = "shared.ctx"
	mem, err := svc.UpsertMemory(ws, m1, botmem.UpsertMemoryInput{KeyName: sharedKey, ValueText: "private-mem"})
	if err != nil {
		t.Fatal(err)
	}
	pin, err := svc.UpsertPin(ws, team.ID, botmem.UpsertPinInput{Key: sharedKey, Value: "team-pin", UpdatedBy: m1})
	if err != nil {
		t.Fatal(err)
	}
	if pin.ID == mem.ID {
		t.Fatal("pin and memory must have distinct ids")
	}

	mem2, err := svc.GetMemory(ws, m1, mem.ID)
	if err != nil {
		t.Fatal(err)
	}
	if mem2.ValueText != "private-mem" || mem2.KeyName != sharedKey {
		t.Fatalf("memory clobbered: %+v", mem2)
	}

	// overwrite pin must not touch memory
	if _, err := svc.UpsertPin(ws, team.ID, botmem.UpsertPinInput{Key: sharedKey, Value: "team-pin-2", UpdatedBy: m1}); err != nil {
		t.Fatal(err)
	}
	mem3, err := svc.GetMemory(ws, m1, mem.ID)
	if err != nil {
		t.Fatal(err)
	}
	if mem3.ValueText != "private-mem" {
		t.Fatalf("memory changed after pin upsert: %+v", mem3)
	}

	// overwrite memory must not touch pin
	if _, err := svc.UpsertMemory(ws, m1, botmem.UpsertMemoryInput{KeyName: sharedKey, ValueText: "private-mem-2"}); err != nil {
		t.Fatal(err)
	}
	pins, err := svc.ListPins(ws, team.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(pins) != 1 || pins[0].Value != "team-pin-2" {
		t.Fatalf("pin changed after memory upsert: %+v", pins)
	}

	if err := svc.DeletePin(ws, team.ID, sharedKey); err != nil {
		t.Fatal(err)
	}
	mem4, err := svc.GetMemory(ws, m1, mem.ID)
	if err != nil {
		t.Fatal(err)
	}
	if mem4.ValueText != "private-mem-2" {
		t.Fatalf("memory deleted by pin delete: %+v", mem4)
	}
}
