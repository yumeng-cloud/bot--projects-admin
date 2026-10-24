package handlers

import (
	"database/sql"
	"net/http"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
)

// resolveMember 优先用上下文（RequireWorkspaceMember / API Key）；否则查 members。
func resolveMember(db *sql.DB, r *http.Request, workspaceID string) (memberID string, err error) {
	if mid := middleware.MemberID(r.Context()); mid != "" {
		return mid, nil
	}
	uid := middleware.UserID(r.Context())
	if uid == "" {
		return "", apperr.Unauthorized("请先登录")
	}
	err = db.QueryRow(`SELECT id FROM members WHERE workspace_id = ? AND user_id = ?`, workspaceID, uid).Scan(&memberID)
	if err == sql.ErrNoRows {
		return "", apperr.Forbidden("你不是该工作空间的成员")
	}
	if err != nil {
		return "", err
	}
	return memberID, nil
}

func requireWorkspaceMember(db *sql.DB, w http.ResponseWriter, r *http.Request) (workspaceID, memberID string, ok bool) {
	workspaceID = r.PathValue("workspaceID")
	if workspaceID == "" {
		workspaceID = middleware.WorkspaceID(r.Context())
	}
	if workspaceID == "" {
		middleware.HandleError(w, apperr.BadRequest("缺少 workspaceID"))
		return "", "", false
	}
	if err := middleware.EnsureWorkspaceActive(db, workspaceID); err != nil {
		middleware.HandleError(w, err)
		return "", "", false
	}
	memberID, err := resolveMember(db, r, workspaceID)
	if err != nil {
		middleware.HandleError(w, err)
		return "", "", false
	}
	return workspaceID, memberID, true
}

// resolveMemberRole 上下文优先，否则查 members.role（ServeMux teams 路由无 RequireWorkspaceMember）。
func resolveMemberRole(db *sql.DB, r *http.Request, memberID string) string {
	if role := middleware.MemberRole(r.Context()); role != "" {
		return role
	}
	if memberID == "" {
		return ""
	}
	var role string
	_ = db.QueryRow(`SELECT role FROM members WHERE id = ?`, memberID).Scan(&role)
	return role
}
