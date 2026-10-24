package handlers

import (
	"net/http"
	"strconv"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/issueboard"
)

type ActivityHandler struct {
	Svc *issueboard.Service
}

func (h *ActivityHandler) ListByIssue(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.Svc.DB, w, r)
	if !ok {
		return
	}
	list, err := h.Svc.ListIssueActivities(ws, r.PathValue("issueID"))
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"activities": list})
}

func (h *ActivityHandler) List(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.Svc.DB, w, r)
	if !ok {
		return
	}
	limit := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	list, err := h.Svc.ListActivities(ws, limit)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"activities": list})
}
