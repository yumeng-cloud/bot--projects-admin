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

func redeemJoinMux(t *testing.T) *http.ServeMux {
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
	// production MountTeams wrap = authAny(wsMem)；redeem 路径不走 wrap（自行可选鉴权）
	wrap := func(next http.Handler) http.Handler { return authAny(wsMem(next)) }
	mux.HandleFunc("POST /api/v1/auth/login", authH.Login)
	mux.Handle("GET /api/v1/workspaces", jwt(http.HandlerFunc(wsH.ListMine)))
	handlers.MountTeams(mux, sqlDB, wrap)
	return mux
}

func createTeamWithInvite(t *testing.T, mux *http.ServeMux, token, wsID string) (teamID, code, seatSlug string) {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"name": "入队自动身份", "template_key": "default"})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+wsID+"/teams/from-template", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated && rr.Code != http.StatusOK {
		// fallback plain create + ensure
		body, _ = json.Marshal(map[string]string{"name": "入队自动身份"})
		rr = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+wsID+"/teams", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		mux.ServeHTTP(rr, req)
		if rr.Code != http.StatusCreated {
			t.Fatalf("create team %d %s", rr.Code, rr.Body.String())
		}
	}
	var teamRes struct {
		Team struct {
			ID string `json:"id"`
		} `json:"team"`
		Roster []struct {
			ID       string `json:"id"`
			RoleSlug string `json:"role_slug"`
			Status   string `json:"status"`
		} `json:"roster"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &teamRes)
	teamID = teamRes.Team.ID
	if teamID == "" {
		t.Fatalf("no team: %s", rr.Body.String())
	}
	if len(teamRes.Roster) == 0 {
		rr = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+wsID+"/teams/"+teamID+"/roster/ensure", bytes.NewReader([]byte("{}")))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		mux.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK && rr.Code != http.StatusCreated {
			t.Fatalf("ensure %d %s", rr.Code, rr.Body.String())
		}
		var ens struct {
			Roster []struct {
				ID       string `json:"id"`
				RoleSlug string `json:"role_slug"`
			} `json:"roster"`
		}
		// ensure may return array or wrapped
		raw := rr.Body.Bytes()
		_ = json.Unmarshal(raw, &ens)
		if len(ens.Roster) == 0 {
			var arr []struct {
				ID       string `json:"id"`
				RoleSlug string `json:"role_slug"`
			}
			if json.Unmarshal(raw, &arr) == nil && len(arr) > 0 {
				seatSlug = arr[0].RoleSlug
			}
		} else {
			seatSlug = ens.Roster[0].RoleSlug
		}
	} else {
		seatSlug = teamRes.Roster[0].RoleSlug
	}
	if seatSlug == "" {
		seatSlug = "frontend"
	}
	invBody, _ := json.Marshal(map[string]any{"seat_slug": seatSlug, "role_hint": seatSlug})
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+wsID+"/teams/"+teamID+"/invites", bytes.NewReader(invBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("invite %d %s", rr.Code, rr.Body.String())
	}
	var invRes struct {
		Code string `json:"code"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &invRes)
	if !strings.HasPrefix(invRes.Code, "ymt_") {
		t.Fatalf("code=%q", invRes.Code)
	}
	return teamID, invRes.Code, seatSlug
}

func TestHTTPRedeemBootstrapNoBearer(t *testing.T) {
	mux := redeemJoinMux(t)
	token, wsID := loginAdmin(t, mux)
	_, code, _ := createTeamWithInvite(t, mux, token, wsID)

	body, _ := json.Marshal(map[string]string{"code": code})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/team-invites/redeem", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// 无 Bearer
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("bootstrap %d %s", rr.Code, rr.Body.String())
	}
	var res map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &res)
	tok, _ := res["token"].(string)
	if !strings.HasPrefix(tok, "ymk_") {
		t.Fatalf("token=%v body=%s", res["token"], rr.Body.String())
	}
	if res["token_issued"] != true {
		t.Fatalf("token_issued=%v", res["token_issued"])
	}
	if res["member_id"] == nil || res["member_id"] == "" {
		t.Fatalf("member_id missing: %s", rr.Body.String())
	}
	if res["seat_id"] == nil || res["seat_id"] == "" {
		t.Fatalf("seat_id missing: %s", rr.Body.String())
	}

	// 同 ymk_ 再 redeem → 不重发
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/team-invites/redeem", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tok)
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("idempotent %d %s", rr.Code, rr.Body.String())
	}
	var res2 map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &res2)
	if res2["token"] != nil && res2["token"] != "" {
		t.Fatalf("reissued token: %v", res2["token"])
	}
	if res2["token_issued"] == true {
		t.Fatal("token_issued true on renew")
	}

	// V1：同 ymt_ 无 Bearer 第二次
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/team-invites/redeem", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("V1 want 409 got %d %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "请粘贴已发的身份钥匙续队") {
		t.Fatalf("V1 msg: %s", rr.Body.String())
	}
}
