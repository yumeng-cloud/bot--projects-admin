package timebilling

import (
	"net/http"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/handlers"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/modules"
)

func init() {
	modules.RegisterTime = Register
}

// Register mounts time-entry and billing routes under /api/v1 (relative paths).
func Register(mux modules.Mux, deps modules.Deps) {
	h := handlers.NewTimeHandler(deps.DB)
	read := wrap(deps, false)
	write := wrap(deps, true)

	mount(mux, "/workspaces/{workspaceID}/time-entries", map[string]http.Handler{
		http.MethodGet:  read(http.HandlerFunc(h.ListEntries)),
		http.MethodPost: write(http.HandlerFunc(h.CreateEntry)),
	})
	// Static suffixes before {entryID} so chi does not treat them as IDs.
	mount(mux, "/workspaces/{workspaceID}/time-entries/export.csv", map[string]http.Handler{
		http.MethodGet: read(http.HandlerFunc(h.ExportCSV)),
	})
	mount(mux, "/workspaces/{workspaceID}/time-entries/summary", map[string]http.Handler{
		http.MethodGet: read(http.HandlerFunc(h.Summary)),
	})
	mount(mux, "/workspaces/{workspaceID}/time-entries/{entryID}", map[string]http.Handler{
		http.MethodGet:    read(http.HandlerFunc(h.GetEntry)),
		http.MethodPatch:  write(http.HandlerFunc(h.UpdateEntry)),
		http.MethodDelete: write(http.HandlerFunc(h.DeleteEntry)),
	})
	mount(mux, "/workspaces/{workspaceID}/issues/{issueID}/time-entries", map[string]http.Handler{
		http.MethodGet:  read(http.HandlerFunc(h.ListIssueEntries)),
		http.MethodPost: write(http.HandlerFunc(h.CreateIssueEntry)),
	})
	mount(mux, "/workspaces/{workspaceID}/billing-rates", map[string]http.Handler{
		http.MethodGet:  read(http.HandlerFunc(h.ListRates)),
		http.MethodPost: write(http.HandlerFunc(h.UpsertRate)),
		http.MethodPut:  write(http.HandlerFunc(h.UpsertRate)),
	})
	mount(mux, "/workspaces/{workspaceID}/billing-rates/{rateID}", map[string]http.Handler{
		http.MethodDelete: write(http.HandlerFunc(h.DeleteRate)),
	})
	mount(mux, "/workspaces/{workspaceID}/billing/summary", map[string]http.Handler{
		http.MethodGet: read(http.HandlerFunc(h.Summary)),
	})
	mount(mux, "/workspaces/{workspaceID}/billing/export.csv", map[string]http.Handler{
		http.MethodGet: read(http.HandlerFunc(h.ExportCSV)),
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
