package handlers

import (
	"net/http"
	"strings"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/issueboard"
)

// SearchHandler workspace-wide search (issues + projects).
type SearchHandler struct {
	Svc *issueboard.Service
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.Svc.DB, w, r)
	if !ok {
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	qParams := r.URL.Query()
	issues, err := h.Svc.ListIssues(ws, issueboard.ListIssuesFilter{
		Q: q, TeamID: qParams.Get("team_id"),
		UnassignedOnly: qParams.Get("unassigned") == "1" || strings.EqualFold(qParams.Get("unassigned"), "true"),
	})
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	projects, err := h.Svc.ListProjects(ws)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	if q != "" {
		ql := strings.ToLower(q)
		filtered := projects[:0]
		for _, p := range projects {
			if strings.Contains(strings.ToLower(p.Name), ql) || strings.Contains(strings.ToLower(p.KeyPrefix), ql) {
				filtered = append(filtered, p)
			}
		}
		projects = filtered
	}
	middleware.JSON(w, http.StatusOK, map[string]any{
		"q":        q,
		"issues":   issues,
		"projects": projects,
	})
}
