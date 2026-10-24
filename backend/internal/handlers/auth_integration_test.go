package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/config"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/handlers"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/ratelimit"
)

func TestRegisterLoginMe(t *testing.T) {
	dir := t.TempDir()
	dsn := filepath.Join(dir, "t.db")
	sqlDB, err := db.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })

	cfg := config.Config{
		JWTSecret: "test-secret",
		JWTTTL:    time.Hour,
		AppEnv:    "test",
	}
	authH := &handlers.AuthHandler{DB: sqlDB, Cfg: cfg, Limiter: ratelimit.New(100, time.Minute)}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/auth/register", authH.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authH.Login)
	mux.Handle("GET /api/v1/auth/me", middleware.AuthJWT(middleware.JWTParser{Secret: cfg.JWTSecret})(http.HandlerFunc(authH.Me)))

	body, _ := json.Marshal(map[string]string{"email": "u@test.com", "password": "password1", "name": "测"})
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body)))
	if rr.Code != http.StatusCreated {
		t.Fatalf("register status=%d body=%s", rr.Code, rr.Body.String())
	}
	var reg map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &reg)
	token, _ := reg["token"].(string)
	if token == "" {
		t.Fatal("missing token")
	}

	rr2 := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	mux.ServeHTTP(rr2, req)
	if rr2.Code != http.StatusOK {
		t.Fatalf("me status=%d body=%s", rr2.Code, rr2.Body.String())
	}

	_ = os.Remove(dsn)
}
