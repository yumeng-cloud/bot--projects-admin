# Bot 连续性（记忆 / 角色 / Handoff）

> **已有** API Key/`ymk_` + Bot REST；**本次补** `ymbot` CLI+MCP（[`BOT_MCP_CLI.md`](./BOT_MCP_CLI.md)）。记忆向量/语义召回：见 [`embed.md`](./embed.md)（upsert 写 `bot_memory` 向量；`GET …/memories?q=&mode=hybrid|semantic|keyword`）。

## 职责

为工作空间内的 **Bot 成员**（`members.is_bot=1`）提供：

- **角色（bot_roles）**：frontend / pm / qa / backend / wiki 等协作分工定义
- **认领（bot_role_claims）**：同一角色同时仅一个 active claim（`released_at IS NULL`）
- **记忆（bot_memories）**：按 `(workspace, member, key_name)` upsert 持久上下文
- **待办（bot_todos）**：Bot 自身 todo|doing|done|canceled
- **Handoff**：只读续工包（无独立表），组装 memories + 未完成 todos + 指派中的开放 Issues

Bot **不得**将 Issue 自动流转到 `done`（需人类确认）。

### 排期叶子领取（与角色认领区分）

| 概念 | 通道 | 含义 |
|------|------|------|
| **角色认领** | `POST …/bot-roles/{id}/claim`；MCP P1 `board_role_claim` | 岗位（frontend/pm/…），表 `bot_role_claims` |
| **Issue 叶子领取** | `ymbot issue claim` / MCP `board_issue_claim`（PATCH `assignee_id=me`） | 干哪条细活；**仅叶子**（无子节点） |

原则（见 [`SCHEDULING_BOARD_P0.md`](./SCHEDULING_BOARD_P0.md) §4.5）：

1. Bot **只领叶子**；父/聚合节点不可 claim（服务端非叶子 → 409）。
2. **禁止**自创无 `parent_id` 的游离 Issue 当待办；`issue create` / `board_issue_create` **须** `parent_id`（通道提前拒）。
3. 新工作由人类（或编排在已有父下）拆叶子 → Bot `issue claim` / `board_issue_claim`。
4. 可领池：`ymbot issue list --leaf-only`（`leaf_only=1`，依赖后端过滤）。
5. **I6**：领取只 PATCH 共享字段 `assignee_id`（与看板/Timeline/抽屉同源）；勿另写平行「认领」表。


## 表

| 表 | 说明 |
|----|------|
| `bot_roles` | 角色定义，`UNIQUE(workspace_id, slug)` |
| `bot_role_claims` | 认领记录；active = `released_at IS NULL`（schema `UNIQUE(role_id)`，释放后重认领走 UPDATE） |
| `bot_memories` | 记忆，`UNIQUE(workspace_id, member_id, key_name)` |
| `bot_todos` | 待办 |

Handoff **无表**，由 service 实时组装。

## REST（均需工作空间成员；写操作支持 `Idempotency-Key`）

鉴权：`Authorization: Bearer <JWT|ymk_…>`（`AuthJWTOrAPIKey`）。

| Method | Path | 说明 |
|--------|------|------|
| GET/POST | `/api/v1/workspaces/{ws}/bot-roles` | 列表 / 创建 |
| POST | `.../bot-roles/{roleID}/claim` | body 可选 `member_id`，默认当前成员；冲突 409 |
| POST | `.../bot-roles/{roleID}/release` | 释放 active claim |
| GET | `.../bots` | Bot 名册（is_bot + claimed_role） |
| GET | `.../bots/me` | 当前鉴权主体（JWT 或 Bot API Key）成员身份；已落地 |
| GET/POST | `.../bots/{memberID}/memories` | list 支持 `?q=&mode=&limit=`（mode 缺省兼容旧 FE）；POST upsert by key_name + 写向量 |
| PATCH/DELETE | `.../bots/{memberID}/memories/{memoryID}` | 更新 / 删除 |
| GET/POST | `.../bots/{memberID}/todos` | list 支持 `?status=` |
| PATCH | `.../bots/{memberID}/todos/{todoID}` | 改 title/status/… |
| GET | `.../bots/{memberID}/handoff` | 续工包（导出） |
| POST | `.../bots/{memberID}/handoff` | 导入续工包：`{ "handoff": {...} }` 或直接 memories/todos；返回 imported_* / claimed / warnings + 刷新后 handoff |

错误体：`{"code","message"}`（snake_case）。

记忆语义召回（Postgres+pgvector）：upsert 写向量、`?mode=` 检索，约定见 [`embed.md`](./embed.md)。本地 SQLite 仅关键词/`keyword_fallback`，不以 sqlite-vec 为生产方案。

### Handoff 字段

```json
{
  "handoff": {
    "workspace_id": "...",
    "member_id": "...",
    "generated_at": "...",
    "claimed_role": { "...": "..." },
    "memories": [],
    "todos": [],
    "issues": []
  }
}
```

- `todos`：status ∈ `todo|doing`
- `issues`：`assignee_id=member` 且 status ∈ `todo|in_progress|in_review`

## 续工流程建议

1. Bot 启动 → `GET .../bots/{me}/handoff` 恢复上下文
2. 认领**角色**（岗位）→ `POST .../bot-roles/{id}/claim`（≠ Issue 叶子 claim）
3. 从面板 **领取叶子** → `ymbot issue claim` / `board_issue_claim`；勿自创无父游离 Issue
4. 工作中 upsert memories、维护 todos；写操作带 `Idempotency-Key`
5. 交接前更新 memories / 打开 todos；下一会话再拉 handoff
6. **换机续工**：源机 `GET .../handoff` 导出 → 目标机 `POST .../handoff` 导入（memories upsert、未完成 todos、尝试 claim；claim 冲突只记 warning）
7. 换岗 → `POST .../release` 后由其他成员 claim

## 挂载方式

Handlers 提供：

```go
handlers.MountBots(mux, db, authMW) // authMW = AuthJWTOrAPIKey
```

主进程通过 `modules.RegisterBotMem` 在 init 中赋值并挂载（见 `backend/internal/modules/botmem`）。**勿**再向 stub 列表加 `/bots`。

Seed：`seed.SeedBotRoles(db, workspaceID)`（在 `seed.Run` 末尾幂等调用），默认 slug：`frontend` / `pm` / `qa` / `backend` / `wiki`。

## 前端

| 路由 | 页面 |
|------|------|
| `/w/:wsId/settings/bots` | Bot 名册 |
| `/w/:wsId/settings/bots/:memberId` | 详情：记忆 / Todo / handoff |

API client：`frontend/src/api/bots.ts`（snake_case）。


## CLI / MCP（通道层）

规格与验收：[`BOT_MCP_CLI.md`](./BOT_MCP_CLI.md)。与 REST 关系：**不改业务语义**，仅 HTTP 调现有接口。

```bash
export YMBOT_API_URL=http://127.0.0.1:8080
export YMBOT_API_KEY=ymk_***
export YMBOT_WORKSPACE_ID=<ws>
# 可选：YMBOT_MEMBER_ID（缺省 GET …/bots/me）
# 可选：YMBOT_IDEMPOTENCY_KEY=<uuid>（写操作未传 --idempotency-key 时）

go run ./cmd/ymbot --help
go run ./cmd/ymbot whoami
go run ./cmd/ymbot issue list --assignee me
go run ./cmd/ymbot comment create --issue <id> -m "进展…" --idempotency-key <uuid>
go run ./cmd/ymbot memory set --key last_focus --value "…" --idempotency-key <uuid>
go run ./cmd/ymbot handoff export --out /tmp/handoff.json
go run ./cmd/ymbot mcp-serve   # Cursor stdio MCP；tools: board_*（写 tool 可传 idempotency_key）
```

共享客户端：`internal/botclient`（Bearer；Idempotency-Key：CLI/MCP 参数 → env → 自动 UUID；一行中文错误含吊销/无效；Bot 禁 `status=done`）。
鉴权主路径：**API Key only**（不主推密码 login）。

## 边界

- 不新建 handoff 表；不改 `schema.sql` 已有 bot_* 列
- 不负责 Issue 状态机 / 看板 UI / API Key CRUD
- Bot 禁止自动 `done` Issue
- claim 冲突返回 `conflict`；记忆按 key upsert

## 相关

- Bot 工作流 MCP/CLI 最小能力规格：[`BOT_MCP_CLI.md`](./BOT_MCP_CLI.md)
