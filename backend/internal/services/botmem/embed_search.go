package botmem

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/models"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/embedder"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/hybridsearch"
)

// SearchMode keyword | semantic | hybrid
type SearchMode string

const (
	ModeKeyword  SearchMode = "keyword"
	ModeSemantic SearchMode = "semantic"
	ModeHybrid   SearchMode = "hybrid"
)

// SearchResult 带后端标记，便于 FE/验收。
type SearchResult struct {
	Memories      []models.BotMemory `json:"memories"`
	Q             string             `json:"q"`
	Mode          string             `json:"mode"`
	SearchBackend string             `json:"search_backend"` // hybrid|semantic|keyword|keyword_fallback
}

// WithEmbedder 启用写入切块嵌入 + 语义/混合检索（需 Postgres+pgvector Store）。
func (s *Service) WithEmbedder(e embedder.Embedder, store *embedder.Store) *Service {
	s.Embed = e
	s.Store = store
	return s
}

func (s *Service) indexMemory(ctx context.Context, workspaceID, memoryID, keyName, valueText string) {
	if s.Embed == nil || s.Store == nil {
		return
	}
	text := strings.TrimSpace(keyName + "\n\n" + valueText)
	if text == "" {
		_ = s.Store.DeleteEntity(ctx, workspaceID, embedder.EntityBotMemory, memoryID)
		return
	}
	chunks := embedder.Chunk(text, 800, 100)
	if len(chunks) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	vecs, err := s.Embed.Embed(ctx, chunks)
	if err != nil {
		log.Printf("botmem embed: memory %s: %v", memoryID, err)
		return
	}
	if err := s.Store.ReplaceEntity(ctx, workspaceID, embedder.EntityBotMemory, memoryID, chunks, vecs, s.Embed.ModelID(), s.Embed.Dims()); err != nil {
		log.Printf("botmem embed store: memory %s: %v", memoryID, err)
	}
}

func (s *Service) dropMemoryIndex(ctx context.Context, workspaceID, memoryID string) {
	if s.Store == nil {
		return
	}
	if err := s.Store.DeleteEntity(ctx, workspaceID, embedder.EntityBotMemory, memoryID); err != nil {
		log.Printf("botmem embed delete: memory %s: %v", memoryID, err)
	}
}

// SearchMemories 支持 mode=keyword|semantic|hybrid；无 Embed/Store 时 semantic/hybrid → keyword_fallback。
// q 为空时仅列出该 member 全部记忆（keyword 路径）。
func (s *Service) SearchMemories(ctx context.Context, workspaceID, memberID, q string, mode SearchMode, limit int) (*SearchResult, error) {
	if err := s.assertMemberInWorkspace(workspaceID, memberID); err != nil {
		return nil, err
	}
	q = strings.TrimSpace(q)
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if mode == "" {
		mode = ModeHybrid
	}
	mode = SearchMode(strings.ToLower(string(mode)))

	if q == "" {
		mems, err := s.ListMemories(workspaceID, memberID, "")
		if err != nil {
			return nil, err
		}
		if len(mems) > limit {
			mems = mems[:limit]
		}
		return &SearchResult{Memories: mems, Q: q, Mode: string(mode), SearchBackend: "keyword"}, nil
	}

	canVector := s.Embed != nil && s.Store != nil
	if (mode == ModeSemantic || mode == ModeHybrid) && !canVector {
		mems, err := s.listMemoriesLimited(workspaceID, memberID, q, limit)
		if err != nil {
			return nil, err
		}
		return &SearchResult{Memories: mems, Q: q, Mode: string(mode), SearchBackend: "keyword_fallback"}, nil
	}

	switch mode {
	case ModeKeyword:
		mems, err := s.listMemoriesLimited(workspaceID, memberID, q, limit)
		if err != nil {
			return nil, err
		}
		return &SearchResult{Memories: mems, Q: q, Mode: "keyword", SearchBackend: "keyword"}, nil
	case ModeSemantic:
		ids, err := s.semanticIDs(ctx, workspaceID, q, limit)
		if err != nil {
			return nil, err
		}
		mems, err := s.memoriesByIDs(workspaceID, memberID, ids)
		if err != nil {
			return nil, err
		}
		return &SearchResult{Memories: mems, Q: q, Mode: "semantic", SearchBackend: "semantic"}, nil
	default: // hybrid
		kw, err := s.listMemoriesLimited(workspaceID, memberID, q, limit)
		if err != nil {
			return nil, err
		}
		kwIDs := make([]string, 0, len(kw))
		for _, m := range kw {
			kwIDs = append(kwIDs, m.ID)
		}
		semIDs, err := s.semanticIDs(ctx, workspaceID, q, limit)
		if err != nil {
			return nil, err
		}
		merged := hybridsearch.RRF(kwIDs, semIDs)
		if len(merged) > limit {
			merged = merged[:limit]
		}
		mems, err := s.memoriesByIDs(workspaceID, memberID, merged)
		if err != nil {
			return nil, err
		}
		return &SearchResult{Memories: mems, Q: q, Mode: "hybrid", SearchBackend: "hybrid"}, nil
	}
}

func (s *Service) listMemoriesLimited(workspaceID, memberID, q string, limit int) ([]models.BotMemory, error) {
	mems, err := s.ListMemories(workspaceID, memberID, q)
	if err != nil {
		return nil, err
	}
	if limit > 0 && len(mems) > limit {
		mems = mems[:limit]
	}
	return mems, nil
}

func (s *Service) semanticIDs(ctx context.Context, workspaceID, q string, limit int) ([]string, error) {
	vecs, err := s.Embed.Embed(ctx, []string{q})
	if err != nil {
		return nil, err
	}
	if len(vecs) == 0 {
		return nil, nil
	}
	scored, err := s.Store.SearchVector(ctx, workspaceID, embedder.EntityBotMemory, vecs[0], limit)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(scored))
	for _, h := range scored {
		ids = append(ids, h.ID)
	}
	return ids, nil
}

func (s *Service) memoriesByIDs(workspaceID, memberID string, ids []string) ([]models.BotMemory, error) {
	out := make([]models.BotMemory, 0, len(ids))
	seen := map[string]bool{}
	for _, id := range ids {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		m, err := s.GetMemory(workspaceID, memberID, id)
		if err != nil {
			continue // 他 member 的向量命中或已删 → 跳过
		}
		out = append(out, *m)
	}
	if out == nil {
		out = []models.BotMemory{}
	}
	return out, nil
}
