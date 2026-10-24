package models

// BotRole 协作角色定义（前端/PM/QA/后端/知识库等）。
type BotRole struct {
	ID                string `json:"id"`
	WorkspaceID       string `json:"workspace_id"`
	Name              string `json:"name"`
	Slug              string `json:"slug"`
	Responsibilities  string `json:"responsibilities"`
	Skills            string `json:"skills"`              // JSON array string
	DefaultProjectIDs string `json:"default_project_ids"` // JSON array string
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
	// 列表扩展：当前认领
	Claim *BotRoleClaim `json:"claim,omitempty"`
}

// BotRoleClaim 角色认领（同一 role 同时仅一个 active：released_at IS NULL）。
type BotRoleClaim struct {
	ID          string  `json:"id"`
	WorkspaceID string  `json:"workspace_id"`
	RoleID      string  `json:"role_id"`
	MemberID    string  `json:"member_id"`
	ClaimedAt   string  `json:"claimed_at"`
	ReleasedAt  *string `json:"released_at"`
}

// BotMemory 按 member+key 持久记忆（upsert）。
type BotMemory struct {
	ID          string  `json:"id"`
	WorkspaceID string  `json:"workspace_id"`
	MemberID    string  `json:"member_id"`
	RoleID      *string `json:"role_id"`
	KeyName     string  `json:"key_name"`
	ValueText   string  `json:"value_text"`
	Tags        string  `json:"tags"` // JSON array string
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// BotTodo Bot 待办。
type BotTodo struct {
	ID          string  `json:"id"`
	WorkspaceID string  `json:"workspace_id"`
	MemberID    string  `json:"member_id"`
	RoleID      *string `json:"role_id"`
	Title       string  `json:"title"`
	Status      string  `json:"status"` // todo|doing|done|canceled
	IssueID     *string `json:"issue_id"`
	SortOrder   float64 `json:"sort_order"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// BotRosterItem 名册一行：is_bot 成员 + 当前认领角色。
type BotRosterItem struct {
	MemberID    string   `json:"member_id"`
	DisplayName string   `json:"display_name"`
	BotEnabled  bool     `json:"bot_enabled"`
	Role        string   `json:"role"` // members.role
	ClaimedRole *BotRole `json:"claimed_role"`
}

// BotHandoff 只读续工包（由 todos + memories + 指派 issue 组装，无独立表）。
type BotHandoff struct {
	WorkspaceID string         `json:"workspace_id"`
	MemberID    string         `json:"member_id"`
	GeneratedAt string         `json:"generated_at"`
	ClaimedRole *BotRole       `json:"claimed_role"`
	Memories    []BotMemory    `json:"memories"`
	Todos       []BotTodo      `json:"todos"`
	Issues      []HandoffIssue `json:"issues"`
}

// HandoffIssue 续工包中的 issue 摘要。
type HandoffIssue struct {
	ID         string  `json:"id"`
	Identifier string  `json:"identifier"`
	Title      string  `json:"title"`
	Status     string  `json:"status"`
	Priority   string  `json:"priority"`
	ProjectID  string  `json:"project_id"`
	CycleID    *string `json:"cycle_id"`
	UpdatedAt  string  `json:"updated_at"`
}

// Team 协作班底（UUID 引导，工作区内可多团队）。
type Team struct {
	ID          string  `json:"id"`
	WorkspaceID string  `json:"workspace_id"`
	Name        string  `json:"name"`
	Slug        *string `json:"slug,omitempty"`
	TemplateKey string  `json:"template_key"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// TeamSeatClaim 席位上的 active bot_role_claim 摘要。
type TeamSeatClaim struct {
	MemberID  string `json:"member_id"`
	ClaimedAt string `json:"claimed_at"`
}

// TeamSeat 班底席位；MemberID 为 BoundMemberID 的 JSON 别名。
type TeamSeat struct {
	ID               string         `json:"id"`
	TeamID           string         `json:"team_id"`
	WorkspaceID      string         `json:"workspace_id"`
	RoleSlug         string         `json:"role_slug"`
	DisplayName      string         `json:"display_name"`
	PlaybookRef      *string        `json:"playbook_ref,omitempty"`
	BotRoleID        *string        `json:"bot_role_id,omitempty"`
	BoundMemberID    *string        `json:"bound_member_id,omitempty"`
	MemberID         *string        `json:"member_id,omitempty"` // alias of bound_member_id
	BoundAt          *string        `json:"bound_at,omitempty"`
	BoundDisplayName *string        `json:"bound_display_name,omitempty"`
	// Status: open | active | stopped | left（空串读出时按 open）
	Status           string         `json:"status"`
	CreatedAt        string         `json:"created_at"`
	UpdatedAt        string         `json:"updated_at"`
	Claim            *TeamSeatClaim `json:"claim,omitempty"`
}

// Seat status 常量（G2）
const (
	SeatStatusOpen    = "open"
	SeatStatusActive  = "active"
	SeatStatusStopped = "stopped"
	SeatStatusLeft    = "left"
)

// TeamInvite 团队加入密钥（ymt_）；库内只存 hash，明文仅签发时返回一次。
type TeamInvite struct {
	ID              string  `json:"id"`
	WorkspaceID     string  `json:"workspace_id"`
	TeamID          string  `json:"team_id"`
	TokenPrefix     string  `json:"token_prefix,omitempty"`
	// CodeTail 列表展示用末四（自 token_prefix 推导）；非明文
	CodeTail        string  `json:"code_tail,omitempty"`
	AllowExpandSeat bool    `json:"allow_expand_seat"`
	SeatSlug        *string `json:"seat_slug,omitempty"`
	RoleHint        *string `json:"role_hint,omitempty"`
	MaxUses         *int    `json:"max_uses,omitempty"`
	UseCount        int     `json:"use_count"`
	// Status: active | revoked | expired | exhausted（由 revoked/expires/uses 推导，不落库）
	Status          string  `json:"status"`
	ExpiresAt       *string `json:"expires_at,omitempty"`
	RevokedAt       *string `json:"revoked_at,omitempty"`
	CreatedBy       *string `json:"created_by,omitempty"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
	// Code 仅 CreateInvite 响应明文，不落库；List 永不返回
	Code string `json:"code,omitempty"`
}

// Invite status 常量（G3 list）
const (
	InviteStatusActive    = "active"
	InviteStatusRevoked   = "revoked"
	InviteStatusExpired   = "expired"
	InviteStatusExhausted = "exhausted"
)

// TeamInviteRedeemResult 兑换响应（无明文、无 hash）
type TeamInviteRedeemResult struct {
	WorkspaceID     string  `json:"workspace_id"`
	TeamID          string  `json:"team_id"`
	AllowExpandSeat bool    `json:"allow_expand_seat"`
	RoleHint        *string `json:"role_hint,omitempty"`
	SeatSlug        *string `json:"seat_slug,omitempty"`
}
