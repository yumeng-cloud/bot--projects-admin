package handlers

import (
	"database/sql"
	"net/http"
	"os"
	"strconv"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/embedder"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/botmem"
)

type BotsHandler struct {
	DB  *sql.DB
	Svc *botmem.Service
}

func NewBotsHandler(sqlDB *sql.DB) *BotsHandler {
	svc := botmem.New(sqlDB)
	// 向量写入/语义检索仅 Postgres+pgvector；SQLite 不注入（见 docs/embed.md）。
	if db.IsPostgres(os.Getenv("DB_DRIVER")) {
		svc.WithEmbedder(embedder.MustNew(embedder.LoadConfig()), embedder.NewStore(sqlDB))
	}
	return &BotsHandler{DB: sqlDB, Svc: svc}
}

// MountBots 挂载 Bot 角色 / 记忆 / 待办 / handoff。调用方用 AuthJWTOrAPIKey。
func MountBots(mux *http.ServeMux, db *sql.DB, wrap func(http.Handler) http.Handler) {
	h := NewBotsHandler(db)
	if wrap == nil {
		wrap = func(next http.Handler) http.Handler { return next }
	}
	reg := func(pattern string, fn http.HandlerFunc) {
		mux.Handle(pattern, wrap(fn))
	}

	reg("GET /api/v1/workspaces/{workspaceID}/bot-roles", h.ListRoles)
	reg("POST /api/v1/workspaces/{workspaceID}/bot-roles", h.Idempotent(h.CreateRole))
	reg("POST /api/v1/workspaces/{workspaceID}/bot-roles/{roleID}/claim", h.Idempotent(h.ClaimRole))
	reg("POST /api/v1/workspaces/{workspaceID}/bot-roles/{roleID}/release", h.Idempotent(h.ReleaseRole))

	reg("GET /api/v1/workspaces/{workspaceID}/bots", h.Roster)
	// /bots/me 须在 /bots/{memberID}/... 之前（精确路径优先）
	reg("GET /api/v1/workspaces/{workspaceID}/bots/me", h.Me)

	reg("GET /api/v1/workspaces/{workspaceID}/bots/{memberID}/memories", h.ListMemories)
	reg("POST /api/v1/workspaces/{workspaceID}/bots/{memberID}/memories", h.Idempotent(h.CreateMemory))
	reg("PATCH /api/v1/workspaces/{workspaceID}/bots/{memberID}/memories/{memoryID}", h.Idempotent(h.PatchMemory))
	reg("DELETE /api/v1/workspaces/{workspaceID}/bots/{memberID}/memories/{memoryID}", h.DeleteMemory)

	reg("GET /api/v1/workspaces/{workspaceID}/bots/{memberID}/todos", h.ListTodos)
	reg("POST /api/v1/workspaces/{workspaceID}/bots/{memberID}/todos", h.Idempotent(h.CreateTodo))
	reg("PATCH /api/v1/workspaces/{workspaceID}/bots/{memberID}/todos/{todoID}", h.Idempotent(h.PatchTodo))

	reg("GET /api/v1/workspaces/{workspaceID}/bots/{memberID}/handoff", h.Handoff)
	reg("POST /api/v1/workspaces/{workspaceID}/bots/{memberID}/handoff", h.Idempotent(h.ImportHandoff))
}

func (h *BotsHandler) Idempotent(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ws := r.PathValue("workspaceID")
		withIdempotency(h.DB, ws, fn)(w, r)
	}
}

func (h *BotsHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	list, err := h.Svc.ListRoles(ws)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"bot_roles": list})
}

func (h *BotsHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	var req struct {
		Name              string `json:"name"`
		Slug              string `json:"slug"`
		Responsibilities  string `json:"responsibilities"`
		Skills            string `json:"skills"`
		DefaultProjectIDs string `json:"default_project_ids"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	role, err := h.Svc.CreateRole(ws, botmem.CreateRoleInput{
		Name: req.Name, Slug: req.Slug, Responsibilities: req.Responsibilities,
		Skills: req.Skills, DefaultProjectIDs: req.DefaultProjectIDs,
	})
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusCreated, map[string]any{"bot_role": role})
}

func (h *BotsHandler) ClaimRole(w http.ResponseWriter, r *http.Request) {
	ws, actorMember, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	var req struct {
		MemberID string `json:"member_id"`
	}
	_ = middleware.DecodeJSON(r, &req)
	memberID := req.MemberID
	if memberID == "" {
		memberID = actorMember
	}
	claim, err := h.Svc.Claim(ws, r.PathValue("roleID"), memberID)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"claim": claim})
}

func (h *BotsHandler) ReleaseRole(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	if err := h.Svc.Release(ws, r.PathValue("roleID")); err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "已释放"})
}

func (h *BotsHandler) Roster(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	list, err := h.Svc.Roster(ws)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"bots": list})
}

// Me 返回当前鉴权主体在本工作区的成员身份（Bot API Key 或人类 JWT）。
// 供 MCP/CLI whoami 使用，避免 Bot 调 JWT-only /auth/me。
func (h *BotsHandler) Me(w http.ResponseWriter, r *http.Request) {
	ws, memberID, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	if memberID == "" {
		middleware.HandleError(w, apperr.BadRequest("缺少成员上下文"))
		return
	}

	var displayName, role string
	var isBot int
	err := h.DB.QueryRow(`
		SELECT display_name, role, is_bot FROM members
		WHERE id = ? AND workspace_id = ?`, memberID, ws).Scan(&displayName, &role, &isBot)
	if err == sql.ErrNoRows {
		middleware.HandleError(w, apperr.Forbidden("成员不存在或不属于该工作空间"))
		return
	}
	if err != nil {
		middleware.HandleError(w, err)
		return
	}

	actorKind := middleware.ActorKind(r.Context())
	if actorKind == "" {
		if isBot == 1 {
			actorKind = "bot"
		} else {
			actorKind = "human"
		}
	}
	if ctxRole := middleware.MemberRole(r.Context()); ctxRole != "" {
		role = ctxRole
	}

	middleware.JSON(w, http.StatusOK, map[string]any{
		"workspace_id": ws,
		"member_id":    memberID,
		"actor_kind":   actorKind,
		"role":         role,
		"display_name": displayName,
	})
}

func (h *BotsHandler) ListMemories(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	q := r.URL.Query().Get("q")
	mode := r.URL.Query().Get("mode")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	memberID := r.PathValue("memberID")

	// 无 mode：兼容旧 FE，仅返回 {memories}
	if mode == "" {
		list, err := h.Svc.ListMemories(ws, memberID, q)
		if err != nil {
			middleware.HandleError(w, err)
			return
		}
		middleware.JSON(w, http.StatusOK, map[string]any{"memories": list})
		return
	}

	res, err := h.Svc.SearchMemories(r.Context(), ws, memberID, q, botmem.SearchMode(mode), limit)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, res)
}

func (h *BotsHandler) CreateMemory(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	var req struct {
		KeyName   string  `json:"key_name"`
		ValueText string  `json:"value_text"`
		Tags      string  `json:"tags"`
		RoleID    *string `json:"role_id"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	m, err := h.Svc.UpsertMemory(ws, r.PathValue("memberID"), botmem.UpsertMemoryInput{
		KeyName: req.KeyName, ValueText: req.ValueText, Tags: req.Tags, RoleID: req.RoleID,
	})
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusCreated, map[string]any{"memory": m})
}

func (h *BotsHandler) PatchMemory(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	var req struct {
		KeyName   *string `json:"key_name"`
		ValueText *string `json:"value_text"`
		Tags      *string `json:"tags"`
		RoleID    *string `json:"role_id"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	m, err := h.Svc.PatchMemory(ws, r.PathValue("memberID"), r.PathValue("memoryID"), botmem.PatchMemoryInput{
		KeyName: req.KeyName, ValueText: req.ValueText, Tags: req.Tags, RoleID: req.RoleID,
	})
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"memory": m})
}

func (h *BotsHandler) DeleteMemory(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	if err := h.Svc.DeleteMemory(ws, r.PathValue("memberID"), r.PathValue("memoryID")); err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "已删除"})
}

func (h *BotsHandler) ListTodos(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	list, err := h.Svc.ListTodos(ws, r.PathValue("memberID"), r.URL.Query().Get("status"))
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"todos": list})
}

func (h *BotsHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	var req struct {
		Title     string  `json:"title"`
		Status    string  `json:"status"`
		RoleID    *string `json:"role_id"`
		IssueID   *string `json:"issue_id"`
		SortOrder float64 `json:"sort_order"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	t, err := h.Svc.CreateTodo(ws, r.PathValue("memberID"), botmem.CreateTodoInput{
		Title: req.Title, Status: req.Status, RoleID: req.RoleID, IssueID: req.IssueID, SortOrder: req.SortOrder,
	})
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusCreated, map[string]any{"todo": t})
}

func (h *BotsHandler) PatchTodo(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	var req struct {
		Title     *string  `json:"title"`
		Status    *string  `json:"status"`
		RoleID    *string  `json:"role_id"`
		IssueID   *string  `json:"issue_id"`
		SortOrder *float64 `json:"sort_order"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	t, err := h.Svc.PatchTodo(ws, r.PathValue("memberID"), r.PathValue("todoID"), botmem.PatchTodoInput{
		Title: req.Title, Status: req.Status, RoleID: req.RoleID, IssueID: req.IssueID, SortOrder: req.SortOrder,
	})
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"todo": t})
}

func (h *BotsHandler) Handoff(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	hdo, err := h.Svc.Handoff(ws, r.PathValue("memberID"))
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"handoff": hdo})
}

func (h *BotsHandler) ImportHandoff(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	var req botmem.ImportHandoffInput
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	res, err := h.Svc.ImportHandoff(ws, r.PathValue("memberID"), req)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, res)
}
