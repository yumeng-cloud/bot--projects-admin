package handlers

import (
	"database/sql"
	"net/http"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/timetrack"
)

type TimeHandler struct {
	DB  *sql.DB
	Svc *timetrack.Service
}

func NewTimeHandler(db *sql.DB) *TimeHandler {
	return &TimeHandler{DB: db, Svc: timetrack.New(db)}
}

// MountTime 挂载工时与计费路由。wrap 为鉴权中间件（JWT / API Key）。
func MountTime(mux *http.ServeMux, db *sql.DB, wrap func(http.Handler) http.Handler) {
	h := NewTimeHandler(db)
	read := chainAuth(db, wrap, false)
	write := chainAuth(db, wrap, true)
	reg := func(chain func(http.Handler) http.Handler, pattern string, fn http.HandlerFunc) {
		mux.Handle(pattern, chain(fn))
	}

	reg(read, "GET /api/v1/workspaces/{workspaceID}/time-entries", h.ListEntries)
	reg(write, "POST /api/v1/workspaces/{workspaceID}/time-entries", h.CreateEntry)
	reg(read, "GET /api/v1/workspaces/{workspaceID}/time-entries/{entryID}", h.GetEntry)
	reg(write, "PATCH /api/v1/workspaces/{workspaceID}/time-entries/{entryID}", h.UpdateEntry)
	reg(write, "DELETE /api/v1/workspaces/{workspaceID}/time-entries/{entryID}", h.DeleteEntry)

	reg(read, "GET /api/v1/workspaces/{workspaceID}/issues/{issueID}/time-entries", h.ListIssueEntries)
	reg(write, "POST /api/v1/workspaces/{workspaceID}/issues/{issueID}/time-entries", h.CreateIssueEntry)

	reg(read, "GET /api/v1/workspaces/{workspaceID}/billing-rates", h.ListRates)
	reg(write, "POST /api/v1/workspaces/{workspaceID}/billing-rates", h.UpsertRate)
	reg(write, "PUT /api/v1/workspaces/{workspaceID}/billing-rates", h.UpsertRate)
	reg(write, "DELETE /api/v1/workspaces/{workspaceID}/billing-rates/{rateID}", h.DeleteRate)

	reg(read, "GET /api/v1/workspaces/{workspaceID}/billing/summary", h.Summary)
	reg(read, "GET /api/v1/workspaces/{workspaceID}/billing/export.csv", h.ExportCSV)
	reg(read, "GET /api/v1/workspaces/{workspaceID}/time-entries/export.csv", h.ExportCSV)
	reg(read, "GET /api/v1/workspaces/{workspaceID}/time-entries/summary", h.Summary)
}

func (h *TimeHandler) ListEntries(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	list, err := h.Svc.ListEntries(ws, timetrack.ListFilter{
		IssueID: q.Get("issue_id"), MemberID: q.Get("member_id"),
		From: q.Get("from"), To: q.Get("to"),
	})
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"time_entries": list})
}

func (h *TimeHandler) ListIssueEntries(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	list, err := h.Svc.ListEntries(ws, timetrack.ListFilter{
		IssueID: r.PathValue("issueID"),
		From:    r.URL.Query().Get("from"),
		To:      r.URL.Query().Get("to"),
	})
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"time_entries": list})
}

func (h *TimeHandler) CreateEntry(w http.ResponseWriter, r *http.Request) {
	ws, mid, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	var req struct {
		IssueID  string   `json:"issue_id"`
		MemberID string   `json:"member_id"`
		Hours    float64  `json:"hours"`
		Note     string   `json:"note"`
		Billable *bool    `json:"billable"`
		Rate     *float64 `json:"rate"`
		WorkedAt string   `json:"worked_at"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	if req.MemberID == "" {
		req.MemberID = mid
	}
	e, err := h.Svc.CreateEntry(ws, timetrack.CreateEntryInput{
		IssueID: req.IssueID, MemberID: req.MemberID, Hours: req.Hours,
		Note: req.Note, Billable: req.Billable, Rate: req.Rate, WorkedAt: req.WorkedAt,
	})
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusCreated, map[string]any{"time_entry": e})
}

func (h *TimeHandler) CreateIssueEntry(w http.ResponseWriter, r *http.Request) {
	ws, mid, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	var req struct {
		MemberID string   `json:"member_id"`
		Hours    float64  `json:"hours"`
		Note     string   `json:"note"`
		Billable *bool    `json:"billable"`
		Rate     *float64 `json:"rate"`
		WorkedAt string   `json:"worked_at"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	if req.MemberID == "" {
		req.MemberID = mid
	}
	e, err := h.Svc.CreateEntry(ws, timetrack.CreateEntryInput{
		IssueID: r.PathValue("issueID"), MemberID: req.MemberID, Hours: req.Hours,
		Note: req.Note, Billable: req.Billable, Rate: req.Rate, WorkedAt: req.WorkedAt,
	})
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusCreated, map[string]any{"time_entry": e})
}

func (h *TimeHandler) GetEntry(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	e, err := h.Svc.GetEntry(ws, r.PathValue("entryID"))
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"time_entry": e})
}

func (h *TimeHandler) UpdateEntry(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	var req struct {
		Hours    *float64 `json:"hours"`
		Note     *string  `json:"note"`
		Billable *bool    `json:"billable"`
		Rate     *float64 `json:"rate"`
		WorkedAt *string  `json:"worked_at"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	e, err := h.Svc.UpdateEntry(ws, r.PathValue("entryID"), timetrack.UpdateEntryInput{
		Hours: req.Hours, Note: req.Note, Billable: req.Billable, Rate: req.Rate, WorkedAt: req.WorkedAt,
	})
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"time_entry": e})
}

func (h *TimeHandler) DeleteEntry(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	if err := h.Svc.DeleteEntry(ws, r.PathValue("entryID")); err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "已删除"})
}

func (h *TimeHandler) ListRates(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	list, err := h.Svc.ListRates(ws)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"billing_rates": list})
}

func (h *TimeHandler) UpsertRate(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	var req struct {
		MemberID   *string `json:"member_id"`
		HourlyRate float64 `json:"hourly_rate"`
		Currency   string  `json:"currency"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	rate, err := h.Svc.UpsertRate(ws, timetrack.UpsertRateInput{
		MemberID: req.MemberID, HourlyRate: req.HourlyRate, Currency: req.Currency,
	})
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"billing_rate": rate})
}

func (h *TimeHandler) DeleteRate(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	if err := h.Svc.DeleteRate(ws, r.PathValue("rateID")); err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "已删除"})
}

func (h *TimeHandler) Summary(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	sum, err := h.Svc.Summary(ws, q.Get("from"), q.Get("to"), q.Get("group_by"))
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"summary": sum})
}

func (h *TimeHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="time-entries.csv"`)
	if err := h.Svc.ExportCSV(w, ws, q.Get("from"), q.Get("to")); err != nil {
		// headers may already be written; best effort
		middleware.HandleError(w, err)
	}
}
