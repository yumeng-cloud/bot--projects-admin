package issueboard

import (
	"database/sql"
	"strings"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/models"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
)

const commentCols = `id, issue_id, author_id, body, created_at, updated_at`

func scanComment(row interface{ Scan(dest ...any) error }) (models.Comment, error) {
	var c models.Comment
	err := row.Scan(&c.ID, &c.IssueID, &c.AuthorID, &c.Body, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (s *Service) ListComments(workspaceID, issueID string) ([]models.Comment, error) {
	if err := s.ensureIssueInWorkspace(workspaceID, issueID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(`SELECT c.id, c.issue_id, c.author_id, c.body, c.created_at, c.updated_at
		FROM comments c
		JOIN issues i ON i.id = c.issue_id
		WHERE c.issue_id = ? AND i.workspace_id = ?
		ORDER BY c.created_at ASC`, issueID, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.Comment{}
	for rows.Next() {
		c, err := scanComment(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

type CreateCommentInput struct {
	Body      string
	AuthorID  string
	ActorKind string
}

func (s *Service) CreateComment(workspaceID, issueID string, in CreateCommentInput) (*models.Comment, error) {
	in.Body = strings.TrimSpace(in.Body)
	if in.Body == "" {
		return nil, apperr.BadRequest("评论内容不能为空")
	}
	if in.AuthorID == "" {
		return nil, apperr.BadRequest("缺少作者")
	}
	if err := s.ensureIssueInWorkspace(workspaceID, issueID); err != nil {
		return nil, err
	}
	if err := s.ensureMember(workspaceID, in.AuthorID); err != nil {
		return nil, err
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	id := idgen.New()
	now := nowUTC()
	_, err = tx.Exec(`INSERT INTO comments (id, issue_id, author_id, body, created_at, updated_at)
		VALUES (?,?,?,?,?,?)`, id, issueID, in.AuthorID, in.Body, now, now)
	if err != nil {
		return nil, err
	}
	if err := appendActivity(tx, workspaceID, issueID, in.AuthorID, in.ActorKind, "comment.created", map[string]any{
		"comment_id": id,
	}); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &models.Comment{
		ID: id, IssueID: issueID, AuthorID: in.AuthorID,
		Body: in.Body, CreatedAt: now, UpdatedAt: now,
	}, nil
}

type UpdateCommentInput struct {
	Body string
	// ExpectedUpdatedAt is optional optimistic-lock token from the client (JSON updated_at).
	// When set, UPDATE matches updated_at; mismatch → VersionConflict.
	// When omitted, keeps last-write-wins for FE that does not send it yet.
	ExpectedUpdatedAt string
	ActorID           string
	ActorKind         string
}

func (s *Service) UpdateComment(workspaceID, issueID, commentID string, in UpdateCommentInput) (*models.Comment, error) {
	in.Body = strings.TrimSpace(in.Body)
	if in.Body == "" {
		return nil, apperr.BadRequest("评论内容不能为空")
	}
	if err := s.ensureIssueInWorkspace(workspaceID, issueID); err != nil {
		return nil, err
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var authorID, createdAt string
	err = tx.QueryRow(`SELECT author_id, created_at FROM comments WHERE id=? AND issue_id=?`,
		commentID, issueID).Scan(&authorID, &createdAt)
	if err == sql.ErrNoRows {
		return nil, apperr.NotFound("评论不存在")
	}
	if err != nil {
		return nil, err
	}
	now := nowUTC()
	var res sql.Result
	if in.ExpectedUpdatedAt != "" {
		res, err = tx.Exec(`UPDATE comments SET body=?, updated_at=? WHERE id=? AND issue_id=? AND updated_at=?`,
			in.Body, now, commentID, issueID, in.ExpectedUpdatedAt)
	} else {
		res, err = tx.Exec(`UPDATE comments SET body=?, updated_at=? WHERE id=? AND issue_id=?`,
			in.Body, now, commentID, issueID)
	}
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		if in.ExpectedUpdatedAt != "" {
			return nil, apperr.VersionConflict("评论已被他人更新")
		}
		return nil, apperr.NotFound("评论不存在")
	}
	actor := in.ActorID
	if actor == "" {
		actor = authorID
	}
	if err := appendActivity(tx, workspaceID, issueID, actor, in.ActorKind, "comment.updated", map[string]any{
		"comment_id": commentID,
	}); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &models.Comment{
		ID: commentID, IssueID: issueID, AuthorID: authorID,
		Body: in.Body, CreatedAt: createdAt, UpdatedAt: now,
	}, nil
}

func (s *Service) DeleteComment(workspaceID, issueID, commentID, actorID, actorKind string) error {
	if err := s.ensureIssueInWorkspace(workspaceID, issueID); err != nil {
		return err
	}
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var id string
	err = tx.QueryRow(`SELECT id FROM comments WHERE id=? AND issue_id=?`, commentID, issueID).Scan(&id)
	if err == sql.ErrNoRows {
		return apperr.NotFound("评论不存在")
	}
	if err != nil {
		return err
	}
	if err := appendActivity(tx, workspaceID, issueID, actorID, actorKind, "comment.deleted", map[string]any{
		"comment_id": commentID,
	}); err != nil {
		return err
	}
	res, err := tx.Exec(`DELETE FROM comments WHERE id=? AND issue_id=?`, commentID, issueID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperr.NotFound("评论不存在")
	}
	return tx.Commit()
}
