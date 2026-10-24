package issueboard

import (
	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"

	"database/sql"
	"strings"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/models"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
)

const projectCols = `id, workspace_id, name, key_prefix, description, archived_at, created_at, updated_at`

func scanProject(row interface{ Scan(dest ...any) error }) (models.Project, error) {
	var p models.Project
	var archived sql.NullString
	err := row.Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.KeyPrefix, &p.Description, &archived, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return p, err
	}
	p.ArchivedAt = strPtr(archived)
	return p, nil
}

type CreateProjectInput struct {
	Name        string
	KeyPrefix   string
	Description string
}

func (s *Service) ListProjects(workspaceID string) ([]models.Project, error) {
	rows, err := s.DB.Query(`SELECT `+projectCols+` FROM projects WHERE workspace_id = ? ORDER BY created_at ASC`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.Project{}
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

func (s *Service) CreateProject(workspaceID string, in CreateProjectInput) (*models.Project, error) {
	in.Name = strings.TrimSpace(in.Name)
	prefix := normalizePrefix(in.KeyPrefix)
	if in.Name == "" {
		return nil, apperr.BadRequest("项目名称不能为空")
	}
	if prefix == "" || len(prefix) > 8 {
		return nil, apperr.BadRequest("key_prefix 须为 1–8 位字母或数字")
	}
	id := idgen.New()
	now := nowUTC()
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`INSERT INTO projects (id, workspace_id, name, key_prefix, description, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?)`, id, workspaceID, in.Name, prefix, in.Description, now, now)
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, apperr.Conflict("该工作空间已存在相同 key_prefix 的项目")
		}
		return nil, err
	}
	_, err = tx.Exec(`INSERT INTO issue_counters (project_id, next_number) VALUES (?, 1)`, id)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.GetProject(workspaceID, id)
}

func (s *Service) GetProject(workspaceID, projectID string) (*models.Project, error) {
	p, err := scanProject(s.DB.QueryRow(`SELECT `+projectCols+` FROM projects WHERE id = ? AND workspace_id = ?`, projectID, workspaceID))
	if err == sql.ErrNoRows {
		return nil, apperr.NotFound("项目不存在")
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

type UpdateProjectInput struct {
	Name        *string
	Description *string
	Archive     *bool
}

func (s *Service) UpdateProject(workspaceID, projectID string, in UpdateProjectInput) (*models.Project, error) {
	p, err := s.GetProject(workspaceID, projectID)
	if err != nil {
		return nil, err
	}
	name, desc := p.Name, p.Description
	if in.Name != nil {
		name = strings.TrimSpace(*in.Name)
		if name == "" {
			return nil, apperr.BadRequest("项目名称不能为空")
		}
	}
	if in.Description != nil {
		desc = *in.Description
	}
	now := nowUTC()
	var archived any
	if in.Archive != nil {
		if *in.Archive {
			archived = now
		} else {
			archived = nil
		}
	} else if p.ArchivedAt != nil {
		archived = *p.ArchivedAt
	}
	_, err = s.DB.Exec(`UPDATE projects SET name=?, description=?, archived_at=?, updated_at=? WHERE id=? AND workspace_id=?`,
		name, desc, archived, now, projectID, workspaceID)
	if err != nil {
		return nil, err
	}
	return s.GetProject(workspaceID, projectID)
}

func (s *Service) DeleteProject(workspaceID, projectID string) error {
	res, err := s.DB.Exec(`DELETE FROM projects WHERE id = ? AND workspace_id = ?`, projectID, workspaceID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperr.NotFound("项目不存在")
	}
	return nil
}
