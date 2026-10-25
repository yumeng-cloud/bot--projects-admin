package handlers

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/botmem"
)

type MemberHandler struct {
	DB *sql.DB
}

type memberDTO struct {
	ID          string  `json:"id"`
	WorkspaceID string  `json:"workspace_id"`
	UserID      *string `json:"user_id"`
	Role        string  `json:"role"`
	DisplayName string  `json:"display_name"`
	IsBot       bool    `json:"is_bot"`
	Email       *string `json:"email,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

func (h *MemberHandler) List(w http.ResponseWriter, r *http.Request) {
	wsID := middleware.WorkspaceID(r.Context())
	rows, err := h.DB.Query(`
		SELECT m.id, m.workspace_id, m.user_id, m.role, m.display_name, m.is_bot, m.created_at, u.email
		FROM members m
		LEFT JOIN users u ON u.id = m.user_id
		WHERE m.workspace_id = ?
		ORDER BY m.is_bot ASC, m.created_at ASC`, wsID)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	defer rows.Close()
	list := []memberDTO{}
	for rows.Next() {
		var m memberDTO
		var uid, email sql.NullString
		var isBot int
		if err := rows.Scan(&m.ID, &m.WorkspaceID, &uid, &m.Role, &m.DisplayName, &isBot, &m.CreatedAt, &email); err != nil {
			middleware.HandleError(w, err)
			return
		}
		m.IsBot = isBot == 1
		if uid.Valid {
			m.UserID = &uid.String
		}
		if email.Valid {
			m.Email = &email.String
		}
		list = append(list, m)
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"members": list})
}

func (h *MemberHandler) Invite(w http.ResponseWriter, r *http.Request) {
	wsID := middleware.WorkspaceID(r.Context())
	if middleware.MemberRole(r.Context()) != "admin" {
		middleware.HandleError(w, apperr.Forbidden("需要管理员权限"))
		return
	}
	var req struct {
		Email       string `json:"email"`
		Role        string `json:"role"`
		DisplayName string `json:"display_name"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		middleware.HandleError(w, apperr.BadRequest("邮箱格式不正确"))
		return
	}
	if req.Role == "" {
		req.Role = "member"
	}
	if req.Role != "admin" && req.Role != "member" {
		middleware.HandleError(w, apperr.BadRequest("角色仅支持 admin 或 member"))
		return
	}
	var userID, name string
	err := h.DB.QueryRow(`SELECT id, name FROM users WHERE email=?`, req.Email).Scan(&userID, &name)
	if err == sql.ErrNoRows {
		middleware.HandleError(w, apperr.NotFound("该邮箱尚未注册，请先让对方注册账号"))
		return
	}
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	display := strings.TrimSpace(req.DisplayName)
	if display == "" {
		display = name
	}
	var existing string
	err = h.DB.QueryRow(`SELECT id FROM members WHERE workspace_id=? AND user_id=?`, wsID, userID).Scan(&existing)
	if err == nil {
		middleware.HandleError(w, apperr.Conflict("该用户已是成员"))
		return
	}
	id := idgen.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = h.DB.Exec(`INSERT INTO members (id, workspace_id, user_id, role, display_name, is_bot, bot_enabled, created_at)
		VALUES (?,?,?,?,?,0,1,?)`, id, wsID, userID, req.Role, display, now)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	email := req.Email
	middleware.JSON(w, http.StatusCreated, map[string]any{
		"member": memberDTO{
			ID: id, WorkspaceID: wsID, UserID: &userID, Role: req.Role,
			DisplayName: display, Email: &email, CreatedAt: now,
		},
	})
}

func (h *MemberHandler) Patch(w http.ResponseWriter, r *http.Request) {
	wsID := middleware.WorkspaceID(r.Context())
	if middleware.MemberRole(r.Context()) != "admin" {
		middleware.HandleError(w, apperr.Forbidden("需要管理员权限"))
		return
	}
	mid := r.PathValue("memberID")
	var req struct {
		Role        *string `json:"role"`
		DisplayName *string `json:"display_name"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	var role, display string
	var uid sql.NullString
	err := h.DB.QueryRow(`SELECT role, display_name, user_id FROM members WHERE id=? AND workspace_id=?`, mid, wsID).
		Scan(&role, &display, &uid)
	if err == sql.ErrNoRows {
		middleware.HandleError(w, apperr.NotFound("成员不存在"))
		return
	}
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	if req.Role != nil {
		if *req.Role != "admin" && *req.Role != "member" && *req.Role != "bot" {
			middleware.HandleError(w, apperr.BadRequest("无效角色"))
			return
		}
		if err := guardLastAdmin(h.DB, wsID, mid, role, *req.Role, false); err != nil {
			middleware.HandleError(w, err)
			return
		}
		role = *req.Role
	}
	if req.DisplayName != nil {
		display = strings.TrimSpace(*req.DisplayName)
	}
	_, err = h.DB.Exec(`UPDATE members SET role=?, display_name=? WHERE id=?`, role, display, mid)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	m := memberDTO{ID: mid, WorkspaceID: wsID, Role: role, DisplayName: display}
	if uid.Valid {
		m.UserID = &uid.String
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"member": m})
}

func (h *MemberHandler) Remove(w http.ResponseWriter, r *http.Request) {
	wsID := middleware.WorkspaceID(r.Context())
	if middleware.MemberRole(r.Context()) != "admin" {
		middleware.HandleError(w, apperr.Forbidden("需要管理员权限"))
		return
	}
	mid := r.PathValue("memberID")
	actor := middleware.MemberID(r.Context())
	if mid == actor {
		middleware.HandleError(w, apperr.BadRequest("不能移除自己"))
		return
	}
	var targetRole string
	err := h.DB.QueryRow(`SELECT role FROM members WHERE id=? AND workspace_id=?`, mid, wsID).Scan(&targetRole)
	if err == sql.ErrNoRows {
		middleware.HandleError(w, apperr.NotFound("成员不存在"))
		return
	}
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	if err := guardLastAdmin(h.DB, wsID, mid, targetRole, "", true); err != nil {
		middleware.HandleError(w, err)
		return
	}
	// U-G1：先级联清该成员占用席（left + release claim + 审计），再删成员行
	svc := botmem.New(h.DB)
	if _, err := svc.CascadeUnbindSeatsForMember(wsID, mid, actor); err != nil {
		middleware.HandleError(w, err)
		return
	}
	res, err := h.DB.Exec(`DELETE FROM members WHERE id=? AND workspace_id=?`, mid, wsID)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		middleware.HandleError(w, apperr.NotFound("成员不存在"))
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "成员已移除"})
}

// countHumanAdmins returns number of non-bot admin members in a workspace.
func countHumanAdmins(db *sql.DB, wsID string) (int, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM members WHERE workspace_id=? AND role='admin' AND is_bot=0`, wsID).Scan(&n)
	return n, err
}

func guardLastAdmin(db *sql.DB, wsID, memberID, currentRole, newRole string, removing bool) error {
	if currentRole != "admin" {
		return nil
	}
	demoting := removing || (newRole != "" && newRole != "admin")
	if !demoting {
		return nil
	}
	n, err := countHumanAdmins(db, wsID)
	if err != nil {
		return err
	}
	if n <= 1 {
		return apperr.Conflict("不能移除或降级最后一个管理员")
	}
	return nil
}
