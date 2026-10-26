package handlers

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
)

// SessionContextHandler serves GET|PUT /api/v1/me/session-context (human JWT only).
// Isolated from Bot handoff / botmem — do not share tables or routes.
type SessionContextHandler struct {
	DB *sql.DB
}

// sessionContextDTO matches MOBILE_V2 debt camelCase field names.
type sessionContextDTO struct {
	WsID      string `json:"wsId"`
	TeamID    string `json:"teamId"`
	ProjectID string `json:"projectId"`
	IssueID   string `json:"issueId"`
	View      string `json:"view"`
	UpdatedAt string `json:"updatedAt"`
}

// sessionContextPutBody accepts the same fields; updatedAt from client is ignored.
type sessionContextPutBody struct {
	WsID      *string `json:"wsId"`
	TeamID    *string `json:"teamId"`
	ProjectID *string `json:"projectId"`
	IssueID   *string `json:"issueId"`
	View      *string `json:"view"`
	UpdatedAt *string `json:"updatedAt"` // ignored; server bumps
}

func emptySessionContext() sessionContextDTO {
	return sessionContextDTO{}
}

// GetSessionContext GET /api/v1/me/session-context — empty defaults when no row yet.
func (h *SessionContextHandler) GetSessionContext(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r.Context())
	if uid == "" {
		middleware.HandleError(w, apperr.Unauthorized("请先登录"))
		return
	}
	dto, err := h.read(uid)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, dto)
}

// PutSessionContext PUT /api/v1/me/session-context — last-write-wins full replace.
func (h *SessionContextHandler) PutSessionContext(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r.Context())
	if uid == "" {
		middleware.HandleError(w, apperr.Unauthorized("请先登录"))
		return
	}
	var req sessionContextPutBody
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	dto := sessionContextDTO{
		WsID:      derefTrim(req.WsID),
		TeamID:    derefTrim(req.TeamID),
		ProjectID: derefTrim(req.ProjectID),
		IssueID:   derefTrim(req.IssueID),
		View:      derefTrim(req.View),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
	_, err := h.DB.Exec(`
		INSERT INTO user_session_contexts (user_id, ws_id, team_id, project_id, issue_id, view, updated_at)
		VALUES (?,?,?,?,?,?,?)
		ON CONFLICT(user_id) DO UPDATE SET
			ws_id = excluded.ws_id,
			team_id = excluded.team_id,
			project_id = excluded.project_id,
			issue_id = excluded.issue_id,
			view = excluded.view,
			updated_at = excluded.updated_at`,
		uid, dto.WsID, dto.TeamID, dto.ProjectID, dto.IssueID, dto.View, dto.UpdatedAt)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, dto)
}

func (h *SessionContextHandler) read(userID string) (sessionContextDTO, error) {
	var dto sessionContextDTO
	err := h.DB.QueryRow(`
		SELECT ws_id, team_id, project_id, issue_id, view, updated_at
		FROM user_session_contexts WHERE user_id = ?`, userID).
		Scan(&dto.WsID, &dto.TeamID, &dto.ProjectID, &dto.IssueID, &dto.View, &dto.UpdatedAt)
	if err == sql.ErrNoRows {
		return emptySessionContext(), nil
	}
	if err != nil {
		return sessionContextDTO{}, err
	}
	return dto, nil
}

func derefTrim(p *string) string {
	if p == nil {
		return ""
	}
	return strings.TrimSpace(*p)
}
