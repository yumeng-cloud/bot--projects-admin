-- Rebuild embedding_chunks for dim change (e.g. 1536 → 384).
-- WARNING: drops all vectors; re-index Wiki/Bot after.
DROP INDEX IF EXISTS idx_embedding_chunks_hnsw;
DROP TABLE IF EXISTS embedding_chunks;

CREATE TABLE embedding_chunks (
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

CREATE INDEX idx_embedding_chunks_entity
  ON embedding_chunks (workspace_id, entity_type, entity_id);

CREATE INDEX idx_embedding_chunks_hnsw
  ON embedding_chunks USING hnsw (embedding vector_cosine_ops);
