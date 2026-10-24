package handlers

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
)

// chainAuth 组合外层鉴权 wrap；write=true 时叠加 Idempotency（依赖上下文 WorkspaceID）。
func chainAuth(db *sql.DB, wrap func(http.Handler) http.Handler, write bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		h := next
		if write {
			h = middleware.Idempotency(db)(h)
		}
		if wrap != nil {
			h = wrap(h)
		}
		return h
	}
}

// withIdempotency 按路径 workspace 做幂等。
func withIdempotency(db *sql.DB, workspaceID string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if workspaceID != "" && middleware.WorkspaceID(ctx) == "" {
			ctx = context.WithValue(ctx, middleware.WorkspaceIDKey, workspaceID)
		}
		middleware.Idempotency(db)(next).ServeHTTP(w, r.WithContext(ctx))
	}
}
