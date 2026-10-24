package handlers

import (
	"net/http"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/issueboard"
)

type CycleHandler struct {
	Svc *issueboard.Service
}

func (h *CycleHandler) List(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.Svc.DB, w, r)
	if !ok {
		return
	}
	projectID := r.PathValue("projectID")
	if projectID == "" {
		projectID = r.URL.Query().Get("project_id")
	}
	list, err := h.Svc.ListCycles(ws, projectID)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"cycles": list})
}

func (h *CycleHandler) Create(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.Svc.DB, w, r)
	if !ok {
		return
	}
	var req struct {
		Name      string  `json:"name"`
		StartDate *string `json:"start_date"`
		EndDate   *string `json:"end_date"`
		Capacity  int     `json:"capacity"`
		Status    string  `json:"status"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	c, err := h.Svc.CreateCycle(ws, r.PathValue("projectID"), issueboard.CreateCycleInput{
		Name: req.Name, StartDate: req.StartDate, EndDate: req.EndDate,
		Capacity: req.Capacity, Status: req.Status,
	})
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusCreated, map[string]any{"cycle": c})
}

func (h *CycleHandler) Get(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.Svc.DB, w, r)
	if !ok {
		return
	}
	c, err := h.Svc.GetCycle(ws, r.PathValue("cycleID"))
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"cycle": c})
}

func (h *CycleHandler) Patch(w http.ResponseWriter, r *http.Request) {
	ws, mid, ok := requireWorkspaceMember(h.Svc.DB, w, r)
	if !ok {
		return
	}
	var req struct {
		Name      *string `json:"name"`
		StartDate *string `json:"start_date"`
		EndDate   *string `json:"end_date"`
		Capacity  *int    `json:"capacity"`
		Status    *string `json:"status"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	in := issueboard.UpdateCycleInput{
		Name: req.Name, StartDate: req.StartDate, EndDate: req.EndDate,
		Capacity: req.Capacity, Status: req.Status,
		ActorID: mid, ActorKind: actorKind(r),
	}
	// explicit null clear via empty pointer already omitted; empty string clears dates
	if req.StartDate != nil && *req.StartDate == "" {
		in.ClearStart = true
		in.StartDate = nil
	}
	if req.EndDate != nil && *req.EndDate == "" {
		in.ClearEnd = true
		in.EndDate = nil
	}
	c, err := h.Svc.UpdateCycle(ws, r.PathValue("cycleID"), in)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"cycle": c})
}
