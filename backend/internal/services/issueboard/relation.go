package issueboard

import (
	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"

	"database/sql"
	"strings"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/models"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
)

const relationCols = `id, workspace_id, from_issue_id, to_issue_id, relation_type, created_at`

func scanRelation(row interface{ Scan(dest ...any) error }) (models.IssueRelation, error) {
	var r models.IssueRelation
	err := row.Scan(&r.ID, &r.WorkspaceID, &r.FromIssueID, &r.ToIssueID, &r.RelationType, &r.CreatedAt)
	return r, err
}

// ListRelations returns edges touching issueID.
// Default (includeArchived=false): omit edges whose from or to issue is archived (WHEEL_FIX B3).
// includeArchived=true: return all edges (no physical delete).
func (s *Service) ListRelations(workspaceID, issueID string, includeArchived bool) ([]models.IssueRelation, error) {
	if err := s.ensureIssueInWorkspace(workspaceID, issueID); err != nil {
		return nil, err
	}
	q := `SELECT r.id, r.workspace_id, r.from_issue_id, r.to_issue_id, r.relation_type, r.created_at
		FROM issue_relations r`
	args := []any{workspaceID, issueID, issueID}
	if !includeArchived {
		q += `
		JOIN issues f ON f.id = r.from_issue_id AND f.workspace_id = r.workspace_id
		JOIN issues t ON t.id = r.to_issue_id AND t.workspace_id = r.workspace_id`
		q += `
		WHERE r.workspace_id = ? AND (r.from_issue_id = ? OR r.to_issue_id = ?)
		  AND f.archived_at IS NULL AND t.archived_at IS NULL`
	} else {
		q += `
		WHERE r.workspace_id = ? AND (r.from_issue_id = ? OR r.to_issue_id = ?)`
	}
	q += `
		ORDER BY r.created_at ASC`
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.IssueRelation{}
	for rows.Next() {
		rel, err := scanRelation(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, rel)
	}
	return list, rows.Err()
}

type CreateRelationInput struct {
	ToIssueID    string
	RelationType string
	ActorID      string
	ActorKind    string
}

func (s *Service) CreateRelation(workspaceID, fromIssueID string, in CreateRelationInput) (*models.IssueRelation, error) {
	in.RelationType = strings.TrimSpace(in.RelationType)
	in.ToIssueID = strings.TrimSpace(in.ToIssueID)
	if in.ToIssueID == "" {
		return nil, apperr.BadRequest("缺少 to_issue_id")
	}
	if !validRelationTypes[in.RelationType] {
		return nil, apperr.BadRequest("无效的关联类型")
	}
	if fromIssueID == in.ToIssueID {
		return nil, apperr.BadRequest("不能将 Issue 关联到自身")
	}
	if err := s.ensureIssueInWorkspace(workspaceID, fromIssueID); err != nil {
		return nil, err
	}
	if err := s.ensureIssueInWorkspace(workspaceID, in.ToIssueID); err != nil {
		return nil, err
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Cycle check runs inside the tx so concurrent inserts cannot both pass then form a cycle.
	if in.RelationType == "blocks" || in.RelationType == "blocked_by" || in.RelationType == "parent" {
		if err := checkRelationCycle(tx, workspaceID, fromIssueID, in.ToIssueID, in.RelationType); err != nil {
			return nil, err
		}
	}

	id := idgen.New()
	now := nowUTC()
	_, err = tx.Exec(`INSERT INTO issue_relations
		(id, workspace_id, from_issue_id, to_issue_id, relation_type, created_at)
		VALUES (?,?,?,?,?,?)`,
		id, workspaceID, fromIssueID, in.ToIssueID, in.RelationType, now)
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, apperr.Conflict("关联关系已存在")
		}
		return nil, err
	}
	if err := appendActivity(tx, workspaceID, fromIssueID, in.ActorID, in.ActorKind, "relation.created", map[string]any{
		"relation_id":   id,
		"to_issue_id":   in.ToIssueID,
		"relation_type": in.RelationType,
	}); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &models.IssueRelation{
		ID: id, WorkspaceID: workspaceID,
		FromIssueID: fromIssueID, ToIssueID: in.ToIssueID,
		RelationType: in.RelationType, CreatedAt: now,
	}, nil
}

func (s *Service) DeleteRelation(workspaceID, relationID, actorID, actorKind string) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var fromID, toID, relType string
	err = tx.QueryRow(`SELECT from_issue_id, to_issue_id, relation_type FROM issue_relations
		WHERE id=? AND workspace_id=?`, relationID, workspaceID).Scan(&fromID, &toID, &relType)
	if err == sql.ErrNoRows {
		return apperr.NotFound("关联关系不存在")
	}
	if err != nil {
		return err
	}
	if err := appendActivity(tx, workspaceID, fromID, actorID, actorKind, "relation.deleted", map[string]any{
		"relation_id":   relationID,
		"to_issue_id":   toID,
		"relation_type": relType,
	}); err != nil {
		return err
	}
	res, err := tx.Exec(`DELETE FROM issue_relations WHERE id=? AND workspace_id=?`, relationID, workspaceID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperr.NotFound("关联关系不存在")
	}
	return tx.Commit()
}

// DeleteRelationOnIssue deletes a relation that belongs to the given from_issue (or either side).
func (s *Service) DeleteRelationOnIssue(workspaceID, issueID, relationID, actorID, actorKind string) error {
	var fromID, toID string
	err := s.DB.QueryRow(`SELECT from_issue_id, to_issue_id FROM issue_relations
		WHERE id=? AND workspace_id=?`, relationID, workspaceID).Scan(&fromID, &toID)
	if err == sql.ErrNoRows {
		return apperr.NotFound("关联关系不存在")
	}
	if err != nil {
		return err
	}
	if fromID != issueID && toID != issueID {
		return apperr.NotFound("关联关系不存在")
	}
	return s.DeleteRelation(workspaceID, relationID, actorID, actorKind)
}

// normalizeBlocksEdge maps a relation to a directed "blocks" edge (from blocks to).
// blocked_by is the reverse of blocks.
func normalizeBlocksEdge(from, to, relType string) (edgeFrom, edgeTo string, ok bool) {
	switch relType {
	case "blocks", "parent":
		return from, to, true
	case "blocked_by":
		return to, from, true
	default:
		return "", "", false
	}
}

// checkRelationCycle rejects if adding from→to (normalized) would create a cycle among
// blocks / blocked_by / parent edges. Message must contain「依赖成环」.
// q should be the open *sql.Tx (or DB) so concurrent CreateRelation cannot race.
func checkRelationCycle(q querier, workspaceID, fromID, toID, relType string) error {
	newFrom, newTo, ok := normalizeBlocksEdge(fromID, toID, relType)
	if !ok {
		return nil
	}
	if newFrom == newTo {
		return apperr.Conflict("依赖成环：不能形成自环依赖")
	}

	rows, err := q.Query(`SELECT from_issue_id, to_issue_id, relation_type FROM issue_relations
		WHERE workspace_id = ? AND relation_type IN ('blocks','blocked_by','parent')`, workspaceID)
	if err != nil {
		return err
	}
	defer rows.Close()

	adj := map[string][]string{}
	for rows.Next() {
		var f, t, rt string
		if err := rows.Scan(&f, &t, &rt); err != nil {
			return err
		}
		ef, et, ok := normalizeBlocksEdge(f, t, rt)
		if !ok {
			continue
		}
		adj[ef] = append(adj[ef], et)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Path from newTo back to newFrom ⇒ cycle after adding newFrom→newTo.
	if hasPath(adj, newTo, newFrom) {
		return apperr.Conflict("依赖成环：blocks/parent 关系形成环路")
	}
	return nil
}

func hasPath(adj map[string][]string, start, target string) bool {
	if start == target {
		return true
	}
	seen := map[string]bool{start: true}
	queue := []string{start}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, next := range adj[cur] {
			if next == target {
				return true
			}
			if seen[next] {
				continue
			}
			seen[next] = true
			queue = append(queue, next)
		}
	}
	return false
}
