package wiki

import (
	"net/http"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/handlers"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/modules"
)

func init() {
	modules.RegisterWiki = Register
}

// Register mounts Wiki routes under /api/v1 (relative paths).
func Register(mux modules.Mux, deps modules.Deps) {
	h := handlers.NewWikiHandler(deps.DB)
	read := wrap(deps, false)
	write := wrap(deps, true)

	mount(mux, "/workspaces/{workspaceID}/wiki/spaces", map[string]http.Handler{
		http.MethodGet:  read(http.HandlerFunc(h.ListSpaces)),
		http.MethodPost: write(http.HandlerFunc(h.CreateSpace)),
	})
	mount(mux, "/workspaces/{workspaceID}/wiki/spaces/{spaceID}", map[string]http.Handler{
		http.MethodGet:    read(http.HandlerFunc(h.GetSpace)),
		http.MethodPatch:  write(http.HandlerFunc(h.UpdateSpace)),
		http.MethodDelete: write(http.HandlerFunc(h.DeleteSpace)),
	})
	mount(mux, "/workspaces/{workspaceID}/wiki/spaces/{spaceID}/pages", map[string]http.Handler{
		http.MethodGet:  read(http.HandlerFunc(h.ListPages)),
		http.MethodPost: write(http.HandlerFunc(h.CreatePage)),
	})
	mount(mux, "/workspaces/{workspaceID}/wiki/pages/{pageID}", map[string]http.Handler{
		http.MethodGet:    read(http.HandlerFunc(h.GetPage)),
		http.MethodPatch:  write(http.HandlerFunc(h.UpdatePage)),
		http.MethodDelete: write(http.HandlerFunc(h.DeletePage)),
	})
	mount(mux, "/workspaces/{workspaceID}/wiki/search", map[string]http.Handler{
		http.MethodGet: read(http.HandlerFunc(h.Search)),
	})
	mount(mux, "/workspaces/{workspaceID}/wiki/tree", map[string]http.Handler{
		http.MethodGet: read(http.HandlerFunc(h.Tree)),
	})
}

func wrap(deps modules.Deps, write bool) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		if write {
			h = middleware.Idempotency(deps.DB)(h)
		}
		if deps.WSMemMW != nil {
			h = deps.WSMemMW(h)
		}
		if deps.AuthMW != nil {
			h = deps.AuthMW(h)
		}
		return h
	}
}

func mount(mux modules.Mux, path string, byMethod map[string]http.Handler) {
	mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h, ok := byMethod[r.Method]
		if !ok {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.ServeHTTP(w, r)
	}))
}
