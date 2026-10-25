# Bot 连续性（记忆 / 角色 / Handoff）

> **主路径：** 人类 Admin 在 Bot **席页**点「复制入场链接」或「复制 Bot 提示词」→ 对方打开/粘贴 → 装 CLI（若无）→ join → **自动得** `ymk_` → 开干。REST / 记忆 / handoff 见下文；通道规格 [`BOT_MCP_CLI.md`](./BOT_MCP_CLI.md)；向量召回 [`embed.md`](./embed.md)。手配 `YMBOT_*` / `mcp.json` 见文末附录，**非推荐**新人入场。

## 职责

为工作空间内的 **Bot 成员**（`members.is_bot=1`）提供：

- **角色（bot_roles）**：frontend / pm / qa / backend / wiki 等协作分工定义
- **认领（bot_role_claims）**：同一角色同时仅一个 active claim（`released_at IS NULL`）
- **记忆（bot_memories）**：按 `(workspace, member, key_name)` upsert 持久上下文
- **待办（bot_todos）**：Bot 自身 todo|doing|done|canceled
- **Handoff**：只读续工包（无独立表），组装 memories + 未完成 todos + 指派中的开放 Issues

Bot **不得**将 Issue 自动流转到 `done`（需人类确认）。

### 入场与三前缀（H4）

| 前缀 | 含义 |
|------|------|
| `ymt_` | 队席入场码（人从席页复制入场链接 / 提示词 / 入场包发出；链与提示词只带 `ymt_`） |
| `ymk_` | Bot 身份钥匙（join/redeem **无钥时自动签发**，当次回入队方；日常可转发提示词**禁止**嵌 `ymk_`） |
| `ymw_` | 人类工作区邀请（与 Bot 分轨；见 [`API.md`](./API.md)） |

**主路径（ONE_TAP · 席页一键）：**

1. 人类 Admin 浏览器建队 → 在 Bot **席页**点 **「复制入场链接」** 或 **「复制 Bot 提示词」**（裸 `ymt_` 入场包可留次要）。
2. 链接形如：`https://bot.api.sn.cn/join?code=ymt_…&role=…`（生产默认 API 根；自托管随 `PUBLIC_API_BASE`）。
3. 提示词内嵌：**一键安装**（主 `curl -fsSL https://bot.api.sn.cn/install.sh | bash`，备 `go install …/ymbot@latest`）+ API 根 + `ymt_` + role + 主航线：装（若无）→ join → `whoami` / `handoff get`。
4. Bot 粘贴提示词或打开链接 → 按需安装 → join → **自动得** `ymk_` → 开干。

公开 `/join` 只是说明页（安装 + 复制提示词）；redeem / 得钥在 CLI、通道或向导的**成功响应**。设置里 Admin 单把签 `ymk_` 是旁路，不是建队主路径。

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

**工作区隔离（XWS-AUTH · 通道 X-M1）：** 调用方非本仓成员 → **403**；path `memberID` ∉ 当前 ws → **404**（与 handoff 齐，**禁** 200 空当成功）。`ymbot`/MCP 对 403/404 人话透传，错误态不打印「(无记忆)」。member∈ws 硬门见 Go X-G1。

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

## 续命（服务端真相源）

怕干活中途中断来不及 export。短原则 [`BOT_CONTINUITY_P0.md`](./BOT_CONTINUITY_P0.md)；**验收 H1–H5** [`BOT_HANDOFF_CONTINUITY_P0.md`](./BOT_HANDOFF_CONTINUITY_P0.md)：

1. **真相源 = 服务端** handoff / memories / todos / 指派 Issue，**不是**本地文件。
2. **干活中自动写服务端**：开干就 `memory set` / todo；关键节点 `comment`；claim 后进度写记忆；**禁止**攒到最后才 `handoff export`。
3. **新 Bot 续工**：`ymbot whoami` → `ymbot handoff get`（**无文件**）即可；MCP：`board_whoami` → `board_handoff_get`。
4. **`handoff export`**：仅换机备份可选；日常续工不需要。

可选：`--autosave` 或 `YMBOT_AUTOSAVE=1` → claim / comment / memory set 成功后 stderr 提醒「已写服务端」（约定，非后台 daemon，非本地文件）。

## 续工流程建议

> **续命权威**：服务端 `GET .../bots/{me}/handoff`（`ymbot handoff get`）。详见 [`BOT_HANDOFF_CONTINUITY_P0.md`](./BOT_HANDOFF_CONTINUITY_P0.md)。  
> `export` / `import` 仅可选备份与救场，**不是**日常换机会话必经。

1. Bot 启动 → **`whoami` → `handoff get`**（无文件）恢复上下文（memories + todos + 开放 Issues）
2. 按指派 / 开放 `YM-n` 继续；只领叶子；禁自 `done`
3. **过程中写回服务端**：焦点/卡点 → `memory set`；步骤 → todo；对人类进展 → `comment create`
4. 断线或换会话：同一 API Key 再 `handoff get` 即可续——**不必**先持有导出文件
5. （可选）冷备：`handoff export --out file.json`；灾难恢复才 `handoff import`


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

**入场先走席页一键**（上节「复制入场链接 / 复制 Bot 提示词」）；装好后日常命令：

```bash
ymbot whoami
ymbot issue list --assignee me          # 默认文本表：YM-n / title / status…
ymbot issue get YM-12                   # 入参接受 YM-n 或 UUID
ymbot comment list --issue YM-12        # 默认摘要；--full 才全量
ymbot comment create --issue YM-12 -m "进展…" --idempotency-key <uuid>
ymbot memory list                       # key + 值摘要
ymbot memory set --key last_focus --value "…" --idempotency-key <uuid>
ymbot handoff get                       # 续工：服务端摘要（无文件）
ymbot handoff export --out /tmp/handoff.json   # 换机备份可选；日常不需要
```

通道 DX（省 token / 少 UUID）：[`BOT_CHANNEL_DX_P0.md`](./BOT_CHANNEL_DX_P0.md)。默认无 32 位 hex；`--id` 附加裸 id；`--full` 接近 REST。Issue 可读名：优先嵌套 `assignee.display_name`，否则 `members` 映射 `assignee_id`；入参 `YM-n`→id 由通道解析。

共享客户端：`internal/botclient`（Bearer；Idempotency-Key：CLI/MCP 参数 → env → 自动 UUID；一行中文错误含吊销/无效；Bot 禁 `status=done`）。
干活鉴权：入场后的 **`ymk_`**（Bearer）；人用浏览器 JWT。设置里单把签钥是旁路。

## 附录 · env / MCP（非推荐）

> **主路径用席页一键链接/提示词。** 手配 `export YMBOT_*` / 手改 `mcp.json` / 长安装文档 **仅救场或自托管**，**非推荐**新人入场。

本地调试或自托管救场时，可显式导出（生产 Bot 应由席页提示词带好 API 根与 join，勿把下列清单当 onboard）：

```bash
export YMBOT_API_URL=http://127.0.0.1:8080   # 或 https://bot.api.sn.cn
# 空手：先 join 拿自动签发的 ymk_；续队/换机：导出已有钥匙
# ymbot team join --code ymt_*** --role frontend
export YMBOT_API_KEY=ymk_***
export YMBOT_WORKSPACE_ID=<ws>
# 可选：YMBOT_MEMBER_ID（缺省 GET …/bots/me）
# 可选：YMBOT_IDEMPOTENCY_KEY=<uuid>

ymbot mcp-serve   # Cursor stdio MCP；tools: board_*；full/ids 同 --full/--id
```

MCP 通道能力与工具表见 [`BOT_MCP_CLI.md`](./BOT_MCP_CLI.md)，勿再以「先配齐 env 再 mcp-serve」当主入场叙事。

## 边界

- Handoff 无独立表；由 service 组装
- Issue 状态机 / 看板 UI 在其它域；Admin 单把 `ymk_` CRUD 在设置旁路
- Bot 禁止自动 `done` Issue
- claim 冲突返回 `conflict`；记忆按 key upsert

## 相关

- Bot 工作流 MCP/CLI 最小能力规格：[`BOT_MCP_CLI.md`](./BOT_MCP_CLI.md)
- 续命短原则：[`BOT_CONTINUITY_P0.md`](./BOT_CONTINUITY_P0.md)
- 续命验收 H1–H5：[`BOT_HANDOFF_CONTINUITY_P0.md`](./BOT_HANDOFF_CONTINUITY_P0.md)

- 人用入场锚点：[`NEXT_WAVE_DEBT_HUMAN_ONBOARD.md`](./NEXT_WAVE_DEBT_HUMAN_ONBOARD.md) §H4
- 一键入场文案债（D1）：[`NEXT_WAVE_DEBT_ONE_TAP_JOIN.md`](./NEXT_WAVE_DEBT_ONE_TAP_JOIN.md) §D1
- 班底席位 / 入场码：[`BOT_TEAM_SEAT_P0.md`](./BOT_TEAM_SEAT_P0.md) · [`BOT_TEAM_JOIN_KEY_P0.md`](./BOT_TEAM_JOIN_KEY_P0.md)（`seat.status`=`open|active|stopped|left`；`GET/POST …/teams/{id}/invites`）
- **T7-3 执行源（已落地 `d9e5a3c`）**：席位作业手册 = `playbook_ref`；roster / bind / handoff 回包含该字段；空或非法 → **400** 中文；升版 T7-4/5 仍后置。开工单 [`NEXT_WAVE_DEBT_T73.md`](./NEXT_WAVE_DEBT_T73.md)（通道 MCP 人话投影见记忆席进度）
- **Issue 归档（实现待 SHA）**：`POST …/issues/archive`；`issue list` 默认无已归档；`include_archived` / `archived=only`（债 [`NEXT_WAVE_DEBT_ISSUE_ARCHIVE.md`](./NEXT_WAVE_DEBT_ISSUE_ARCHIVE.md)）

