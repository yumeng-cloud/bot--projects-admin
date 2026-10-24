package botmem

import (
	"database/sql"
	"strings"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/models"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
)

// RequireTeamPinAccess 本队成员（席位 bound）或 Admin；错队/无队 → 404。
func (s *Service) RequireTeamPinAccess(workspaceID, teamID, memberID string, isAdmin bool) error {
	teamID = strings.TrimSpace(teamID)
	if teamID == "" {
		return apperr.NotFound("团队不存在或无权访问")
	}
	if _, err := s.GetTeam(workspaceID, teamID); err != nil {
		// 统一 404，不泄露「存在但无权」
		return apperr.NotFound("团队不存在或无权访问")
	}
	if isAdmin {
		return nil
	}
	if memberID == "" {
		return apperr.NotFound("团队不存在或无权访问")
	}
	var n int
	err := s.DB.QueryRow(`SELECT COUNT(1) FROM team_seats
		WHERE workspace_id = ? AND team_id = ? AND bound_member_id = ?`,
		workspaceID, teamID, memberID).Scan(&n)
	if err != nil {
		return err
	}
	if n == 0 {
		return apperr.NotFound("团队不存在或无权访问")
	}
	return nil
}

func (s *Service) ListPins(workspaceID, teamID string) ([]models.TeamPin, error) {
	rows, err := s.DB.Query(`SELECT id, workspace_id, team_id, key, value, updated_by, updated_at
		FROM team_pins WHERE workspace_id = ? AND team_id = ? ORDER BY key`, workspaceID, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.TeamPin
	for rows.Next() {
		p, err := scanPin(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if out == nil {
		out = []models.TeamPin{}
	}
	return out, nil
}

type UpsertPinInput struct {
	Key       string
	Value     string
	UpdatedBy string
}

func (s *Service) UpsertPin(workspaceID, teamID string, in UpsertPinInput) (*models.TeamPin, error) {
	key := strings.TrimSpace(in.Key)
	if key == "" {
		return nil, apperr.BadRequest("key 不能为空")
	}
	now := nowUTC()
	var byPtr *string
	if ub := strings.TrimSpace(in.UpdatedBy); ub != "" {
		byPtr = &ub
	}
	var existingID string
	err := s.DB.QueryRow(`SELECT id FROM team_pins WHERE team_id = ? AND key = ?`, teamID, key).Scan(&existingID)
	if err == nil {
		_, err = s.DB.Exec(`UPDATE team_pins SET value = ?, updated_by = ?, updated_at = ?, workspace_id = ?
			WHERE id = ?`, in.Value, nullStr(byPtr), now, workspaceID, existingID)
		if err != nil {
			return nil, err
		}
		return s.getPinByID(workspaceID, teamID, existingID)
	}
	if err != sql.ErrNoRows {
		return nil, err
	}
	id := idgen.New()
	_, err = s.DB.Exec(`INSERT INTO team_pins (id, workspace_id, team_id, key, value, updated_by, updated_at)
		VALUES (?,?,?,?,?,?,?)`, id, workspaceID, teamID, key, in.Value, nullStr(byPtr), now)
	if err != nil {
		return nil, err
	}
	return s.getPinByID(workspaceID, teamID, id)
}


func (s *Service) GetPin(workspaceID, teamID, key string) (*models.TeamPin, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, apperr.BadRequest("key 不能为空")
	}
	p, err := scanPin(s.DB.QueryRow(`SELECT id, workspace_id, team_id, key, value, updated_by, updated_at
		FROM team_pins WHERE workspace_id = ? AND team_id = ? AND key = ?`, workspaceID, teamID, key))
	if err == sql.ErrNoRows {
		return nil, apperr.NotFound("pin 不存在")
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Service) DeletePin(workspaceID, teamID, key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return apperr.BadRequest("key 不能为空")
	}
	res, err := s.DB.Exec(`DELETE FROM team_pins WHERE workspace_id = ? AND team_id = ? AND key = ?`,
		workspaceID, teamID, key)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperr.NotFound("pin 不存在")
	}
	return nil
}

func (s *Service) getPinByID(workspaceID, teamID, id string) (*models.TeamPin, error) {
	p, err := scanPin(s.DB.QueryRow(`SELECT id, workspace_id, team_id, key, value, updated_by, updated_at
		FROM team_pins WHERE id = ? AND workspace_id = ? AND team_id = ?`, id, workspaceID, teamID))
	if err == sql.ErrNoRows {
		return nil, apperr.NotFound("pin 不存在")
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

type pinScanner interface {
	Scan(dest ...any) error
}

func scanPin(row pinScanner) (models.TeamPin, error) {
	var p models.TeamPin
	var updatedBy sql.NullString
	err := row.Scan(&p.ID, &p.WorkspaceID, &p.TeamID, &p.Key, &p.Value, &updatedBy, &p.UpdatedAt)
	if err != nil {
		return p, err
	}
	if updatedBy.Valid && updatedBy.String != "" {
		s := updatedBy.String
		p.UpdatedBy = &s
	}
	return p, nil
}
