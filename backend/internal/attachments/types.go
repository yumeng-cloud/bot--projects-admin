package attachments

import "time"

// MaxBytes 单文件上限（10MB）。可在测试中覆盖为更小值。
var MaxBytes int64 = 10 << 20

// Attachment 附件元数据（文件本体在本地磁盘，不入 DB BLOB）。
type Attachment struct {
	ID          string  `json:"id"`
	WorkspaceID string  `json:"workspace_id"`
	IssueID     string  `json:"issue_id"`
	CommentID   *string `json:"comment_id,omitempty"`
	UploaderID  string  `json:"uploader_id"`
	Filename    string  `json:"filename"`
	ContentType string  `json:"content_type"`
	SizeBytes   int64   `json:"size_bytes"`
	StorageKey  string  `json:"-"` // 不对前端暴露绝对/相对盘路径
	CreatedAt   string  `json:"created_at"`
	URL         string  `json:"url"`
}

func (a *Attachment) withURL() *Attachment {
	if a == nil {
		return nil
	}
	a.URL = contentURL(a.WorkspaceID, a.ID)
	return a
}

func contentURL(wsID, id string) string {
	return "/api/v1/workspaces/" + wsID + "/attachments/" + id + "/content"
}

func nowUTC() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}
