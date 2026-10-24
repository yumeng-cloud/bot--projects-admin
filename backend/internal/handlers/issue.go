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
	list, err := h.Svc.ListIssues(ws, issueboard.ListIssuesFilter{
		ProjectID:  q.Get("project_id"),
		Status:     q.Get("status"),
		AssigneeID: q.Get("assignee_id"),
		LabelID:    q.Get("label_id"),
		CycleID:    q.Get("cycle_id"),
		Q:          q.Get("q"),
		LeafOnly:   leafOnly,
		Blocked:    blocked,
		Overdue:    overdue,
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
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	iss, err := h.Svc.CreateIssue(ws, issueboard.CreateIssueInput{
		ProjectID: req.ProjectID, Title: req.Title, Description: req.Description,
		Status: req.Status, Priority: req.Priority, Estimate: req.Estimate,
		AssigneeID: req.AssigneeID, CycleID: req.CycleID, ParentID: req.ParentID,
		StartAt: req.StartAt, DueAt: req.DueAt,
		BoardOrder: req.BoardOrder, LabelIDs: req.LabelIDs,
		ActorID: mid, ActorKind: actorKind(r),
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
	iss, err := h.Svc.GetIssue(ws, r.PathValue("issueID"))
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
	in := issueboard.UpdateIssueInput{ActorID: mid, ActorKind: actorKind(r)}
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
