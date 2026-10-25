package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/auth"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/config"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/handlers"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/ratelimit"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/seed"
)

func teamPatchEnv(t *testing.T) (*http.ServeMux, *sql.DB) {
	t.Helper()
	dir := t.TempDir()
	dsn := filepath.Join(dir, "t.db")
	sqlDB, err := db.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	cfg := config.Config{
		JWTSecret:       "test-secret-team-patch",
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
	mux.HandleFunc("POST /api/v1/auth/register", authH.Register)
	mux.Handle("GET /api/v1/workspaces", jwt(http.HandlerFunc(wsH.ListMine)))
	handlers.MountTeams(mux, sqlDB, wrap)
	return mux, sqlDB
}

func createTeamForPatch(t *testing.T, mux *http.ServeMux, token, wsID, name string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"name": name})
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
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"team"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &teamRes)
	if teamRes.Team.ID == "" {
		t.Fatalf("no team id: %s", rr.Body.String())
	}
	return teamRes.Team.ID
}

func patchTeam(t *testing.T, mux *http.ServeMux, token, wsID, teamID string, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/workspaces/"+wsID+"/teams/"+teamID, bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(rr, req)
	return rr
}

func TestPatchTeamRenameOK(t *testing.T) {
	mux, _ := teamPatchEnv(t)
	token, wsID := loginAdmin(t, mux)
	teamID := createTeamForPatch(t, mux, token, wsID, "旧队名")

	rr := patchTeam(t, mux, token, wsID, teamID, map[string]string{"name": "新队名"})
	if rr.Code != http.StatusOK {
		t.Fatalf("patch %d %s", rr.Code, rr.Body.String())
	}
	var res struct {
		Team struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"team"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &res)
	if res.Team.ID != teamID || res.Team.Name != "新队名" {
		t.Fatalf("got %+v body=%s", res.Team, rr.Body.String())
	}

	// GET 确认持久化
	rr2 := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+wsID+"/teams/"+teamID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	mux.ServeHTTP(rr2, req)
	if rr2.Code != http.StatusOK {
		t.Fatalf("get %d %s", rr2.Code, rr2.Body.String())
	}
	_ = json.Unmarshal(rr2.Body.Bytes(), &res)
	if res.Team.Name != "新队名" {
		t.Fatalf("persisted name=%q", res.Team.Name)
	}
}

func TestPatchTeamEmptyNameRejected(t *testing.T) {
	mux, _ := teamPatchEnv(t)
	token, wsID := loginAdmin(t, mux)
	teamID := createTeamForPatch(t, mux, token, wsID, "可改")

	for _, name := range []string{"", "   "} {
		rr := patchTeam(t, mux, token, wsID, teamID, map[string]string{"name": name})
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("empty %q: want 400 got %d %s", name, rr.Code, rr.Body.String())
		}
		if !strings.Contains(rr.Body.String(), "团队名称不能为空") {
			t.Fatalf("empty %q: want 中文空名错误, body=%s", name, rr.Body.String())
		}
	}
}

func TestPatchTeamNonAdminForbidden(t *testing.T) {
	mux, sqlDB := teamPatchEnv(t)
	adminTok, wsID := loginAdmin(t, mux)
	teamID := createTeamForPatch(t, mux, adminTok, wsID, "权限队")

	// 注册普通用户并插入 member 角色
	hash, err := auth.HashPassword("Member123!")
	if err != nil {
		t.Fatal(err)
	}
	uid := idgen.New()
	mid := idgen.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := sqlDB.Exec(`INSERT INTO users (id, email, password_hash, name, created_at, updated_at) VALUES (?,?,?,?,?,?)`,
		uid, "member@patch.test", hash, "普通成员", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := sqlDB.Exec(`INSERT INTO members (id, workspace_id, user_id, role, display_name, is_bot, bot_enabled) VALUES (?,?,?,?,?,?,?)`,
		mid, wsID, uid, "member", "普通成员", 0, 1); err != nil {
		t.Fatal(err)
	}

	body, _ := json.Marshal(map[string]string{"email": "member@patch.test", "password": "Member123!"})
	rrLogin := httptest.NewRecorder()
	mux.ServeHTTP(rrLogin, httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body)))
	if rrLogin.Code != http.StatusOK {
		t.Fatalf("member login %d %s", rrLogin.Code, rrLogin.Body.String())
	}
	var loginRes map[string]any
	_ = json.Unmarshal(rrLogin.Body.Bytes(), &loginRes)
	memberTok, _ := loginRes["token"].(string)
	if memberTok == "" {
		t.Fatal("no member token")
	}

	rr := patchTeam(t, mux, memberTok, wsID, teamID, map[string]string{"name": "偷改"})
	if rr.Code != http.StatusForbidden {
		t.Fatalf("want 403 got %d %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "需要管理员权限") {
		t.Fatalf("want 人话 403, body=%s", rr.Body.String())
	}
}
