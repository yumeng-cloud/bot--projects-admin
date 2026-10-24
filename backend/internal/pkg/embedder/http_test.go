package embedder

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPEmbed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/embeddings" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("auth %q", got)
		}
		var req embedReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.Model != "all-MiniLM-L6-v2" || len(req.Input) != 2 {
			t.Fatalf("req %+v", req)
		}
		_ = json.NewEncoder(w).Encode(embedResp{Data: []struct {
			Embedding []float32 `json:"embedding"`
			Index     int       `json:"index"`
		}{
			{Index: 0, Embedding: []float32{0.1, 0.2, 0.3}},
			{Index: 1, Embedding: []float32{0.4, 0.5, 0.6}},
		}})
	}))
	defer srv.Close()

	h := NewHTTP(Config{
		Provider: "http",
		BaseURL:  srv.URL,
		APIKey:   "test-key",
		Model:    "all-MiniLM-L6-v2",
		Dims:     3,
	})
	vecs, err := h.Embed(context.Background(), []string{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}
	if len(vecs) != 2 || len(vecs[0]) != 3 || vecs[1][0] != 0.4 {
		t.Fatalf("vecs %+v", vecs)
	}
}

func TestHTTPEmbedNoKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Errorf("expected no Authorization, got %q", r.Header.Get("Authorization"))
		}
		_ = json.NewEncoder(w).Encode(embedResp{Data: []struct {
			Embedding []float32 `json:"embedding"`
			Index     int       `json:"index"`
		}{
			{Index: 0, Embedding: []float32{0.1, 0.2}},
		}})
	}))
	defer srv.Close()
	h := NewHTTP(Config{BaseURL: srv.URL, Model: "m", Dims: 2})
	vecs, err := h.Embed(context.Background(), []string{"x"})
	if err != nil {
		t.Fatal(err)
	}
	if len(vecs) != 1 || len(vecs[0]) != 2 {
		t.Fatalf("vecs %+v", vecs)
	}
}

func TestNewRequiresKeyForOpenAI(t *testing.T) {
	_, err := New(Config{Provider: "openai", Model: "m", Dims: 8})
	if err == nil {
		t.Fatal("expected error without API key")
	}
}

func TestNewHTTPLocalNoKeyOK(t *testing.T) {
	e, err := New(Config{Provider: "http", BaseURL: "http://127.0.0.1:8081/v1", Model: "all-MiniLM-L6-v2", Dims: 384})
	if err != nil {
		t.Fatal(err)
	}
	if e.Dims() != 384 {
		t.Fatalf("dims %d", e.Dims())
	}
}

func TestLoadConfigDefaultsLocalHTTP(t *testing.T) {
	t.Setenv("EMBED_PROVIDER", "")
	t.Setenv("EMBED_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("EMBED_BASE_URL", "")
	t.Setenv("EMBED_MODEL", "")
	t.Setenv("EMBED_DIMS", "")
	cfg := LoadConfig()
	if cfg.Provider != "http" {
		t.Fatalf("provider %q want http (no hash fallback)", cfg.Provider)
	}
	if cfg.BaseURL != DefaultBaseURL {
		t.Fatalf("baseURL %q want %s", cfg.BaseURL, DefaultBaseURL)
	}
	if cfg.Model != DefaultModelID || cfg.Dims != DefaultDims {
		t.Fatalf("model/dims %q %d", cfg.Model, cfg.Dims)
	}
	e, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := e.(*HTTP); !ok {
		t.Fatalf("want *HTTP, got %T", e)
	}
}

func TestLoadConfigExplicitOpenAINoKey(t *testing.T) {
	t.Setenv("EMBED_PROVIDER", "openai")
	t.Setenv("EMBED_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	cfg := LoadConfig()
	if cfg.Provider != "openai" {
		t.Fatalf("provider %q want openai", cfg.Provider)
	}
	_, err := New(cfg)
	if err == nil {
		t.Fatal("expected New error")
	}
}
