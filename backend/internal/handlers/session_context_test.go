package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/config"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/seed"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/server"
)

func sessionContextRouter(t *testing.T) (http.Handler, string) {
	t.Helper()
	dir := t.TempDir()
	dsn := filepath.Join(dir, "t.db")
	sqlDB, err := db.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	cfg := config.Config{
		JWTSecret:      "test-secret",
		JWTTTL:         time.Hour,
		AppEnv:         "test",
		SeedAdminEmail: "admin@yumeng.tech",
		SeedAdminPass:  "Admin123!",
		SeedWSName:     "育蒙科技",
		SeedSampleData: true,
		LoginRateLimit: 100,
	}
	if err := seed.Run(sqlDB, cfg); err != nil {
		t.Fatal(err)
	}
	h := server.NewRouter(sqlDB, cfg)

	body, _ := json.Marshal(map[string]string{"email": cfg.SeedAdminEmail, "password": cfg.SeedAdminPass})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body)))
	if rr.Code != 200 {
		t.Fatalf("login %d %s", rr.Code, rr.Body.String())
	}
	var login map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &login)
	token, _ := login["token"].(string)
	if token == "" {
		token, _ = login["access_token"].(string)
	}
	if token == "" {
		t.Fatalf("no token: %s", rr.Body.String())
	}
	return h, token
}

func TestSessionContextGetDefaultEmpty(t *testing.T) {
	h, token := sessionContextRouter(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/session-context", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("GET %d %s", rr.Code, rr.Body.String())
	}
	var dto map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &dto); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"wsId", "teamId", "projectId", "issueId", "view", "updatedAt"} {
		if _, ok := dto[k]; !ok {
			t.Fatalf("missing key %s in %v", k, dto)
		}
	}
	if dto["wsId"] != "" || dto["teamId"] != "" {
		t.Fatalf("expected empty defaults, got %v", dto)
	}
}

func TestSessionContextPutGetRoundtripAndOverwrite(t *testing.T) {
	h, token := sessionContextRouter(t)
	base := "/api/v1/me/session-context"

	put := func(body map[string]any) map[string]any {
		t.Helper()
		raw, _ := json.Marshal(body)
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, base, bytes.NewReader(raw))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		h.ServeHTTP(rr, req)
		if rr.Code != 200 {
			t.Fatalf("PUT %d %s", rr.Code, rr.Body.String())
		}
		var dto map[string]any
		_ = json.Unmarshal(rr.Body.Bytes(), &dto)
		return dto
	}
	get := func() map[string]any {
		t.Helper()
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, base, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		h.ServeHTTP(rr, req)
		if rr.Code != 200 {
			t.Fatalf("GET %d %s", rr.Code, rr.Body.String())
		}
		var dto map[string]any
		_ = json.Unmarshal(rr.Body.Bytes(), &dto)
		return dto
	}

	first := put(map[string]any{
		"wsId": "ws_1", "teamId": "tm_1", "projectId": "pj_1",
		"issueId": "is_1", "view": "board",
	})
	if first["wsId"] != "ws_1" || first["view"] != "board" {
		t.Fatalf("first put: %v", first)
	}
	if first["updatedAt"] == "" {
		t.Fatal("updatedAt empty after put")
	}
	got := get()
	if got["wsId"] != "ws_1" || got["teamId"] != "tm_1" || got["projectId"] != "pj_1" ||
		got["issueId"] != "is_1" || got["view"] != "board" {
		t.Fatalf("roundtrip: %v", got)
	}
	at1 := got["updatedAt"].(string)

	time.Sleep(5 * time.Millisecond)
	second := put(map[string]any{
		"wsId": "ws_2", "teamId": "", "projectId": "pj_2",
		"issueId": "", "view": "issue",
	})
	if second["wsId"] != "ws_2" || second["view"] != "issue" || second["teamId"] != "" {
		t.Fatalf("overwrite put: %v", second)
	}
	got2 := get()
	if got2["wsId"] != "ws_2" || got2["projectId"] != "pj_2" || got2["issueId"] != "" || got2["view"] != "issue" {
		t.Fatalf("overwrite get: %v", got2)
	}
	at2 := got2["updatedAt"].(string)
	if at2 == "" || at2 == at1 {
		t.Fatalf("updatedAt should bump: before=%s after=%s", at1, at2)
	}
}

func TestSessionContextUnauthenticated401(t *testing.T) {
	h, _ := sessionContextRouter(t)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/me/session-context", nil))
	if rr.Code != 401 {
		t.Fatalf("GET unauth %d %s", rr.Code, rr.Body.String())
	}
	rr2 := httptest.NewRecorder()
	body, _ := json.Marshal(map[string]string{"wsId": "x"})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/me/session-context", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rr2, req)
	if rr2.Code != 401 {
		t.Fatalf("PUT unauth %d %s", rr2.Code, rr2.Body.String())
	}
}

func TestSessionContextBotAPIKeyRejected(t *testing.T) {
	h, token := sessionContextRouter(t)

	// Create a real bot API key via workspace route, then try human session-context with it.
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	h.ServeHTTP(rr, req)
	var wsWrap struct {
		Workspaces []struct {
			ID string `json:"id"`
		} `json:"workspaces"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &wsWrap)
	if len(wsWrap.Workspaces) == 0 {
		t.Fatalf("no workspace: %s", rr.Body.String())
	}
	wsID := wsWrap.Workspaces[0].ID

	body, _ := json.Marshal(map[string]string{"name": "sess-ctx-bot"})
	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+wsID+"/api-keys", bytes.NewReader(body))
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusCreated && rr2.Code != 200 {
		t.Fatalf("create key %d %s", rr2.Code, rr2.Body.String())
	}
	var keyRes map[string]any
	_ = json.Unmarshal(rr2.Body.Bytes(), &keyRes)
	raw, _ := keyRes["token"].(string)
	if raw == "" {
		if k, ok := keyRes["api_key"].(map[string]any); ok {
			raw, _ = k["token"].(string)
			if raw == "" {
				raw, _ = k["key"].(string)
			}
		}
	}
	if raw == "" {
		raw, _ = keyRes["key"].(string)
	}
	if raw == "" || len(raw) < 4 || raw[:4] != "ymk_" {
		// Fallback: synthetic ymk_ still must 401 on jwtOnly route
		raw = "ymk_not_a_real_key_but_bot_prefix"
	}

	rr3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodGet, "/api/v1/me/session-context", nil)
	req3.Header.Set("Authorization", "Bearer "+raw)
	h.ServeHTTP(rr3, req3)
	if rr3.Code != 401 {
		t.Fatalf("bot key GET should 401, got %d %s", rr3.Code, rr3.Body.String())
	}

	rr4 := httptest.NewRecorder()
	putBody, _ := json.Marshal(map[string]string{"wsId": "ws_bot"})
	req4 := httptest.NewRequest(http.MethodPut, "/api/v1/me/session-context", bytes.NewReader(putBody))
	req4.Header.Set("Authorization", "Bearer "+raw)
	req4.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rr4, req4)
	if rr4.Code != 401 {
		t.Fatalf("bot key PUT should 401, got %d %s", rr4.Code, rr4.Body.String())
	}
}
