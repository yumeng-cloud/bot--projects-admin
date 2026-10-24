package issueboard

import (
	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"

	"database/sql"
	"strings"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/models"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
)

func (s *Service) ListLabels(workspaceID string) ([]models.Label, error) {
	rows, err := s.DB.Query(`SELECT id, workspace_id, name, color, created_at FROM labels WHERE workspace_id=? ORDER BY name`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.Label{}
	for rows.Next() {
		var l models.Label
		if err := rows.Scan(&l.ID, &l.WorkspaceID, &l.Name, &l.Color, &l.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, l)
	}
	return list, rows.Err()
}

type CreateLabelInput struct {
	Name  string
	Color string
}

func (s *Service) CreateLabel(workspaceID string, in CreateLabelInput) (*models.Label, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return nil, apperr.BadRequest("标签名称不能为空")
	}
	color := in.Color
	if color == "" {
		color = "#1677ff"
	}
	id := idgen.New()
	now := nowUTC()
	_, err := s.DB.Exec(`INSERT INTO labels (id, workspace_id, name, color, created_at) VALUES (?,?,?,?,?)`,
		id, workspaceID, in.Name, color, now)
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, apperr.Conflict("标签名称已存在")
		}
		return nil, err
	}
	return s.GetLabel(workspaceID, id)
}

func (s *Service) GetLabel(workspaceID, labelID string) (*models.Label, error) {
	var l models.Label
	err := s.DB.QueryRow(`SELECT id, workspace_id, name, color, created_at FROM labels WHERE id=? AND workspace_id=?`,
		labelID, workspaceID).Scan(&l.ID, &l.WorkspaceID, &l.Name, &l.Color, &l.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, apperr.NotFound("标签不存在")
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

type UpdateLabelInput struct {
	Name  *string
	Color *string
}

func (s *Service) UpdateLabel(workspaceID, labelID string, in UpdateLabelInput) (*models.Label, error) {
	cur, err := s.GetLabel(workspaceID, labelID)
	if err != nil {
		return nil, err
	}
	name, color := cur.Name, cur.Color
	if in.Name != nil {
		name = strings.TrimSpace(*in.Name)
		if name == "" {
			return nil, apperr.BadRequest("标签名称不能为空")
		}
	}
	if in.Color != nil {
		color = *in.Color
	}
	_, err = s.DB.Exec(`UPDATE labels SET name=?, color=? WHERE id=? AND workspace_id=?`, name, color, labelID, workspaceID)
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, apperr.Conflict("标签名称已存在")
		}
		return nil, err
	}
	return s.GetLabel(workspaceID, labelID)
}

func (s *Service) DeleteLabel(workspaceID, labelID string) error {
	res, err := s.DB.Exec(`DELETE FROM labels WHERE id=? AND workspace_id=?`, labelID, workspaceID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperr.NotFound("标签不存在")
	}
	return nil
}

func (s *Service) AttachLabel(workspaceID, issueID, labelID, actorID, actorKind string) error {
	if _, err := s.GetIssue(workspaceID, issueID); err != nil {
		return err
	}
	if _, err := s.GetLabel(workspaceID, labelID); err != nil {
		return err
	}
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`INSERT INTO issue_labels (issue_id, label_id) VALUES (?,?) ON CONFLICT(issue_id, label_id) DO NOTHING`, issueID, labelID)
	if err != nil {
		return err
	}
	if err := appendActivity(tx, workspaceID, issueID, actorID, actorKind, "issue.label_attached", map[string]any{"label_id": labelID}); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) DetachLabel(workspaceID, issueID, labelID, actorID, actorKind string) error {
	if _, err := s.GetIssue(workspaceID, issueID); err != nil {
		return err
	}
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.Exec(`DELETE FROM issue_labels WHERE issue_id=? AND label_id=?`, issueID, labelID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperr.NotFound("Issue 上不存在该标签")
	}
	if err := appendActivity(tx, workspaceID, issueID, actorID, actorKind, "issue.label_detached", map[string]any{"label_id": labelID}); err != nil {
		return err
	}
	return tx.Commit()
}
