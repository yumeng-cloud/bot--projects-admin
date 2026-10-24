package embedder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTP 调用 OpenAI 兼容 /embeddings 接口。
type HTTP struct {
	cfg    Config
	client *http.Client
}

func NewHTTP(cfg Config) *HTTP {
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}
	return &HTTP{
		cfg:    cfg,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (h *HTTP) ModelID() string { return h.cfg.Model }
func (h *HTTP) Dims() int       { return h.cfg.Dims }

type embedReq struct {
	Model      string   `json:"model"`
	Input      []string `json:"input"`
	Dimensions int      `json:"dimensions,omitempty"`
}

type embedResp struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (h *HTTP) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	body, err := json.Marshal(embedReq{Model: h.cfg.Model, Input: texts, Dimensions: h.cfg.Dims})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.cfg.BaseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if h.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+h.cfg.APIKey)
	}
	res, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("embedder http %d: %s", res.StatusCode, truncate(string(raw), 200))
	}
	var out embedResp
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	if out.Error != nil {
		return nil, fmt.Errorf("embedder: %s", out.Error.Message)
	}
	vecs := make([][]float32, len(texts))
	for _, d := range out.Data {
		if d.Index < 0 || d.Index >= len(vecs) {
			continue
		}
		vecs[d.Index] = d.Embedding
	}
	for i, v := range vecs {
		if v == nil {
			return nil, fmt.Errorf("embedder: missing embedding for index %d", i)
		}
		if len(v) != h.cfg.Dims && h.cfg.Dims > 0 {
			// 部分模型忽略 dimensions；以实际长度为准但仍校验非空
			if len(v) == 0 {
				return nil, fmt.Errorf("embedder: empty vector at %d", i)
			}
		}
	}
	return vecs, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
