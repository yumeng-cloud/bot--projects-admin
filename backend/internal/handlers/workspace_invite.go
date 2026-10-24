package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
)

// WorkspaceInviteHandler human workspace invite codes (ymw_ — isolated from team ymt_).
type WorkspaceInviteHandler struct {
	DB *sql.DB
}

const workspaceInvitePrefix = "ymw_"

type workspaceInviteDTO struct {
	ID          string  `json:"id"`
	WorkspaceID string  `json:"workspace_id"`
	Email       *string `json:"email,omitempty"`
	Role        string  `json:"role"`
	TokenPrefix string  `json:"token_prefix"`
	CodeTail    string  `json:"code_tail"`
	MaxUses     *int    `json:"max_uses,omitempty"`
	UseCount    int     `json:"use_count"`
	ExpiresAt   *string `json:"expires_at,omitempty"`
	RevokedAt   *string `json:"revoked_at,omitempty"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
	Code        string  `json:"code,omitempty"` // plaintext only on create
}

func hashWSInviteToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func newWSInviteCode() (plain, prefix, hash string, err error) {
	b := make([]byte, 24)
	if _, err = rand.Read(b); err != nil {
		return "", "", "", err
	}
	plain = workspaceInvitePrefix + hex.EncodeToString(b)
	prefix = plain
	if len(prefix) > 12 {
		prefix = prefix[:12]
	}
	hash = hashWSInviteToken(plain)
	return plain, prefix, hash, nil
}

func inviteCodeTail(prefix string) string {
	if len(prefix) <= 4 {
		return prefix
	}
	return prefix[len(prefix)-4:]
}

func inviteStatus(revoked, expires sql.NullString, maxUses sql.NullInt64, useCount int) string {
	if revoked.Valid && revoked.String != "" {
		return "revoked"
	}
	if expires.Valid && expires.String != "" {
		if t, err := time.Parse(time.RFC3339Nano, expires.String); err == nil && time.Now().UTC().After(t) {
			return "expired"
		}
		if t, err := time.Parse(time.RFC3339, expires.String); err == nil && time.Now().UTC().After(t) {
			return "expired"
		}
	}
	if maxUses.Valid && maxUses.Int64 > 0 && int64(useCount) >= maxUses.Int64 {
		return "exhausted"
	}
	return "active"
}

// Create POST /workspaces/{id}/invites
func (h *WorkspaceInviteHandler) Create(w http.ResponseWriter, r *http.Request) {
	wsID := middleware.WorkspaceID(r.Context())
	var req struct {
		Email     string `json:"email"`
		Role      string `json:"role"`
		MaxUses   *int   `json:"max_uses"`
		ExpiresAt string `json:"expires_at"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Role == "" {
		req.Role = "member"
	}
	if req.Role != "admin" && req.Role != "member" {
		middleware.HandleError(w, apperr.BadRequest("角色仅支持 admin 或 member"))
		return
	}
	if req.Email != "" && !strings.Contains(req.Email, "@") {
		middleware.HandleError(w, apperr.BadRequest("邮箱格式不正确"))
		return
	}
	var expires any
	if v := strings.TrimSpace(req.ExpiresAt); v != "" {
		if _, err := time.Parse(time.RFC3339, v); err != nil {
			if _, err2 := time.Parse(time.RFC3339Nano, v); err2 != nil {
				middleware.HandleError(w, apperr.BadRequest("expires_at 须为 RFC3339"))
				return
			}
		}
		expires = v
	}
	var maxUses any
	if req.MaxUses != nil {
		if *req.MaxUses < 0 {
			middleware.HandleError(w, apperr.BadRequest("max_uses 不能为负"))
			return
		}
		maxUses = *req.MaxUses
	}
	plain, prefix, hash, err := newWSInviteCode()
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	id := idgen.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	createdBy := middleware.MemberID(r.Context())
	var emailAny any
	if req.Email != "" {
		emailAny = req.Email
	}
	var createdByAny any
	if createdBy != "" {
		createdByAny = createdBy
	}
	_, err = h.DB.Exec(`INSERT INTO workspace_invites
		(id, workspace_id, email, role, token_hash, token_prefix, max_uses, use_count, expires_at, revoked_at, created_by, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,0,?,NULL,?,?,?)`,
		id, wsID, emailAny, req.Role, hash, prefix, maxUses, expires, createdByAny, now, now)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	dto := workspaceInviteDTO{
		ID: id, WorkspaceID: wsID, Role: req.Role, TokenPrefix: prefix,
		CodeTail: inviteCodeTail(prefix), UseCount: 0, Status: "active", CreatedAt: now, Code: plain,
	}
	if req.Email != "" {
		dto.Email = &req.Email
	}
	if req.MaxUses != nil {
		dto.MaxUses = req.MaxUses
	}
	if exp, ok := expires.(string); ok {
		dto.ExpiresAt = &exp
	}
	middleware.JSON(w, http.StatusCreated, map[string]any{"invite": dto, "code": plain})
}

// List GET /workspaces/{id}/invites
func (h *WorkspaceInviteHandler) List(w http.ResponseWriter, r *http.Request) {
	wsID := middleware.WorkspaceID(r.Context())
	rows, err := h.DB.Query(`
		SELECT id, workspace_id, email, role, token_prefix, max_uses, use_count, expires_at, revoked_at, created_at
		FROM workspace_invites WHERE workspace_id=? ORDER BY created_at DESC`, wsID)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	defer rows.Close()
	list := []workspaceInviteDTO{}
	for rows.Next() {
		var it workspaceInviteDTO
		var email, expires, revoked sql.NullString
		var maxUses sql.NullInt64
		if err := rows.Scan(&it.ID, &it.WorkspaceID, &email, &it.Role, &it.TokenPrefix, &maxUses, &it.UseCount, &expires, &revoked, &it.CreatedAt); err != nil {
			middleware.HandleError(w, err)
			return
		}
		it.CodeTail = inviteCodeTail(it.TokenPrefix)
		it.Status = inviteStatus(revoked, expires, maxUses, it.UseCount)
		if email.Valid && email.String != "" {
			it.Email = &email.String
		}
		if maxUses.Valid {
			v := int(maxUses.Int64)
			it.MaxUses = &v
		}
		if expires.Valid && expires.String != "" {
			it.ExpiresAt = &expires.String
		}
		if revoked.Valid && revoked.String != "" {
			it.RevokedAt = &revoked.String
		}
		list = append(list, it)
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"invites": list})
}

// Revoke POST /workspaces/{id}/invites/{inviteID}/revoke
func (h *WorkspaceInviteHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	wsID := middleware.WorkspaceID(r.Context())
	inviteID := r.PathValue("inviteID")
	var email, expires, revoked sql.NullString
	var maxUses sql.NullInt64
	var it workspaceInviteDTO
	err := h.DB.QueryRow(`
		SELECT id, workspace_id, email, role, token_prefix, max_uses, use_count, expires_at, revoked_at, created_at
		FROM workspace_invites WHERE id=? AND workspace_id=?`, inviteID, wsID).
		Scan(&it.ID, &it.WorkspaceID, &email, &it.Role, &it.TokenPrefix, &maxUses, &it.UseCount, &expires, &revoked, &it.CreatedAt)
	if err == sql.ErrNoRows {
		middleware.HandleError(w, apperr.NotFound("邀请不存在"))
		return
	}
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if !(revoked.Valid && revoked.String != "") {
		_, err = h.DB.Exec(`UPDATE workspace_invites SET revoked_at=?, updated_at=? WHERE id=? AND workspace_id=?`,
			now, now, inviteID, wsID)
		if err != nil {
			middleware.HandleError(w, err)
			return
		}
		revoked = sql.NullString{String: now, Valid: true}
	}
	it.CodeTail = inviteCodeTail(it.TokenPrefix)
	it.Status = inviteStatus(revoked, expires, maxUses, it.UseCount)
	if email.Valid && email.String != "" {
		it.Email = &email.String
	}
	if maxUses.Valid {
		v := int(maxUses.Int64)
		it.MaxUses = &v
	}
	if expires.Valid && expires.String != "" {
		it.ExpiresAt = &expires.String
	}
	if revoked.Valid && revoked.String != "" {
		it.RevokedAt = &revoked.String
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"invite": it})
}

// Redeem POST /api/v1/workspace-invites/redeem — JWT login user only
func (h *WorkspaceInviteHandler) Redeem(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r.Context())
	if uid == "" {
		middleware.HandleError(w, apperr.Unauthorized("请先登录后再兑换邀请"))
		return
	}
	var req struct {
		Code string `json:"code"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	code := strings.TrimSpace(req.Code)
	if !strings.HasPrefix(code, workspaceInvitePrefix) {
		middleware.HandleError(w, apperr.BadRequest("邀请码须以 ymw_ 开头（与队席 ymt_ 隔离）"))
		return
	}
	hash := hashWSInviteToken(code)
	var id, wsID, role string
	var email, expires, revoked sql.NullString
	var maxUses sql.NullInt64
	var useCount int
	err := h.DB.QueryRow(`
		SELECT id, workspace_id, email, role, max_uses, use_count, expires_at, revoked_at
		FROM workspace_invites WHERE token_hash=?`, hash).
		Scan(&id, &wsID, &email, &role, &maxUses, &useCount, &expires, &revoked)
	if err == sql.ErrNoRows {
		middleware.HandleError(w, apperr.NotFound("邀请码无效"))
		return
	}
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	status := inviteStatus(revoked, expires, maxUses, useCount)
	if status != "active" {
		msg := "邀请已不可用"
		switch status {
		case "revoked":
			msg = "邀请已吊销"
		case "expired":
			msg = "邀请已过期"
		case "exhausted":
			msg = "邀请已达使用上限"
		}
		middleware.HandleError(w, apperr.BadRequest(msg))
		return
	}
	if err := middleware.EnsureWorkspaceActive(h.DB, wsID); err != nil {
		middleware.HandleError(w, err)
		return
	}
	var userEmail, userName string
	err = h.DB.QueryRow(`SELECT email, name FROM users WHERE id=?`, uid).Scan(&userEmail, &userName)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	if email.Valid && email.String != "" && !strings.EqualFold(email.String, userEmail) {
		middleware.HandleError(w, apperr.Forbidden("该邀请仅限指定邮箱兑换"))
		return
	}
	var existing string
	err = h.DB.QueryRow(`SELECT id FROM members WHERE workspace_id=? AND user_id=?`, wsID, uid).Scan(&existing)
	if err == nil {
		middleware.HandleError(w, apperr.Conflict("您已是该工作空间成员"))
		return
	}
	if err != nil && err != sql.ErrNoRows {
		middleware.HandleError(w, err)
		return
	}
	memID := idgen.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	tx, err := h.DB.Begin()
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	defer tx.Rollback()
	_, err = tx.Exec(`INSERT INTO members (id, workspace_id, user_id, role, display_name, is_bot, bot_enabled, created_at)
		VALUES (?,?,?,?,?,0,1,?)`, memID, wsID, uid, role, userName, now)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	_, err = tx.Exec(`UPDATE workspace_invites SET use_count = use_count + 1, updated_at=? WHERE id=?`, now, id)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	if err := tx.Commit(); err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{
		"workspace_id": wsID,
		"member": memberDTO{
			ID: memID, WorkspaceID: wsID, UserID: &uid, Role: role,
			DisplayName: userName, Email: &userEmail, CreatedAt: now,
		},
	})
}
