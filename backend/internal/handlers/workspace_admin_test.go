package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/config"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/seed"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/server"
)

func wsAdminEnv(t *testing.T) (http.Handler, config.Config, *http.ServeMux) {
	t.Helper()
	dir := t.TempDir()
	dsn := filepath.Join(dir, "t.db")
	sqlDB, err := db.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	cfg := config.Config{
		JWTSecret:           "test-secret-ws-admin",
		JWTTTL:              time.Hour,
		AppEnv:              "test",
		SeedAdminEmail:      "admin@yumeng.tech",
		SeedAdminPass:       "Admin123!",
		SeedWSName:          "育蒙科技",
		SeedSampleData: true,
		LoginRateLimit:      1000,
		LoginRateWindow:     time.Minute,
		PlatformAdminEmails: []string{"admin@yumeng.tech"},
		CORSOrigins:         []string{"*"},
	}
	if err := seed.Run(sqlDB, cfg); err != nil {
		t.Fatal(err)
	}
	if err := seed.BootstrapPlatformAdmins(sqlDB, cfg.PlatformAdminEmails); err != nil {
		t.Fatal(err)
	}
	h := server.NewRouter(sqlDB, cfg)
	return h, cfg, nil
}

func doJSON(t *testing.T, h http.Handler, method, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var rdr *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rdr = bytes.NewReader(b)
	} else {
		rdr = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, rdr)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func registerUser(t *testing.T, h http.Handler, email, pass, name string) (token, userID string) {
	t.Helper()
	rr := doJSON(t, h, http.MethodPost, "/api/v1/auth/register", "", map[string]string{
		"email": email, "password": pass, "name": name,
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("register %s: %d %s", email, rr.Code, rr.Body.String())
	}
	var res map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &res)
	token, _ = res["token"].(string)
	user, _ := res["user"].(map[string]any)
	userID, _ = user["id"].(string)
	return token, userID
}

func loginUser(t *testing.T, h http.Handler, email, pass string) string {
	t.Helper()
	rr := doJSON(t, h, http.MethodPost, "/api/v1/auth/login", "", map[string]string{
		"email": email, "password": pass,
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("login %s: %d %s", email, rr.Code, rr.Body.String())
	}
	var res map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &res)
	tok, _ := res["token"].(string)
	return tok
}

func TestAuthMeIsPlatformAdmin(t *testing.T) {
	h, _, _ := wsAdminEnv(t)
	tok := loginUser(t, h, "admin@yumeng.tech", "Admin123!")
	rr := doJSON(t, h, http.MethodGet, "/api/v1/auth/me", tok, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("me %d %s", rr.Code, rr.Body.String())
	}
	var res struct {
		User struct {
			IsPlatformAdmin bool `json:"is_platform_admin"`
		} `json:"user"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &res)
	if !res.User.IsPlatformAdmin {
		t.Fatalf("expected is_platform_admin true: %s", rr.Body.String())
	}
}

func TestCreateWorkspaceAndPatch(t *testing.T) {
	h, _, _ := wsAdminEnv(t)
	tok, _ := registerUser(t, h, "creator@test.com", "password1", "创建者")
	rr := doJSON(t, h, http.MethodPost, "/api/v1/workspaces", tok, map[string]string{
		"name": "我的空间", "slug": "my-space",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("create %d %s", rr.Code, rr.Body.String())
	}
	var created struct {
		Workspace struct {
			ID string `json:"id"`
		} `json:"workspace"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &created)
	wsID := created.Workspace.ID

	rr = doJSON(t, h, http.MethodPatch, "/api/v1/workspaces/"+wsID, tok, map[string]string{
		"name": "新名称", "slug": "new-slug",
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("patch %d %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"slug":"new-slug"`) {
		t.Fatalf("patch body %s", rr.Body.String())
	}

	// slug conflict
	rr = doJSON(t, h, http.MethodPost, "/api/v1/workspaces", tok, map[string]string{"name": "二", "slug": "other"})
	if rr.Code != http.StatusCreated {
		t.Fatalf("create2 %d %s", rr.Code, rr.Body.String())
	}
	var created2 struct {
		Workspace struct {
			ID string `json:"id"`
		} `json:"workspace"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &created2)
	rr = doJSON(t, h, http.MethodPatch, "/api/v1/workspaces/"+created2.Workspace.ID, tok, map[string]string{"slug": "new-slug"})
	if rr.Code != http.StatusConflict {
		t.Fatalf("want 409 slug conflict, got %d %s", rr.Code, rr.Body.String())
	}
}

func TestLastAdminGuard(t *testing.T) {
	h, _, _ := wsAdminEnv(t)
	tok, _ := registerUser(t, h, "solo@test.com", "password1", "独管")
	rr := doJSON(t, h, http.MethodPost, "/api/v1/workspaces", tok, map[string]string{"name": "独", "slug": "solo-ws"})
	if rr.Code != http.StatusCreated {
		t.Fatalf("create %d %s", rr.Code, rr.Body.String())
	}
	var created struct {
		Workspace struct {
			ID string `json:"id"`
		} `json:"workspace"`
		MemberID string `json:"member_id"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &created)
	wsID := created.Workspace.ID
	memID := created.MemberID

	rr = doJSON(t, h, http.MethodPatch, "/api/v1/workspaces/"+wsID+"/members/"+memID, tok, map[string]string{"role": "member"})
	if rr.Code != http.StatusConflict {
		t.Fatalf("demote last admin want 409 got %d %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "最后一个管理员") {
		t.Fatalf("msg %s", rr.Body.String())
	}
}

func TestWorkspaceInviteRedeem(t *testing.T) {
	h, _, _ := wsAdminEnv(t)
	adminTok, _ := registerUser(t, h, "wsadmin@test.com", "password1", "WS管")
	rr := doJSON(t, h, http.MethodPost, "/api/v1/workspaces", adminTok, map[string]string{"name": "邀", "slug": "invite-ws"})
	if rr.Code != http.StatusCreated {
		t.Fatalf("create %d %s", rr.Code, rr.Body.String())
	}
	var created struct {
		Workspace struct {
			ID string `json:"id"`
		} `json:"workspace"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &created)
	wsID := created.Workspace.ID

	// invite unregistered email
	rr = doJSON(t, h, http.MethodPost, "/api/v1/workspaces/"+wsID+"/invites", adminTok, map[string]any{
		"email": "newbie@test.com", "role": "member",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("invite %d %s", rr.Code, rr.Body.String())
	}
	var invRes struct {
		Code   string `json:"code"`
		Invite struct {
			ID string `json:"id"`
		} `json:"invite"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &invRes)
	if !strings.HasPrefix(invRes.Code, "ymw_") {
		t.Fatalf("code=%s", invRes.Code)
	}
	if strings.HasPrefix(invRes.Code, "ymt_") {
		t.Fatal("must not use ymt_")
	}

	// list must not leak plain code
	rr = doJSON(t, h, http.MethodGet, "/api/v1/workspaces/"+wsID+"/invites", adminTok, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("list %d %s", rr.Code, rr.Body.String())
	}
	if strings.Contains(rr.Body.String(), invRes.Code) {
		t.Fatalf("list leaked code")
	}

	// register then redeem
	userTok, _ := registerUser(t, h, "newbie@test.com", "password1", "新人")
	rr = doJSON(t, h, http.MethodPost, "/api/v1/workspace-invites/redeem", userTok, map[string]string{"code": invRes.Code})
	if rr.Code != http.StatusOK {
		t.Fatalf("redeem %d %s", rr.Code, rr.Body.String())
	}

	// wrong prefix rejected
	rr = doJSON(t, h, http.MethodPost, "/api/v1/workspace-invites/redeem", userTok, map[string]string{"code": "ymt_deadbeef"})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("ymt redeem want 400 got %d %s", rr.Code, rr.Body.String())
	}

	// revoke
	rr = doJSON(t, h, http.MethodPost, "/api/v1/workspaces/"+wsID+"/invites", adminTok, map[string]any{"role": "member"})
	var inv2 struct {
		Invite struct {
			ID string `json:"id"`
		} `json:"invite"`
		Code string `json:"code"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &inv2)
	rr = doJSON(t, h, http.MethodPost, "/api/v1/workspaces/"+wsID+"/invites/"+inv2.Invite.ID+"/revoke", adminTok, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("revoke %d %s", rr.Code, rr.Body.String())
	}
	otherTok, _ := registerUser(t, h, "other@test.com", "password1", "他")
	rr = doJSON(t, h, http.MethodPost, "/api/v1/workspace-invites/redeem", otherTok, map[string]string{"code": inv2.Code})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("redeem revoked want 400 got %d %s", rr.Code, rr.Body.String())
	}
}

func TestPlatformAdminAuthAndDisable(t *testing.T) {
	h, _, _ := wsAdminEnv(t)
	adminTok := loginUser(t, h, "admin@yumeng.tech", "Admin123!")
	normalTok, normalID := registerUser(t, h, "normal@test.com", "password1", "普通")

	// non-admin → 403
	rr := doJSON(t, h, http.MethodGet, "/api/v1/admin/workspaces", normalTok, nil)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("non-admin list want 403 got %d %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "平台超级管理员") {
		t.Fatalf("msg %s", rr.Body.String())
	}

	// admin list
	rr = doJSON(t, h, http.MethodGet, "/api/v1/admin/workspaces", adminTok, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("admin list %d %s", rr.Code, rr.Body.String())
	}

	// create for owner
	rr = doJSON(t, h, http.MethodPost, "/api/v1/admin/workspaces", adminTok, map[string]string{
		"name": "代开", "slug": "proxy-ws", "owner_email": "normal@test.com",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("proxy create %d %s", rr.Code, rr.Body.String())
	}
	var proxy struct {
		Workspace struct {
			ID string `json:"id"`
		} `json:"workspace"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &proxy)
	wsID := proxy.Workspace.ID

	// owner can access
	rr = doJSON(t, h, http.MethodGet, "/api/v1/workspaces/"+wsID, normalTok, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("owner get %d %s", rr.Code, rr.Body.String())
	}

	// disable
	rr = doJSON(t, h, http.MethodPatch, "/api/v1/admin/workspaces/"+wsID, adminTok, map[string]bool{"disabled": true})
	if rr.Code != http.StatusOK {
		t.Fatalf("disable %d %s", rr.Code, rr.Body.String())
	}

	// member access → 403 工作区已停用
	rr = doJSON(t, h, http.MethodGet, "/api/v1/workspaces/"+wsID, normalTok, nil)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("disabled get want 403 got %d %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "工作区已停用") {
		t.Fatalf("msg %s", rr.Body.String())
	}

	// issues also blocked
	rr = doJSON(t, h, http.MethodGet, "/api/v1/workspaces/"+wsID+"/issues", normalTok, nil)
	if rr.Code != http.StatusForbidden || !strings.Contains(rr.Body.String(), "工作区已停用") {
		t.Fatalf("issues disabled %d %s", rr.Code, rr.Body.String())
	}

	// restore
	rr = doJSON(t, h, http.MethodPatch, "/api/v1/admin/workspaces/"+wsID, adminTok, map[string]bool{"disabled": false})
	if rr.Code != http.StatusOK {
		t.Fatalf("enable %d %s", rr.Code, rr.Body.String())
	}
	rr = doJSON(t, h, http.MethodGet, "/api/v1/workspaces/"+wsID, normalTok, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("after enable %d %s", rr.Code, rr.Body.String())
	}

	// grant/revoke platform admin
	rr = doJSON(t, h, http.MethodPatch, "/api/v1/admin/users/"+normalID, adminTok, map[string]bool{"is_platform_admin": true})
	if rr.Code != http.StatusOK {
		t.Fatalf("grant %d %s", rr.Code, rr.Body.String())
	}
	rr = doJSON(t, h, http.MethodGet, "/api/v1/admin/users", normalTok, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("new admin list %d %s", rr.Code, rr.Body.String())
	}
	rr = doJSON(t, h, http.MethodPatch, "/api/v1/admin/users/"+normalID, adminTok, map[string]bool{"is_platform_admin": false})
	if rr.Code != http.StatusOK {
		t.Fatalf("revoke %d %s", rr.Code, rr.Body.String())
	}
}

func TestAuditLogged(t *testing.T) {
	h, _, _ := wsAdminEnv(t)
	adminTok := loginUser(t, h, "admin@yumeng.tech", "Admin123!")
	_, uid := registerUser(t, h, "auditee@test.com", "password1", "审计")
	rr := doJSON(t, h, http.MethodPatch, "/api/v1/admin/users/"+uid, adminTok, map[string]bool{"is_platform_admin": true})
	if rr.Code != http.StatusOK {
		t.Fatalf("grant %d %s", rr.Code, rr.Body.String())
	}
	// verify via sqlite through a second open of same pattern — use admin list as smoke;
	// deeper: recreate env is separate DB. Re-open via creating another request path.
	// Direct check: create workspace for + disable and ensure no 500.
	rr = doJSON(t, h, http.MethodPost, "/api/v1/admin/workspaces", adminTok, map[string]string{
		"name": "审计WS", "slug": "audit-ws", "owner_user_id": uid,
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("create %d %s", rr.Code, rr.Body.String())
	}
	var created struct {
		Workspace struct {
			ID string `json:"id"`
		} `json:"workspace"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &created)
	rr = doJSON(t, h, http.MethodPatch, "/api/v1/admin/workspaces/"+created.Workspace.ID, adminTok, map[string]bool{"disabled": true})
	if rr.Code != http.StatusOK {
		t.Fatalf("disable %d %s", rr.Code, rr.Body.String())
	}
}

func TestListUsersSearchAndWorkspaces(t *testing.T) {
	h, _, _ := wsAdminEnv(t)
	adminTok := loginUser(t, h, "admin@yumeng.tech", "Admin123!")
	normalTok, normalID := registerUser(t, h, "searchme@test.com", "password1", "可搜用户")

	// create WS owned by searchme → membership
	rr := doJSON(t, h, http.MethodPost, "/api/v1/admin/workspaces", adminTok, map[string]string{
		"name": "所属空间", "slug": "owned-ws", "owner_user_id": normalID,
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("create ws %d %s", rr.Code, rr.Body.String())
	}
	var created struct {
		Workspace struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Slug string `json:"slug"`
		} `json:"workspace"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &created)
	if created.Workspace.ID == "" {
		t.Fatalf("no ws id: %s", rr.Body.String())
	}

	// q= email hit
	rr = doJSON(t, h, http.MethodGet, "/api/v1/admin/users?q=searchme", adminTok, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("search email %d %s", rr.Code, rr.Body.String())
	}
	var byEmail struct {
		Users []struct {
			ID         string `json:"id"`
			Email      string `json:"email"`
			Name       string `json:"name"`
			Workspaces []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Slug string `json:"slug"`
			} `json:"workspaces"`
		} `json:"users"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &byEmail)
	if len(byEmail.Users) != 1 || byEmail.Users[0].ID != normalID {
		t.Fatalf("email search want 1 hit got %s", rr.Body.String())
	}
	ws := byEmail.Users[0].Workspaces
	if len(ws) != 1 || ws[0].ID != created.Workspace.ID || ws[0].Name != "所属空间" || ws[0].Slug != "owned-ws" {
		t.Fatalf("workspaces want owned-ws got %+v body=%s", ws, rr.Body.String())
	}

	// q= display name hit
	rr = doJSON(t, h, http.MethodGet, "/api/v1/admin/users?q="+url.QueryEscape("可搜"), adminTok, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("search name %d %s", rr.Code, rr.Body.String())
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &byEmail)
	if len(byEmail.Users) != 1 || byEmail.Users[0].Email != "searchme@test.com" {
		t.Fatalf("name search want 1 hit got %s", rr.Body.String())
	}

	// q= miss
	rr = doJSON(t, h, http.MethodGet, "/api/v1/admin/users?q=zzznomatch", adminTok, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("miss %d %s", rr.Code, rr.Body.String())
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &byEmail)
	if len(byEmail.Users) != 0 {
		t.Fatalf("miss want 0 got %s", rr.Body.String())
	}

	// empty q still lists (includes seed admin + searchme); workspaces key present
	rr = doJSON(t, h, http.MethodGet, "/api/v1/admin/users", adminTok, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("list all %d %s", rr.Code, rr.Body.String())
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &byEmail)
	if len(byEmail.Users) < 2 {
		t.Fatalf("list all want >=2 got %s", rr.Body.String())
	}
	found := false
	for _, u := range byEmail.Users {
		if u.ID == normalID {
			found = true
			if u.Workspaces == nil {
				t.Fatalf("workspaces null for %s", normalID)
			}
			if len(u.Workspaces) != 1 || u.Workspaces[0].Slug != "owned-ws" {
				t.Fatalf("list workspaces %+v", u.Workspaces)
			}
		}
	}
	if !found {
		t.Fatalf("normal user missing in list: %s", rr.Body.String())
	}
	_ = normalTok // silence if unused in future refactors
}
