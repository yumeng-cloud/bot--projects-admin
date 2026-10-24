package issueboard

import (
	"errors"
	"strings"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
)

// BatchStatusIssueResult is one id outcome for POST …/issues/batch-status.
type BatchStatusIssueResult struct {
	ID      string `json:"id"`
	OK      bool   `json:"ok"`
	Status  string `json:"status,omitempty"`
	Version int    `json:"version,omitempty"`
	Error   string `json:"error,omitempty"`
}

// BatchStatusIssuesInput batch-updates issue status (PRO_ADMIN C2-2).
// TeamID non-empty: wrong-team → per-id "Issue 不存在" (same as Archive/UpdateIssue.RequireTeamID).
type BatchStatusIssuesInput struct {
	IDs       []string
	Status    string
	TeamID    string
	ActorID   string
	ActorKind string
}

// BatchStatusIssues sets status for each id via the same rules as UpdateIssue (validStatuses,
// bot→done forbidden, RequireTeamID, optimistic lock, issue.status_changed activity).
// Request-level: empty ids / invalid status / bot+done. Per-id partial failure otherwise.
func (s *Service) BatchStatusIssues(workspaceID string, in BatchStatusIssuesInput) ([]BatchStatusIssueResult, error) {
	if len(in.IDs) == 0 {
		return nil, apperr.BadRequest("ids 不能为空")
	}
	status := strings.TrimSpace(in.Status)
	if !validStatuses[status] {
		return nil, apperr.BadRequest("无效的状态")
	}
	if in.ActorKind == "bot" && status == "done" {
		return nil, apperr.Forbidden("Bot 不能单独将 Issue 设为完成（done），请设为 in_review 并由人工确认")
	}

	out := make([]BatchStatusIssueResult, 0, len(in.IDs))
	seen := map[string]bool{}
	teamID := strings.TrimSpace(in.TeamID)

	for _, raw := range in.IDs {
		id := strings.TrimSpace(raw)
		if id == "" {
			out = append(out, BatchStatusIssueResult{ID: raw, OK: false, Error: "id 无效"})
			continue
		}
		if seen[id] {
			continue
		}
		seen[id] = true

		cur, err := s.GetIssueInTeam(workspaceID, id, teamID)
		if err != nil {
			out = append(out, BatchStatusIssueResult{ID: id, OK: false, Error: appErrMessage(err)})
			continue
		}
		st := status
		upd, err := s.UpdateIssue(workspaceID, id, UpdateIssueInput{
			Version:       cur.Version,
			Status:        &st,
			RequireTeamID: teamID,
			ActorID:       in.ActorID,
			ActorKind:     in.ActorKind,
		})
		if err != nil {
			out = append(out, BatchStatusIssueResult{ID: id, OK: false, Error: appErrMessage(err)})
			continue
		}
		out = append(out, BatchStatusIssueResult{ID: id, OK: true, Status: upd.Status, Version: upd.Version})
	}
	return out, nil
}

func appErrMessage(err error) string {
	var ae *apperr.AppError
	if errors.As(err, &ae) && ae.Message != "" {
		return ae.Message
	}
	if err != nil {
		return err.Error()
	}
	return "未知错误"
}
