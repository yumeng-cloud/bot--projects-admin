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
	"github.com/yumeng-tech/ai-collab-board/backend/internal/handlers"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/ratelimit"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/seed"
)

func botsMeMux(t *testing.T) *http.ServeMux {
	t.Helper()
	dir := t.TempDir()
	dsn := filepath.Join(dir, "t.db")
	sqlDB, err := db.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	cfg := config.Config{
		JWTSecret:       "test-secret",
		JWTTTL:          time.Hour,
		AppEnv:          "test",
		SeedAdminEmail:  "admin@yumeng.tech",
		SeedAdminPass:   "Admin123!",
		SeedWSName:      "育蒙科技",
		SeedSampleData: true,
		LoginRateLimit:  100,
		LoginRateWindow: time.Minute,
	}
	if err := seed.Run(sqlDB, cfg); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	authH := &handlers.AuthHandler{DB: sqlDB, Cfg: cfg, Limiter: ratelimit.New(100, time.Minute)}
	wsH := &handlers.WorkspaceHandler{DB: sqlDB}

	jwt := middleware.AuthJWT(middleware.JWTParser{Secret: cfg.JWTSecret})
	authAny := middleware.AuthJWTOrAPIKey(middleware.JWTParser{Secret: cfg.JWTSecret}, sqlDB)
	wsMem := middleware.RequireWorkspaceMember(sqlDB)
	wrap := func(next http.Handler) http.Handler {
		return authAny(wsMem(next))
	}

	mux.HandleFunc("POST /api/v1/auth/login", authH.Login)
	mux.Handle("GET /api/v1/workspaces", jwt(http.HandlerFunc(wsH.ListMine)))
	handlers.RegisterIntegrations(mux, sqlDB, wrap)
	handlers.MountBots(mux, sqlDB, wrap)
	return mux
}

func TestBotsMeAPIKey(t *testing.T) {
	mux := botsMeMux(t)
	token, wsID := loginAdmin(t, mux)

	body, _ := json.Marshal(map[string]string{"name": "me-key"})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+wsID+"/api-keys", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create key %d %s", rr.Code, rr.Body.String())
	}
	var keyRes map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &keyRes)
	raw, _ := keyRes["token"].(string)
	apiKey, _ := keyRes["api_key"].(map[string]any)
	boundMember, _ := apiKey["member_id"].(string)
	if raw == "" || boundMember == "" {
		t.Fatalf("bad create key response: %v", keyRes)
	}

	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+wsID+"/bots/me", nil)
	req2.Header.Set("Authorization", "Bearer "+raw)
	mux.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("bots/me %d %s", rr2.Code, rr2.Body.String())
	}
	var me map[string]any
	_ = json.Unmarshal(rr2.Body.Bytes(), &me)
	if me["workspace_id"] != wsID {
		t.Fatalf("workspace_id: %v", me["workspace_id"])
	}
	if me["member_id"] != boundMember {
		t.Fatalf("member_id got %v want %v", me["member_id"], boundMember)
	}
	if me["actor_kind"] != "bot" {
		t.Fatalf("actor_kind: %v", me["actor_kind"])
	}
	if me["display_name"] == "" {
		t.Fatalf("display_name empty: %v", me)
	}
	if me["role"] == "" {
		t.Fatalf("role empty: %v", me)
	}
}

func TestBotsMeHumanJWT(t *testing.T) {
	mux := botsMeMux(t)
	token, wsID := loginAdmin(t, mux)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+wsID+"/bots/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("bots/me jwt %d %s", rr.Code, rr.Body.String())
	}
	var me map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &me)
	if me["workspace_id"] != wsID {
		t.Fatalf("workspace_id: %v", me["workspace_id"])
	}
	if me["member_id"] == "" {
		t.Fatalf("member_id empty: %v", me)
	}
	if me["actor_kind"] != "human" {
		t.Fatalf("actor_kind: %v", me["actor_kind"])
	}
	if me["display_name"] == "" || me["role"] == "" {
		t.Fatalf("incomplete: %v", me)
	}
}
