package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/embedder"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/wiki"
)

type WikiHandler struct {
	DB  *sql.DB
	Svc *wiki.Service
}

func NewWikiHandler(db *sql.DB) *WikiHandler {
	svc := wiki.New(db)
	// 向量仅 Postgres+pgvector；与 Bot 记忆共用 embedder/Store。
	if d := strings.ToLower(os.Getenv("DB_DRIVER")); d == "postgres" || d == "postgresql" || d == "pgx" {
		cfg := embedder.LoadConfig()
		e := embedder.MustNew(cfg)
		log.Printf("wiki embedder: provider=%s model=%s dims=%d", cfg.ProviderName(), e.ModelID(), e.Dims())
		svc.WithEmbedder(e, embedder.NewStore(db))
	}
	return &WikiHandler{DB: db, Svc: svc}
}

// MountWiki 将 Wiki 路由挂到 mux。wrap 为鉴权中间件（JWT / API Key）。
func MountWiki(mux *http.ServeMux, db *sql.DB, wrap func(http.Handler) http.Handler) {
	h := NewWikiHandler(db)
	read := chainAuth(db, wrap, false)
	write := chainAuth(db, wrap, true)
	reg := func(chain func(http.Handler) http.Handler, pattern string, fn http.HandlerFunc) {
		mux.Handle(pattern, chain(fn))
	}

	reg(read, "GET /api/v1/workspaces/{workspaceID}/wiki/spaces", h.ListSpaces)
	reg(write, "POST /api/v1/workspaces/{workspaceID}/wiki/spaces", h.CreateSpace)
	reg(read, "GET /api/v1/workspaces/{workspaceID}/wiki/spaces/{spaceID}", h.GetSpace)
	reg(write, "PATCH /api/v1/workspaces/{workspaceID}/wiki/spaces/{spaceID}", h.UpdateSpace)
	reg(write, "DELETE /api/v1/workspaces/{workspaceID}/wiki/spaces/{spaceID}", h.DeleteSpace)

	reg(read, "GET /api/v1/workspaces/{workspaceID}/wiki/spaces/{spaceID}/pages", h.ListPages)
	reg(write, "POST /api/v1/workspaces/{workspaceID}/wiki/spaces/{spaceID}/pages", h.CreatePage)

	reg(read, "GET /api/v1/workspaces/{workspaceID}/wiki/pages/{pageID}", h.GetPage)
	reg(write, "PATCH /api/v1/workspaces/{workspaceID}/wiki/pages/{pageID}", h.UpdatePage)
	reg(write, "DELETE /api/v1/workspaces/{workspaceID}/wiki/pages/{pageID}", h.DeletePage)

	reg(read, "GET /api/v1/workspaces/{workspaceID}/wiki/search", h.Search)
	reg(read, "GET /api/v1/workspaces/{workspaceID}/wiki/tree", h.Tree)
}

func (h *WikiHandler) ListSpaces(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	list, err := h.Svc.ListSpaces(ws)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"spaces": list})
}

func (h *WikiHandler) CreateSpace(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	var req struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	sp, err := h.Svc.CreateSpace(ws, wiki.CreateSpaceInput{Name: req.Name, Slug: req.Slug, Description: req.Description})
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusCreated, map[string]any{"space": sp})
}

func (h *WikiHandler) GetSpace(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	sp, err := h.Svc.GetSpace(ws, r.PathValue("spaceID"))
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"space": sp})
}

func (h *WikiHandler) UpdateSpace(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	var req struct {
		Name        *string `json:"name"`
		Slug        *string `json:"slug"`
		Description *string `json:"description"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	sp, err := h.Svc.UpdateSpace(ws, r.PathValue("spaceID"), wiki.UpdateSpaceInput{
		Name: req.Name, Slug: req.Slug, Description: req.Description,
	})
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"space": sp})
}

func (h *WikiHandler) DeleteSpace(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	if err := h.Svc.DeleteSpace(ws, r.PathValue("spaceID")); err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "已删除"})
}

func (h *WikiHandler) ListPages(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	list, err := h.Svc.ListPages(ws, r.PathValue("spaceID"))
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"pages": list})
}

func (h *WikiHandler) CreatePage(w http.ResponseWriter, r *http.Request) {
	ws, mid, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	var req struct {
		Title     string  `json:"title"`
		Slug      string  `json:"slug"`
		BodyMD    string  `json:"body_md"`
		ParentID  *string `json:"parent_id"`
		ProjectID *string `json:"project_id"`
		IssueID   *string `json:"issue_id"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	p, err := h.Svc.CreatePage(ws, r.PathValue("spaceID"), wiki.CreatePageInput{
		Title: req.Title, Slug: req.Slug, BodyMD: req.BodyMD,
		ParentID: req.ParentID, ProjectID: req.ProjectID, IssueID: req.IssueID, MemberID: mid,
	})
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusCreated, map[string]any{"page": p})
}

func (h *WikiHandler) GetPage(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	p, err := h.Svc.GetPage(ws, r.PathValue("pageID"))
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"page": p})
}

func (h *WikiHandler) UpdatePage(w http.ResponseWriter, r *http.Request) {
	ws, mid, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	var req struct {
		Title     *string `json:"title"`
		Slug      *string `json:"slug"`
		BodyMD    *string `json:"body_md"`
		ParentID  *string `json:"parent_id"`
		ProjectID *string `json:"project_id"`
		IssueID   *string `json:"issue_id"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	in := wiki.UpdatePageInput{
		Title: req.Title, Slug: req.Slug, BodyMD: req.BodyMD,
		ParentID: req.ParentID, ProjectID: req.ProjectID, IssueID: req.IssueID, MemberID: mid,
	}
	// 显式 null：JSON null 解码后指针非 nil 但值为空串时由 service 处理；
	// 若客户端传 "" 表示清除 parent。
	p, err := h.Svc.UpdatePage(ws, r.PathValue("pageID"), in)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"page": p})
}

func (h *WikiHandler) DeletePage(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	if err := h.Svc.DeletePage(ws, r.PathValue("pageID")); err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "已删除"})
}

func (h *WikiHandler) Search(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	q := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	mode := wiki.SearchMode(r.URL.Query().Get("mode"))
	if mode == "" {
		mode = wiki.ModeHybrid
	}
	res, err := h.Svc.SearchEx(r.Context(), ws, q, mode, limit)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, res)
}

func (h *WikiHandler) Tree(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	spaceID := r.URL.Query().Get("space_id")
	tree, err := h.Svc.Tree(ws, spaceID)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"tree": tree, "space_id": spaceID})
}
