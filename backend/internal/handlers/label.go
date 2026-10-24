package handlers

import (
	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"

	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
)

type LabelHandler struct {
	DB *sql.DB
}

type labelDTO struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	CreatedAt   string `json:"created_at"`
}

func (h *LabelHandler) List(w http.ResponseWriter, r *http.Request) {
	wsID := middleware.WorkspaceID(r.Context())
	rows, err := h.DB.Query(`SELECT id, workspace_id, name, color, created_at FROM labels WHERE workspace_id=? ORDER BY name`, wsID)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	defer rows.Close()
	list := []labelDTO{}
	for rows.Next() {
		var l labelDTO
		if err := rows.Scan(&l.ID, &l.WorkspaceID, &l.Name, &l.Color, &l.CreatedAt); err != nil {
			middleware.HandleError(w, err)
			return
		}
		list = append(list, l)
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"labels": list})
}

func (h *LabelHandler) Create(w http.ResponseWriter, r *http.Request) {
	wsID := middleware.WorkspaceID(r.Context())
	var req struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		middleware.HandleError(w, apperr.BadRequest("标签名称不能为空"))
		return
	}
	if req.Color == "" {
		req.Color = "#1677ff"
	}
	id := idgen.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := h.DB.Exec(`INSERT INTO labels (id, workspace_id, name, color, created_at) VALUES (?,?,?,?,?)`,
		id, wsID, req.Name, req.Color, now)
	if err != nil {
		if db.IsUniqueViolation(err) {
			middleware.HandleError(w, apperr.Conflict("标签名称已存在"))
			return
		}
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusCreated, map[string]any{
		"label": labelDTO{ID: id, WorkspaceID: wsID, Name: req.Name, Color: req.Color, CreatedAt: now},
	})
}

func (h *LabelHandler) Patch(w http.ResponseWriter, r *http.Request) {
	wsID := middleware.WorkspaceID(r.Context())
	lid := r.PathValue("labelID")
	var req struct {
		Name  *string `json:"name"`
		Color *string `json:"color"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	var name, color, created string
	err := h.DB.QueryRow(`SELECT name, color, created_at FROM labels WHERE id=? AND workspace_id=?`, lid, wsID).
		Scan(&name, &color, &created)
	if err == sql.ErrNoRows {
		middleware.HandleError(w, apperr.NotFound("标签不存在"))
		return
	}
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	if req.Name != nil {
		name = strings.TrimSpace(*req.Name)
	}
	if req.Color != nil {
		color = *req.Color
	}
	_, err = h.DB.Exec(`UPDATE labels SET name=?, color=? WHERE id=?`, name, color, lid)
	if err != nil {
		if db.IsUniqueViolation(err) {
			middleware.HandleError(w, apperr.Conflict("标签名称已存在"))
			return
		}
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{
		"label": labelDTO{ID: lid, WorkspaceID: wsID, Name: name, Color: color, CreatedAt: created},
	})
}

func (h *LabelHandler) Delete(w http.ResponseWriter, r *http.Request) {
	wsID := middleware.WorkspaceID(r.Context())
	lid := r.PathValue("labelID")
	res, err := h.DB.Exec(`DELETE FROM labels WHERE id=? AND workspace_id=?`, lid, wsID)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		middleware.HandleError(w, apperr.NotFound("标签不存在"))
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "标签已删除"})
}
