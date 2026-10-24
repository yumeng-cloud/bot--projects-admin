package handlers

import (
	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"

	"database/sql"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
)

type WorkspaceHandler struct {
	DB *sql.DB
}

var slugRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func (h *WorkspaceHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r.Context())
	rows, err := h.DB.Query(`
		SELECT w.id, w.name, w.slug, w.disabled_at, m.role, m.id
		FROM workspaces w
		JOIN members m ON m.workspace_id = w.id
		WHERE m.user_id = ?
		ORDER BY w.created_at ASC`, uid)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	defer rows.Close()
	type item struct {
		ID          string  `json:"id"`
		Name        string  `json:"name"`
		Slug        string  `json:"slug"`
		DisabledAt  *string `json:"disabled_at"`
		Role        string  `json:"role"`
		MemberID    string  `json:"member_id"`
	}
	list := []item{}
	for rows.Next() {
		var it item
		var disabled sql.NullString
		if err := rows.Scan(&it.ID, &it.Name, &it.Slug, &disabled, &it.Role, &it.MemberID); err != nil {
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

func (h *WorkspaceHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r.Context())
	var req struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Slug = strings.ToLower(strings.TrimSpace(req.Slug))
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
	var display string
	_ = h.DB.QueryRow(`SELECT name FROM users WHERE id=?`, uid).Scan(&display)
	if display == "" {
		display = "管理员"
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
		VALUES (?,?,?,?,?,0,1,?)`, memID, wsID, uid, "admin", display, now)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	if err := tx.Commit(); err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusCreated, map[string]any{
		"workspace": map[string]string{"id": wsID, "name": req.Name, "slug": req.Slug},
		"member_id": memID,
		"role":      "admin",
	})
}

func (h *WorkspaceHandler) Get(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r.Context())
	wsID := r.PathValue("workspaceID")
	// bots may have workspace already
	if middleware.ActorKind(r.Context()) == "bot" {
		var name, slug string
		err := h.DB.QueryRow(`SELECT name, slug FROM workspaces WHERE id=?`, wsID).Scan(&name, &slug)
		if err == sql.ErrNoRows {
			middleware.HandleError(w, apperr.NotFound("工作空间不存在"))
			return
		}
		if err != nil {
			middleware.HandleError(w, err)
			return
		}
		middleware.JSON(w, http.StatusOK, map[string]any{
			"workspace": map[string]string{"id": wsID, "name": name, "slug": slug},
			"role":      middleware.MemberRole(r.Context()),
			"member_id": middleware.MemberID(r.Context()),
		})
		return
	}
	var name, slug, role, memberID string
	err := h.DB.QueryRow(`
		SELECT w.name, w.slug, m.role, m.id FROM workspaces w
		JOIN members m ON m.workspace_id = w.id
		WHERE w.id = ? AND m.user_id = ?`, wsID, uid).Scan(&name, &slug, &role, &memberID)
	if err == sql.ErrNoRows {
		middleware.HandleError(w, apperr.NotFound("工作空间不存在或无权访问"))
		return
	}
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{
		"workspace": map[string]string{"id": wsID, "name": name, "slug": slug},
		"role":      role,
		"member_id": memberID,
	})
}

func (h *WorkspaceHandler) Current(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r.Context())
	var wsID, name, slug, role, memberID string
	err := h.DB.QueryRow(`
		SELECT w.id, w.name, w.slug, m.role, m.id
		FROM workspaces w
		JOIN members m ON m.workspace_id = w.id
		WHERE m.user_id = ?
		ORDER BY w.created_at ASC LIMIT 1`, uid).Scan(&wsID, &name, &slug, &role, &memberID)
	if err == sql.ErrNoRows {
		middleware.HandleError(w, apperr.NotFound("尚未加入任何工作空间"))
		return
	}
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{
		"workspace": map[string]string{"id": wsID, "name": name, "slug": slug},
		"role":      role,
		"member_id": memberID,
	})
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
		} else if !prevDash {
			b.WriteByte('-')
			prevDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		out = "ws-" + idgen.New()[:8]
	}
	return out
}

// Patch updates workspace name/slug. Requires WS admin (caller must be behind RequireAdmin).
func (h *WorkspaceHandler) Patch(w http.ResponseWriter, r *http.Request) {
	wsID := middleware.WorkspaceID(r.Context())
	if wsID == "" {
		wsID = r.PathValue("workspaceID")
	}
	var req struct {
		Name *string `json:"name"`
		Slug *string `json:"slug"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	var name, slug string
	var disabled sql.NullString
	err := h.DB.QueryRow(`SELECT name, slug, disabled_at FROM workspaces WHERE id=?`, wsID).Scan(&name, &slug, &disabled)
	if err == sql.ErrNoRows {
		middleware.HandleError(w, apperr.NotFound("工作空间不存在"))
		return
	}
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	if req.Name != nil {
		n := strings.TrimSpace(*req.Name)
		if n == "" {
			middleware.HandleError(w, apperr.BadRequest("工作空间名称不能为空"))
			return
		}
		name = n
	}
	if req.Slug != nil {
		s := strings.ToLower(strings.TrimSpace(*req.Slug))
		if s == "" {
			middleware.HandleError(w, apperr.BadRequest("slug 不能为空"))
			return
		}
		if !slugRe.MatchString(s) {
			middleware.HandleError(w, apperr.BadRequest("slug 仅允许小写字母、数字与连字符"))
			return
		}
		slug = s
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = h.DB.Exec(`UPDATE workspaces SET name=?, slug=?, updated_at=? WHERE id=?`, name, slug, now, wsID)
	if err != nil {
		if db.IsUniqueViolation(err) {
			middleware.HandleError(w, apperr.Conflict("该 slug 已被占用"))
			return
		}
		middleware.HandleError(w, err)
		return
	}
	out := map[string]any{"id": wsID, "name": name, "slug": slug}
	if disabled.Valid && disabled.String != "" {
		out["disabled_at"] = disabled.String
	} else {
		out["disabled_at"] = nil
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"workspace": out})
}
