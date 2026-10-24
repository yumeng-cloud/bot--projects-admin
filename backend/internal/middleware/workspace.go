package middleware

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
)

// RequireWorkspaceMember resolves workspace from chi/std path param "workspaceID"
// or header X-Workspace-Id, verifies membership, injects MemberID/WorkspaceID/MemberRole.
// Bots already have WorkspaceID from API key; still verifies path matches.
// Disabled workspaces (disabled_at set) → 403「工作区已停用」for all members/bots.
func RequireWorkspaceMember(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wsID := r.PathValue("workspaceID")
			if wsID == "" {
				wsID = r.Header.Get("X-Workspace-Id")
			}
			if wsID == "" {
				writeErr(w, apperr.BadRequest("缺少工作空间 ID"))
				return
			}

			if err := EnsureWorkspaceActive(db, wsID); err != nil {
				if ae, ok := err.(*apperr.AppError); ok {
					writeErr(w, ae)
				} else {
					writeErr(w, apperr.Internal("校验工作空间失败"))
				}
				return
			}

			// Bot already scoped
			if ActorKind(r.Context()) == "bot" {
				bound := WorkspaceID(r.Context())
				if bound != "" && bound != wsID {
					writeErr(w, apperr.Forbidden("API Key 不属于该工作空间"))
					return
				}
				ctx := context.WithValue(r.Context(), WorkspaceIDKey, wsID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			uid := UserID(r.Context())
			var memberID, role string
			err := db.QueryRow(`
				SELECT id, role FROM members
				WHERE workspace_id = ? AND user_id = ?`, wsID, uid).Scan(&memberID, &role)
			if err == sql.ErrNoRows {
				writeErr(w, apperr.Forbidden("您不是该工作空间的成员"))
				return
			}
			if err != nil {
				writeErr(w, apperr.Internal("校验成员失败"))
				return
			}
			ctx := context.WithValue(r.Context(), WorkspaceIDKey, wsID)
			ctx = context.WithValue(ctx, MemberIDKey, memberID)
			ctx = context.WithValue(ctx, MemberRoleKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// EnsureWorkspaceActive returns 403 if the workspace is disabled or missing.
func EnsureWorkspaceActive(db *sql.DB, wsID string) error {
	var disabled sql.NullString
	err := db.QueryRow(`SELECT disabled_at FROM workspaces WHERE id = ?`, wsID).Scan(&disabled)
	if err == sql.ErrNoRows {
		return apperr.NotFound("工作空间不存在")
	}
	if err != nil {
		return err
	}
	if disabled.Valid && disabled.String != "" {
		return apperr.Forbidden("工作区已停用")
	}
	return nil
}

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if MemberRole(r.Context()) != "admin" {
			writeErr(w, apperr.Forbidden("需要管理员权限"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequirePlatformAdmin ensures the JWT user has users.is_platform_admin = 1.
func RequirePlatformAdmin(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			uid := UserID(r.Context())
			if uid == "" {
				writeErr(w, apperr.Unauthorized("请先登录"))
				return
			}
			var flag int
			err := db.QueryRow(`SELECT is_platform_admin FROM users WHERE id = ?`, uid).Scan(&flag)
			if err == sql.ErrNoRows {
				writeErr(w, apperr.Unauthorized("用户不存在"))
				return
			}
			if err != nil {
				writeErr(w, apperr.Internal("校验平台管理员失败"))
				return
			}
			if flag != 1 {
				writeErr(w, apperr.Forbidden("需要平台超级管理员权限"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
