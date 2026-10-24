# 协作分工 · AI 协作看板

Tech lead 已落地 **共享地基**。各 Bot 按下方所有权并行开发，**禁止大面积改动他人目录/路由**；跨模块改动先在本文件或 PR 说明里同步。

## 仓库路径

`/workspace/ai-collab-board`

## 目录所有权

| 路径 | Owner | 说明 |
|------|--------|------|
| `backend/internal/db/` schema 核心表 + migrate | **Tech lead (冻结)** | 新增表可 append；改已有列需同步全员 |
| `backend/internal/auth/` `middleware/` `config/` `pkg/` | **Tech lead (冻结)** | JWT / argon2id / CORS / ratelimit |
| `backend/internal/handlers/auth.go` `health.go` `workspace.go` | **Tech lead** | 仅扩展，勿重写 |
| `backend/internal/handlers/projects*.go` `issues*.go` `cycles*.go` | **Issue/Board Bot** | 项目、Issue、状态机、看板、关联、搜索 |
| `backend/internal/handlers/comments*.go` `activity*.go` `labels*.go` `members*.go` | **Collab Bot** | 评论、活动、标签、成员 |
| `backend/internal/handlers/apikeys*.go` `webhooks*.go` | **Integrations Bot** | API Key、Webhook HMAC + delivery log |
| `backend/internal/handlers/wiki*.go` | **Wiki Bot** | 知识库 spaces/pages CRUD + 搜索 |
| `backend/internal/handlers/time*.go` `billing*.go` | **Time Bot** | 工时、费率、报表、CSV |
| `backend/internal/handlers/bot_*.go` | **Bot Continuity Bot** | roles/claims、memory、todos、handoff |
| `backend/internal/seed/` | **Tech lead + 各域追加** | 追加 seed 用独立函数，保持幂等 |
| `frontend/src/pages/Login*` `Home*` `api/client.ts` | **Tech lead** | 登录壳 |
| `frontend/src/pages/Project*` `Board*` `Issue*` `Cycle*` | **Issue/Board FE** | |
| `frontend/src/pages/Wiki*` | **Wiki FE** | |
| `frontend/src/pages/Time*` `Reports*` | **Time FE** | |
| `frontend/src/pages/Bot*` `Settings*` | **Bot/Settings FE** | |
| `docs/BOT.md` `docs/API.md` | 各域追加章节 | Tech lead 维护目录 |

## `/api/v1` 路由预留（已实现 / 待实现）

### 已实现（地基）

| Method | Path | 说明 |
|--------|------|------|
| GET | `/healthz` `/api/v1/healthz` | 健康检查 |
| POST | `/api/v1/auth/register` | 注册 |
| POST | `/api/v1/auth/login` | 登录（有速率限制） |
| POST | `/api/v1/auth/logout` | 退出 |
| GET | `/api/v1/auth/me` | 当前用户（JWT） |
| GET | `/api/v1/workspaces` | 我的工作空间 |
| GET | `/api/v1/workspaces/{workspaceID}` | 工作空间详情 |

### 预留 · Issue / Board

| Method | Path |
|--------|------|
| CRUD | `/api/v1/workspaces/{ws}/projects` `.../projects/{id}` |
| CRUD + transition | `/api/v1/workspaces/{ws}/issues` `.../issues/{id}` `.../issues/{id}/transition` |
| CRUD | `/api/v1/workspaces/{ws}/projects/{id}/cycles` |
| GET | `/api/v1/workspaces/{ws}/search?q=` |
| POST/DELETE | `/api/v1/workspaces/{ws}/issues/{id}/relations` |

### 预留 · Collab

| Method | Path |
|--------|------|
| CRUD | `.../issues/{id}/comments` |
| GET | `.../issues/{id}/activities` · `.../activities` |
| CRUD | `.../labels` · issue label attach |
| CRUD | `.../members` |

### 预留 · Integrations

| Method | Path |
|--------|------|
| CRUD | `.../api-keys`（创建明文仅创建时返回） |
| CRUD | `.../webhooks` · `.../webhooks/{id}/deliveries` |

### 预留 · Wiki

| Method | Path |
|--------|------|
| CRUD | `.../wiki/spaces` · `.../wiki/pages` |
| GET | `.../wiki/search?q=` |

### 预留 · Time / Billing

| Method | Path |
|--------|------|
| CRUD | `.../time-entries` |
| GET | `.../time-entries/report` · `.../time-entries/export.csv` |
| CRUD | `.../billing-rates` |

### 预留 · Bot continuity

| Method | Path |
|--------|------|
| CRUD | `.../bot-roles` · `.../bot-roles/{id}/claim` · `.../release` |
| CRUD | `.../bots/{memberID}/memories` · search |
| CRUD | `.../bots/{memberID}/todos` |
| GET | `.../bots/{memberID}/handoff` |

## 约定

1. **鉴权**：人类 JWT（`Authorization: Bearer` 或 httpOnly `session` cookie）。地基已提供 `middleware.AuthJWT` 与 `AuthJWTOrAPIKey`（Bot Key 前缀 `ymk_`，校验 `api_keys.key_hash`）。API Key CRUD 仍由 Integrations Bot 实现。
2. **错误体**（顶层，非嵌套）：`{"code":"...","message":"中文"}`
3. **Bot 写操作**：支持 `Idempotency-Key`；Issue **PATCH**（及看板 move）必须带当前 `version`（乐观锁）。
   - 缺 `version` / 非法 → **400** `{"code":"bad_request","message":"…"}`。
   - 版本落后（他人已更新）→ **409** `{"code":"conflict","message":"乐观锁冲突：Issue 已被他人更新，请刷新后重试"}`（`apperr.Conflict`）；客户端刷新后带新 `version` 重试，禁止静默覆盖。
   - 其它业务冲突（slug 占用、依赖成环、角色已认领等）同为 **409** + `code=conflict`，以 `message` 区分。
4. **时区展示**：Asia/Shanghai；DB 存 UTC RFC3339。
5. **状态枚举**：`backlog|todo|in_progress|in_review|done|canceled`；Bot **不得**自动转到 `done`。
6. **Schema / 迁移**：`backend/internal/db/schema.sql` 已含 wiki / time / bot_* 等；对应 Bot 直接用，勿重复建表。`DB_DRIVER=sqlite` 时 `db.Open` 自动执行该 schema。`DB_DRIVER=postgres` 时须自行：`CREATE EXTENSION vector` → 应用 schema（方言适配）→ 应用 [`docs/embed.md`](./docs/embed.md) 中 `embedding_chunks`；细节与 env 见 `.env.example`。
7. **搜索**：当前 `?q=` 为关键词/模糊（title/body 等，本地 SQLite 可跑）。**向量 / 语义召回**的生产方案是 **PostgreSQL + pgvector**（见 [`docs/embed.md`](./docs/embed.md)）。**勿**把 sqlite-vec 写成生产方案；本地 SQLite 仅开发用，不做向量扩展依赖。
8. 路由在 `cmd/server/main.go` 注册；替换 stub 时删除对应 501 占位。

## Seed 账号

- Email: `admin@yumeng.tech`
- Password: `Admin123!`
- Workspace slug: `yumeng`

服务启动默认 **不** AUTO_SEED（设 `AUTO_SEED=1` 才跑结构种子）。演示项目/Issue/标签另需 `SEED_SAMPLE_DATA=1`；生产保持两者关闭。

## UI bar
对标 Linear：克制、紧凑、细边框；由「UI 视觉」主导主题，React/看板配合。

## 相关规格

- Bot 工作流 MCP/CLI 最小能力：[`docs/BOT_MCP_CLI.md`](./docs/BOT_MCP_CLI.md)（产品体验；落地见文内拆分；**§11** help/import/Inspector）
- 班底 T6/T7 最小说明：[`docs/BOT_TEAM_TEMPLATE_T6_T7.md`](./docs/BOT_TEAM_TEMPLATE_T6_T7.md)
- 向量嵌入 / pgvector：[`docs/embed.md`](./docs/embed.md)

- 全局排期 Timeline P0（**已裁定**，可开工）：[`docs/SCHEDULING_BOARD_P0.md`](./docs/SCHEDULING_BOARD_P0.md)
- 成品套装 P0（包装视图 / 列模板 / 与排期·抽屉一体）：[`docs/PRODUCT_SUITE_P0.md`](./docs/PRODUCT_SUITE_P0.md)
