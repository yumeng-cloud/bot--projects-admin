# 数据库：SQLite ↔ PostgreSQL

## 默认（开发）

```bash
export DB_DRIVER=sqlite
export DATABASE_URL=./data/ai-collab.db
# 不需要 AUTO_MIGRATE；打开时自动执行 backend/internal/db/schema.sql
```

## Postgres + pgvector（生产 / 向量检索目标）

```bash
export DB_DRIVER=postgres          # 或 pgx
export DATABASE_URL='postgres://user:pass@localhost:5432/ai_collab?sslmode=disable'
export AUTO_MIGRATE=1              # 执行 internal/db/postgres/*.sql（含 vector 扩展与业务表）
```

迁移文件（**勿**把 SQLite `schema.sql` 原样 Exec 到 PG）：

| 文件 | 内容 |
|------|------|
| `backend/internal/db/postgres/001_extensions.sql` | `CREATE EXTENSION IF NOT EXISTS vector;` |
| `backend/internal/db/postgres/002_schema.sql` | 全业务表（Postgres 方言） |
| `backend/internal/db/postgres/003_embeddings.sql` | `embedding_chunks`（维度默认 384，见 `docs/embed.md`） |

占位符：业务 SQL 仍写 `?`；Postgres 经 `postgres-rebind` 驱动自动 Rebind 为 `$1…`。`embedder.Store` 可继续直接写 `$n`。

## 搜索约定

- Wiki / Botmem / 工作区 `?q=`：**LIKE 模糊为开发兜底**（双驱动均可）。
- **语义 / 向量召回**：归属知识库，经 `embedder.Store` + `embedding_chunks`；禁止 sqlite-vec 作生产方案。

## 已知限制

- `AUTO_MIGRATE=0`（默认未设）时，Postgres 只 Ping，不跑 SQL——需自行 psql 应用 `postgres/*.sql`。
- IVFFlat/HNSW 近似索引在 `003` 中注释，有数据后由知识库创建。
- 完整向量写入管线 / 相似度 HTTP API 见知识库与 `docs/embed.md`（本层只保证连接、迁移与表结构）。

## 并发

Issue 更新 / 看板 `board_order`+`status` 使用事务 + `version` 乐观锁（冲突 **409** 中文）；创建 Issue 对 `issue_counters`：Postgres 使用 `SELECT … FOR UPDATE`；SQLite 依赖单连接事务。

## Postgres 首启（QA / 干净库）

PG15+ 默认应用用户对 `public` schema 无 CREATE，直接 `AUTO_MIGRATE=1` 会报：

`permission denied for schema public (SQLSTATE 42501)`

先跑一次（box 已装 postgresql-17）：

```bash
/workspace/ai-collab-board/scripts/pg-bootstrap.sh
# 或手动（超级用户）：
# GRANT ALL ON SCHEMA public TO aicollab;
# ALTER SCHEMA public OWNER TO aicollab;
# ALTER DATABASE ai_collab OWNER TO aicollab;
```

再：

```bash
set -a && source /workspace/ai-collab-board/.env.postgres && set +a
cd /workspace/ai-collab-board/backend && go run ./cmd/server
```

可复制环境见仓库根 `.env.postgres`。

