package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
)

// boardColumnPrefItem is one ordered column override (PRODUCT_SUITE S4).
type boardColumnPrefItem struct {
	Status string  `json:"status"`
	Hidden *bool   `json:"hidden,omitempty"`
	Label  *string `json:"label,omitempty"`
}

// boardColumnPrefsBody is the stored JSON shape and PUT body.
type boardColumnPrefsBody struct {
	TemplateID *string               `json:"template_id,omitempty"`
	Columns    []boardColumnPrefItem `json:"columns"`
}

type boardColumnPrefsDTO struct {
	ProjectID   string                `json:"project_id"`
	WorkspaceID string                `json:"workspace_id"`
	TemplateID  *string               `json:"template_id,omitempty"`
	Columns     []boardColumnPrefItem `json:"columns"`
	UpdatedAt   string                `json:"updated_at,omitempty"`
}

var issueStatusOK = map[string]bool{
	"backlog": true, "todo": true, "in_progress": true,
	"in_review": true, "done": true, "canceled": true,
}

// defaultBoardColumns matches FE BOARD_COLUMNS (默认五列, no canceled).
func defaultBoardColumns() []boardColumnPrefItem {
	return []boardColumnPrefItem{
		{Status: "backlog"},
		{Status: "todo"},
		{Status: "in_progress"},
		{Status: "in_review"},
		{Status: "done"},
	}
}

// GetBoardColumnPrefs GET …/projects/{projectID}/board-column-prefs
func (h *ProjectHandler) GetBoardColumnPrefs(w http.ResponseWriter, r *http.Request) {
	p, err := h.loadProjectForActor(r)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	if ws := middleware.WorkspaceID(r.Context()); ws != "" && ws != p.WorkspaceID {
		middleware.HandleError(w, apperr.Forbidden("无权访问该项目"))
		return
	}
	dto, err := h.readBoardColumnPrefs(p)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"prefs": dto})
}

// PutBoardColumnPrefs PUT …/projects/{projectID}/board-column-prefs — full replace.
// Empty columns (or null body columns) clears prefs (= 恢复默认五列).
func (h *ProjectHandler) PutBoardColumnPrefs(w http.ResponseWriter, r *http.Request) {
	p, err := h.loadProjectForActor(r)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	if ws := middleware.WorkspaceID(r.Context()); ws != "" && ws != p.WorkspaceID {
		middleware.HandleError(w, apperr.Forbidden("无权访问该项目"))
		return
	}
	var req boardColumnPrefsBody
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	if len(req.Columns) == 0 {
		if err := h.clearBoardColumnPrefs(p.ID); err != nil {
			middleware.HandleError(w, err)
			return
		}
		p.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
		dto := boardColumnPrefsDTO{
			ProjectID:   p.ID,
			WorkspaceID: p.WorkspaceID,
			Columns:     defaultBoardColumns(),
			UpdatedAt:   p.UpdatedAt,
		}
		middleware.JSON(w, http.StatusOK, map[string]any{"prefs": dto})
		return
	}
	if err := validateBoardColumnPrefs(&req); err != nil {
		middleware.HandleError(w, err)
		return
	}
	raw, err := json.Marshal(req)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = h.DB.Exec(`UPDATE projects SET board_column_prefs=?, updated_at=? WHERE id=?`, string(raw), now, p.ID)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	dto := boardColumnPrefsDTO{
		ProjectID:   p.ID,
		WorkspaceID: p.WorkspaceID,
		TemplateID:  req.TemplateID,
		Columns:     req.Columns,
		UpdatedAt:   now,
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"prefs": dto})
}

// DeleteBoardColumnPrefs DELETE …/projects/{projectID}/board-column-prefs — 恢复默认.
func (h *ProjectHandler) DeleteBoardColumnPrefs(w http.ResponseWriter, r *http.Request) {
	p, err := h.loadProjectForActor(r)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	if ws := middleware.WorkspaceID(r.Context()); ws != "" && ws != p.WorkspaceID {
		middleware.HandleError(w, apperr.Forbidden("无权访问该项目"))
		return
	}
	if err := h.clearBoardColumnPrefs(p.ID); err != nil {
		middleware.HandleError(w, err)
		return
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	dto := boardColumnPrefsDTO{
		ProjectID:   p.ID,
		WorkspaceID: p.WorkspaceID,
		Columns:     defaultBoardColumns(),
		UpdatedAt:   now,
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"prefs": dto})
}

func (h *ProjectHandler) clearBoardColumnPrefs(projectID string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := h.DB.Exec(`UPDATE projects SET board_column_prefs=NULL, updated_at=? WHERE id=?`, now, projectID)
	return err
}

func (h *ProjectHandler) readBoardColumnPrefs(p projectDTO) (boardColumnPrefsDTO, error) {
	var raw sql.NullString
	err := h.DB.QueryRow(`SELECT board_column_prefs FROM projects WHERE id=?`, p.ID).Scan(&raw)
	if err != nil {
		return boardColumnPrefsDTO{}, err
	}
	dto := boardColumnPrefsDTO{
		ProjectID:   p.ID,
		WorkspaceID: p.WorkspaceID,
		UpdatedAt:   p.UpdatedAt,
	}
	if !raw.Valid || strings.TrimSpace(raw.String) == "" || raw.String == "null" {
		dto.Columns = defaultBoardColumns()
		return dto, nil
	}
	var body boardColumnPrefsBody
	if err := json.Unmarshal([]byte(raw.String), &body); err != nil {
		return dto, apperr.Internal("列偏好数据损坏")
	}
	if len(body.Columns) == 0 {
		dto.Columns = defaultBoardColumns()
		return dto, nil
	}
	dto.TemplateID = body.TemplateID
	dto.Columns = body.Columns
	return dto, nil
}

func validateBoardColumnPrefs(req *boardColumnPrefsBody) error {
	seen := map[string]bool{}
	for i, c := range req.Columns {
		st := strings.TrimSpace(c.Status)
		if st == "" {
			return apperr.BadRequest("第 " + strconv.Itoa(i+1) + " 列缺少 status")
		}
		if !issueStatusOK[st] {
			return apperr.BadRequest("不支持的状态: " + st)
		}
		if seen[st] {
			return apperr.BadRequest("状态重复: " + st)
		}
		seen[st] = true
		req.Columns[i].Status = st
		if c.Label != nil {
			lb := strings.TrimSpace(*c.Label)
			if lb == "" {
				req.Columns[i].Label = nil
			} else {
				req.Columns[i].Label = &lb
			}
		}
	}
	return nil
}

