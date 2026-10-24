package models

// WikiSpace 知识库空间。
type WikiSpace struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// WikiPage 知识库页面（Markdown body）。
type WikiPage struct {
	ID          string  `json:"id"`
	SpaceID     string  `json:"space_id"`
	WorkspaceID string  `json:"workspace_id"`
	ParentID    *string `json:"parent_id"`
	Title       string  `json:"title"`
	Slug        string  `json:"slug"`
	BodyMD      string  `json:"body_md"`
	ProjectID   *string `json:"project_id"`
	IssueID     *string `json:"issue_id"`
	CreatedBy   *string `json:"created_by"`
	UpdatedBy   *string `json:"updated_by"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// WikiTreeNode 树形目录节点。
type WikiTreeNode struct {
	ID       string         `json:"id"`
	Title    string         `json:"title"`
	Slug     string         `json:"slug"`
	ParentID *string        `json:"parent_id"`
	Children []WikiTreeNode `json:"children"`
}
