package handlers

import (
	"database/sql"
	"net/http"
)

// RegisterWikiAndTime 一次性挂载 Wiki + 工时/计费路由。
// wrap 通常为 AuthJWT / AuthJWTOrAPIKey；可为 nil。
func RegisterWikiAndTime(mux *http.ServeMux, db *sql.DB, wrap func(http.Handler) http.Handler) {
	MountWiki(mux, db, wrap)
	MountTime(mux, db, wrap)
}

// RegisterProjects 挂载工作空间下项目 CRUD。
// wrap 须注入 WorkspaceID（建议 Auth* + RequireWorkspaceMember）。
func RegisterProjects(mux *http.ServeMux, db *sql.DB, wrap func(http.Handler) http.Handler) {
	if wrap == nil {
		wrap = func(next http.Handler) http.Handler { return next }
	}
	h := &ProjectHandler{DB: db}
	reg := func(pattern string, fn http.HandlerFunc) {
		mux.Handle(pattern, wrap(fn))
	}
	reg("GET /api/v1/workspaces/{workspaceID}/projects", h.List)
	reg("POST /api/v1/workspaces/{workspaceID}/projects", h.Create)
	reg("GET /api/v1/workspaces/{workspaceID}/projects/{projectID}", h.Get)
	reg("PATCH /api/v1/workspaces/{workspaceID}/projects/{projectID}", h.Patch)
	reg("DELETE /api/v1/workspaces/{workspaceID}/projects/{projectID}", h.Delete)
}

// RegisterIntegrations 挂载 API Key + Webhook。
func RegisterIntegrations(mux *http.ServeMux, db *sql.DB, wrap func(http.Handler) http.Handler) {
	if wrap == nil {
		wrap = func(next http.Handler) http.Handler { return next }
	}
	keys := &APIKeyHandler{DB: db}
	hooks := &WebhookHandler{DB: db}
	reg := func(pattern string, fn http.HandlerFunc) {
		mux.Handle(pattern, wrap(fn))
	}
	reg("GET /api/v1/workspaces/{workspaceID}/api-keys", keys.List)
	reg("POST /api/v1/workspaces/{workspaceID}/api-keys", keys.Create)
	reg("DELETE /api/v1/workspaces/{workspaceID}/api-keys/{keyID}", keys.Revoke)
	reg("POST /api/v1/workspaces/{workspaceID}/api-keys/{keyID}/revoke", keys.Revoke)

	reg("GET /api/v1/workspaces/{workspaceID}/webhooks", hooks.List)
	reg("POST /api/v1/workspaces/{workspaceID}/webhooks", hooks.Create)
	reg("PATCH /api/v1/workspaces/{workspaceID}/webhooks/{webhookID}", hooks.Patch)
	reg("DELETE /api/v1/workspaces/{workspaceID}/webhooks/{webhookID}", hooks.Delete)
	reg("GET /api/v1/workspaces/{workspaceID}/webhooks/{webhookID}/deliveries", hooks.ListDeliveries)
}
