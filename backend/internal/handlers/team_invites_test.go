package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/config"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/handlers"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/ratelimit"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/seed"
)

func teamInvitesMux(t *testing.T) *http.ServeMux {
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
	handlers.MountTeams(mux, sqlDB, wrap)
	return mux
}

func TestListInvitesAdmin(t *testing.T) {
	mux := teamInvitesMux(t)
	token, wsID := loginAdmin(t, mux)

	body, _ := json.Marshal(map[string]string{"name": "班底-邀请列表"})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+wsID+"/teams", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create team %d %s", rr.Code, rr.Body.String())
	}
	var teamRes struct {
		Team struct {
			ID string `json:"id"`
		} `json:"team"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &teamRes)
	teamID := teamRes.Team.ID
	if teamID == "" {
		t.Fatal("no team id")
	}

	// empty list
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+wsID+"/teams/"+teamID+"/invites", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("list empty %d %s", rr.Code, rr.Body.String())
	}
	var empty struct {
		Invites []any `json:"invites"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &empty)
	if empty.Invites == nil || len(empty.Invites) != 0 {
		t.Fatalf("want empty slice, got %#v", empty.Invites)
	}

	invBody, _ := json.Marshal(map[string]any{"allow_expand_seat": true, "role_hint": "frontend", "max_uses": 3})
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+wsID+"/teams/"+teamID+"/invites", bytes.NewReader(invBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create invite %d %s", rr.Code, rr.Body.String())
	}
	var createRes map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &createRes)
	code, _ := createRes["code"].(string)
	if !strings.HasPrefix(code, "ymt_") {
		t.Fatalf("create code=%v", createRes["code"])
	}
	invObj, _ := createRes["invite"].(map[string]any)
	if invObj["code"] == nil || invObj["code"] == "" {
		t.Fatal("create invite.code missing")
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+wsID+"/teams/"+teamID+"/invites", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("list %d %s", rr.Code, rr.Body.String())
	}
	raw := rr.Body.String()
	if strings.Contains(raw, code) {
		t.Fatalf("list leaked plain code: %s", raw)
	}
	if strings.Contains(raw, `"token_hash"`) {
		t.Fatalf("list leaked token_hash: %s", raw)
	}
	var listRes struct {
		Invites []map[string]any `json:"invites"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &listRes)
	if len(listRes.Invites) != 1 {
		t.Fatalf("len=%d body=%s", len(listRes.Invites), raw)
	}
	item := listRes.Invites[0]
	if item["status"] != "active" {
		t.Fatalf("status=%v", item["status"])
	}
	if item["code"] != nil && item["code"] != "" {
		t.Fatalf("list invite.code=%v", item["code"])
	}
	if item["allow_expand_seat"] != true {
		t.Fatalf("allow_expand_seat=%v", item["allow_expand_seat"])
	}
	uc, _ := item["use_count"].(float64)
	if uc != 0 {
		t.Fatalf("use_count=%v", item["use_count"])
	}
	if item["code_tail"] == nil || item["code_tail"] == "" {
		t.Fatalf("code_tail missing: %#v", item)
	}
}
