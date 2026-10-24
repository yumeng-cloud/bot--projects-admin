package issueboard

import (
	"database/sql"
	"strings"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/models"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
)

const cycleCols = `id, project_id, name, start_date, end_date, capacity, status, created_at, updated_at`

func scanCycle(row interface{ Scan(dest ...any) error }) (models.Cycle, error) {
	var c models.Cycle
	var start, end sql.NullString
	err := row.Scan(&c.ID, &c.ProjectID, &c.Name, &start, &end, &c.Capacity, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return c, err
	}
	c.StartDate = strPtr(start)
	c.EndDate = strPtr(end)
	return c, nil
}

func (s *Service) enrichCycle(c *models.Cycle) error {
	var sum sql.NullInt64
	err := s.DB.QueryRow(`SELECT COALESCE(SUM(estimate),0) FROM issues WHERE cycle_id = ?`, c.ID).Scan(&sum)
	if err != nil {
		return err
	}
	c.EstimateSum = int(sum.Int64)
	c.OverCapacity = c.Capacity > 0 && c.EstimateSum > c.Capacity
	return nil
}

type CreateCycleInput struct {
	Name      string
	StartDate *string
	EndDate   *string
	Capacity  int
	Status    string
}

func (s *Service) ListCycles(workspaceID, projectID string) ([]models.Cycle, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if projectID != "" {
		if _, err := s.GetProject(workspaceID, projectID); err != nil {
			return nil, err
		}
		rows, err = s.DB.Query(`SELECT `+cycleCols+` FROM cycles WHERE project_id = ? ORDER BY created_at ASC`, projectID)
	} else {
		rows, err = s.DB.Query(`SELECT c.id, c.project_id, c.name, c.start_date, c.end_date, c.capacity, c.status, c.created_at, c.updated_at
			FROM cycles c
			JOIN projects p ON p.id = c.project_id
			WHERE p.workspace_id = ?
			ORDER BY c.created_at ASC`, workspaceID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.Cycle{}
	for rows.Next() {
		c, err := scanCycle(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range list {
		if err := s.enrichCycle(&list[i]); err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (s *Service) CreateCycle(workspaceID, projectID string, in CreateCycleInput) (*models.Cycle, error) {
	if _, err := s.GetProject(workspaceID, projectID); err != nil {
		return nil, err
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return nil, apperr.BadRequest("Cycle 名称不能为空")
	}
	status := in.Status
	if status == "" {
		status = "planned"
	}
	if !validCycleStatuses[status] {
		return nil, apperr.BadRequest("无效的 Cycle 状态")
	}
	id := idgen.New()
	now := nowUTC()
	_, err := s.DB.Exec(`INSERT INTO cycles (id, project_id, name, start_date, end_date, capacity, status, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?)`,
		id, projectID, in.Name, nullStr(in.StartDate), nullStr(in.EndDate), in.Capacity, status, now, now)
	if err != nil {
		return nil, err
	}
	return s.GetCycle(workspaceID, id)
}

func (s *Service) GetCycle(workspaceID, cycleID string) (*models.Cycle, error) {
	c, err := scanCycle(s.DB.QueryRow(`SELECT c.id, c.project_id, c.name, c.start_date, c.end_date, c.capacity, c.status, c.created_at, c.updated_at
		FROM cycles c JOIN projects p ON p.id = c.project_id
		WHERE c.id = ? AND p.workspace_id = ?`, cycleID, workspaceID))
	if err == sql.ErrNoRows {
		return nil, apperr.NotFound("Cycle 不存在")
	}
	if err != nil {
		return nil, err
	}
	if err := s.enrichCycle(&c); err != nil {
		return nil, err
	}
	return &c, nil
}

type UpdateCycleInput struct {
	Name       *string
	StartDate  *string
	ClearStart bool
	EndDate    *string
	ClearEnd   bool
	Capacity   *int
	Status     *string
	ActorID    string
	ActorKind  string
}

func (s *Service) UpdateCycle(workspaceID, cycleID string, in UpdateCycleInput) (*models.Cycle, error) {
	cur, err := s.GetCycle(workspaceID, cycleID)
	if err != nil {
		return nil, err
	}
	name, status := cur.Name, cur.Status
	start, end := cur.StartDate, cur.EndDate
	capacity := cur.Capacity
	completing := false

	if in.Name != nil {
		name = strings.TrimSpace(*in.Name)
		if name == "" {
			return nil, apperr.BadRequest("Cycle 名称不能为空")
		}
	}
	if in.ClearStart {
		start = nil
	} else if in.StartDate != nil {
		start = in.StartDate
	}
	if in.ClearEnd {
		end = nil
	} else if in.EndDate != nil {
		end = in.EndDate
	}
	if in.Capacity != nil {
		capacity = *in.Capacity
	}
	if in.Status != nil {
		if !validCycleStatuses[*in.Status] {
			return nil, apperr.BadRequest("无效的 Cycle 状态")
		}
		if *in.Status == "completed" && cur.Status != "completed" {
			completing = true
		}
		status = *in.Status
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := nowUTC()
	_, err = tx.Exec(`UPDATE cycles SET name=?, start_date=?, end_date=?, capacity=?, status=?, updated_at=? WHERE id=?`,
		name, nullStr(start), nullStr(end), capacity, status, now, cycleID)
	if err != nil {
		return nil, err
	}

	if completing {
		rows, err := tx.Query(`SELECT id, identifier, status FROM issues WHERE cycle_id = ? AND status NOT IN ('done','canceled')`, cycleID)
		if err != nil {
			return nil, err
		}
		type moved struct {
			ID, Identifier, Status string
		}
		var toMove []moved
		for rows.Next() {
			var m moved
			if err := rows.Scan(&m.ID, &m.Identifier, &m.Status); err != nil {
				rows.Close()
				return nil, err
			}
			toMove = append(toMove, m)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}
		for _, m := range toMove {
			_, err = tx.Exec(`UPDATE issues SET status='backlog', cycle_id=NULL, version=version+1, updated_at=? WHERE id=?`, now, m.ID)
			if err != nil {
				return nil, err
			}
			if err := appendActivity(tx, workspaceID, m.ID, in.ActorID, in.ActorKind, "issue.cycle_completed_moved", map[string]any{
				"cycle_id": cycleID, "from_status": m.Status, "to_status": "backlog", "identifier": m.Identifier,
			}); err != nil {
				return nil, err
			}
		}
		if err := appendActivity(tx, workspaceID, "", in.ActorID, in.ActorKind, "cycle.completed", map[string]any{
			"cycle_id": cycleID, "moved_count": len(toMove),
		}); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.GetCycle(workspaceID, cycleID)
}
