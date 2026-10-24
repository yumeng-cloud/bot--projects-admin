# Wiki 向量短报（Postgres+pgvector）

**时间**：2026-10-24 Asia/Shanghai  
**状态**：统一 HTTP 嵌入（本地 MiniLM sidecar / 生产 OpenAI 兼容）；禁 hash 运行时默认

## 选型
- **存向量**：Postgres 17 + **pgvector**（禁 SQLite 向量）
- **关键词**：title/body `LIKE`；混合 **RRF**（`pkg/hybridsearch`）
- **嵌入**：唯一路径 OpenAI 兼容 HTTP；开发默认 `scripts/local-embed`（`all-MiniLM-L6-v2` → **384**）；生产只改 `EMBED_BASE_URL`/`MODEL`/`DIMS`/`KEY`。`NewHash` 仅单测。

## DDL
- `schema_postgres.sql` + `postgres/003_embeddings.sql` → `embedding_chunks`（`vector(384)`）
- 1536→384：`scripts/rebuild-embedding-chunks.sql`
- 连接：`.env.postgres`

## 并发 / 注入 / API
同前：`ReplaceEntity` advisory lock；`NewWikiHandler`/`bots` `MustNew(LoadConfig())`；  
`GET .../wiki/search?mode=hybrid|semantic|keyword`

## 验收
```bash
# 起本地 embed
cd scripts/local-embed && source .venv/bin/activate && python server.py &

set -a && source .env.postgres && set +a
# 离线 smoke 可显式：export EMBED_PROVIDER=hash
cd backend && go test ./internal/pkg/embedder/ ./internal/pkg/hybridsearch/ && go build ./...
go test ./internal/services/wiki/ -run TestPostgresHybridSearchSmoke -v
```

约定：`docs/embed.md`
