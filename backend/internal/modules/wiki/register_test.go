package wiki_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/modules"
	_ "github.com/yumeng-tech/ai-collab-board/backend/internal/modules/wiki"
)

type chiMux struct{ r chi.Router }

func (c chiMux) Handle(pattern string, handler http.Handler) { c.r.Handle(pattern, handler) }
func (c chiMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	c.r.HandleFunc(pattern, handler)
}

func TestRegisterWikiSpacesNot501(t *testing.T) {
	if modules.RegisterWiki == nil {
		t.Fatal("modules.RegisterWiki is nil; init did not assign")
	}
	sqlDB, err := db.Open("sqlite", filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	api := chi.NewRouter()
	deps := modules.Deps{
		DB:      sqlDB,
		AuthMW:  middleware.AuthJWT(middleware.JWTParser{Secret: "test"}),
		WSMemMW: middleware.RequireWorkspaceMember(sqlDB),
	}
	modules.RegisterWiki(chiMux{r: api}, deps)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/workspaces/ws1/wiki/spaces", nil)
	api.ServeHTTP(rr, req)
	if rr.Code == http.StatusNotImplemented {
		t.Fatalf("got 501 stub; body=%s", rr.Body.String())
	}
	// unauthenticated → 401 (route is live)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d body=%s", rr.Code, rr.Body.String())
	}
}
