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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := s.attachCommentAuthors(workspaceID, list); err != nil {
		return nil, err
	}
	return list, nil
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
	// body/summary/comment: FE + Bot lean-read human text (align seat.force_bind payload keys)
	if err := appendActivity(tx, workspaceID, issueID, in.AuthorID, in.ActorKind, "comment.created", map[string]any{
		"comment_id": id,
		"body":       in.Body,
		"summary":    in.Body,
		"comment":    in.Body,
	}); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	c := models.Comment{
		ID: id, IssueID: issueID, AuthorID: in.AuthorID,
		Body: in.Body, CreatedAt: now, UpdatedAt: now,
	}
	one := []models.Comment{c}
	if err := s.attachCommentAuthors(workspaceID, one); err != nil {
		return nil, err
	}
	return &one[0], nil
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
		"body":       in.Body,
		"summary":    in.Body,
		"comment":    in.Body,
	}); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	c := models.Comment{
		ID: commentID, IssueID: issueID, AuthorID: authorID,
		Body: in.Body, CreatedAt: createdAt, UpdatedAt: now,
	}
	one := []models.Comment{c}
	if err := s.attachCommentAuthors(workspaceID, one); err != nil {
		return nil, err
	}
	return &one[0], nil
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

// attachCommentAuthors batch-loads members.display_name onto Comment.Author (workspace-scoped).
func (s *Service) attachCommentAuthors(workspaceID string, list []models.Comment) error {
	if len(list) == 0 {
		return nil
	}
	idSet := map[string]struct{}{}
	for i := range list {
		if strings.TrimSpace(list[i].AuthorID) != "" {
			idSet[list[i].AuthorID] = struct{}{}
		}
	}
	names := map[string]string{}
	if len(idSet) > 0 {
		ids := make([]string, 0, len(idSet))
		for id := range idSet {
			ids = append(ids, id)
		}
		placeholders := make([]string, len(ids))
		args := make([]any, 0, len(ids)+1)
		args = append(args, workspaceID)
		for i, id := range ids {
			placeholders[i] = "?"
			args = append(args, id)
		}
		q := `SELECT id, display_name FROM members WHERE workspace_id=? AND id IN (` + strings.Join(placeholders, ",") + `)`
		rows, err := s.DB.Query(q, args...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id, name string
			if err := rows.Scan(&id, &name); err != nil {
				return err
			}
			names[id] = name
		}
		if err := rows.Err(); err != nil {
			return err
		}
	}
	for i := range list {
		if strings.TrimSpace(list[i].AuthorID) == "" {
			list[i].Author = nil
			continue
		}
		id := list[i].AuthorID
		list[i].Author = &models.MemberRef{ID: id, DisplayName: names[id]}
	}
	return nil
}
