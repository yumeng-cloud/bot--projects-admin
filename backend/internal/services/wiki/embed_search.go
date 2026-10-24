package wiki

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/models"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
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
	Pages         []models.WikiPage `json:"pages"`
	Q             string            `json:"q"`
	Mode          string            `json:"mode"`
	SearchBackend string            `json:"search_backend"` // hybrid|semantic|keyword|keyword_fallback
}

// WithEmbedder 启用写入切块嵌入 + 语义/混合检索（需 Postgres+pgvector Store）。
func (s *Service) WithEmbedder(e embedder.Embedder, store *embedder.Store) *Service {
	s.Embed = e
	s.Store = store
	return s
}

func (s *Service) indexPage(ctx context.Context, workspaceID, pageID, title, body string) {
	if s.Embed == nil || s.Store == nil {
		return
	}
	text := strings.TrimSpace(title + "\n\n" + body)
	if text == "" {
		_ = s.Store.DeleteEntity(ctx, workspaceID, embedder.EntityWikiPage, pageID)
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
		log.Printf("wiki embed: page %s: %v", pageID, err)
		return
	}
	if err := s.Store.ReplaceEntity(ctx, workspaceID, embedder.EntityWikiPage, pageID, chunks, vecs, s.Embed.ModelID(), s.Embed.Dims()); err != nil {
		log.Printf("wiki embed store: page %s: %v", pageID, err)
	}
}

func (s *Service) dropPageIndex(ctx context.Context, workspaceID, pageID string) {
	if s.Store == nil {
		return
	}
	if err := s.Store.DeleteEntity(ctx, workspaceID, embedder.EntityWikiPage, pageID); err != nil {
		log.Printf("wiki embed delete: page %s: %v", pageID, err)
	}
}

// SearchEx 支持 mode=keyword|semantic|hybrid。
func (s *Service) SearchEx(ctx context.Context, workspaceID, q string, mode SearchMode, limit int) (*SearchResult, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil, apperr.BadRequest("搜索关键词不能为空")
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if mode == "" {
		mode = ModeHybrid
	}
	mode = SearchMode(strings.ToLower(string(mode)))

	canVector := s.Embed != nil && s.Store != nil
	if (mode == ModeSemantic || mode == ModeHybrid) && !canVector {
		pages, err := s.Search(workspaceID, q, limit)
		if err != nil {
			return nil, err
		}
		return &SearchResult{Pages: pages, Q: q, Mode: string(mode), SearchBackend: "keyword_fallback"}, nil
	}

	switch mode {
	case ModeKeyword:
		pages, err := s.Search(workspaceID, q, limit)
		if err != nil {
			return nil, err
		}
		return &SearchResult{Pages: pages, Q: q, Mode: "keyword", SearchBackend: "keyword"}, nil
	case ModeSemantic:
		ids, err := s.semanticIDs(ctx, workspaceID, q, limit)
		if err != nil {
			return nil, err
		}
		pages, err := s.pagesByIDs(workspaceID, ids)
		if err != nil {
			return nil, err
		}
		return &SearchResult{Pages: pages, Q: q, Mode: "semantic", SearchBackend: "semantic"}, nil
	default: // hybrid
		kw, err := s.Search(workspaceID, q, limit)
		if err != nil {
			return nil, err
		}
		kwIDs := make([]string, 0, len(kw))
		for _, p := range kw {
			kwIDs = append(kwIDs, p.ID)
		}
		semIDs, err := s.semanticIDs(ctx, workspaceID, q, limit)
		if err != nil {
			return nil, err
		}
		merged := hybridsearch.RRF(kwIDs, semIDs)
		if len(merged) > limit {
			merged = merged[:limit]
		}
		pages, err := s.pagesByIDs(workspaceID, merged)
		if err != nil {
			return nil, err
		}
		return &SearchResult{Pages: pages, Q: q, Mode: "hybrid", SearchBackend: "hybrid"}, nil
	}
}

func (s *Service) semanticIDs(ctx context.Context, workspaceID, q string, limit int) ([]string, error) {
	vecs, err := s.Embed.Embed(ctx, []string{q})
	if err != nil {
		return nil, err
	}
	if len(vecs) == 0 {
		return nil, nil
	}
	scored, err := s.Store.SearchVector(ctx, workspaceID, embedder.EntityWikiPage, vecs[0], limit)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(scored))
	for _, h := range scored {
		ids = append(ids, h.ID)
	}
	return ids, nil
}

func (s *Service) pagesByIDs(workspaceID string, ids []string) ([]models.WikiPage, error) {
	out := make([]models.WikiPage, 0, len(ids))
	seen := map[string]bool{}
	for _, id := range ids {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		p, err := s.GetPage(workspaceID, id)
		if err != nil {
			continue // 已删页的残留向量跳过
		}
		out = append(out, *p)
	}
	if out == nil {
		out = []models.WikiPage{}
	}
	return out, nil
}
