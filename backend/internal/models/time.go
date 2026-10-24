package models

// TimeEntry 工时记录。
type TimeEntry struct {
	ID          string   `json:"id"`
	WorkspaceID string   `json:"workspace_id"`
	IssueID     string   `json:"issue_id"`
	MemberID    string   `json:"member_id"`
	Hours       float64  `json:"hours"`
	Note        string   `json:"note"`
	Billable    bool     `json:"billable"`
	Rate        *float64 `json:"rate"`
	WorkedAt    string   `json:"worked_at"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
	// 汇总扩展（可选）
	ProjectID   string `json:"project_id,omitempty"`
	IssueIdent  string `json:"issue_identifier,omitempty"`
	IssueTitle  string `json:"issue_title,omitempty"`
	MemberName  string `json:"member_name,omitempty"`
}

// BillingRate 成员/工作区默认费率。
type BillingRate struct {
	ID          string  `json:"id"`
	WorkspaceID string  `json:"workspace_id"`
	MemberID    *string `json:"member_id"`
	HourlyRate  float64 `json:"hourly_rate"`
	Currency    string  `json:"currency"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// BillingSummaryRow 汇总行。
type BillingSummaryRow struct {
	GroupKey   string  `json:"group_key"`
	GroupLabel string  `json:"group_label"`
	Hours      float64 `json:"hours"`
	BillableHours float64 `json:"billable_hours"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	Entries    int     `json:"entries"`
}

// BillingSummary 汇总结果。
type BillingSummary struct {
	From    string              `json:"from"`
	To      string              `json:"to"`
	GroupBy string              `json:"group_by"`
	Rows    []BillingSummaryRow `json:"rows"`
	TotalHours float64          `json:"total_hours"`
	TotalAmount float64         `json:"total_amount"`
	Currency string             `json:"currency"`
}
