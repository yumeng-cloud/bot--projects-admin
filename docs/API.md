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

投递头：`X-Signature-256: sha256=<hmac hex>`，`X-Webhook-Event`。

## Issue 域约定（由同事实现，此处仅契约）

建议路径（以实际挂载为准）：

- `GET/POST /workspaces/{ws}/projects/{id}/issues` 或 `/workspaces/{ws}/issues`
- `GET/PATCH /issues/{id}`；`POST /issues/{id}/move` `{status,board_order,version}`
- Relations 防环 → `409 conflict`；乐观锁见上文「竞态 / 乐观锁」
- Cycles / Comments / Activities / Search

写操作建议支持 Header `Idempotency-Key`。

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

