package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
)

// AdminHandler platform super-admin APIs (/api/v1/admin/*).
type AdminHandler struct {
	DB *sql.DB
}

type adminWorkspaceDTO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	MemberCount int     `json:"member_count"`
	DisabledAt  *string `json:"disabled_at"`
	CreatedAt   string  `json:"created_at"`
}

type adminUserDTO struct {
	ID              string `json:"id"`
	Email           string `json:"email"`
	Name            string `json:"name"`
	IsPlatformAdmin bool   `json:"is_platform_admin"`
	CreatedAt       string `json:"created_at"`
}

func (h *AdminHandler) writeAudit(actorUserID, action, targetType, targetID string, payload any) {
	raw := "{}"
	if payload != nil {
		if b, err := json.Marshal(payload); err == nil {
			raw = string(b)
		}
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, _ = h.DB.Exec(`INSERT INTO platform_audit_logs (id, actor_user_id, action, target_type, target_id, payload, created_at)
		VALUES (?,?,?,?,?,?,?)`, idgen.New(), actorUserID, action, targetType, targetID, raw, now)
}

// ListWorkspaces GET /api/v1/admin/workspaces
func (h *AdminHandler) ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`
		SELECT w.id, w.name, w.slug, w.disabled_at, w.created_at,
			(SELECT COUNT(*) FROM members m WHERE m.workspace_id = w.id) AS member_count
		FROM workspaces w
		ORDER BY w.created_at DESC`)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	defer rows.Close()
	list := []adminWorkspaceDTO{}
	for rows.Next() {
		var it adminWorkspaceDTO
		var disabled sql.NullString
		if err := rows.Scan(&it.ID, &it.Name, &it.Slug, &disabled, &it.CreatedAt, &it.MemberCount); err != nil {
			middleware.HandleError(w, err)
			return
		}
		if disabled.Valid && disabled.String != "" {
			it.DisabledAt = &disabled.String
		}
		list = append(list, it)
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"workspaces": list})
}

// PatchWorkspace PATCH /api/v1/admin/workspaces/{id} — disable / restore
func (h *AdminHandler) PatchWorkspace(w http.ResponseWriter, r *http.Request) {
	wsID := r.PathValue("id")
	if wsID == "" {
		wsID = r.PathValue("workspaceID")
	}
	var req struct {
		Disabled *bool `json:"disabled"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	if req.Disabled == nil {
		middleware.HandleError(w, apperr.BadRequest("须提供 disabled 布尔字段"))
		return
	}
	var name, slug, created string
	var disabled sql.NullString
	err := h.DB.QueryRow(`SELECT name, slug, disabled_at, created_at FROM workspaces WHERE id=?`, wsID).
		Scan(&name, &slug, &disabled, &created)
	if err == sql.ErrNoRows {
		middleware.HandleError(w, apperr.NotFound("工作空间不存在"))
		return
	}
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	actor := middleware.UserID(r.Context())
	if *req.Disabled {
		_, err = h.DB.Exec(`UPDATE workspaces SET disabled_at=?, updated_at=? WHERE id=?`, now, now, wsID)
		if err != nil {
			middleware.HandleError(w, err)
			return
		}
		h.writeAudit(actor, "workspace.disable", "workspace", wsID, map[string]any{"name": name, "slug": slug})
		disabledAt := now
		middleware.JSON(w, http.StatusOK, map[string]any{
			"workspace": adminWorkspaceDTO{ID: wsID, Name: name, Slug: slug, DisabledAt: &disabledAt, CreatedAt: created},
		})
		return
	}
	_, err = h.DB.Exec(`UPDATE workspaces SET disabled_at=NULL, updated_at=? WHERE id=?`, now, wsID)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	h.writeAudit(actor, "workspace.enable", "workspace", wsID, map[string]any{"name": name, "slug": slug})
	middleware.JSON(w, http.StatusOK, map[string]any{
		"workspace": adminWorkspaceDTO{ID: wsID, Name: name, Slug: slug, DisabledAt: nil, CreatedAt: created},
	})
}

// CreateWorkspace POST /api/v1/admin/workspaces — create on behalf of owner
func (h *AdminHandler) CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		OwnerEmail  string `json:"owner_email"`
		OwnerUserID string `json:"owner_user_id"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Slug = strings.ToLower(strings.TrimSpace(req.Slug))
	req.OwnerEmail = strings.ToLower(strings.TrimSpace(req.OwnerEmail))
	req.OwnerUserID = strings.TrimSpace(req.OwnerUserID)
	if req.Name == "" {
		middleware.HandleError(w, apperr.BadRequest("工作空间名称不能为空"))
		return
	}
	if req.Slug == "" {
		req.Slug = slugify(req.Name)
	}
	if !slugRe.MatchString(req.Slug) {
		middleware.HandleError(w, apperr.BadRequest("slug 仅允许小写字母、数字与连字符"))
		return
	}
	if req.OwnerEmail == "" && req.OwnerUserID == "" {
		middleware.HandleError(w, apperr.BadRequest("须指定 owner_email 或 owner_user_id"))
		return
	}
	var ownerID, ownerName string
	var err error
	if req.OwnerUserID != "" {
		err = h.DB.QueryRow(`SELECT id, name FROM users WHERE id=?`, req.OwnerUserID).Scan(&ownerID, &ownerName)
	} else {
		err = h.DB.QueryRow(`SELECT id, name FROM users WHERE email=?`, req.OwnerEmail).Scan(&ownerID, &ownerName)
	}
	if err == sql.ErrNoRows {
		middleware.HandleError(w, apperr.NotFound("指定的所有者用户不存在"))
		return
	}
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	wsID := idgen.New()
	memID := idgen.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	tx, err := h.DB.Begin()
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	defer tx.Rollback()
	_, err = tx.Exec(`INSERT INTO workspaces (id, name, slug, created_at, updated_at) VALUES (?,?,?,?,?)`,
		wsID, req.Name, req.Slug, now, now)
	if err != nil {
		if db.IsUniqueViolation(err) {
			middleware.HandleError(w, apperr.Conflict("该 slug 已被占用"))
			return
		}
		middleware.HandleError(w, err)
		return
	}
	_, err = tx.Exec(`INSERT INTO members (id, workspace_id, user_id, role, display_name, is_bot, bot_enabled, created_at)
		VALUES (?,?,?,?,?,0,1,?)`, memID, wsID, ownerID, "admin", ownerName, now)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	if err := tx.Commit(); err != nil {
		middleware.HandleError(w, err)
		return
	}
	actor := middleware.UserID(r.Context())
	h.writeAudit(actor, "workspace.create_for", "workspace", wsID, map[string]any{
		"name": req.Name, "slug": req.Slug, "owner_user_id": ownerID,
	})
	middleware.JSON(w, http.StatusCreated, map[string]any{
		"workspace": map[string]any{"id": wsID, "name": req.Name, "slug": req.Slug, "disabled_at": nil, "created_at": now},
		"owner":     map[string]string{"user_id": ownerID, "member_id": memID, "role": "admin"},
	})
}

// ListUsers GET /api/v1/admin/users
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	var rows *sql.Rows
	var err error
	if q == "" {
		rows, err = h.DB.Query(`SELECT id, email, name, COALESCE(is_platform_admin,0), created_at FROM users ORDER BY created_at DESC`)
	} else {
		like := "%" + strings.ToLower(q) + "%"
		rows, err = h.DB.Query(`SELECT id, email, name, COALESCE(is_platform_admin,0), created_at FROM users
			WHERE lower(email) LIKE ? OR lower(name) LIKE ?
			ORDER BY created_at DESC`, like, like)
	}
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	defer rows.Close()
	list := []adminUserDTO{}
	for rows.Next() {
		var it adminUserDTO
		var flag int
		if err := rows.Scan(&it.ID, &it.Email, &it.Name, &flag, &it.CreatedAt); err != nil {
			middleware.HandleError(w, err)
			return
		}
		it.IsPlatformAdmin = flag == 1
		list = append(list, it)
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"users": list})
}

// PatchUser PATCH /api/v1/admin/users/{id} — grant/revoke is_platform_admin
func (h *AdminHandler) PatchUser(w http.ResponseWriter, r *http.Request) {
	uid := r.PathValue("id")
	if uid == "" {
		uid = r.PathValue("userID")
	}
	var req struct {
		IsPlatformAdmin *bool `json:"is_platform_admin"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	if req.IsPlatformAdmin == nil {
		middleware.HandleError(w, apperr.BadRequest("须提供 is_platform_admin 布尔字段"))
		return
	}
	var email, name, created string
	var flag int
	err := h.DB.QueryRow(`SELECT email, name, COALESCE(is_platform_admin,0), created_at FROM users WHERE id=?`, uid).
		Scan(&email, &name, &flag, &created)
	if err == sql.ErrNoRows {
		middleware.HandleError(w, apperr.NotFound("用户不存在"))
		return
	}
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	newFlag := 0
	if *req.IsPlatformAdmin {
		newFlag = 1
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = h.DB.Exec(`UPDATE users SET is_platform_admin=?, updated_at=? WHERE id=?`, newFlag, now, uid)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	actor := middleware.UserID(r.Context())
	action := "user.revoke_platform_admin"
	if newFlag == 1 {
		action = "user.grant_platform_admin"
	}
	h.writeAudit(actor, action, "user", uid, map[string]any{"email": email, "is_platform_admin": newFlag == 1})
	middleware.JSON(w, http.StatusOK, map[string]any{
		"user": adminUserDTO{ID: uid, Email: email, Name: name, IsPlatformAdmin: newFlag == 1, CreatedAt: created},
	})
}
