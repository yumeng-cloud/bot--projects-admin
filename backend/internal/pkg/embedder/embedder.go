// Package embedder 提供可复用的文本嵌入接口，供 Wiki 与 Bot 记忆共用。
//
// 存储约定见 docs/embed.md（Postgres + pgvector 表 embedding_chunks）。
// 运行时唯一路径：OpenAI 兼容 HTTP；开发默认打本地 sidecar（见 scripts/local-embed）。
package embedder

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// 默认维度：对齐本地 all-MiniLM-L6-v2（384）。改维度须同步 pgvector DDL。
const DefaultDims = 384

// 默认模型 ID（写入 embedding_chunks.model_id）；本地 sidecar 同名。
const DefaultModelID = "all-MiniLM-L6-v2"

// 默认本地 OpenAI 兼容 embeddings 基址（scripts/local-embed 监听 8081）。
const DefaultBaseURL = "http://127.0.0.1:8081/v1"

// Entity types stored in embedding_chunks.entity_type.
const (
	EntityWikiPage  = "wiki_page"
	EntityBotMemory = "bot_memory"
)

// Embedder 将文本转为固定维度向量。实现必须线程安全。
type Embedder interface {
	// Embed 批量嵌入；返回切片与 texts 等长，每个向量长度 = Dims()。
	Embed(ctx context.Context, texts []string) ([][]float32, error)
	ModelID() string
	Dims() int
}

// Config 从环境变量加载嵌入配置。
type Config struct {
	// Provider: "http"（默认）| "openai"（HTTP 别名，强制 key）| "hash"（仅显式/单测，禁止作运行时默认）
	Provider string
	// BaseURL 兼容 OpenAI embeddings API，如 http://127.0.0.1:8081/v1
	BaseURL string
	// APIKey 对应 Authorization Bearer；本地 http 可空，openai 必填
	APIKey string
	// Model 写入 model_id，并作为 API 请求的 model 字段
	Model string
	// Dims 向量维度；须与 pgvector 列一致
	Dims int
}

// LoadConfig 读取环境变量（未设 EMBED_* 时默认本地 HTTP，绝不回退 hash）：
//
//	EMBED_PROVIDER   默认 http（openai 为显式别名）
//	EMBED_BASE_URL   默认 http://127.0.0.1:8081/v1；provider=openai 且未设时用 https://api.openai.com/v1
//	EMBED_API_KEY    或 OPENAI_API_KEY（本地可空）
//	EMBED_MODEL      默认 all-MiniLM-L6-v2
//	EMBED_DIMS       默认 384
func LoadConfig() Config {
	dims, _ := strconv.Atoi(getenv("EMBED_DIMS", strconv.Itoa(DefaultDims)))
	if dims <= 0 {
		dims = DefaultDims
	}
	key := getenv("EMBED_API_KEY", "")
	if key == "" {
		key = getenv("OPENAI_API_KEY", "")
	}
	provider := strings.ToLower(strings.TrimSpace(os.Getenv("EMBED_PROVIDER")))
	if provider == "" {
		provider = "http"
	}
	defaultURL := DefaultBaseURL
	if provider == "openai" {
		defaultURL = "https://api.openai.com/v1"
	}
	return Config{
		Provider: provider,
		BaseURL:  strings.TrimRight(getenv("EMBED_BASE_URL", defaultURL), "/"),
		APIKey:   key,
		Model:    getenv("EMBED_MODEL", DefaultModelID),
		Dims:     dims,
	}
}

// New 按 Config 构造 Embedder。运行时认 http/openai；hash 仅显式（单测/离线 smoke）。
func New(cfg Config) (Embedder, error) {
	if cfg.Dims <= 0 {
		cfg.Dims = DefaultDims
	}
	if cfg.Model == "" {
		cfg.Model = DefaultModelID
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}
	switch strings.ToLower(cfg.Provider) {
	case "openai":
		if cfg.APIKey == "" {
			return nil, fmt.Errorf("embedder: EMBED_API_KEY/OPENAI_API_KEY required for provider %q", cfg.Provider)
		}
		return NewHTTP(cfg), nil
	case "http", "":
		// 本地 sidecar 不强制 key；生产可仍设 Bearer
		return NewHTTP(cfg), nil
	case "hash", "local":
		// 仅显式保留：单元测 / 离线 smoke；禁止作 LoadConfig 默认
		return NewHash(cfg.Model, cfg.Dims), nil
	default:
		return nil, fmt.Errorf("embedder: unknown provider %q (http|openai; hash 仅单测显式)", cfg.Provider)
	}
}

// MustNew 同 New；失败则 panic（不再静默落到 hash）。
func MustNew(cfg Config) Embedder {
	e, err := New(cfg)
	if err != nil {
		panic("embedder: " + err.Error())
	}
	return e
}

// ProviderName 返回当前配置的 provider 字符串（日志用）。
func (c Config) ProviderName() string {
	if c.Provider == "" {
		return "http"
	}
	return c.Provider
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
