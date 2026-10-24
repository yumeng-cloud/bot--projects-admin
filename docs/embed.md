# 嵌入与混合检索约定（Wiki + Bot 记忆共用）

> 定稿：**Postgres + pgvector** 存向量；**tsvector/ILIKE + 向量** 双路，**RRF** 融合。  
> **唯一运行时路径**：OpenAI 兼容 HTTP Embedder。开发默认本地 sidecar；生产只换 endpoint。  
> SQLite 仍可用于无向量的本地单测；**向量写入/语义检索仅在 Postgres 驱动下启用**。

## 一句话方案

开发起 `scripts/local-embed`（CPU MiniLM→384）→ Go 默认打 `127.0.0.1:8081`；生产同一 HTTP 客户端，只改 `EMBED_BASE_URL` / `MODEL` / `DIMS` / `KEY`。无「未配 key→hash」双轨。

## 包路径

| 包 | 路径 | 职责 |
|----|------|------|
| Embedder | `backend/internal/pkg/embedder` | `Embed` / `ModelID` / `Dims`；`NewHTTP` |
| Chunk / vector 字面量 | 同上 | `Chunk` · `EncodeVector` |
| Store | `embedder.Store` | `ReplaceEntity` / `DeleteEntity` / `SearchVector` |
| RRF | `backend/internal/pkg/hybridsearch` | `RRF` / `RRFK` |
| 本地 sidecar | `scripts/local-embed/` | OpenAI 兼容 `/v1/embeddings` |

`entity_type`：`wiki_page` / `bot_memory`。`NewHash` **仅单测/显式** `EMBED_PROVIDER=hash`，禁止作运行时默认。

### 环境变量

| 变量 | 默认 | 说明 |
|------|------|------|
| `DB_DRIVER` | `sqlite` | 向量功能请设 `postgres` |
| `DATABASE_URL` | `./data/ai-collab.db` | Postgres DSN |
| `EMBED_PROVIDER` | **`http`** | `http`（默认）/ `openai`（同 HTTP，强制 key）/ `hash`（仅显式单测） |
| `EMBED_BASE_URL` | **`http://127.0.0.1:8081/v1`** | `openai` 且未设时 → `https://api.openai.com/v1` |
| `EMBED_API_KEY` 或 `OPENAI_API_KEY` | 空 | 本地 http 可空；`openai` 必填 |
| `EMBED_MODEL` | **`all-MiniLM-L6-v2`** | 写入 `model_id` |
| `EMBED_DIMS` | **384** | 须与 DDL `vector(N)` 一致 |

生产示例：

```bash
export EMBED_PROVIDER=openai   # 或 http
export EMBED_BASE_URL=https://api.openai.com/v1
export EMBED_API_KEY=sk-...
export EMBED_MODEL=text-embedding-3-small
export EMBED_DIMS=1536         # 若用 1536 模型须同步改 DDL
```

## 本地 sidecar

```bash
cd scripts/local-embed
python3 -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
python server.py
```

详见 `scripts/local-embed/README.md`。

## 表 DDL（pgvector）

维度默认 **384**（MiniLM）。若改 `EMBED_DIMS`，须同步改列类型并重建索引。

```sql
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS embedding_chunks (
  id            TEXT PRIMARY KEY,
  workspace_id  TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  entity_type   TEXT NOT NULL,  -- wiki_page | bot_memory
  entity_id     TEXT NOT NULL,
  chunk_idx     INT  NOT NULL DEFAULT 0,
  content       TEXT NOT NULL DEFAULT '',
  embedding     vector(384) NOT NULL,
  model_id      TEXT NOT NULL,
  dims          INT  NOT NULL,
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (workspace_id, entity_type, entity_id, chunk_idx)
);

CREATE INDEX IF NOT EXISTS idx_embedding_chunks_entity
  ON embedding_chunks (workspace_id, entity_type, entity_id);

CREATE INDEX IF NOT EXISTS idx_embedding_chunks_hnsw
  ON embedding_chunks USING hnsw (embedding vector_cosine_ops);
```

### 已有库从 1536 → 384（重建）

`ALTER` 无法直接改 vector 维度；空表或可丢数据时：

```sql
DROP INDEX IF EXISTS idx_embedding_chunks_hnsw;
DROP TABLE IF EXISTS embedding_chunks;
-- 再跑 003_embeddings.sql / 或重启 AUTO_MIGRATE（见 schema_postgres.sql）
```

有数据：导出文本 → 重建表 → 全量重索引（Wiki/Bot 再写一遍）。脚本提示见 `scripts/rebuild-embedding-chunks.sql`。

## 并发策略

1. **写入**：`ReplaceEntity` / `DeleteEntity` 在单事务内 advisory lock → DELETE → INSERT。  
2. **检索**：READ COMMITTED；RRF 在应用层合并。  
3. **Wiki**：先提交页再 embed；失败可下次更新重试。

## Wiki Search API

`GET /api/v1/workspaces/{ws}/wiki/search?q=&mode=hybrid|semantic|keyword&limit=`

未配 Postgres/embed 时 `semantic`/`hybrid` 降级 keyword（`search_backend=keyword_fallback`）。

## 验收

```bash
# 单元
cd backend && go test ./internal/pkg/embedder/ ./internal/pkg/hybridsearch/ && go build ./...

# 本地 sidecar smoke
curl -s http://127.0.0.1:8081/v1/embeddings -H 'Content-Type: application/json' \
  -d '{"model":"all-MiniLM-L6-v2","input":["ping"]}'
```
