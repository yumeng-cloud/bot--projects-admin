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

func prefsRouter(t *testing.T) (http.Handler, string, string, string) {
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
	token, _ := login["access_token"].(string)
	if token == "" {
		token, _ = login["token"].(string)
	}

	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	h.ServeHTTP(rr2, req2)
	var wsWrap struct {
		Workspaces []struct {
			ID string `json:"id"`
		} `json:"workspaces"`
	}
	_ = json.Unmarshal(rr2.Body.Bytes(), &wsWrap)
	wsID := ""
	if len(wsWrap.Workspaces) > 0 {
		wsID = wsWrap.Workspaces[0].ID
	} else {
		var alt map[string]any
		_ = json.Unmarshal(rr2.Body.Bytes(), &alt)
		if list, ok := alt["workspaces"].([]any); ok && len(list) > 0 {
			wsID, _ = list[0].(map[string]any)["id"].(string)
		}
	}
	if wsID == "" {
		t.Fatalf("no workspace: %s", rr2.Body.String())
	}

	rr3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+wsID+"/projects", nil)
	req3.Header.Set("Authorization", "Bearer "+token)
	h.ServeHTTP(rr3, req3)
	var projWrap struct {
		Projects []struct {
			ID string `json:"id"`
		} `json:"projects"`
	}
	_ = json.Unmarshal(rr3.Body.Bytes(), &projWrap)
	if len(projWrap.Projects) == 0 {
		t.Fatalf("no project: %s", rr3.Body.String())
	}
	return h, token, wsID, projWrap.Projects[0].ID
}

func TestBoardColumnPrefsGetDefaultPutDelete(t *testing.T) {
	h, token, wsID, projID := prefsRouter(t)
	base := "/api/v1/workspaces/" + wsID + "/projects/" + projID + "/board-column-prefs"

	// GET default five
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, base, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("GET %d %s", rr.Code, rr.Body.String())
	}
	var getResp struct {
		Prefs struct {
			ProjectID string `json:"project_id"`
			Columns   []struct {
				Status string `json:"status"`
			} `json:"columns"`
		} `json:"prefs"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &getResp); err != nil {
		t.Fatal(err)
	}
	if getResp.Prefs.ProjectID != projID {
		t.Fatalf("project_id=%s", getResp.Prefs.ProjectID)
	}
	if len(getResp.Prefs.Columns) != 5 {
		t.Fatalf("want 5 default cols, got %d", len(getResp.Prefs.Columns))
	}

	// PUT custom
	putBody := []byte(`{"template_id":"simple-3","columns":[{"status":"todo"},{"status":"in_progress","label":"Doing"},{"status":"done","hidden":true}]}`)
	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPut, base, bytes.NewReader(putBody))
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Idempotency-Key", "prefs-put-1")
	h.ServeHTTP(rr2, req2)
	if rr2.Code != 200 {
		t.Fatalf("PUT %d %s", rr2.Code, rr2.Body.String())
	}
	var putResp struct {
		Prefs struct {
			TemplateID *string `json:"template_id"`
			Columns    []struct {
				Status string  `json:"status"`
				Label  *string `json:"label"`
				Hidden *bool   `json:"hidden"`
			} `json:"columns"`
		} `json:"prefs"`
	}
	_ = json.Unmarshal(rr2.Body.Bytes(), &putResp)
	if putResp.Prefs.TemplateID == nil || *putResp.Prefs.TemplateID != "simple-3" {
		t.Fatalf("template_id=%v", putResp.Prefs.TemplateID)
	}
	if len(putResp.Prefs.Columns) != 3 {
		t.Fatalf("cols=%d", len(putResp.Prefs.Columns))
	}

	// invalid status
	rrBad := httptest.NewRecorder()
	reqBad := httptest.NewRequest(http.MethodPut, base, bytes.NewReader([]byte(`{"columns":[{"status":"nope"}]}`)))
	reqBad.Header.Set("Authorization", "Bearer "+token)
	reqBad.Header.Set("Content-Type", "application/json")
	reqBad.Header.Set("Idempotency-Key", "prefs-put-bad")
	h.ServeHTTP(rrBad, reqBad)
	if rrBad.Code != 400 {
		t.Fatalf("want 400, got %d %s", rrBad.Code, rrBad.Body.String())
	}

	// DELETE reset
	rr3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodDelete, base, nil)
	req3.Header.Set("Authorization", "Bearer "+token)
	req3.Header.Set("Idempotency-Key", "prefs-del-1")
	h.ServeHTTP(rr3, req3)
	if rr3.Code != 200 {
		t.Fatalf("DELETE %d %s", rr3.Code, rr3.Body.String())
	}
	var delResp struct {
		Prefs struct {
			Columns []struct {
				Status string `json:"status"`
			} `json:"columns"`
		} `json:"prefs"`
	}
	_ = json.Unmarshal(rr3.Body.Bytes(), &delResp)
	if len(delResp.Prefs.Columns) != 5 {
		t.Fatalf("after delete want 5, got %d", len(delResp.Prefs.Columns))
	}
}
