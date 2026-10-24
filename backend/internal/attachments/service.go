package attachments

import (
	"context"
	"database/sql"
	"io"
	"path/filepath"
	"strings"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
)

// Service 附件元数据 + 本地落盘。Issue 域可 import 调用 List*/Delete/CanDelete。
type Service struct {
	DB    *sql.DB
	Store *DiskStore
}

func New(db *sql.DB, storeRoot string) *Service {
	return &Service{DB: db, Store: NewDiskStore(storeRoot)}
}

const selectCols = `id, workspace_id, issue_id, comment_id, uploader_id, filename, content_type, size_bytes, storage_key, created_at`

func scanAttachment(row interface{ Scan(dest ...any) error }) (*Attachment, error) {
	var a Attachment
	var commentID sql.NullString
	err := row.Scan(
		&a.ID, &a.WorkspaceID, &a.IssueID, &commentID, &a.UploaderID,
		&a.Filename, &a.ContentType, &a.SizeBytes, &a.StorageKey, &a.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if commentID.Valid && commentID.String != "" {
		s := commentID.String
		a.CommentID = &s
	}
	return a.withURL(), nil
}

func (s *Service) ensureIssue(workspaceID, issueID string) error {
	var id string
	err := s.DB.QueryRow(`SELECT id FROM issues WHERE id = ? AND workspace_id = ?`, issueID, workspaceID).Scan(&id)
	if err == sql.ErrNoRows {
		return apperr.NotFound("Issue 不存在")
	}
	return err
}

func (s *Service) ensureComment(workspaceID, issueID, commentID string) error {
	var id string
	err := s.DB.QueryRow(`
		SELECT c.id FROM comments c
		JOIN issues i ON i.id = c.issue_id
		WHERE c.id = ? AND c.issue_id = ? AND i.workspace_id = ?`,
		commentID, issueID, workspaceID).Scan(&id)
	if err == sql.ErrNoRows {
		return apperr.NotFound("评论不存在")
	}
	return err
}

// UploadIssue 上传 Issue 级附件（comment_id 为空）。
func (s *Service) UploadIssue(ctx context.Context, wsID, issueID, uploaderID, filename, contentType string, r io.Reader, size int64) (*Attachment, error) {
	return s.upload(ctx, wsID, issueID, "", uploaderID, filename, contentType, r, size)
}

// UploadComment 上传评论附件；校验 comment 属于该 issue/workspace。
func (s *Service) UploadComment(ctx context.Context, wsID, issueID, commentID, uploaderID, filename, contentType string, r io.Reader, size int64) (*Attachment, error) {
	if strings.TrimSpace(commentID) == "" {
		return nil, apperr.BadRequest("缺少评论 ID")
	}
	return s.upload(ctx, wsID, issueID, commentID, uploaderID, filename, contentType, r, size)
}

func (s *Service) upload(ctx context.Context, wsID, issueID, commentID, uploaderID, filename, contentType string, r io.Reader, size int64) (*Attachment, error) {
	_ = ctx
	if uploaderID == "" {
		return nil, apperr.Unauthorized("请先登录")
	}
	filename = normalizeFilename(filename)
	if err := checkDeclaredSize(size); err != nil {
		return nil, err
	}
	_, mime, err := resolveContentType(filename, contentType)
	if err != nil {
		return nil, err
	}
	if err := s.ensureIssue(wsID, issueID); err != nil {
		return nil, err
	}
	if commentID != "" {
		if err := s.ensureComment(wsID, issueID, commentID); err != nil {
			return nil, err
		}
	}

	id := idgen.New()
	storageKey := filepath.ToSlash(filepath.Join(wsID, id))

	// 先落盘再写库；落盘失败不落库。写库失败则删盘（F10）。
	written, err := s.Store.Save(storageKey, r, MaxBytes)
	if err != nil {
		return nil, err
	}

	now := nowUTC()
	var commentArg any
	if commentID != "" {
		commentArg = commentID
	}
	_, err = s.DB.Exec(`
		INSERT INTO attachments
			(id, workspace_id, issue_id, comment_id, uploader_id, filename, content_type, size_bytes, storage_key, created_at)
		VALUES (?,?,?,?,?,?,?,?,?,?)`,
		id, wsID, issueID, commentArg, uploaderID, filename, mime, written, storageKey, now)
	if err != nil {
		_ = s.Store.Delete(storageKey)
		return nil, apperr.Internal("保存附件元数据失败")
	}

	att := &Attachment{
		ID: id, WorkspaceID: wsID, IssueID: issueID, UploaderID: uploaderID,
		Filename: filename, ContentType: mime, SizeBytes: written,
		StorageKey: storageKey, CreatedAt: now,
	}
	if commentID != "" {
		c := commentID
		att.CommentID = &c
	}
	return att.withURL(), nil
}

// Get 返回附件元数据（须属该 workspace）。
func (s *Service) Get(ctx context.Context, wsID, id string) (*Attachment, error) {
	_ = ctx
	a, err := scanAttachment(s.DB.QueryRow(`SELECT `+selectCols+` FROM attachments WHERE id = ? AND workspace_id = ?`, id, wsID))
	if err == sql.ErrNoRows {
		return nil, apperr.NotFound("附件不存在")
	}
	if err != nil {
		return nil, err
	}
	return a, nil
}

// OpenContent 打开文件流；调用方负责 Close。
func (s *Service) OpenContent(ctx context.Context, wsID, id string) (*Attachment, io.ReadCloser, error) {
	a, err := s.Get(ctx, wsID, id)
	if err != nil {
		return nil, nil, err
	}
	f, err := s.Store.Open(a.StorageKey)
	if err != nil {
		return nil, nil, err
	}
	return a, f, nil
}

// ListByIssue 列出该 Issue 下全部附件（含评论附件）。
func (s *Service) ListByIssue(ctx context.Context, wsID, issueID string) ([]Attachment, error) {
	_ = ctx
	if err := s.ensureIssue(wsID, issueID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(`
		SELECT `+selectCols+` FROM attachments
		WHERE workspace_id = ? AND issue_id = ?
		ORDER BY created_at ASC`, wsID, issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collect(rows)
}

// ListByComment 列出某评论下的附件。
func (s *Service) ListByComment(ctx context.Context, wsID, issueID, commentID string) ([]Attachment, error) {
	_ = ctx
	if err := s.ensureComment(wsID, issueID, commentID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(`
		SELECT `+selectCols+` FROM attachments
		WHERE workspace_id = ? AND issue_id = ? AND comment_id = ?
		ORDER BY created_at ASC`, wsID, issueID, commentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collect(rows)
}

func collect(rows *sql.Rows) ([]Attachment, error) {
	out := []Attachment{}
	for rows.Next() {
		a, err := scanAttachment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *a)
	}
	return out, rows.Err()
}

// CanDelete：上传者本人或 workspace Admin 可删；Bot 以绑定 member_id / role 计。
func CanDelete(actorMemberID, role string, att *Attachment) bool {
	if att == nil || actorMemberID == "" {
		return false
	}
	if att.UploaderID == actorMemberID {
		return true
	}
	return role == "admin"
}

// Delete 删除元数据行并尝试删盘。权限由调用方用 CanDelete 判定后再调。
func (s *Service) Delete(ctx context.Context, wsID, id string) error {
	a, err := s.Get(ctx, wsID, id)
	if err != nil {
		return err
	}
	res, err := s.DB.Exec(`DELETE FROM attachments WHERE id = ? AND workspace_id = ?`, id, wsID)
	if err != nil {
		return apperr.Internal("删除附件失败")
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperr.NotFound("附件不存在")
	}
	_ = s.Store.Delete(a.StorageKey) // 盘文件缺失不阻断
	return nil
}
