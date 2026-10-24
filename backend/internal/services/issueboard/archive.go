package issueboard

import (
	"database/sql"
	"strings"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
)

// ArchiveIssueResult is one id outcome for POST …/issues/archive.
type ArchiveIssueResult struct {
	ID         string  `json:"id"`
	OK         bool    `json:"ok"`
	ArchivedAt *string `json:"archived_at,omitempty"`
	Error      string  `json:"error,omitempty"`
}

// ArchiveIssuesInput batch-archives done issues (ISSUE_ARCHIVE).
// TeamID non-empty: wrong-team / not-in-pool → per-id "Issue 不存在" (T6-I).
type ArchiveIssuesInput struct {
	IDs       []string
	TeamID    string
	ActorID   string
	ActorKind string
}

// ArchiveIssues sets archived_at=now for status=done issues only.
// Idempotent: already archived → ok no-op. Non-done → per-id failure (Chinese).
// Does not change status (orthogonal to archive).
func (s *Service) ArchiveIssues(workspaceID string, in ArchiveIssuesInput) ([]ArchiveIssueResult, error) {
	if len(in.IDs) == 0 {
		return nil, apperr.BadRequest("ids 不能为空")
	}
	teamID := strings.TrimSpace(in.TeamID)
	out := make([]ArchiveIssueResult, 0, len(in.IDs))
	seen := map[string]bool{}

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	for _, raw := range in.IDs {
		id := strings.TrimSpace(raw)
		if id == "" {
			out = append(out, ArchiveIssueResult{ID: raw, OK: false, Error: "id 无效"})
			continue
		}
		if seen[id] {
			// duplicate id in request: treat as idempotent success after first
			continue
		}
		seen[id] = true

		var status string
		var archived sql.NullString
		var issTeam sql.NullString
		var identifier string
		err := tx.QueryRow(
			`SELECT status, archived_at, team_id, identifier FROM issues WHERE id=? AND workspace_id=?`,
			id, workspaceID,
		).Scan(&status, &archived, &issTeam, &identifier)
		if err == sql.ErrNoRows {
			out = append(out, ArchiveIssueResult{ID: id, OK: false, Error: "Issue 不存在"})
			continue
		}
		if err != nil {
			return nil, err
		}
		if teamID != "" {
			// T6-I: assigned to another team → not in pool
			if issTeam.Valid && strings.TrimSpace(issTeam.String) != "" && issTeam.String != teamID {
				out = append(out, ArchiveIssueResult{ID: id, OK: false, Error: "Issue 不存在"})
				continue
			}
		}
		if archived.Valid && strings.TrimSpace(archived.String) != "" {
			at := archived.String
			out = append(out, ArchiveIssueResult{ID: id, OK: true, ArchivedAt: &at})
			continue
		}
		if status != "done" {
			out = append(out, ArchiveIssueResult{
				ID: id, OK: false,
				Error: "仅完成(done)状态的 Issue 可归档",
			})
			continue
		}
		now := nowUTC()
		res, err := tx.Exec(
			`UPDATE issues SET archived_at=?, updated_at=? WHERE id=? AND workspace_id=? AND archived_at IS NULL`,
			now, now, id, workspaceID,
		)
		if err != nil {
			return nil, err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			// raced to archive — re-read
			var at sql.NullString
			_ = tx.QueryRow(`SELECT archived_at FROM issues WHERE id=? AND workspace_id=?`, id, workspaceID).Scan(&at)
			if at.Valid {
				s := at.String
				out = append(out, ArchiveIssueResult{ID: id, OK: true, ArchivedAt: &s})
				continue
			}
			out = append(out, ArchiveIssueResult{ID: id, OK: false, Error: "归档失败，请重试"})
			continue
		}
		if err := appendActivity(tx, workspaceID, id, in.ActorID, in.ActorKind, "issue.archived", map[string]any{
			"identifier":  identifier,
			"status":      status,
			"archived_at": now,
		}); err != nil {
			return nil, err
		}
		at := now
		out = append(out, ArchiveIssueResult{ID: id, OK: true, ArchivedAt: &at})
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}

// UnarchiveIssueResult is one id outcome for POST …/issues/unarchive.
type UnarchiveIssueResult struct {
	ID    string `json:"id"`
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// UnarchiveIssuesInput batch-clears archived_at (WHEEL_FIX B1).
type UnarchiveIssuesInput struct {
	IDs       []string
	TeamID    string
	ActorID   string
	ActorKind string
}

// UnarchiveIssues clears archived_at for archived issues.
// Idempotent: already unarchived (archived_at NULL) → ok no-op.
// Does not change status.
func (s *Service) UnarchiveIssues(workspaceID string, in UnarchiveIssuesInput) ([]UnarchiveIssueResult, error) {
	if len(in.IDs) == 0 {
		return nil, apperr.BadRequest("ids 不能为空")
	}
	teamID := strings.TrimSpace(in.TeamID)
	out := make([]UnarchiveIssueResult, 0, len(in.IDs))
	seen := map[string]bool{}

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	for _, raw := range in.IDs {
		id := strings.TrimSpace(raw)
		if id == "" {
			out = append(out, UnarchiveIssueResult{ID: raw, OK: false, Error: "id 无效"})
			continue
		}
		if seen[id] {
			continue
		}
		seen[id] = true

		var status string
		var archived sql.NullString
		var issTeam sql.NullString
		var identifier string
		err := tx.QueryRow(
			`SELECT status, archived_at, team_id, identifier FROM issues WHERE id=? AND workspace_id=?`,
			id, workspaceID,
		).Scan(&status, &archived, &issTeam, &identifier)
		if err == sql.ErrNoRows {
			out = append(out, UnarchiveIssueResult{ID: id, OK: false, Error: "Issue 不存在"})
			continue
		}
		if err != nil {
			return nil, err
		}
		if teamID != "" {
			if issTeam.Valid && strings.TrimSpace(issTeam.String) != "" && issTeam.String != teamID {
				out = append(out, UnarchiveIssueResult{ID: id, OK: false, Error: "Issue 不存在"})
				continue
			}
		}
		if !archived.Valid || strings.TrimSpace(archived.String) == "" {
			// already unarchived — idempotent success
			out = append(out, UnarchiveIssueResult{ID: id, OK: true})
			continue
		}
		now := nowUTC()
		res, err := tx.Exec(
			`UPDATE issues SET archived_at=NULL, updated_at=? WHERE id=? AND workspace_id=? AND archived_at IS NOT NULL`,
			now, id, workspaceID,
		)
		if err != nil {
			return nil, err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			// raced to unarchive — treat as ok
			out = append(out, UnarchiveIssueResult{ID: id, OK: true})
			continue
		}
		if err := appendActivity(tx, workspaceID, id, in.ActorID, in.ActorKind, "issue.unarchived", map[string]any{
			"identifier": identifier,
			"status":     status,
		}); err != nil {
			return nil, err
		}
		out = append(out, UnarchiveIssueResult{ID: id, OK: true})
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}
