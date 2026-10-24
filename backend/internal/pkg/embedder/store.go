package embedder

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
)

// ChunkRow 对应 embedding_chunks 一行（pgvector）。
type ChunkRow struct {
	ID          string
	WorkspaceID string
	EntityType  string
	EntityID    string
	ChunkIdx    int
	Content     string
	Embedding   []float32
	ModelID     string
	Dims        int
	UpdatedAt   string
}

// Store 写入/删除/按向量检索 embedding_chunks。
// Wiki 与 Bot 记忆共用；并发：ReplaceEntity 在单事务内 DELETE+INSERT。
type Store struct {
	DB *sql.DB
}

func NewStore(db *sql.DB) *Store { return &Store{DB: db} }

// ReplaceEntity 事务内替换某实体全部 chunk（先锁再删再插，避免 search 半旧半新）。
// 使用 pg_advisory_xact_lock(hashtext(workspace||type||id)) + DELETE + INSERT。
func (s *Store) ReplaceEntity(ctx context.Context, workspaceID, entityType, entityID string, contents []string, vectors [][]float32, modelID string, dims int) error {
	if len(contents) != len(vectors) {
		return fmt.Errorf("embedder store: contents/vectors length mismatch")
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// 事务级咨询锁：同实体并发写入串行化
	lockKey := workspaceID + "|" + entityType + "|" + entityID
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, lockKey); err != nil {
		return fmt.Errorf("embedder store lock: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM embedding_chunks WHERE workspace_id=$1 AND entity_type=$2 AND entity_id=$3`,
		workspaceID, entityType, entityID); err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	for i, content := range contents {
		id := idgen.New()
		vecLit := EncodeVector(vectors[i])
		_, err := tx.ExecContext(ctx, `
			INSERT INTO embedding_chunks
				(id, workspace_id, entity_type, entity_id, chunk_idx, content, embedding, model_id, dims, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7::vector,$8,$9,$10)`,
			id, workspaceID, entityType, entityID, i, content, vecLit, modelID, dims, now)
		if err != nil {
			return fmt.Errorf("embedder store insert chunk %d: %w", i, err)
		}
	}
	return tx.Commit()
}

// DeleteEntity 删除实体全部向量（页面删除时调用）。
func (s *Store) DeleteEntity(ctx context.Context, workspaceID, entityType, entityID string) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	lockKey := workspaceID + "|" + entityType + "|" + entityID
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, lockKey); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx,
		`DELETE FROM embedding_chunks WHERE workspace_id=$1 AND entity_type=$2 AND entity_id=$3`,
		workspaceID, entityType, entityID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// SearchVector 余弦相似度召回，返回 entity_id 去重后的有序列表（最佳 chunk 分）。
// 先按 <=> 排序 LIMIT 过量召回以利用 HNSW，再在应用层按 entity_id 去重（避免 GROUP BY 挡索引）。
// 使用读已提交；与 ReplaceEntity 的事务锁配合，避免看到部分替换。
func (s *Store) SearchVector(ctx context.Context, workspaceID, entityType string, query []float32, limit int) ([]ScoredID, error) {
	if limit <= 0 {
		limit = 20
	}
	overfetch := limit * 8
	if overfetch < 64 {
		overfetch = 64
	}
	vecLit := EncodeVector(query)
	rows, err := s.DB.QueryContext(ctx, `
		SELECT entity_id, (embedding <=> $1::vector) AS dist
		FROM embedding_chunks
		WHERE workspace_id = $2 AND entity_type = $3 AND model_id <> ''
		ORDER BY embedding <=> $1::vector
		LIMIT $4`, vecLit, workspaceID, entityType, overfetch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	seen := map[string]bool{}
	var out []ScoredID
	for rows.Next() {
		var id string
		var dist float64
		if err := rows.Scan(&id, &dist); err != nil {
			return nil, err
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, ScoredID{ID: id, Score: 1.0 - dist})
		if len(out) >= limit {
			break
		}
	}
	return out, rows.Err()
}

// ScoredID 检索命中。
type ScoredID struct {
	ID    string
	Score float64
}
