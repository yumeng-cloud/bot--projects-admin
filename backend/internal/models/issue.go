package models

// Project 工作区内项目。
type Project struct {
	ID          string  `json:"id"`
	WorkspaceID string  `json:"workspace_id"`
	Name        string  `json:"name"`
	KeyPrefix   string  `json:"key_prefix"`
	Description string  `json:"description"`
	ArchivedAt  *string `json:"archived_at"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// Label 工作区标签。
type Label struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	CreatedAt   string `json:"created_at"`
}

// Cycle 项目迭代。
type Cycle struct {
	ID           string  `json:"id"`
	ProjectID    string  `json:"project_id"`
	Name         string  `json:"name"`
	StartDate    *string `json:"start_date"`
	EndDate      *string `json:"end_date"`
	Capacity     int     `json:"capacity"`
	Status       string  `json:"status"`
	EstimateSum  int     `json:"estimate_sum"`
	OverCapacity bool    `json:"over_capacity"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// Issue 看板事项。
type Issue struct {
	ID          string     `json:"id"`
	ProjectID   string     `json:"project_id"`
	WorkspaceID string     `json:"workspace_id"`
	TeamID      *string    `json:"team_id"` // nil = 全工作区可见（旧单）；新建必填
	Number      int        `json:"number"`
	Identifier  string     `json:"identifier"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	Estimate    *int       `json:"estimate"`
	AssigneeID  *string    `json:"assignee_id"`
	Assignee    *MemberRef `json:"assignee"`
	ReporterID  *string    `json:"reporter_id"`
	Reporter    *MemberRef `json:"reporter"`
	CycleID     *string    `json:"cycle_id"`
	ParentID    *string    `json:"parent_id"`
	StartAt     *string    `json:"start_at"`
	DueAt       *string    `json:"due_at"`
	IsLeaf      bool       `json:"is_leaf"`
	ChildCount  int        `json:"child_count"`
	BoardOrder  float64    `json:"board_order"`
	Version     int        `json:"version"`
	LabelIDs    []string   `json:"label_ids,omitempty"`
	Labels      []Label    `json:"labels,omitempty"`
	CreatedAt   string     `json:"created_at"`
	UpdatedAt   string     `json:"updated_at"`
}

// MemberRef is a minimal nested member reference for Issue assignee/reporter.
type MemberRef struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// IssueRelation 依赖关系。
type IssueRelation struct {
	ID           string `json:"id"`
	WorkspaceID  string `json:"workspace_id"`
	FromIssueID  string `json:"from_issue_id"`
	ToIssueID    string `json:"to_issue_id"`
	RelationType string `json:"relation_type"`
	CreatedAt    string `json:"created_at"`
}

// Comment Issue 评论。
type Comment struct {
	ID        string     `json:"id"`
	IssueID   string     `json:"issue_id"`
	AuthorID  string     `json:"author_id"`
	Author    *MemberRef `json:"author"`
	Body      string     `json:"body"`
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
}

// Activity 活动流条目。
type Activity struct {
	ID          string  `json:"id"`
	WorkspaceID string  `json:"workspace_id"`
	IssueID     *string `json:"issue_id"`
	ActorID     *string `json:"actor_id"`
	ActorKind   string  `json:"actor_kind"`
	Action      string  `json:"action"`
	Payload     string  `json:"payload"`
	CreatedAt   string  `json:"created_at"`
}
