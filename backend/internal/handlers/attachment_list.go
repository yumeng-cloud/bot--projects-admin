package handlers

import (
	"net/http"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/attachments"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
)

// ListByIssue GET …/issues/{issueID}/attachments
func (h *AttachmentHandler) ListByIssue(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	list, err := h.Svc.ListByIssue(r.Context(), ws, r.PathValue("issueID"))
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	if list == nil {
		list = []attachments.Attachment{}
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"attachments": list})
}

// ListByComment GET …/issues/{issueID}/comments/{commentID}/attachments
func (h *AttachmentHandler) ListByComment(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	list, err := h.Svc.ListByComment(r.Context(), ws, r.PathValue("issueID"), r.PathValue("commentID"))
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	if list == nil {
		list = []attachments.Attachment{}
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"attachments": list})
}

// Delete DELETE …/attachments/{attachmentID}
// 流程：Get → CanDelete → Delete（权限在 handler）。
func (h *AttachmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ws, mid, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	id := r.PathValue("attachmentID")
	att, err := h.Svc.Get(r.Context(), ws, id)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	role := middleware.MemberRole(r.Context())
	if !attachments.CanDelete(mid, role, att) {
		middleware.HandleError(w, apperr.Forbidden("无权删除该附件"))
		return
	}
	if err := h.Svc.Delete(r.Context(), ws, id); err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"ok": true})
}
