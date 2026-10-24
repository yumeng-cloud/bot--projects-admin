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

func testMux(t *testing.T) (*http.ServeMux, config.Config, string) {
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
		LoginRateWindow: time.Minute,
	}
	if err := seed.Run(sqlDB, cfg); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	authH := &handlers.AuthHandler{DB: sqlDB, Cfg: cfg, Limiter: ratelimit.New(100, time.Minute)}
	wsH := &handlers.WorkspaceHandler{DB: sqlDB}
	memH := &handlers.MemberHandler{DB: sqlDB}

	jwt := middleware.AuthJWT(middleware.JWTParser{Secret: cfg.JWTSecret})
	authAny := middleware.AuthJWTOrAPIKey(middleware.JWTParser{Secret: cfg.JWTSecret}, sqlDB)
	wsMem := middleware.RequireWorkspaceMember(sqlDB)
	wrap := func(next http.Handler) http.Handler {
		return authAny(wsMem(next))
	}

	mux.HandleFunc("POST /api/v1/auth/login", authH.Login)
	mux.HandleFunc("POST /api/v1/auth/register", authH.Register)
	mux.Handle("GET /api/v1/workspaces", jwt(http.HandlerFunc(wsH.ListMine)))
	mux.Handle("GET /api/v1/workspaces/{workspaceID}/members", wrap(http.HandlerFunc(memH.List)))
	handlers.RegisterProjects(mux, sqlDB, wrap)
	handlers.RegisterIntegrations(mux, sqlDB, wrap)
	return mux, cfg, dsn
}

func loginAdmin(t *testing.T, mux *http.ServeMux) (token, wsID string) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"email": "admin@yumeng.tech", "password": "Admin123!"})
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body)))
	if rr.Code != http.StatusOK {
		t.Fatalf("login %d %s", rr.Code, rr.Body.String())
	}
	var res map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &res)
	token, _ = res["token"].(string)

	rr2 := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	mux.ServeHTTP(rr2, req)
	if rr2.Code != http.StatusOK {
		t.Fatalf("workspaces %d %s", rr2.Code, rr2.Body.String())
	}
	var wsRes struct {
		Workspaces []struct {
			ID string `json:"id"`
		} `json:"workspaces"`
	}
	_ = json.Unmarshal(rr2.Body.Bytes(), &wsRes)
	if len(wsRes.Workspaces) == 0 {
		t.Fatal("no workspace")
	}
	return token, wsRes.Workspaces[0].ID
}

func TestProjectCreateAndList(t *testing.T) {
	mux, _, _ := testMux(t)
	token, wsID := loginAdmin(t, mux)

	body, _ := json.Marshal(map[string]string{"name": "新项目", "key_prefix": "NP"})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+wsID+"/projects", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create project %d %s", rr.Code, rr.Body.String())
	}

	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+wsID+"/projects", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	mux.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("list %d %s", rr2.Code, rr2.Body.String())
	}
	var list struct {
		Projects []map[string]any `json:"projects"`
	}
	_ = json.Unmarshal(rr2.Body.Bytes(), &list)
	if len(list.Projects) < 2 { // seed YM + NP
		t.Fatalf("expected >=2 projects, got %d", len(list.Projects))
	}
}

func TestAPIKeyAuthAndNonMember403(t *testing.T) {
	mux, cfg, _ := testMux(t)
	token, wsID := loginAdmin(t, mux)

	// create API key
	body, _ := json.Marshal(map[string]string{"name": "bot-key"})
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
	if raw == "" || raw[:4] != "ymk_" {
		t.Fatalf("bad token: %v", keyRes["token"])
	}

	// API key can list projects
	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+wsID+"/projects", nil)
	req2.Header.Set("Authorization", "Bearer "+raw)
	mux.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("bot list projects %d %s", rr2.Code, rr2.Body.String())
	}

	// register outsider, expect 403 on workspace projects
	ob, _ := json.Marshal(map[string]string{"email": "out@test.com", "password": "password1", "name": "外"})
	rr3 := httptest.NewRecorder()
	mux.ServeHTTP(rr3, httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(ob)))
	if rr3.Code != http.StatusCreated {
		t.Fatalf("register outsider %d %s", rr3.Code, rr3.Body.String())
	}
	var oreg map[string]any
	_ = json.Unmarshal(rr3.Body.Bytes(), &oreg)
	otok, _ := oreg["token"].(string)

	rr4 := httptest.NewRecorder()
	req4 := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+wsID+"/projects", nil)
	req4.Header.Set("Authorization", "Bearer "+otok)
	mux.ServeHTTP(rr4, req4)
	if rr4.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d %s", rr4.Code, rr4.Body.String())
	}
	var errBody map[string]string
	_ = json.Unmarshal(rr4.Body.Bytes(), &errBody)
	if errBody["code"] != "forbidden" {
		t.Fatalf("error body: %v", errBody)
	}
	_ = cfg
}
