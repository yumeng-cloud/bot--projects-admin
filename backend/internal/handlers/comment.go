package handlers

import (
	"net/http"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/issueboard"
)

type CommentHandler struct {
	Svc *issueboard.Service
}

func (h *CommentHandler) List(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.Svc.DB, w, r)
	if !ok {
		return
	}
	list, err := h.Svc.ListComments(ws, r.PathValue("issueID"))
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"comments": list})
}

func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	ws, mid, ok := requireWorkspaceMember(h.Svc.DB, w, r)
	if !ok {
		return
	}
	var req struct {
		Body string `json:"body"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	c, err := h.Svc.CreateComment(ws, r.PathValue("issueID"), issueboard.CreateCommentInput{
		Body: req.Body, AuthorID: mid, ActorKind: actorKind(r),
	})
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusCreated, map[string]any{"comment": c})
}

func (h *CommentHandler) Patch(w http.ResponseWriter, r *http.Request) {
	ws, mid, ok := requireWorkspaceMember(h.Svc.DB, w, r)
	if !ok {
		return
	}
	var req struct {
		Body      string `json:"body"`
		UpdatedAt string `json:"updated_at"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	c, err := h.Svc.UpdateComment(ws, r.PathValue("issueID"), r.PathValue("commentID"), issueboard.UpdateCommentInput{
		Body: req.Body, ExpectedUpdatedAt: req.UpdatedAt, ActorID: mid, ActorKind: actorKind(r),
	})
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"comment": c})
}

func (h *CommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ws, mid, ok := requireWorkspaceMember(h.Svc.DB, w, r)
	if !ok {
		return
	}
	if err := h.Svc.DeleteComment(ws, r.PathValue("issueID"), r.PathValue("commentID"), mid, actorKind(r)); err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"ok": true})
}
