package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/issueboard"
)

type IssueHandler struct {
	Svc *issueboard.Service
}

func (h *IssueHandler) List(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.Svc.DB, w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	leafOnly := q.Get("leaf_only") == "1" || strings.EqualFold(q.Get("leaf_only"), "true")
	blocked := q.Get("blocked") == "1" || strings.EqualFold(q.Get("blocked"), "true")
	overdue := q.Get("overdue") == "1" || strings.EqualFold(q.Get("overdue"), "true")
	unassigned := q.Get("unassigned") == "1" || strings.EqualFold(q.Get("unassigned"), "true")
	includeArchived, archivedOnly := parseIssueArchiveQuery(q.Get)
	list, err := h.Svc.ListIssues(ws, issueboard.ListIssuesFilter{
		ProjectID:       q.Get("project_id"),
		TeamID:          q.Get("team_id"),
		UnassignedOnly:  unassigned,
		Status:          q.Get("status"),
		AssigneeID:      q.Get("assignee_id"),
		LabelID:         q.Get("label_id"),
		CycleID:         q.Get("cycle_id"),
		Q:               q.Get("q"),
		LeafOnly:        leafOnly,
		Blocked:         blocked,
		Overdue:         overdue,
		IncludeArchived: includeArchived,
		ArchivedOnly:    archivedOnly,
	})
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"issues": list})
}

func (h *IssueHandler) Create(w http.ResponseWriter, r *http.Request) {
	ws, mid, ok := requireWorkspaceMember(h.Svc.DB, w, r)
	if !ok {
		return
	}
	var req struct {
		ProjectID   string   `json:"project_id"`
		TeamID      string   `json:"team_id"`
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Status      string   `json:"status"`
		Priority    string   `json:"priority"`
		Estimate    *int     `json:"estimate"`
		AssigneeID  *string  `json:"assignee_id"`
		CycleID     *string  `json:"cycle_id"`
		ParentID    *string  `json:"parent_id"`
		StartAt     *string  `json:"start_at"`
		DueAt       *string  `json:"due_at"`
		BoardOrder  *float64 `json:"board_order"`
		LabelIDs    []string `json:"label_ids"`
		SkipNotify  bool     `json:"skip_notify"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	// body 优先；亦可 query ?team_id=
	teamID := strings.TrimSpace(req.TeamID)
	if teamID == "" {
		teamID = strings.TrimSpace(r.URL.Query().Get("team_id"))
	}
	iss, err := h.Svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: req.ProjectID, TeamID: teamID, Title: req.Title, Description: req.Description,
		Status: req.Status, Priority: req.Priority, Estimate: req.Estimate,
		AssigneeID: req.AssigneeID, CycleID: req.CycleID, ParentID: req.ParentID,
		StartAt: req.StartAt, DueAt: req.DueAt,
		BoardOrder: req.BoardOrder, LabelIDs: req.LabelIDs,
		ActorID: mid, ActorKind: actorKind(r),
		SkipNotify: req.SkipNotify,
	})
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusCreated, map[string]any{"issue": iss})
}

func (h *IssueHandler) Get(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.Svc.DB, w, r)
	if !ok {
		return
	}
	teamID := strings.TrimSpace(r.URL.Query().Get("team_id"))
	iss, err := h.Svc.GetIssueInTeam(ws, r.PathValue("issueID"), teamID)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"issue": iss})
}

func (h *IssueHandler) Patch(w http.ResponseWriter, r *http.Request) {
	ws, mid, ok := requireWorkspaceMember(h.Svc.DB, w, r)
	if !ok {
		return
	}
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	in := issueboard.UpdateIssueInput{
		ActorID: mid, ActorKind: actorKind(r),
		RequireTeamID: strings.TrimSpace(r.URL.Query().Get("team_id")),
	}
	if v, ok := raw["version"]; !ok {
		middleware.HandleError(w, apperr.BadRequest("更新 Issue 必须提供 version（乐观锁）"))
		return
	} else if err := json.Unmarshal(v, &in.Version); err != nil {
		middleware.HandleError(w, apperr.BadRequest("version 无效"))
		return
	}
	decodePtr := func(key string, dst any) {
		if v, ok := raw[key]; ok {
			_ = json.Unmarshal(v, dst)
		}
	}
	decodePtr("title", &in.Title)
	decodePtr("description", &in.Description)
	decodePtr("status", &in.Status)
	decodePtr("priority", &in.Priority)
	decodePtr("board_order", &in.BoardOrder)
	if v, ok := raw["estimate"]; ok {
		if string(v) == "null" {
			in.ClearEstimate = true
		} else {
			decodePtr("estimate", &in.Estimate)
		}
	}
	if v, ok := raw["assignee_id"]; ok {
		if string(v) == "null" || string(v) == `""` {
			in.ClearAssignee = true
		} else {
			decodePtr("assignee_id", &in.AssigneeID)
		}
	}
	if v, ok := raw["cycle_id"]; ok {
		if string(v) == "null" || string(v) == `""` {
			in.ClearCycle = true
		} else {
			decodePtr("cycle_id", &in.CycleID)
		}
	}
	if v, ok := raw["parent_id"]; ok {
		if string(v) == "null" || string(v) == `""` {
			in.ClearParent = true
		} else {
			decodePtr("parent_id", &in.ParentID)
		}
	}
	if v, ok := raw["start_at"]; ok {
		if string(v) == "null" || string(v) == `""` {
			in.ClearStartAt = true
		} else {
			decodePtr("start_at", &in.StartAt)
		}
	}
	if v, ok := raw["due_at"]; ok {
		if string(v) == "null" || string(v) == `""` {
			in.ClearDueAt = true
		} else {
			decodePtr("due_at", &in.DueAt)
		}
	}
	if v, ok := raw["force_assignee"]; ok {
		_ = json.Unmarshal(v, &in.ForceAssignee)
	}
	if v, ok := raw["team_id"]; ok {
		decodePtr("team_id", &in.TeamID)
		_ = v
	}
	if v, ok := raw["label_ids"]; ok {
		var ids []string
		if err := json.Unmarshal(v, &ids); err == nil {
			in.LabelIDs = &ids
		}
	}
	iss, err := h.Svc.UpdateIssue(ws, r.PathValue("issueID"), in)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"issue": iss})
}

func (h *IssueHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ws, mid, ok := requireWorkspaceMember(h.Svc.DB, w, r)
	if !ok {
		return
	}
	if err := h.Svc.DeleteIssue(ws, r.PathValue("issueID"), mid, actorKind(r)); err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"ok": true})
}

func actorKind(r *http.Request) string {
	k := middleware.ActorKind(r.Context())
	if k == "" {
		return "human"
	}
	return k
}

// parseIssueArchiveQuery maps list/search/board query params (ISSUE_ARCHIVE).
// include_archived=1|true → include archived; archived=only → archived only (wins).
// parseIssueArchiveQuery: include_archived=1|true; archived=only (wins). NOT archived_only.
func parseIssueArchiveQuery(get func(string) string) (includeArchived, archivedOnly bool) {
	includeArchived = get("include_archived") == "1" || strings.EqualFold(get("include_archived"), "true")
	archivedOnly = strings.EqualFold(strings.TrimSpace(get("archived")), "only")
	return includeArchived, archivedOnly
}

func (h *IssueHandler) Archive(w http.ResponseWriter, r *http.Request) {
	ws, mid, ok := requireWorkspaceMember(h.Svc.DB, w, r)
	if !ok {
		return
	}
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	results, err := h.Svc.ArchiveIssues(ws, issueboard.ArchiveIssuesInput{
		IDs: req.IDs, TeamID: strings.TrimSpace(r.URL.Query().Get("team_id")),
		ActorID: mid, ActorKind: actorKind(r),
	})
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	// Per-id results (200). Non-done / wrong-team carry Chinese error on that id.
	// Single-id all-non-done → 409 for clear client UX (debt A3 中文 4xx).
	if len(results) == 1 && !results[0].OK && results[0].Error == "仅完成(done)状态的 Issue 可归档" {
		middleware.HandleError(w, apperr.Conflict("仅完成(done)状态的 Issue 可归档"))
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"results": results})
}

func (h *IssueHandler) Unarchive(w http.ResponseWriter, r *http.Request) {
	ws, mid, ok := requireWorkspaceMember(h.Svc.DB, w, r)
	if !ok {
		return
	}
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	results, err := h.Svc.UnarchiveIssues(ws, issueboard.UnarchiveIssuesInput{
		IDs: req.IDs, TeamID: strings.TrimSpace(r.URL.Query().Get("team_id")),
		ActorID: mid, ActorKind: actorKind(r),
	})
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"results": results})
}
