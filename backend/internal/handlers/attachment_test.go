package handlers_test

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/config"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/seed"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/server"
)

func attachmentRouter(t *testing.T) (http.Handler, config.Config, string) {
	t.Helper()
	dir := t.TempDir()
	dsn := filepath.Join(dir, "t.db")
	sqlDB, err := db.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	storeDir := filepath.Join(dir, "attachments")
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
		AttachmentsDir:  storeDir,
		CORSOrigins:     []string{"*"},
	}
	if err := seed.Run(sqlDB, cfg); err != nil {
		t.Fatal(err)
	}
	h := server.NewRouter(sqlDB, cfg)
	return h, cfg, storeDir
}

func loginAndWS(t *testing.T, h http.Handler) (token, wsID, projectID string) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"email": "admin@yumeng.tech", "password": "Admin123!"})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body)))
	if rr.Code != http.StatusOK {
		t.Fatalf("login %d %s", rr.Code, rr.Body.String())
	}
	var res map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &res)
	token, _ = res["token"].(string)

	rr2 := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	h.ServeHTTP(rr2, req)
	var wsRes struct {
		Workspaces []struct {
			ID string `json:"id"`
		} `json:"workspaces"`
	}
	_ = json.Unmarshal(rr2.Body.Bytes(), &wsRes)
	if len(wsRes.Workspaces) == 0 {
		t.Fatal("no ws")
	}
	wsID = wsRes.Workspaces[0].ID

	rr3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+wsID+"/projects", nil)
	req3.Header.Set("Authorization", "Bearer "+token)
	h.ServeHTTP(rr3, req3)
	var projRes struct {
		Projects []struct {
			ID string `json:"id"`
		} `json:"projects"`
	}
	_ = json.Unmarshal(rr3.Body.Bytes(), &projRes)
	if len(projRes.Projects) == 0 {
		t.Fatal("no project")
	}
	return token, wsID, projRes.Projects[0].ID
}

func createIssue(t *testing.T, h http.Handler, token, wsID, projectID string) string {
	t.Helper()
	tbody, _ := json.Marshal(map[string]any{"name": "附件队", "slug": "attach-team"})
	trr := httptest.NewRecorder()
	treq := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+wsID+"/teams", bytes.NewReader(tbody))
	treq.Header.Set("Authorization", "Bearer "+token)
	treq.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(trr, treq)
	if trr.Code != http.StatusCreated && trr.Code != http.StatusOK {
		t.Fatalf("create team %d %s", trr.Code, trr.Body.String())
	}
	var tres map[string]any
	_ = json.Unmarshal(trr.Body.Bytes(), &tres)
	team, _ := tres["team"].(map[string]any)
	teamID, _ := team["id"].(string)
	if teamID == "" {
		teamID, _ = tres["id"].(string)
	}
	if teamID == "" {
		t.Fatalf("no team id in %s", trr.Body.String())
	}
	body, _ := json.Marshal(map[string]any{"project_id": projectID, "team_id": teamID, "title": "附件测"})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+wsID+"/issues", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated && rr.Code != http.StatusOK {
		t.Fatalf("create issue %d %s", rr.Code, rr.Body.String())
	}
	var res map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &res)
	iss, _ := res["issue"].(map[string]any)
	if iss == nil {
		// some handlers return issue at top
		if id, _ := res["id"].(string); id != "" {
			return id
		}
		t.Fatalf("no issue in %s", rr.Body.String())
	}
	id, _ := iss["id"].(string)
	if id == "" {
		t.Fatal("empty issue id")
	}
	return id
}

func multipartPNG(t *testing.T, filename string, data []byte) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write(data); err != nil {
		t.Fatal(err)
	}
	_ = w.Close()
	return &buf, w.FormDataContentType()
}

func tinyPNGBytes() []byte {
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xde, 0x00, 0x00, 0x00,
		0x0c, 0x49, 0x44, 0x41, 0x54, 0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0x00,
		0x00, 0x00, 0x03, 0x00, 0x01, 0x00, 0x05, 0xfe, 0xd4, 0xef, 0x00, 0x00,
		0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
	}
}

func TestAttachmentUploadDownloadAuth(t *testing.T) {
	h, _, storeDir := attachmentRouter(t)
	token, wsID, projectID := loginAndWS(t, h)
	issueID := createIssue(t, h, token, wsID, projectID)

	png := tinyPNGBytes()
	body, ctype := multipartPNG(t, "a.png", png)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+wsID+"/issues/"+issueID+"/attachments", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", ctype)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("upload %d %s", rr.Code, rr.Body.String())
	}
	var up struct {
		Attachment map[string]any `json:"attachment"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &up)
	attID, _ := up.Attachment["id"].(string)
	urlPath, _ := up.Attachment["url"].(string)
	if attID == "" || !strings.Contains(urlPath, attID+"/content") {
		t.Fatalf("bad attachment %+v", up.Attachment)
	}

	// meta
	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+wsID+"/attachments/"+attID, nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("meta %d %s", rr2.Code, rr2.Body.String())
	}

	// content
	rr3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodGet, urlPath, nil)
	req3.Header.Set("Authorization", "Bearer "+token)
	h.ServeHTTP(rr3, req3)
	if rr3.Code != http.StatusOK {
		t.Fatalf("content %d %s", rr3.Code, rr3.Body.String())
	}
	if ct := rr3.Header().Get("Content-Type"); !strings.Contains(ct, "image/png") {
		t.Fatalf("ct=%s", ct)
	}
	if disp := rr3.Header().Get("Content-Disposition"); !strings.Contains(disp, "inline") {
		t.Fatalf("disp=%s", disp)
	}
	got, _ := io.ReadAll(rr3.Body)
	if !bytes.Equal(got, png) {
		t.Fatal("body mismatch")
	}

	// 401 without auth
	rr4 := httptest.NewRecorder()
	h.ServeHTTP(rr4, httptest.NewRequest(http.MethodGet, urlPath, nil))
	if rr4.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rr4.Code)
	}

	// 403 non-member
	ob, _ := json.Marshal(map[string]string{"email": "outsider@test.com", "password": "password1", "name": "外"})
	rr5 := httptest.NewRecorder()
	h.ServeHTTP(rr5, httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(ob)))
	var oreg map[string]any
	_ = json.Unmarshal(rr5.Body.Bytes(), &oreg)
	otok, _ := oreg["token"].(string)
	rr6 := httptest.NewRecorder()
	req6 := httptest.NewRequest(http.MethodGet, urlPath, nil)
	req6.Header.Set("Authorization", "Bearer "+otok)
	h.ServeHTTP(rr6, req6)
	if rr6.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d %s", rr6.Code, rr6.Body.String())
	}

	// exe reject
	exeBody, exeCT := multipartPNG(t, "x.exe", []byte{0x4d, 0x5a})
	rr7 := httptest.NewRecorder()
	req7 := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+wsID+"/issues/"+issueID+"/attachments", exeBody)
	req7.Header.Set("Authorization", "Bearer "+token)
	req7.Header.Set("Content-Type", exeCT)
	h.ServeHTTP(rr7, req7)
	if rr7.Code != http.StatusBadRequest {
		t.Fatalf("exe want 400, got %d %s", rr7.Code, rr7.Body.String())
	}
	var errBody map[string]string
	_ = json.Unmarshal(rr7.Body.Bytes(), &errBody)
	if errBody["message"] != "不支持的文件类型" {
		t.Fatalf("msg=%v", errBody)
	}

	// store should only have the png
	n := 0
	_ = filepath.Walk(storeDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && info != nil && !info.IsDir() {
			n++
		}
		return nil
	})
	if n != 1 {
		t.Fatalf("want 1 file on disk, got %d", n)
	}
}

func TestAttachmentCommentUpload(t *testing.T) {
	h, _, _ := attachmentRouter(t)
	token, wsID, projectID := loginAndWS(t, h)
	issueID := createIssue(t, h, token, wsID, projectID)

	cbody, _ := json.Marshal(map[string]string{"body": "带附件"})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+wsID+"/issues/"+issueID+"/comments", bytes.NewReader(cbody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated && rr.Code != http.StatusOK {
		t.Fatalf("comment %d %s", rr.Code, rr.Body.String())
	}
	var cres map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &cres)
	cmt, _ := cres["comment"].(map[string]any)
	commentID, _ := cmt["id"].(string)
	if commentID == "" {
		t.Fatalf("no comment id: %s", rr.Body.String())
	}

	txt := []byte("hello md")
	body, ctype := multipartPNG(t, "n.md", txt)
	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost,
		"/api/v1/workspaces/"+wsID+"/issues/"+issueID+"/comments/"+commentID+"/attachments", body)
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Content-Type", ctype)
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusCreated {
		t.Fatalf("upload comment att %d %s", rr2.Code, rr2.Body.String())
	}
}
