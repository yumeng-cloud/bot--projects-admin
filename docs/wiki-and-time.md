# Wiki 知识库与工时计费 API

模块归属：知识库工时。前缀均为 `/api/v1`。鉴权：`Authorization: Bearer <JWT|API Key>`（`AuthJWTOrAPIKey` + `RequireWorkspaceMember`；写操作套 `Idempotency`）。JSON 字段 snake_case。写操作可带 `Idempotency-Key`（薄幂等回放）。

成员解析：从 JWT `user_id` + 路径 `workspaceID` 查 `members`；非成员返回 403。

---

## Wiki

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/workspaces/{ws}/wiki/spaces` | 列出空间 |
| POST | `/workspaces/{ws}/wiki/spaces` | 创建空间 `{name,slug?,description?}` |
| GET | `/workspaces/{ws}/wiki/spaces/{spaceID}` | 空间详情 |
| PATCH | `/workspaces/{ws}/wiki/spaces/{spaceID}` | 更新 |
| DELETE | `/workspaces/{ws}/wiki/spaces/{spaceID}` | 删除（级联页面） |
| GET | `/workspaces/{ws}/wiki/spaces/{spaceID}/pages` | 列出页面 |
| POST | `/workspaces/{ws}/wiki/spaces/{spaceID}/pages` | 创建页 `{title,slug?,body_md,parent_id?,project_id?,issue_id?}` |
| GET | `/workspaces/{ws}/wiki/pages/{pageID}` | 页面详情（`body_md` 为 Markdown） |
| PATCH | `/workspaces/{ws}/wiki/pages/{pageID}` | 更新 |
| DELETE | `/workspaces/{ws}/wiki/pages/{pageID}` | 删除 |
| GET | `/workspaces/{ws}/wiki/search?q=` | 按 title/body **关键词/模糊**搜索（当前实现） |
| GET | `/workspaces/{ws}/wiki/tree?space_id=` | 树形目录 |

**搜索与向量**：`GET .../wiki/search?q=&mode=hybrid|semantic|keyword&limit=`。生产：**PostgreSQL + pgvector**（见 [`embed.md`](./embed.md)、`.env.postgres`）；SQLite 仅 keyword。`DB_DRIVER=postgres` 时 `NewWikiHandler` 自动 `WithEmbedder`（默认本地 HTTP embed；见 embed.md）。写入页后同步切块嵌入；并发：`pg_advisory_xact_lock` + 单事务 DELETE/INSERT。勿用 sqlite-vec。

种子：`seed.Wiki` 写入空间 `handbook`（协作手册）及页面：产品约定、Issue 状态机、Bot 协作规范。

挂载：通过 `modules.RegisterWiki` / `RegisterTime`（`cmd/server` blank import `modules/wiki`、`modules/timebilling`）挂到 `/api/v1`；亦可 `handlers.MountWiki` / `RegisterWikiAndTime`（std ServeMux）。

---

## 工时

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/workspaces/{ws}/time-entries` | 列表（`issue_id`/`member_id`/`from`/`to`） |
| POST | `/workspaces/{ws}/time-entries` | 记工时 `{issue_id,hours,note?,billable?,rate?,worked_at?,member_id?}` |
| GET/PATCH/DELETE | `/workspaces/{ws}/time-entries/{entryID}` | 单条 |
| GET/POST | `/workspaces/{ws}/issues/{issueID}/time-entries` | 按 Issue |

`hours > 0`；`billable` 默认 true；未传 `rate` 时汇总用成员/工作区默认费率。

---

## 计费

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/workspaces/{ws}/billing-rates` | 费率列表 |
| POST/PUT | `/workspaces/{ws}/billing-rates` | upsert `{member_id?,hourly_rate,currency?}`；`member_id` 空=工作区默认 |
| DELETE | `/workspaces/{ws}/billing-rates/{rateID}` | 删除 |
| GET | `/workspaces/{ws}/billing/summary?from=&to=&group_by=project\|member\|date` | 汇总；金额 = hours × (entry.rate \|\| 默认费率)，仅 billable |
| GET | `/workspaces/{ws}/billing/export.csv?from=&to=` | CSV 导出 |
| GET | `/workspaces/{ws}/time-entries/summary` / `export.csv` | 同上（兼容） |

挂载：同上，`modules.RegisterTime`（blank import `modules/timebilling`）；亦可 `handlers.MountTime`。

---

## 包路径

- `internal/models/wiki.go`、`time.go`
- `internal/services/wiki`、`internal/services/timetrack`
- `internal/handlers/wiki.go`、`time.go`、`mount.go`
- `internal/seed/wiki.go`

前端页面本次未实现，可由 React 同事对接上述 API。
