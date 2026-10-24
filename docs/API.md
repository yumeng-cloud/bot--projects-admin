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
| GET | `/auth/me` | 当前用户（JWT）；`user.is_platform_admin` bool（已落地） |

## Workspaces

| Method | Path | 说明 |
|--------|------|------|
| GET | `/workspaces` | 我的工作空间列表（已落地） |
| POST | `/workspaces` | 创建（JWT；创建者 → WS `admin`；body `{name, slug?}`；已落地） |
| GET | `/workspaces/current` | 第一个工作空间（已落地） |
| GET | `/workspaces/{workspaceID}` | 详情（须为成员；已落地） |
| PATCH | `/workspaces/{workspaceID}` | WS Admin 改 `{name?, slug?}`；slug 规则与唯一（已落地） |

也可传 Header `X-Workspace-Id`（与路径二选一，路径优先）。

生产开通主路径：**注册 / 登录 → `POST /workspaces`**（无需 seed）。本地演示仍可用 `make seed` / `AUTO_SEED`。

## Members / Labels

| Method | Path | 说明 |
|--------|------|------|
| GET | `/workspaces/{ws}/members` | 成员列表（已落地） |
| POST | `/workspaces/{ws}/members` | WS Admin：`{email, role}` 直接入席（**仅已注册**邮箱；已落地） |
| PATCH | `/workspaces/{ws}/members/{memberID}` | WS Admin：`{role?, display_name?}`；不可降级最后一个 admin → **409**（已落地） |
| DELETE | `/workspaces/{ws}/members/{memberID}` | WS Admin 移除；不可摘掉最后一个 admin → **409**（已落地） |
| GET/POST | `/workspaces/{ws}/labels` | 标签 |
| PATCH/DELETE | `/workspaces/{ws}/labels/{labelID}` | |

### Workspace invites（人类邀请链 · `ymw_` · 已落地 `05a1d35`）

> **勿与**队席入场码 `ymt_`（`POST …/teams/{id}/invites` · `/team-invites/redeem`）混淆。本段 = **工作区人类成员**邀请链接。

| Method | Path | Body / 响应 | 鉴权 | 说明 |
|--------|------|-------------|------|------|
| POST | `/workspaces/{ws}/invites` | body `{email?, role, expires_at?}` → `{invite, code}`（**明文 `code` 仅此一次**） | WS Admin | 未注册邮箱亦可签发；对方注册/登录后 redeem |
| GET | `/workspaces/{ws}/invites` | `{invites:[{id,email?,role,status,code_tail,use_count,expires_at,…}]}`（**无**明文 code/hash） | WS Admin | 列表 |
| POST | `/workspaces/{ws}/invites/{inviteID}/revoke` | `{invite}` / `{ok:true}`；已吊销幂等 200 | WS Admin | 吊销 |
| POST | `/workspace-invites/redeem` | `{code}` → `{workspace_id, role, member_id?}` | JWT（已登录用户） | 不在 ws 路径下；入席当前用户 |

**字段 / 错误（约定）**

| 项 | 约定 |
|----|------|
| `role` | `admin` \| `member` |
| `status`（派生） | `active` \| `revoked` \| `expired` \| `exhausted` |
| 未知/非法 code | **404**；已吊销/过期/用尽 | **400** 中文；`ymt_` 前缀 → **400** |
| 非 WS Admin 签发/列表/吊销 | **403** |
| 非成员访问他 WS | **404**（不泄露） |

### Platform admin（全站超管 · 已落地 `05a1d35`）

> Bootstrap：环境变量 `PLATFORM_ADMIN_EMAILS`（逗号分隔；见 [`.env.example`](../.env.example)）。冷启将匹配邮箱用户授 `users.is_platform_admin`；亦可 seed 升旗。非超管打 `/admin/*` → **403** 中文。

| Method | Path | Body / 响应 | 说明 |
|--------|------|-------------|------|
| GET | `/admin/workspaces` | `{workspaces:[{id,name,slug,member_count,disabled_at?,created_at,…}]}` | 全站列表 |
| POST | `/admin/workspaces` | `{name, slug?, owner_email\|owner_user_id}` → `{workspace,…}` | 代开；owner → 该 WS admin |
| PATCH | `/admin/workspaces/{id}` | `{disabled:true\|false}` | 停用/恢复；停用后该 WS 成员业务 API → **403**「工作区已停用」 |
| GET | `/admin/users` | `{users:[…]}`（含 `is_platform_admin`、所属 WS 摘要） | 全站用户 |
| PATCH | `/admin/users/{id}` | `{is_platform_admin}` | 授/撤平台超管 |

管理操作写入 `platform_audit_logs`（停用/恢复/代开/授撤超管；已落地）。邀请码前缀 **`ymw_`**（与队席 `ymt_` 隔离）。

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
| GET/POST | `/workspaces/{ws}/teams` | 团队列表 / 创建（`template_key`；T7-R 见下） |
| GET | `/workspaces/{ws}/teams/{teamID}` | 详情 |
| GET/POST | `/workspaces/{ws}/teams/{teamID}/roster` · `…/roster/ensure` | 班底席位；ensure 幂等（可带 `playbook_ref`；T7-R 见下） |
| POST | `/workspaces/{ws}/teams/{teamID}/seats` | 加席；`status` 初始 `open` |
| POST | `…/seats/{seatID}/bind` · `unbind` · `stop` | 绑定 / 解绑→`left` / 停工→`stopped` |
| POST | `…/seats/expand` · `…/expand` | 给空闲席派叶子 Issue（≠加新席） |
| POST | `/workspaces/{ws}/teams/{teamID}/invites` | Admin 签发 `ymt_`；**明文 `code` 仅此一次**；body 字段名 `allow_expand_seat` |
| GET | `/workspaces/{ws}/teams/{teamID}/invites` | Admin 列已签发；**无**明文 code / hash；含 `status`/`use_count`/`expires_at`/`code_tail` 等 |
| POST | `/workspaces/{ws}/teams/{teamID}/invites/{inviteID}/revoke` | Admin 吊销；幂等已吊销仍 200 + `status=revoked` |
| POST | `/team-invites/redeem` | `{code}` → `{workspace_id,team_id,allow_expand_seat,seat_slug?,role_hint?}`（不在 ws 路径下） |

### Team templates（T7-R · 已落地 `716c3d0`）

> 权威开工单：[`NEXT_WAVE_DEBT_T7.md`](./NEXT_WAVE_DEBT_T7.md)（契约基线 `0999576` / 债文 `95e27c6`）。手册索引：[`BOT_TEAM_TEMPLATE_INDEX.md`](./BOT_TEAM_TEMPLATE_INDEX.md)。  
> **本波仅 R1–R2**（选模板 → 空席 + `playbook_ref` + 错误码）。**升版 dry-run/apply（T7-4/5）后置**，勿当本波必验。  
> 验收对齐 [`BOT_TEAM_TEMPLATE_T6_T7.md`](./BOT_TEAM_TEMPLATE_T6_T7.md) **T7-1 / T7-2 / T7-3 / T7-6 / T7-7**（非整表 T7）。

**现状（已落地 `716c3d0`）**：内置 `default` 嵌入；`GET …/team-templates` 可读清单；`POST …/teams` / `…/from-template` 带 `template_key` 幂等建空席（slug/显示名/`playbook_ref`，**无**自动绑 Key）；`POST …/roster/ensure` 仅 `template_key`、空 `seats` → 按模板展开。未知模板 / 非法 `playbook_ref` → **400** 中文。

| Method | Path | Body / Query | 响应 | 说明 |
|--------|------|--------------|------|------|
| GET | `/workspaces/{ws}/team-templates` | — | `{templates:[{key,name,description?,seats:[{slug,display_name,playbook_ref}]}]}` | **已落地 `716c3d0`**：可读内置清单（现网 `default`；`bot-team-template` 等键仍见索引约定） |
| POST | `/workspaces/{ws}/teams` | `{name, slug?, template_key}` | `{team}`（含 `template_key`） | 已有创建；T7-R：**从模板创建**时 `template_key` 必填且须为清单内 key（`716c3d0`） |
| POST | `/workspaces/{ws}/teams/{teamID}/roster/ensure` | `{template_key?, seats?:[{slug,name,playbook_ref}]}` | `{roster: seats[]}` | 已有幂等；T7-R：仅 `template_key`、空 `seats` → 按索引展开空席+`playbook_ref`（**已落地 `716c3d0`**）；显式 `seats` 仍可用 |
| POST | `/workspaces/{ws}/teams/from-template`（可选别名） | `{name, slug?, template_key}` | `{team, roster}` | **已落地 `716c3d0`**：创建+ensure 一次完成；亦可客户端串两步 |

**字段**

| 字段 | 位置 | 约定 |
|------|------|------|
| `template_key` | team / ensure body | 如 `default`、`bot-team-template`；未知 → **400** 人话 |
| `role_slug` / `slug` | seat | 短 slug（`frontend`…）；与手册席位一致 |
| `display_name` / `name` | seat | 显示名 |
| `playbook_ref` | seat | 字符串 `repo@ref:path`（见索引）；空席创建后**未**绑 Key（`status=open`）；**T7-3 执行源**（`d9e5a3c`）：席位唯一作业手册入口；`GET roster` / `bind` / `GET …/handoff` 回包含该字段；写入/绑定时空或非法 → **400** 中文（不默默换硬编码手册） |

**错误（R2）**

| 情况 | HTTP |
|------|------|
| 未知 / 空（从模板路径）`template_key` | **400** 中文 |
| 空 / 非法 `playbook_ref`（无法解析 path；T7-3 · `d9e5a3c`） | **400** 中文 |
| 团队不存在 / 非本 WS | **404** |
| 他队 path | **404**（不泄露） |

**不做（本波）**：升版 dry-run diff / apply；自动绑 Key；模板写回用户 git；静默覆盖 memory。

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

