package handlers

import (
	"net/http"
	"strings"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/botmem"
)

func (h *TeamHandler) pinAccess(w http.ResponseWriter, r *http.Request) (ws, teamID, memberID string, ok bool) {
	ws, memberID, ok = requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return "", "", "", false
	}
	teamID = r.PathValue("teamID")
	isAdmin := resolveMemberRole(h.DB, r, memberID) == "admin"
	if err := h.Svc.RequireTeamPinAccess(ws, teamID, memberID, isAdmin); err != nil {
		middleware.HandleError(w, err)
		return "", "", "", false
	}
	return ws, teamID, memberID, true
}

// ListPins GET …/teams/{teamID}/pins — 本队记忆墙列表。
func (h *TeamHandler) ListPins(w http.ResponseWriter, r *http.Request) {
	ws, teamID, _, ok := h.pinAccess(w, r)
	if !ok {
		return
	}
	list, err := h.Svc.ListPins(ws, teamID)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"pins": list})
}

// GetPin GET …/teams/{teamID}/pins/{key}
func (h *TeamHandler) GetPin(w http.ResponseWriter, r *http.Request) {
	ws, teamID, _, ok := h.pinAccess(w, r)
	if !ok {
		return
	}
	pin, err := h.Svc.GetPin(ws, teamID, r.PathValue("key"))
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"pin": pin})
}

// PutPin PUT …/teams/{teamID}/pins — upsert by key；body {key, value}（与 API.md 契约一致）。
func (h *TeamHandler) PutPin(w http.ResponseWriter, r *http.Request) {
	ws, teamID, memberID, ok := h.pinAccess(w, r)
	if !ok {
		return
	}
	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	// 若路由带 {key}（兼容别名），优先 path
	if pathKey := strings.TrimSpace(r.PathValue("key")); pathKey != "" {
		req.Key = pathKey
	}
	pin, err := h.Svc.UpsertPin(ws, teamID, botmem.UpsertPinInput{
		Key: req.Key, Value: req.Value, UpdatedBy: memberID,
	})
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"pin": pin})
}

// DeletePin DELETE …/teams/{teamID}/pins/{key}
func (h *TeamHandler) DeletePin(w http.ResponseWriter, r *http.Request) {
	ws, teamID, _, ok := h.pinAccess(w, r)
	if !ok {
		return
	}
	if err := h.Svc.DeletePin(ws, teamID, r.PathValue("key")); err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"ok": true})
}
