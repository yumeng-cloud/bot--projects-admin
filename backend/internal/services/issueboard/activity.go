package issueboard

import (
	"database/sql"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/models"
)

const activityCols = `id, workspace_id, issue_id, actor_id, actor_kind, action, payload, created_at`

func scanActivity(row interface{ Scan(dest ...any) error }) (models.Activity, error) {
	var a models.Activity
	var issueID, actorID sql.NullString
	err := row.Scan(&a.ID, &a.WorkspaceID, &issueID, &actorID, &a.ActorKind, &a.Action, &a.Payload, &a.CreatedAt)
	if err != nil {
		return a, err
	}
	a.IssueID = strPtr(issueID)
	a.ActorID = strPtr(actorID)
	return a, nil
}

func (s *Service) ListIssueActivities(workspaceID, issueID string) ([]models.Activity, error) {
	if err := s.ensureIssueInWorkspace(workspaceID, issueID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(`SELECT `+activityCols+` FROM activities
		WHERE workspace_id = ? AND issue_id = ?
		ORDER BY created_at DESC`, workspaceID, issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.Activity{}
	for rows.Next() {
		a, err := scanActivity(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

// ListActivities returns recent workspace activity feed. limit <= 0 defaults to 50; capped at 200.
func (s *Service) ListActivities(workspaceID string, limit int) ([]models.Activity, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	rows, err := s.DB.Query(`SELECT `+activityCols+` FROM activities
		WHERE workspace_id = ?
		ORDER BY created_at DESC
		LIMIT ?`, workspaceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.Activity{}
	for rows.Next() {
		a, err := scanActivity(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}
