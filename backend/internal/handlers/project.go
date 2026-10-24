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

type ProjectHandler struct {
	DB *sql.DB
}

type projectDTO struct {
	ID          string  `json:"id"`
	WorkspaceID string  `json:"workspace_id"`
	Name        string  `json:"name"`
	KeyPrefix   string  `json:"key_prefix"`
	Description string  `json:"description"`
	ArchivedAt  *string `json:"archived_at"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	wsID := middleware.WorkspaceID(r.Context())
	includeArchived := r.URL.Query().Get("include_archived") == "1" || r.URL.Query().Get("include_archived") == "true"
	q := `SELECT id, workspace_id, name, key_prefix, COALESCE(description,''), archived_at, created_at, updated_at
		FROM projects WHERE workspace_id = ?`
	if !includeArchived {
		q += ` AND archived_at IS NULL`
	}
	q += ` ORDER BY created_at ASC`
	rows, err := h.DB.Query(q, wsID)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	defer rows.Close()
	list := []projectDTO{}
	for rows.Next() {
		var p projectDTO
		var arch sql.NullString
		if err := rows.Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.KeyPrefix, &p.Description, &arch, &p.CreatedAt, &p.UpdatedAt); err != nil {
			middleware.HandleError(w, err)
			return
		}
		if arch.Valid {
			p.ArchivedAt = &arch.String
		}
		list = append(list, p)
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"projects": list})
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	wsID := middleware.WorkspaceID(r.Context())
	var req struct {
		Name        string `json:"name"`
		KeyPrefix   string `json:"key_prefix"`
		Description string `json:"description"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.KeyPrefix = strings.ToUpper(strings.TrimSpace(req.KeyPrefix))
	if req.Name == "" {
		middleware.HandleError(w, apperr.BadRequest("项目名称不能为空"))
		return
	}
	if req.KeyPrefix == "" || len(req.KeyPrefix) > 10 {
		middleware.HandleError(w, apperr.BadRequest("项目前缀须为 1–10 位字母数字"))
		return
	}
	for _, c := range req.KeyPrefix {
		if !((c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			middleware.HandleError(w, apperr.BadRequest("项目前缀仅允许大写字母与数字"))
			return
		}
	}
	id := idgen.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	tx, err := h.DB.Begin()
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	defer tx.Rollback()
	_, err = tx.Exec(`INSERT INTO projects (id, workspace_id, name, key_prefix, description, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?)`, id, wsID, req.Name, req.KeyPrefix, req.Description, now, now)
	if err != nil {
		if db.IsUniqueViolation(err) {
			middleware.HandleError(w, apperr.Conflict("该项目前缀已存在"))
			return
		}
		middleware.HandleError(w, err)
		return
	}
	_, err = tx.Exec(`INSERT INTO issue_counters (project_id, next_number) VALUES (?, 1)`, id)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	if err := tx.Commit(); err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusCreated, map[string]any{
		"project": projectDTO{
			ID: id, WorkspaceID: wsID, Name: req.Name, KeyPrefix: req.KeyPrefix,
			Description: req.Description, CreatedAt: now, UpdatedAt: now,
		},
	})
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	p, err := h.loadProjectForActor(r)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"project": p})
}

func (h *ProjectHandler) Patch(w http.ResponseWriter, r *http.Request) {
	p, err := h.loadProjectForActor(r)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Archive     *bool   `json:"archive"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	role := middleware.MemberRole(r.Context())
	if role == "" {
		uid := middleware.UserID(r.Context())
		_ = h.DB.QueryRow(`SELECT role FROM members WHERE workspace_id=? AND user_id=?`, p.WorkspaceID, uid).Scan(&role)
	}
	if req.Archive != nil && *req.Archive && role != "admin" {
		middleware.HandleError(w, apperr.Forbidden("仅管理员可归档项目"))
		return
	}
	name, desc := p.Name, p.Description
	if req.Name != nil {
		name = strings.TrimSpace(*req.Name)
		if name == "" {
			middleware.HandleError(w, apperr.BadRequest("项目名称不能为空"))
			return
		}
	}
	if req.Description != nil {
		desc = *req.Description
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	var archived any
	if req.Archive != nil {
		if *req.Archive {
			archived = now
		} else {
			archived = nil
		}
	} else if p.ArchivedAt != nil {
		archived = *p.ArchivedAt
	} else {
		archived = nil
	}
	_, err = h.DB.Exec(`UPDATE projects SET name=?, description=?, archived_at=?, updated_at=? WHERE id=?`,
		name, desc, archived, now, p.ID)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	p.Name, p.Description, p.UpdatedAt = name, desc, now
	if req.Archive != nil {
		if *req.Archive {
			p.ArchivedAt = &now
		} else {
			p.ArchivedAt = nil
		}
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"project": p})
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if middleware.MemberRole(r.Context()) != "admin" && middleware.ActorKind(r.Context()) != "bot" {
		// resolve role via project membership if coming from /projects/{id}
	}
	p, err := h.loadProjectForActor(r)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	role := middleware.MemberRole(r.Context())
	if role == "" {
		uid := middleware.UserID(r.Context())
		_ = h.DB.QueryRow(`SELECT role FROM members WHERE workspace_id=? AND user_id=?`, p.WorkspaceID, uid).Scan(&role)
	}
	if role != "admin" {
		middleware.HandleError(w, apperr.Forbidden("仅管理员可删除项目"))
		return
	}
	_, err = h.DB.Exec(`DELETE FROM projects WHERE id=?`, p.ID)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "项目已删除"})
}

func (h *ProjectHandler) loadProjectForActor(r *http.Request) (projectDTO, error) {
	pid := r.PathValue("projectID")
	if pid == "" {
		pid = r.PathValue("id")
	}
	var p projectDTO
	var arch sql.NullString
	err := h.DB.QueryRow(`
		SELECT id, workspace_id, name, key_prefix, COALESCE(description,''), archived_at, created_at, updated_at
		FROM projects WHERE id = ?`, pid).Scan(
		&p.ID, &p.WorkspaceID, &p.Name, &p.KeyPrefix, &p.Description, &arch, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return p, apperr.NotFound("项目不存在")
	}
	if err != nil {
		return p, err
	}
	if arch.Valid {
		p.ArchivedAt = &arch.String
	}

	ctx := r.Context()
	if middleware.ActorKind(ctx) == "bot" {
		if middleware.WorkspaceID(ctx) != "" && middleware.WorkspaceID(ctx) != p.WorkspaceID {
			return p, apperr.Forbidden("无权访问该项目")
		}
		return p, nil
	}
	uid := middleware.UserID(ctx)
	var role, mid string
	err = h.DB.QueryRow(`SELECT id, role FROM members WHERE workspace_id=? AND user_id=?`, p.WorkspaceID, uid).Scan(&mid, &role)
	if err == sql.ErrNoRows {
		return p, apperr.Forbidden("您不是该工作空间的成员")
	}
	if err != nil {
		return p, err
	}
	return p, nil
}
