# API 概览（核心域）

基础路径：`/api/v1`  
错误体（顶层）：`{"code":"string","message":"中文说明"}`  
鉴权：`Authorization: Bearer <jwt|ymk_…>` 或 httpOnly cookie `session`  
JSON 字段：`snake_case`；时间：UTC ISO8601  

### 竞态 / 乐观锁（Issue）

| HTTP | `code` | 何时 |
|------|--------|------|
| **400** | `bad_request` | PATCH/move **未带**或非法 `version` |
| **409** | `conflict` | 请求 `version` ≠ 库中当前值（他人已更新）；message 含「乐观锁冲突…请刷新后重试」 |

成功更新后 `version` 自增；客户端须用响应中的新 `version` 再写。其它 409（成环、slug、claim）同 `code=conflict`，靠 message 区分。

完整路由清单与分工见根目录 `COORDINATION.md`。Wiki/工时见 [`wiki-and-time.md`](./wiki-and-time.md)（`AuthJWTOrAPIKey` + 工作区成员）。下文为 **Auth / Workspace / Members / Labels / Projects / API Keys / Webhooks**（本域已实现）。Issue/Cycle/Comments 等由 Issue 域同事实现。

## Auth

| Method | Path | 说明 |
|--------|------|------|
| POST | `/auth/register` | `{email,password,name?}` → `{token,user}` |
| POST | `/auth/login` | 同上；有登录限流 |
| POST | `/auth/logout` | 清除 session cookie |
| GET | `/auth/me` | 当前用户（JWT） |

## Workspaces

| Method | Path | 说明 |
|--------|------|------|
| GET | `/workspaces` | 我的工作空间列表 |
| POST | `/workspaces` | 创建（创建者成为 admin） |
| GET | `/workspaces/current` | 第一个工作空间 |
| GET | `/workspaces/{workspaceID}` | 详情（须为成员） |

也可传 Header `X-Workspace-Id`（与路径二选一，路径优先）。

## Members / Labels

| Method | Path | 说明 |
|--------|------|------|
| GET | `/workspaces/{ws}/members` | 成员列表 |
| POST | `/workspaces/{ws}/members` | Admin 邀请（已注册邮箱） |
| PATCH/DELETE | `/workspaces/{ws}/members/{memberID}` | Admin |
| GET/POST | `/workspaces/{ws}/labels` | 标签 |
| PATCH/DELETE | `/workspaces/{ws}/labels/{labelID}` | |

## Projects

| Method | Path | 说明 |
|--------|------|------|
| GET/POST | `/workspaces/{ws}/projects` | 列表 / 创建（`name`,`key_prefix`） |
| GET/PATCH/DELETE | `/projects/{projectID}` | 详情；PATCH 可 `{name,description,archive}`；删/归档需 Admin |
| GET/PATCH/DELETE | `/workspaces/{ws}/projects/{projectID}` | 同上（workspace 作用域挂载） |

`?include_archived=1` 可列出已归档项目。

## API Keys（Admin）

| Method | Path | 说明 |
|--------|------|------|
| GET/POST | `/workspaces/{ws}/api-keys` | 创建返回一次明文 `token`（`ymk_…`），库内 sha256 |
| DELETE 或 POST …/revoke | `/workspaces/{ws}/api-keys/{keyID}` 或 `/api-keys/{keyID}` | 吊销 |

Bot 使用：`Authorization: Bearer ymk_…` → `actor_kind=bot`，绑定 `member_id`。

## Webhooks（Admin）

| Method | Path | 说明 |
|--------|------|------|
| GET/POST | `/workspaces/{ws}/webhooks` | 创建返回一次 `secret` |
| PATCH/DELETE | `/workspaces/{ws}/webhooks/{id}` | |
| GET | `/workspaces/{ws}/webhooks/{id}/deliveries` | 投递日志 |
| POST | `/workspaces/{ws}/issues/{issueID}/notify` | 同步扇出 `issue.assigned`（N4 立即推送）；返回 delivered/failed |

投递头：`X-Signature-256: sha256=<hmac hex>`，`X-Webhook-Event`。

## Issue 域约定（由同事实现，此处仅契约）

建议路径（以实际挂载为准）：

- `GET/POST /workspaces/{ws}/projects/{id}/issues` 或 `/workspaces/{ws}/issues`
- `GET/PATCH /issues/{id}`；`POST /issues/{id}/move` `{status,board_order,version}`
- Relations 防环 → `409 conflict`；乐观锁见上文「竞态 / 乐观锁」
- Cycles / Comments / Activities / Search

**T6-I `team_id`（Issue 池按队 · 已落地 `75aadce`）**：`POST` 创建 **必须**带 `team_id`（body 或 `?team_id=`；缺/错 → 4xx 中文）。`GET` list/board/search 传 `?team_id=<当前队>` 时 **排除** `team_id IS NULL`（仅本队）；`?unassigned=1` 仅未分队旧单。错队 `GET/PATCH ?team_id=` → **404**。详见 [`NEXT_WAVE_DEBT_T6I.md`](./NEXT_WAVE_DEBT_T6I.md)。

写操作建议支持 Header `Idempotency-Key`。

### Issue · `team_id`（T6-I · 已落地 `75aadce`）

> 权威开工单：[`NEXT_WAVE_DEBT_T6I.md`](./NEXT_WAVE_DEBT_T6I.md)。**已落地**：功能 `fc8a446` · List 默认排除 NULL **`75aadce`**。`issues.team_id` 可空 + FK；Create 必带；List/board/search 按队过滤；I1/I2 单测绿。

| | |
|--|--|
| **列** | `issues.team_id` 可空，`REFERENCES teams(id)`；索引 `(workspace_id, team_id)`；PG `010_issues_team_id.sql` |
| **JSON** | Issue 对象可选 `team_id`（`snake_case`；未分队为 `null`） |
| **新建** | `POST …/issues` **必填** `team_id`（body 或 `?team_id=`）；缺/错/非本 WS 队 → **4xx** 中文 |
| **列表** | `GET …/issues?team_id=<uuid>`：**仅** `team_id=` 匹配行（**排除** `team_id IS NULL`）；`?unassigned=1` → **仅**未分队旧单（`team_id IS NULL`） |
| **NULL 策略（已裁定 · `75aadce`）** | 旧行 `NULL` = **未分队**；带 `team_id` 的 list **排除** NULL；未分队池只用 `?unassigned=1` |
| **读写隔离** | 错队 `GET` / `PATCH`（含改 `team_id` 跨队）→ **404**；不跨 workspace |
| **Update** | `PATCH` 可写 `team_id`（本 WS 内）；跨队目标 → **404** |
| **实现 SHA** | **`75aadce`**（List 排除 NULL）· 功能基线 `fc8a446` |

**不做（本包）**：T7；pins 返工；FE 全局切换器大改；历史单批量搬家 UI。

## Teams / Seats / Invites（Bot 班底）

鉴权同工作区：`AuthJWTOrAPIKey` + 成员；签发/列表入场码需 **Admin**。

| Method | Path | 说明 |
|--------|------|------|
| GET/POST | `/workspaces/{ws}/teams` | 团队列表 / 创建 |
| GET | `/workspaces/{ws}/teams/{teamID}` | 详情 |
| GET/POST | `/workspaces/{ws}/teams/{teamID}/roster` · `…/roster/ensure` | 班底席位；ensure 幂等 |
| POST | `/workspaces/{ws}/teams/{teamID}/seats` | 加席；`status` 初始 `open` |
| POST | `…/seats/{seatID}/bind` · `unbind` · `stop` | 绑定 / 解绑→`left` / 停工→`stopped` |
| POST | `…/seats/expand` · `…/expand` | 给空闲席派叶子 Issue（≠加新席） |
| POST | `/workspaces/{ws}/teams/{teamID}/invites` | Admin 签发 `ymt_`；**明文 `code` 仅此一次**；body 字段名 `allow_expand_seat` |
| GET | `/workspaces/{ws}/teams/{teamID}/invites` | Admin 列已签发；**无**明文 code / hash；含 `status`/`use_count`/`expires_at`/`code_tail` 等 |
| POST | `/workspaces/{ws}/teams/{teamID}/invites/{inviteID}/revoke` | Admin 吊销；幂等已吊销仍 200 + `status=revoked` |
| POST | `/team-invites/redeem` | `{code}` → `{workspace_id,team_id,allow_expand_seat,seat_slug?,role_hint?}`（不在 ws 路径下） |

**seat.status**：`open` | `active` | `stopped` | `left`（空串读出按 `open`）。`bind`/`claim`/enter 成功后写 **`active`**；活跃席 `bind(force)` 硬拒；force 仅 `stopped|left`（及脏绑）。**force 成功**另写人对人可见审计：`activities.action=seat.force_bind`（payload 人话 `comment`/`summary`）+ 关联开放 Issue **comment**（原/新占用者 assignee）。

**invite.status**（派生、不落库）：`active` | `revoked` | `expired` | `exhausted`。

**隔离（T6-R 最小）**：roster / seats / bind / expand 均以 path `teamID` 为界，跨 `team_id` 不串 claim；通道切队 = 换 `team_uuid`（见 [`BOT_MCP_CLI.md`](./BOT_MCP_CLI.md) §11.4）。**已落地（见上「Issue · team_id」· [`NEXT_WAVE_DEBT_T6I.md`](./NEXT_WAVE_DEBT_T6I.md)）**：Issue 列表/创建按 `team_id` 隔离（`75aadce`：`?team_id=` **排除** NULL；`?unassigned=1` 仅未分队）。**`team_pins` REST 已落地**（见下「Team pins」· `6915db4`）；MCP/CLI `board_pin_*` 已落地（[`BOT_MCP_CLI.md`](./BOT_MCP_CLI.md) §11.5 · `63e69c6`）。全文边界：[`BOT_TEAM_TEMPLATE_T6_T7.md`](./BOT_TEAM_TEMPLATE_T6_T7.md) · 开工单：[`NEXT_WAVE_DEBT_PINS.md`](./NEXT_WAVE_DEBT_PINS.md)。

详见 [`BOT_TEAM_JOIN_KEY_P0.md`](./BOT_TEAM_JOIN_KEY_P0.md) · [`BOT_TEAM_SEAT_P0.md`](./BOT_TEAM_SEAT_P0.md)。

## Team pins（共享记忆墙）

> **状态：REST + MCP/CLI 已落地。** REST `6915db4`；通道 `board_pin_*` / `ymbot pin` · `63e69c6`。验收见 [`NEXT_WAVE_DEBT_PINS.md`](./NEXT_WAVE_DEBT_PINS.md) P1–P3（Go）· P5 已勾；P4 另席。  
> **不是** 席位私有 `bot_memories`；同名 `key` 分表不串、互不覆盖。

表：`team_pins`，`UNIQUE(team_id, key)`。字段至少：`id, workspace_id, team_id, key, value, updated_by, updated_at`（JSON `snake_case`）。

鉴权：`AuthJWTOrAPIKey` + **本队成员（或 workspace Admin）**。错队 / 无队 / 非成员 → **404**（不泄露他队是否存在该 key）。不跨 workspace。

| Method | Path | 说明 |
|--------|------|------|
| GET | `/workspaces/{ws}/teams/{teamID}/pins` | 列**本队** pins → `{pins:[]}` |
| GET | `/workspaces/{ws}/teams/{teamID}/pins/{key}` | 单条；本队无此 key → **404**（`pin 不存在`） |
| PUT | `/workspaces/{ws}/teams/{teamID}/pins` | upsert by `key`；body `{key, value}` → `{pin}`；`updated_by` = 当前 actor |
| PUT | `/workspaces/{ws}/teams/{teamID}/pins/{key}` | 同上别名；path `key` 优先于 body |
| DELETE | `/workspaces/{ws}/teams/{teamID}/pins/{key}` | 按 key 删；成功 `{ok:true}`；本队无此 key → **404**；错队/无队 → **404**（`团队不存在或无权访问`） |

写操作建议支持 Header `Idempotency-Key`（与其它 workspace 写路由一致）。

**不做（pins 包）**：用 pin 替代 handoff；FE 全局团队切换器大改；跨 workspace 墙。

## 数据库

- 本地默认：`DB_DRIVER=sqlite`，打开时自动 migrate `schema.sql`。
- 生产 / 向量：`DB_DRIVER=postgres` + **pgvector**；迁移步骤见根目录 `.env.example` 与 [`embed.md`](./embed.md)。

## 种子账号

- Email: `admin@yumeng.tech`
- Password: `Admin123!`
- Workspace: `育蒙科技`（slug `yumeng`）
- 示例项目前缀：`YM`

```bash
cd backend && go run ./cmd/server   # :8080，默认 AUTO_SEED=1
cd backend && go run ./cmd/seed
cd backend && go test ./... -count=1
```

### curl 示例

```bash
# 登录
curl -s -X POST http://127.0.0.1:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@yumeng.tech","password":"Admin123!"}'

# 列项目（替换 TOKEN / WS）
curl -s http://127.0.0.1:8080/api/v1/workspaces/$WS/projects \
  -H "Authorization: Bearer $TOKEN"

# 创建 API Key（明文仅一次）
curl -s -X POST http://127.0.0.1:8080/api/v1/workspaces/$WS/api-keys \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"grok"}'
```

## 附件（Attachments）

Issue / 评论可上传图片与文档；文件落本地磁盘（`ATTACHMENTS_DIR`，默认 `./data/attachments`），元数据在 `attachments` 表。

鉴权：与其它工作区路由相同——`AuthJWTOrAPIKey` + 工作区成员。**无公开永久 URL**；`url` 字段为相对路径，前端请求时仍须带 `Authorization: Bearer <jwt|ymk_…>`。

### HTTP

| Method | Path | 说明 | Owner |
|--------|------|------|-------|
| POST | `/workspaces/{ws}/issues/{issueID}/attachments` | `multipart/form-data` 字段 `file`（可选 `filename`）→ `{attachment}` | Go 附件 |
| POST | `/workspaces/{ws}/issues/{issueID}/comments/{commentID}/attachments` | 同上，挂到评论 | Go 附件 |
| GET | `/workspaces/{ws}/attachments/{attachmentID}` | 元数据 `{attachment}` | Go 附件 |
| GET | `/workspaces/{ws}/attachments/{attachmentID}/content` | 文件流；图片 `inline`，其它 `attachment` | Go 附件 |
| GET | `/workspaces/{ws}/issues/{issueID}/attachments` | `{attachments:[]}`（含评论附件） | Issue 域 |
| GET | `/workspaces/{ws}/issues/{issueID}/comments/{commentID}/attachments` | `{attachments:[]}` | Issue 域 |
| DELETE | `/workspaces/{ws}/attachments/{attachmentID}` | 上传者或 Admin；`{ok:true}` | Issue 域 |

可复用包：`internal/attachments`（`UploadIssue`/`UploadComment`/`Get`/`OpenContent`/`ListByIssue`/`ListByComment`/`CanDelete`/`Delete`）。

### Attachment JSON

`id, workspace_id, issue_id, comment_id?, uploader_id, filename, content_type, size_bytes, created_at, url`

- `url` 例：`/api/v1/workspaces/{ws}/attachments/{id}/content`
- `comment_id`：Issue 级附件为省略/空；评论附件为评论 ID

### 限制

- 单文件 ≤ **10MB**；超限 `message`：`文件过大，上限 10MB`
- 允许：`png/jpg/jpeg/gif/webp/pdf/docx` 以及 `txt/md/csv`
- 拒绝：可执行等 → `不支持的文件类型`；校验失败不落库、不留盘文件
- 删除策略（service）：**上传者本人**或 **workspace admin**；Bot 以绑定 `member_id` / `role` 计

