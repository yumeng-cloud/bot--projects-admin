package server_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/config"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/server"
)

func TestJoinMountedAtAPIRoot(t *testing.T) {
	dir := t.TempDir()
	sqlDB, err := db.Open("sqlite", filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	cfg := config.Config{
		JWTSecret: "t", JWTTTL: time.Hour, AppEnv: "test",
		PublicAPIBase: "https://bot.api.sn.cn",
		CORSOrigins:   []string{"http://localhost:5173"},
		LoginRateLimit: 10, LoginRateWindow: time.Minute,
	}
	h := server.NewRouter(sqlDB, cfg)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/join?code=ymt_routerok&role=go", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("%d %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "ymbot team join --code ymt_routerok --role go") {
		t.Fatalf("body missing cli: %s", rr.Body.String())
	}
}


func TestInstallMountedAtAPIRoot(t *testing.T) {
	dir := t.TempDir()
	sqlDB, err := db.Open("sqlite", filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	cfg := config.Config{
		JWTSecret: "t", JWTTTL: time.Hour, AppEnv: "test",
		PublicAPIBase: "https://bot.api.sn.cn",
		CORSOrigins:   []string{"http://localhost:5173"},
		LoginRateLimit: 10, LoginRateWindow: time.Minute,
	}
	h := server.NewRouter(sqlDB, cfg)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/install.sh", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("%d %s", rr.Code, rr.Body.String())
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/plain") {
		t.Fatalf("content-type %s", ct)
	}
	body := rr.Body.String()
	if strings.TrimSpace(body) == "" || !strings.Contains(body, "ymbot") {
		t.Fatalf("install script body bad: %q", body)
	}
}
