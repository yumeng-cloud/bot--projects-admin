package server

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/config"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/handlers"
	mw "github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/modules"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/ratelimit"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/issueboard"
)

// NewRouter builds the HTTP handler used by cmd/server and integration tests.
func NewRouter(sqlDB *sql.DB, cfg config.Config) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.Recoverer)
	r.Use(mw.CORS(cfg.CORSOrigins))

	health := &handlers.HealthHandler{DB: sqlDB}
	authH := &handlers.AuthHandler{
		DB:      sqlDB,
		Cfg:     cfg,
		Limiter: ratelimit.New(cfg.LoginRateLimit, cfg.LoginRateWindow),
	}
	wsH := &handlers.WorkspaceHandler{DB: sqlDB}
	projH := &handlers.ProjectHandler{DB: sqlDB}
	memH := &handlers.MemberHandler{DB: sqlDB}
	adminH := &handlers.AdminHandler{DB: sqlDB}
	wsInvH := &handlers.WorkspaceInviteHandler{DB: sqlDB}
	labelH := &handlers.LabelHandler{DB: sqlDB}
	keyH := &handlers.APIKeyHandler{DB: sqlDB}
	hookH := &handlers.WebhookHandler{DB: sqlDB}
	boardSvc := issueboard.New(sqlDB)
	boardSvc.DispatchEvent = hookH.DispatchEvent
	issueH := &handlers.IssueHandler{Svc: boardSvc}
	cycleH := &handlers.CycleHandler{Svc: boardSvc}
	searchH := &handlers.SearchHandler{Svc: boardSvc}
	relH := &handlers.RelationHandler{Svc: boardSvc}
	commentH := &handlers.CommentHandler{Svc: boardSvc}
	activityH := &handlers.ActivityHandler{Svc: boardSvc}
	attachH := handlers.NewAttachmentHandler(sqlDB, cfg.AttachmentsDir)

	jwtOnly := mw.AuthJWT(mw.JWTParser{Secret: cfg.JWTSecret})
	authAny := mw.AuthJWTOrAPIKey(mw.JWTParser{Secret: cfg.JWTSecret}, sqlDB)
	wsMem := mw.RequireWorkspaceMember(sqlDB)
	platformAdmin := mw.RequirePlatformAdmin(sqlDB)
	idem := mw.Idempotency(sqlDB)

	r.Get("/healthz", health.Healthz)
	r.Get("/api/v1/healthz", health.Healthz)

	r.Route("/api/v1", func(api chi.Router) {
		api.Post("/auth/register", authH.Register)
		api.Post("/auth/login", authH.Login)
		api.Post("/auth/logout", authH.Logout)
		api.With(jwtOnly).Get("/auth/me", authH.Me)

		api.With(jwtOnly).Get("/workspaces", wsH.ListMine)
		api.With(jwtOnly).Post("/workspaces", wsH.Create)
		api.With(jwtOnly).Get("/workspaces/current", wsH.Current)
		api.With(jwtOnly).Post("/workspace-invites/redeem", wsInvH.Redeem)

		// Platform super-admin (no WS membership required)
		api.Route("/admin", func(adm chi.Router) {
			adm.Use(jwtOnly)
			adm.Use(platformAdmin)
			adm.Get("/workspaces", adminH.ListWorkspaces)
			adm.Post("/workspaces", adminH.CreateWorkspace)
			adm.Patch("/workspaces/{id}", adminH.PatchWorkspace)
			adm.Get("/users", adminH.ListUsers)
			adm.Patch("/users/{id}", adminH.PatchUser)
		})

		api.Route("/workspaces/{workspaceID}", func(ws chi.Router) {
			ws.Use(authAny)
			ws.Use(wsMem)
			ws.Use(idem)

			ws.Get("/", wsH.Get)
			ws.With(mw.RequireAdmin).Patch("/", wsH.Patch)

			ws.Get("/members", memH.List)
			ws.With(mw.RequireAdmin).Post("/members", memH.Invite)
			ws.With(mw.RequireAdmin).Patch("/members/{memberID}", memH.Patch)
			ws.With(mw.RequireAdmin).Delete("/members/{memberID}", memH.Remove)

			ws.With(mw.RequireAdmin).Get("/invites", wsInvH.List)
			ws.With(mw.RequireAdmin).Post("/invites", wsInvH.Create)
			ws.With(mw.RequireAdmin).Post("/invites/{inviteID}/revoke", wsInvH.Revoke)

			ws.Get("/labels", labelH.List)
			ws.Post("/labels", labelH.Create)
			ws.Patch("/labels/{labelID}", labelH.Patch)
			ws.Delete("/labels/{labelID}", labelH.Delete)

			ws.Get("/projects", projH.List)
			ws.Post("/projects", projH.Create)

			ws.Get("/projects/{projectID}/board-column-prefs", projH.GetBoardColumnPrefs)
			ws.Put("/projects/{projectID}/board-column-prefs", projH.PutBoardColumnPrefs)
			ws.Delete("/projects/{projectID}/board-column-prefs", projH.DeleteBoardColumnPrefs)

			ws.With(mw.RequireAdmin).Get("/api-keys", keyH.List)
			ws.With(mw.RequireAdmin).Post("/api-keys", keyH.Create)

			ws.With(mw.RequireAdmin).Get("/webhooks", hookH.List)
			ws.With(mw.RequireAdmin).Post("/webhooks", hookH.Create)
			ws.With(mw.RequireAdmin).Patch("/webhooks/{webhookID}", hookH.Patch)
			ws.With(mw.RequireAdmin).Delete("/webhooks/{webhookID}", hookH.Delete)
			ws.With(mw.RequireAdmin).Get("/webhooks/{webhookID}/deliveries", hookH.ListDeliveries)

			// Issue 域由 modules.RegisterIssues 挂载；未挂前 501
			ws.Get("/issues", issueH.List)
			ws.Post("/issues", issueH.Create)
			ws.Get("/issues/{issueID}", issueH.Get)
			ws.Patch("/issues/{issueID}", issueH.Patch)
			ws.Delete("/issues/{issueID}", issueH.Delete)
			ws.Post("/issues/{issueID}/notify", hookH.NotifyIssue)

			ws.Get("/cycles", cycleH.List)
			ws.Get("/projects/{projectID}/cycles", cycleH.List)
			ws.Post("/projects/{projectID}/cycles", cycleH.Create)
			ws.Get("/cycles/{cycleID}", cycleH.Get)
			ws.Patch("/cycles/{cycleID}", cycleH.Patch)

			ws.Get("/search", searchH.Search)

			ws.Get("/issues/{issueID}/relations", relH.List)
			ws.Post("/issues/{issueID}/relations", relH.Create)
			ws.Delete("/issues/{issueID}/relations", relH.DeleteCollection)
			ws.Delete("/issues/{issueID}/relations/{relationID}", relH.DeleteOnIssue)
			ws.Delete("/relations/{relationID}", relH.Delete)

			ws.Get("/issues/{issueID}/comments", commentH.List)
			ws.Post("/issues/{issueID}/comments", commentH.Create)
			ws.Patch("/issues/{issueID}/comments/{commentID}", commentH.Patch)
			ws.Delete("/issues/{issueID}/comments/{commentID}", commentH.Delete)

			ws.Get("/issues/{issueID}/activities", activityH.ListByIssue)
			ws.Get("/activities", activityH.List)

			// Attachments: upload + meta + content（Go）; List×2 + Delete（Issue 域）
			ws.Post("/issues/{issueID}/attachments", attachH.UploadIssue)
			ws.Post("/issues/{issueID}/comments/{commentID}/attachments", attachH.UploadComment)
			ws.Get("/issues/{issueID}/attachments", attachH.ListByIssue)
			ws.Get("/issues/{issueID}/comments/{commentID}/attachments", attachH.ListByComment)
			ws.Get("/attachments/{attachmentID}", attachH.GetMeta)
			ws.Get("/attachments/{attachmentID}/content", attachH.GetContent)
			ws.Delete("/attachments/{attachmentID}", attachH.Delete)
		})

		api.With(authAny).Get("/projects/{projectID}", projH.Get)
		api.With(authAny).Patch("/projects/{projectID}", projH.Patch)
		api.With(authAny).Delete("/projects/{projectID}", projH.Delete)

		api.With(authAny).Post("/api-keys/{keyID}/revoke", keyH.Revoke)
		api.With(authAny).Delete("/api-keys/{keyID}", keyH.Revoke)

		deps := modules.Deps{DB: sqlDB, AuthMW: authAny, WSMemMW: wsMem}
		adapter := chiMux{r: api}
		// wiki/time/bot 走下方 Mount*（std ServeMux METHOD 路由）；此处只挂 Issue 域
		for _, fn := range []modules.RegisterFunc{modules.RegisterIssues, modules.RegisterCollab} {
			if fn != nil {
				fn(adapter, deps)
			}
		}
	})

	root := http.NewServeMux()
	handlers.MountWiki(root, sqlDB, authAny)
	handlers.MountTime(root, sqlDB, authAny)
	handlers.MountBots(root, sqlDB, authAny)
	handlers.MountTeams(root, sqlDB, authAny, hookH.DispatchEvent)
	root.Handle("/", r)
	return root
}

type chiMux struct{ r chi.Router }

func (c chiMux) Handle(pattern string, handler http.Handler) { c.r.Handle(pattern, handler) }
func (c chiMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	c.r.HandleFunc(pattern, handler)
}

func stub501(r chi.Router, method, path string) {
	h := func(w http.ResponseWriter, _ *http.Request) {
		mw.JSON(w, http.StatusNotImplemented, map[string]string{
			"code":    "not_implemented",
			"message": "该路由已预留，由对应模块负责人实现，见 COORDINATION.md",
		})
	}
	r.Method(method, path, http.HandlerFunc(h))
}
