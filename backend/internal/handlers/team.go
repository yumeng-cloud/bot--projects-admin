package handlers

import (
	"database/sql"
	"net/http"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/botmem"
)

type TeamHandler struct {
	DB  *sql.DB
	Svc *botmem.Service
}

func NewTeamHandler(sqlDB *sql.DB) *TeamHandler {
	return &TeamHandler{DB: sqlDB, Svc: botmem.New(sqlDB)}
}

// resolveBindMember：Bot → 用其 member；人类须 body member_id 或管理员。
func (h *TeamHandler) resolveBindMember(w http.ResponseWriter, r *http.Request, ws, actorMember, bodyMemberID string) (string, bool) {
	if middleware.ActorKind(r.Context()) == "bot" {
		if actorMember == "" {
			middleware.HandleError(w, apperr.BadRequest("缺少成员上下文"))
			return "", false
		}
		return actorMember, true
	}
	mid := bodyMemberID
	if mid == "" {
		mid = actorMember
	}
	if mid == "" {
		middleware.HandleError(w, apperr.BadRequest("member_id 不能为空"))
		return "", false
	}
	if mid != actorMember && middleware.MemberRole(r.Context()) != "admin" {
		middleware.HandleError(w, apperr.Forbidden("需要管理员权限才能代绑其他成员"))
		return "", false
	}
	return mid, true
}

func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	var req struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		TemplateKey string `json:"template_key"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	team, err := h.Svc.CreateTeam(ws, botmem.CreateTeamInput{
		ID: req.ID, Name: req.Name, Slug: req.Slug, TemplateKey: req.TemplateKey,
	})
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusCreated, map[string]any{"team": team})
}

func (h *TeamHandler) ListTeams(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	list, err := h.Svc.ListTeams(ws)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"teams": list})
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	team, err := h.Svc.GetTeam(ws, r.PathValue("teamID"))
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"team": team})
}

func (h *TeamHandler) GetRoster(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	roster, err := h.Svc.GetRoster(ws, r.PathValue("teamID"))
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"roster": roster})
}

func (h *TeamHandler) EnsureRoster(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	var req struct {
		TemplateKey string `json:"template_key"`
		Seats       []struct {
			Slug        string `json:"slug"`
			Name        string `json:"name"`
			PlaybookRef string `json:"playbook_ref"`
		} `json:"seats"`
	}
	_ = middleware.DecodeJSON(r, &req)
	in := botmem.EnsureRosterInput{TemplateKey: req.TemplateKey}
	for _, s := range req.Seats {
		in.Seats = append(in.Seats, botmem.EnsureSeatInput{Slug: s.Slug, Name: s.Name, PlaybookRef: s.PlaybookRef})
	}
	roster, err := h.Svc.EnsureRoster(ws, r.PathValue("teamID"), in)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"roster": roster})
}

func (h *TeamHandler) BindSeat(w http.ResponseWriter, r *http.Request) {
	ws, actor, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	var req struct {
		MemberID string `json:"member_id"`
		Force    bool   `json:"force"`
	}
	_ = middleware.DecodeJSON(r, &req)
	memberID, ok := h.resolveBindMember(w, r, ws, actor, req.MemberID)
	if !ok {
		return
	}
	seat, err := h.Svc.BindSeat(ws, r.PathValue("teamID"), r.PathValue("seatID"), memberID, req.Force)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"seat": seat})
}

func (h *TeamHandler) UnbindSeat(w http.ResponseWriter, r *http.Request) {
	ws, actor, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	var req struct {
		MemberID string `json:"member_id"`
	}
	_ = middleware.DecodeJSON(r, &req)
	asAdmin := middleware.MemberRole(r.Context()) == "admin"
	memberID := req.MemberID
	if memberID == "" {
		memberID = actor
	}
	if middleware.ActorKind(r.Context()) == "bot" {
		memberID = actor
		asAdmin = false
	}
	seat, err := h.Svc.UnbindSeat(ws, r.PathValue("teamID"), r.PathValue("seatID"), memberID, asAdmin)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"seat": seat})
}

func (h *TeamHandler) CreateSeat(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	var req struct {
		Slug        string `json:"slug"`
		RoleSlug    string `json:"role_slug"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		PlaybookRef string `json:"playbook_ref"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	slug := req.Slug
	if slug == "" {
		slug = req.RoleSlug
	}
	name := req.Name
	if name == "" {
		name = req.DisplayName
	}
	seat, err := h.Svc.CreateSeat(ws, r.PathValue("teamID"), botmem.EnsureSeatInput{
		Slug: slug, Name: name, PlaybookRef: req.PlaybookRef,
	})
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusCreated, map[string]any{"seat": seat})
}

func (h *TeamHandler) StopSeat(w http.ResponseWriter, r *http.Request) {
	ws, actor, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	if resolveMemberRole(h.DB, r, actor) != "admin" && middleware.ActorKind(r.Context()) != "bot" {
		middleware.HandleError(w, apperr.Forbidden("需要管理员权限才能标记停工"))
		return
	}
	seat, err := h.Svc.StopSeat(ws, r.PathValue("teamID"), r.PathValue("seatID"))
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"seat": seat})
}


func (h *TeamHandler) ListInvites(w http.ResponseWriter, r *http.Request) {
	ws, actor, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	if resolveMemberRole(h.DB, r, actor) != "admin" {
		middleware.HandleError(w, apperr.Forbidden("需要管理员权限才能查看入场码列表"))
		return
	}
	list, err := h.Svc.ListInvites(ws, r.PathValue("teamID"))
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"invites": list})
}

func (h *TeamHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	ws, actor, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	if resolveMemberRole(h.DB, r, actor) != "admin" {
		middleware.HandleError(w, apperr.Forbidden("需要管理员权限才能签发入场码"))
		return
	}
	var req struct {
		AllowExpandSeat bool   `json:"allow_expand_seat"`
		SeatSlug        string `json:"seat_slug"`
		RoleHint        string `json:"role_hint"`
		RoleSlug        string `json:"role_slug"`
		MaxUses         *int   `json:"max_uses"`
		ExpiresAt       string `json:"expires_at"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	hint := req.RoleHint
	if hint == "" {
		hint = req.RoleSlug
	}
	inv, err := h.Svc.CreateInvite(ws, r.PathValue("teamID"), botmem.CreateInviteInput{
		AllowExpandSeat: req.AllowExpandSeat,
		SeatSlug:        req.SeatSlug,
		RoleHint:        hint,
		MaxUses:         req.MaxUses,
		ExpiresAt:       req.ExpiresAt,
		CreatedBy:       actor,
	})
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusCreated, map[string]any{"invite": inv, "code": inv.Code})
}

func (h *TeamHandler) RedeemInvite(w http.ResponseWriter, r *http.Request) {
	// AuthJWTOrAPIKey 已校验身份；无需 workspace 路径
	var req struct {
		Code string `json:"code"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	res, err := h.Svc.RedeemInvite(req.Code)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, res)
}

func (h *TeamHandler) ExpandSeats(w http.ResponseWriter, r *http.Request) {
	ws, _, ok := requireWorkspaceMember(h.DB, w, r)
	if !ok {
		return
	}
	var req struct {
		ProjectID string `json:"project_id"`
		Limit     int    `json:"limit"`
	}
	_ = middleware.DecodeJSON(r, &req)
	res, err := h.Svc.ExpandSeats(ws, r.PathValue("teamID"), botmem.ExpandInput{
		ProjectID: req.ProjectID, Limit: req.Limit,
	})
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, res)
}

// MountTeams 挂载 teams roster / seat create·bind·stop / invites / expand。auth 同 bots。
// onIssueEvent 可选：ExpandSeats 改派时扇出 Webhook（通常传 WebhookHandler.DispatchEvent）。
func MountTeams(mux *http.ServeMux, db *sql.DB, wrap func(http.Handler) http.Handler, onIssueEvent ...func(workspaceID, event string, payload any)) {
	h := NewTeamHandler(db)
	if len(onIssueEvent) > 0 && onIssueEvent[0] != nil {
		h.Svc.IssueDispatch = onIssueEvent[0]
	}
	if wrap == nil {
		wrap = func(next http.Handler) http.Handler { return next }
	}
	reg := func(pattern string, fn http.HandlerFunc) {
		mux.Handle(pattern, wrap(fn))
	}
	reg("POST /api/v1/workspaces/{workspaceID}/teams", h.Idempotent(h.CreateTeam))
	reg("GET /api/v1/workspaces/{workspaceID}/teams", h.ListTeams)
	reg("GET /api/v1/workspaces/{workspaceID}/teams/{teamID}", h.GetTeam)
	reg("GET /api/v1/workspaces/{workspaceID}/teams/{teamID}/roster", h.GetRoster)
	reg("POST /api/v1/workspaces/{workspaceID}/teams/{teamID}/roster/ensure", h.Idempotent(h.EnsureRoster))
	reg("POST /api/v1/workspaces/{workspaceID}/teams/{teamID}/seats", h.Idempotent(h.CreateSeat))
	reg("POST /api/v1/workspaces/{workspaceID}/teams/{teamID}/seats/{seatID}/bind", h.Idempotent(h.BindSeat))
	reg("POST /api/v1/workspaces/{workspaceID}/teams/{teamID}/seats/{seatID}/unbind", h.Idempotent(h.UnbindSeat))
	reg("POST /api/v1/workspaces/{workspaceID}/teams/{teamID}/seats/{seatID}/stop", h.Idempotent(h.StopSeat))
	reg("POST /api/v1/workspaces/{workspaceID}/teams/{teamID}/seats/expand", h.Idempotent(h.ExpandSeats))
	reg("POST /api/v1/workspaces/{workspaceID}/teams/{teamID}/expand", h.Idempotent(h.ExpandSeats))
	reg("GET /api/v1/workspaces/{workspaceID}/teams/{teamID}/invites", h.ListInvites)
	reg("POST /api/v1/workspaces/{workspaceID}/teams/{teamID}/invites", h.Idempotent(h.CreateInvite))
	// G3 redeem：不在 workspace 路径下（Bot 持 ymk_ 定位团队）
	reg("POST /api/v1/team-invites/redeem", h.RedeemInvite)
}

func (h *TeamHandler) Idempotent(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ws := r.PathValue("workspaceID")
		withIdempotency(h.DB, ws, fn)(w, r)
	}
}
