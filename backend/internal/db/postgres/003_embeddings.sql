-- 003: embedding_chunks for Wiki + Bot memory (pgvector).
-- Dimension default 384 (= embedder.DefaultDims / EMBED_DIMS). Change DDL + EMBED_DIMS together.
-- Canonical copy also in schema_postgres.sql (db.Open migrates that file).
-- Concurrent writes: embedder.Store uses pg_advisory_xact_lock + DELETE+INSERT in one tx.

CREATE TABLE IF NOT EXISTS embedding_chunks (
  id TEXT PRIMARY KEY,
  workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  entity_type TEXT NOT NULL,
  entity_id TEXT NOT NULL,
  chunk_idx INT NOT NULL DEFAULT 0,
  content TEXT NOT NULL DEFAULT '',
  embedding vector(384) NOT NULL,
  model_id TEXT NOT NULL,
  dims INT NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (workspace_id, entity_type, entity_id, chunk_idx)
);

CREATE INDEX IF NOT EXISTS idx_embedding_chunks_entity
  ON embedding_chunks (workspace_id, entity_type, entity_id);

CREATE INDEX IF NOT EXISTS idx_embedding_chunks_hnsw
  ON embedding_chunks USING hnsw (embedding vector_cosine_ops);
