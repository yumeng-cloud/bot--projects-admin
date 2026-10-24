# AI 协作看板（育蒙科技）

Linear-like 协作看板 + Grok Bot。中文 UI，时区 Asia/Shanghai。

> 多 Bot 并行开发中。共享地基由 Tech lead 维护，详见 [COORDINATION.md](./COORDINATION.md)。

## 技术栈

| 层 | 选型 |
|----|------|
| Backend | Go HTTP（标准库 `ServeMux`） |
| Frontend | React + TypeScript + **Ant Design 6.6.7** + Vite |
| DB | SQLite（`modernc.org/sqlite`）默认开发库，启动自动 migrate `schema.sql`；生产 / **向量·语义搜索**用 **PostgreSQL + pgvector**（`DB_DRIVER` / 迁移见 `.env.example`、[`docs/embed.md`](./docs/embed.md)；勿以 sqlite-vec 为生产方案） |
| Auth | 邮箱密码 + argon2id；JWT（Bearer / cookie `session`）；Bot Key 前缀 `ymk_`（见 docs/BOT.md） |

## 快速开始

```bash
cd /workspace/ai-collab-board
cp .env.example .env   # 可选

# 后端依赖
cd backend && go mod tidy && cd ..

# 前端依赖
cd frontend && npm install && cd ..

# 仅 API
make run-backend
# 或 make seed && cd backend && go run ./cmd/server

# 仅前端（代理 /api → :8080）
make run-frontend

# 并行（需已 install）
make dev
```

- API: http://127.0.0.1:8080  
- UI:  http://127.0.0.1:5173  
- Health: `GET /healthz`

## Seed

```bash
make seed   # 可重复执行；按邮箱 / workspace slug / wiki slug 幂等
```

服务默认自动 seed（`AUTO_SEED`，见 `.env.example`）。

## Seed 登录

| 字段 | 值 |
|------|-----|
| 邮箱 | `admin@yumeng.tech` |
| 密码 | `Admin123!` |
| 工作空间 | 育蒙科技（slug: `yumeng`） |

## `/api/v1` 合同摘要（地基已实现）

```
POST /api/v1/auth/register   {email,password,name?}
POST /api/v1/auth/login      {email,password} → {token,user}
POST /api/v1/auth/logout
GET  /api/v1/auth/me         Authorization: Bearer <jwt>
GET  /api/v1/workspaces
GET  /api/v1/workspaces/{id}
GET  /healthz
```

完整路由见 [docs/API.md](./docs/API.md)；协作约定见 [COORDINATION.md](./COORDINATION.md)。

**Bot 接入（CLI/MCP）**：见 [docs/BOT.md](./docs/BOT.md) · [docs/BOT_MCP_CLI.md](./docs/BOT_MCP_CLI.md)（`go run ./cmd/ymbot`）。

## 架构（简图）

```
frontend (Vite/AntD) --proxy--> backend :8080
                                ├─ auth / workspaces (地基)
                                ├─ projects/issues/cycles (Board Bot)
                                ├─ wiki / time / bot-* (各域 Bot)
                                └─ SQLite ./data/ai-collab.db  或  Postgres+pgvector（DB_DRIVER=postgres）
```

## 测试

```bash
cd backend && go test ./...
cd frontend && npm run build
```

## 文档

- [COORDINATION.md](./COORDINATION.md) — 多 Bot 目录归属与约定
- [docs/API.md](./docs/API.md) — 已挂载路由 / 鉴权 / 错误格式
- [docs/BOT.md](./docs/BOT.md) — Bot API Key、幂等、角色记忆（含待实现）
- [docs/ACCEPTANCE.md](./docs/ACCEPTANCE.md) — QA 验收
- [docs/database.md](./docs/database.md) — SQLite↔Postgres 切换与迁移
- [docs/embed.md](./docs/embed.md) — pgvector / embedding_chunks

## 模块边界（摘要）

| 负责人 | 范围 |
|--------|------|
| Go 后端 | auth、workspaces、seed、路由总装 |
| Issue 域 / 看板 | projects、issues、cycles、看板排序 |
| 知识库工时 | wiki、time_entries、billing |
| Bot 记忆角色 | bot_roles / memories / todos / handoff、`docs/BOT.md` |
| React / 看板 UI | `frontend/` |
| 种子与文档 | 幂等 seed、本 README、API/BOT 文档、`.env.example` |

## 超出范围（仅注释，不实现）

SSO、MCP、完整 Roadmap Gantt、原生移动端、计费订阅。

## License

Proprietary · 育蒙科技
