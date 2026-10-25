package handlers_test

import (
	"bytes"
	"database/sql"
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
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/ratelimit"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/seed"
)

func botsIsolationMux(t *testing.T) (*http.ServeMux, *sql.DB, config.Config) {
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
		SeedSampleData:  true,
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
	mux.Handle("POST /api/v1/workspaces", jwt(http.HandlerFunc(wsH.Create)))
	handlers.RegisterIntegrations(mux, sqlDB, wrap)
	handlers.MountBots(mux, sqlDB, wrap)
	return mux, sqlDB, cfg
}

func seedForeignBot(t *testing.T, sqlDB *sql.DB) (ws2, midWS1inForeign string) {
	t.Helper()
	// mid that belongs to ws1 will be created by caller; here create ws2 + its own bot
	ws2 = idgen.New()
	if _, err := sqlDB.Exec(`INSERT INTO workspaces (id, name, slug) VALUES (?,?,?)`, ws2, "他仓", "other-"+ws2[:8]); err != nil {
		t.Fatal(err)
	}
	mid := idgen.New()
	if _, err := sqlDB.Exec(`INSERT INTO members (id, workspace_id, user_id, role, display_name, is_bot, bot_enabled) VALUES (?,?,NULL,'bot','ForeignBot',1,1)`, mid, ws2); err != nil {
		t.Fatal(err)
	}
	return ws2, mid
}

func firstBotMember(t *testing.T, sqlDB *sql.DB, wsID string) string {
	t.Helper()
	var mid string
	err := sqlDB.QueryRow(`SELECT id FROM members WHERE workspace_id = ? AND is_bot = 1 LIMIT 1`, wsID).Scan(&mid)
	if err != nil {
		t.Fatalf("no bot member in %s: %v", wsID, err)
	}
	return mid
}

func TestWSIsolationMemoriesTodosHTTP(t *testing.T) {
	mux, sqlDB, _ := botsIsolationMux(t)
	token, ws1 := loginAdmin(t, mux)
	mid1 := firstBotMember(t, sqlDB, ws1)
	ws2, mid2 := seedForeignBot(t, sqlDB)

	auth := func(req *http.Request) {
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
	}

	// 同仓 POST memory → 201
	body, _ := json.Marshal(map[string]string{"key_name": "ok", "value_text": "v"})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+ws1+"/bots/"+mid1+"/memories", bytes.NewReader(body))
	auth(req)
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("same-ws POST memory %d %s", rr.Code, rr.Body.String())
	}

	// 同仓 GET memories → 200
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+ws1+"/bots/"+mid1+"/memories", nil)
	auth(req)
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("same-ws GET memories %d %s", rr.Code, rr.Body.String())
	}

	// 错仓 mid：ws2 path + mid1 → 404（禁 200 空 / 201）
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+ws2+"/bots/"+mid1+"/memories", nil)
	auth(req)
	mux.ServeHTTP(rr, req)
	// admin 非 ws2 成员 → 403；先把 admin 加进 ws2 再测 path mid 404
	if rr.Code != http.StatusForbidden && rr.Code != http.StatusNotFound {
		t.Fatalf("cross-ws GET without membership: want 403/404, got %d %s", rr.Code, rr.Body.String())
	}

	// 将 admin 加入 ws2，再测 path mid∉ws → 404
	uid := ""
	_ = sqlDB.QueryRow(`SELECT id FROM users WHERE email = ?`, "admin@yumeng.tech").Scan(&uid)
	adminMid2 := idgen.New()
	if _, err := sqlDB.Exec(`INSERT INTO members (id, workspace_id, user_id, role, display_name, is_bot) VALUES (?,?,?,'admin','Admin',0)`, adminMid2, ws2, uid); err != nil {
		t.Fatal(err)
	}

	must404 := func(method, path string, body []byte) {
		t.Helper()
		rr := httptest.NewRecorder()
		var rdr *bytes.Reader
		if body != nil {
			rdr = bytes.NewReader(body)
		} else {
			rdr = bytes.NewReader(nil)
		}
		req := httptest.NewRequest(method, path, rdr)
		auth(req)
		mux.ServeHTTP(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("%s %s: want 404, got %d %s", method, path, rr.Code, rr.Body.String())
		}
		var er map[string]string
		_ = json.Unmarshal(rr.Body.Bytes(), &er)
		if er["message"] != "成员不存在" {
			t.Fatalf("%s %s: message=%q want 成员不存在 body=%s", method, path, er["message"], rr.Body.String())
		}
	}

	memBody, _ := json.Marshal(map[string]string{"key_name": "evil", "value_text": "x"})
	todoBody, _ := json.Marshal(map[string]string{"title": "evil"})
	must404(http.MethodGet, "/api/v1/workspaces/"+ws2+"/bots/"+mid1+"/memories", nil)
	must404(http.MethodPost, "/api/v1/workspaces/"+ws2+"/bots/"+mid1+"/memories", memBody)
	must404(http.MethodGet, "/api/v1/workspaces/"+ws2+"/bots/"+mid1+"/todos", nil)
	must404(http.MethodPost, "/api/v1/workspaces/"+ws2+"/bots/"+mid1+"/todos", todoBody)
	must404(http.MethodGet, "/api/v1/workspaces/"+ws1+"/bots/"+mid2+"/memories", nil)
	must404(http.MethodPost, "/api/v1/workspaces/"+ws1+"/bots/"+mid2+"/todos", todoBody)

	// 同仓 todos 201
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+ws1+"/bots/"+mid1+"/todos", bytes.NewReader(todoBody))
	auth(req)
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("same-ws POST todo %d %s", rr.Code, rr.Body.String())
	}
}

func TestBotKeyWrongWorkspace403(t *testing.T) {
	mux, sqlDB, _ := botsIsolationMux(t)
	token, ws1 := loginAdmin(t, mux)
	ws2, _ := seedForeignBot(t, sqlDB)

	body, _ := json.Marshal(map[string]string{"name": "iso-key"})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+ws1+"/api-keys", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create key %d %s", rr.Code, rr.Body.String())
	}
	var keyRes map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &keyRes)
	raw, _ := keyRes["token"].(string)
	if raw == "" {
		t.Fatal("empty api key token")
	}

	mid1 := firstBotMember(t, sqlDB, ws1)
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+ws2+"/bots/"+mid1+"/memories", nil)
	req.Header.Set("Authorization", "Bearer "+raw)
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("bot key wrong ws: want 403, got %d %s", rr.Code, rr.Body.String())
	}
	var er map[string]string
	_ = json.Unmarshal(rr.Body.Bytes(), &er)
	if er["message"] != "API Key 不属于该工作空间" {
		t.Fatalf("message=%q want API Key 不属于该工作空间", er["message"])
	}
}
