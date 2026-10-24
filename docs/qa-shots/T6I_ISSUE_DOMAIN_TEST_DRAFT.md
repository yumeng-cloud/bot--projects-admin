# T6-I · Issue 域测草（NULL 旧单 + 错队 404）

> **DRAFT** · 2026-10-24 Asia/Shanghai  
> 等 Go 落地 `issues.team_id` 列 + Create/List/Board 过滤后再把骨架贴进  
> `backend/internal/services/issueboard/*_test.go`（或独立 `team_pool_test.go`）。  
> **在列未进 `origin/main` 前勿 `go test` 本骨架**（会引用尚不存在的 `TeamID` / `UnassignedOnly` 符号）。

权威：`docs/NEXT_WAVE_DEBT_T6I.md` · `docs/NEXT_WAVE_DEBT_ISSUE_POOL.md` · `docs/API.md`「Issue · team_id」  
产品裁定（本草遵循）：

| 规则 | 口径 |
|------|------|
| 新建 | `team_id` **必填**（当前队）；缺省 / 非本 WS 队 → **4xx 人话**（建议 400） |
| 旧单 NULL | = **未分队**；**默认 list（`team_id=<current>`）不含 NULL** |
| 未分队滤镜 | 可选 `unassigned=1`（或 `UnassignedOnly`）→ **仅** `team_id IS NULL` |
| 错队 | GET / PATCH（带队上下文）→ **404**（非 403；不泄露存在性） |
| 不做 | T7 · pins · auth · Issue 域改 `internal/db/**` |

---

## 风格锚（本仓已有）

- 包名：`package issueboard_test`
- 开库：`setup(t)` → `db.Open("sqlite", t.TempDir()/t.db)` + `workspaces` insert + `CreateProject`
- 断言：`err.(*apperr.AppError)` + `StatusCode`；人话用 `strings.Contains(ae.Message, …)`
- 参考：`issue_test.go` / `concurrency_test.go` 的 `setup` · `insertMember*` · 409/403 断言

队种子（Issue 域不引 `botmem`）：与 `insertMember` 同风格 raw SQL：

```go
func insertTeam(t *testing.T, svc *issueboard.Service, wsID, name string) string {
	t.Helper()
	tid := idgen.New()
	_, err := svc.DB.Exec(
		`INSERT INTO teams (id, workspace_id, name, slug) VALUES (?,?,?,?)`,
		tid, wsID, name, "slug-"+tid[:8],
	)
	if err != nil {
		t.Fatal(err)
	}
	return tid
}
```

旧 NULL 单（列落地后）：Create 带队后，或测试直接 `UPDATE issues SET team_id=NULL WHERE id=?`（模拟迁移前遗留）。

---

## 假设（文档歧义处 · 测前请 Go 确认）

1. **默认 list 是否含 NULL**  
   - **本草 / API.md / ISSUE_POOL 产品裁定**：**不含**（仅 `i.team_id = ?`）。  
   - T6I 正文曾写「NULL=全 WS 可见或挂默认队，证据写明」——若 Go 选「队 list 并含 NULL」，本草 Case B 需改期望并在证据注明。  
2. **Board** 与 List 共用 `ListIssuesFilter.TeamID` / 同等过滤；无独立 board SQL 时复用 List 用例即可。  
3. **错队 GET**：调用方传当前队（handler `?team_id=` → `GetIssueInTeam` / `RequireTeamID`）；无队上下文的裸 `GetIssue` 可仍按 WS 可见（本草测「带队」路径）。  
4. **NULL 旧单 GET/PATCH 带队**：产品未强制 404；建议仍可读/可改（未分队），与「错队=已分到他队」区分。  
5. **无/错 `team_id` list**：API 写错/无队 → 404；服务层若只收 filter 字符串，404 可能在 handler（队不存在校验）——骨架在服务层标 `TODO`。

---

## 骨架（可复制）

```go
package issueboard_test

import (
	"strings"
	"testing"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/issueboard"
)

// insertTeam — see helper above (paste next to insertMember in issue_test.go).

func TestTeamPool_ListHidesOtherTeam(t *testing.T) {
	svc, ws, pid := setup(t)
	teamA := insertTeam(t, svc, ws, "A")
	teamB := insertTeam(t, svc, ws, "B")

	a, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, TeamID: teamA, Title: "A-only", ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	b, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, TeamID: teamB, Title: "B-only", ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}

	listA, err := svc.ListIssues(ws, issueboard.ListIssuesFilter{ProjectID: pid, TeamID: teamA})
	if err != nil {
		t.Fatal(err)
	}
	listB, err := svc.ListIssues(ws, issueboard.ListIssuesFilter{ProjectID: pid, TeamID: teamB})
	if err != nil {
		t.Fatal(err)
	}

	seenA := map[string]bool{}
	for _, it := range listA {
		seenA[it.ID] = true
	}
	seenB := map[string]bool{}
	for _, it := range listB {
		seenB[it.ID] = true
	}
	if !seenA[a.ID] || seenA[b.ID] {
		t.Fatalf("team A list want A not B: %+v", seenA)
	}
	if !seenB[b.ID] || seenB[a.ID] {
		t.Fatalf("team B list want B not A: %+v", seenB)
	}
}

func TestTeamPool_DefaultListHidesNULL(t *testing.T) {
	svc, ws, pid := setup(t)
	teamA := insertTeam(t, svc, ws, "A")

	owned, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, TeamID: teamA, Title: "owned", ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, TeamID: teamA, Title: "legacy-null", ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DB.Exec(`UPDATE issues SET team_id=NULL WHERE id=?`, legacy.ID); err != nil {
		t.Fatal(err)
	}

	list, err := svc.ListIssues(ws, issueboard.ListIssuesFilter{ProjectID: pid, TeamID: teamA})
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, it := range list {
		seen[it.ID] = true
	}
	// PRODUCT: default team filter excludes NULL
	if !seen[owned.ID] {
		t.Fatal("want owned issue in default team list")
	}
	if seen[legacy.ID] {
		t.Fatal("default team list must hide team_id IS NULL (use unassigned filter)")
	}
}

func TestTeamPool_UnassignedFilterOnlyNULL(t *testing.T) {
	svc, ws, pid := setup(t)
	teamA := insertTeam(t, svc, ws, "A")

	owned, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, TeamID: teamA, Title: "owned", ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, TeamID: teamA, Title: "legacy", ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DB.Exec(`UPDATE issues SET team_id=NULL WHERE id=?`, legacy.ID); err != nil {
		t.Fatal(err)
	}

	list, err := svc.ListIssues(ws, issueboard.ListIssuesFilter{
		ProjectID: pid, UnassignedOnly: true, // handler: ?unassigned=1
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ID != legacy.ID {
		t.Fatalf("unassigned filter want only NULL legacy, got %+v", list)
	}
	for _, it := range list {
		if it.ID == owned.ID {
			t.Fatal("unassigned must not return assigned-team issues")
		}
		if it.TeamID != nil && *it.TeamID != "" {
			t.Fatalf("unassigned row must have nil team_id, got %+v", it.TeamID)
		}
	}
}

func TestTeamPool_CreateRequiresTeamID(t *testing.T) {
	svc, ws, pid := setup(t)
	_, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, Title: "no-team", ActorKind: "human",
		// TeamID omitted / empty
	})
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.StatusCode < 400 || ae.StatusCode >= 500 {
		t.Fatalf("want 4xx, got %v", err)
	}
	// human message (zh); adjust needle to final Go copy
	if !strings.Contains(ae.Message, "team_id") && !strings.Contains(ae.Message, "团队") {
		t.Fatalf("want human message mentioning team_id/团队, got %q", ae.Message)
	}

	// wrong-WS / unknown team → also 4xx human (not silent insert)
	_, err = svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, TeamID: idgen.New(), Title: "bad-team", ActorKind: "human",
	})
	ae, ok = err.(*apperr.AppError)
	if !ok || ae.StatusCode < 400 || ae.StatusCode >= 500 {
		t.Fatalf("want 4xx for foreign team_id, got %v", err)
	}
}

func TestTeamPool_GetPatchWrongTeam404(t *testing.T) {
	svc, ws, pid := setup(t)
	teamA := insertTeam(t, svc, ws, "A")
	teamB := insertTeam(t, svc, ws, "B")

	iss, err := svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: pid, TeamID: teamA, Title: "A-iss", ActorKind: "human",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Prefer GetIssueInTeam / UpdateIssue{RequireTeamID} once Go exports them.
	// Fallback: if only handler wires ?team_id=, keep this as service-level contract test.
	_, err = svc.GetIssueInTeam(ws, iss.ID, teamB)
	ae, ok := err.(*apperr.AppError)
	if !ok || ae.StatusCode != 404 {
		t.Fatalf("wrong-team GET want 404 not 403, got %v", err)
	}
	if ae.StatusCode == 403 {
		t.Fatal("must not use 403 for wrong-team (existence leak)")
	}

	title := "hijack"
	_, err = svc.UpdateIssue(ws, iss.ID, issueboard.UpdateIssueInput{
		Version: iss.Version, Title: &title, ActorKind: "human",
		RequireTeamID: teamB,
	})
	ae, ok = err.(*apperr.AppError)
	if !ok || ae.StatusCode != 404 {
		t.Fatalf("wrong-team PATCH want 404, got %v", err)
	}

	// same-team still OK
	_, err = svc.GetIssueInTeam(ws, iss.ID, teamA)
	if err != nil {
		t.Fatalf("same-team GET: %v", err)
	}
}
```

---

## Case 对照表

| Case | 期望 | 对应函数 |
|------|------|----------|
| List 队 A 不见队 B | A list 含 A 单、不含 B 单；对称 | `TestTeamPool_ListHidesOtherTeam` |
| 默认 list 藏 NULL | `TeamID=A` 不含 `UPDATE→NULL` 旧单 | `TestTeamPool_DefaultListHidesNULL` |
| 未分队滤镜 | `UnassignedOnly` 仅 NULL | `TestTeamPool_UnassignedFilterOnlyNULL` |
| Create 无 team_id | 4xx + 人话含 team/团队 | `TestTeamPool_CreateRequiresTeamID` |
| GET/PATCH 错队 | **404**（非 403） | `TestTeamPool_GetPatchWrongTeam404` |

---

## 落地 checklist（Go 交件后）

- [ ] 符号已在 `models.Issue` / `CreateIssueInput` / `ListIssuesFilter`  
- [ ] 骨架去掉 `ids` 占位，`import models` 若需要  
- [ ] `go test ./internal/services/issueboard/...` 绿  
- [ ] 若默认 list **含** NULL：改 Case B 期望 + 在 `T6I-issue-pool-*.txt` 写明口径  
- [ ] 本文件标题去掉 **DRAFT** 或另开正式测文件

