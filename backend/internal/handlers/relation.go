package handlers

import (
	"net/http"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/issueboard"
)

type RelationHandler struct {
	Svc *issueboard.Service
}

func (h *RelationHandler) List(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.Svc.DB, w, r)
	if !ok {
		return
	}
	list, err := h.Svc.ListRelations(ws, r.PathValue("issueID"))
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"relations": list})
}

func (h *RelationHandler) Create(w http.ResponseWriter, r *http.Request) {
	ws, mid, ok := requireWorkspaceMember(h.Svc.DB, w, r)
	if !ok {
		return
	}
	var req struct {
		ToIssueID    string `json:"to_issue_id"`
		RelationType string `json:"relation_type"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	rel, err := h.Svc.CreateRelation(ws, r.PathValue("issueID"), issueboard.CreateRelationInput{
		ToIssueID: req.ToIssueID, RelationType: req.RelationType,
		ActorID: mid, ActorKind: actorKind(r),
	})
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusCreated, map[string]any{"relation": rel})
}

func (h *RelationHandler) DeleteOnIssue(w http.ResponseWriter, r *http.Request) {
	ws, mid, ok := requireWorkspaceMember(h.Svc.DB, w, r)
	if !ok {
		return
	}
	if err := h.Svc.DeleteRelationOnIssue(ws, r.PathValue("issueID"), r.PathValue("relationID"), mid, actorKind(r)); err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *RelationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ws, mid, ok := requireWorkspaceMember(h.Svc.DB, w, r)
	if !ok {
		return
	}
	if err := h.Svc.DeleteRelation(ws, r.PathValue("relationID"), mid, actorKind(r)); err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"ok": true})
}

// DeleteCollection supports DELETE .../issues/{issueID}/relations with ?relation_id= or JSON {"relation_id"|"id"}.
func (h *RelationHandler) DeleteCollection(w http.ResponseWriter, r *http.Request) {
	ws, mid, ok := requireWorkspaceMember(h.Svc.DB, w, r)
	if !ok {
		return
	}
	rid := r.URL.Query().Get("relation_id")
	if rid == "" {
		rid = r.URL.Query().Get("id")
	}
	if rid == "" {
		var req struct {
			RelationID string `json:"relation_id"`
			ID         string `json:"id"`
		}
		_ = middleware.DecodeJSON(r, &req)
		rid = req.RelationID
		if rid == "" {
			rid = req.ID
		}
	}
	if rid == "" {
		middleware.HandleError(w, apperr.BadRequest("请提供 relation_id"))
		return
	}
	if err := h.Svc.DeleteRelationOnIssue(ws, r.PathValue("issueID"), rid, mid, actorKind(r)); err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"ok": true})
}
