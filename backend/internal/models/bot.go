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
	ID         string  `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	RoleID     string  `json:"role_id"`
	MemberID   string  `json:"member_id"`
	ClaimedAt  string  `json:"claimed_at"`
	ReleasedAt *string `json:"released_at"`
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
	WorkspaceID  string         `json:"workspace_id"`
	MemberID     string         `json:"member_id"`
	GeneratedAt  string         `json:"generated_at"`
	ClaimedRole  *BotRole       `json:"claimed_role"`
	Memories     []BotMemory    `json:"memories"`
	Todos        []BotTodo      `json:"todos"`
	Issues       []HandoffIssue `json:"issues"`
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
