package issueboard

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/models"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
)

// issueCols — parent_id is the tree authority (issue_relations "parent" is NOT).
const issueCols = `id, project_id, workspace_id, team_id, number, identifier, title, description, status, priority,
	estimate, assignee_id, reporter_id, cycle_id, parent_id, start_at, due_at, board_order, version, created_at, updated_at`

func scanIssue(row interface{ Scan(dest ...any) error }) (models.Issue, error) {
	var iss models.Issue
	var est sql.NullInt64
	var teamID, assignee, reporter, cycle, parent, startAt, dueAt sql.NullString
	err := row.Scan(
		&iss.ID, &iss.ProjectID, &iss.WorkspaceID, &teamID, &iss.Number, &iss.Identifier,
		&iss.Title, &iss.Description, &iss.Status, &iss.Priority,
		&est, &assignee, &reporter, &cycle, &parent, &startAt, &dueAt,
		&iss.BoardOrder, &iss.Version, &iss.CreatedAt, &iss.UpdatedAt,
	)
	if err != nil {
		return iss, err
	}
	iss.TeamID = strPtr(teamID)
	iss.Estimate = intPtr(est)
	iss.AssigneeID = strPtr(assignee)
	iss.ReporterID = strPtr(reporter)
	iss.CycleID = strPtr(cycle)
	iss.ParentID = strPtr(parent)
	iss.StartAt = strPtr(startAt)
	iss.DueAt = strPtr(dueAt)
	return iss, nil
}

// validateCalendarDay accepts only YYYY-MM-DD (calendar day, not a TZ instant).
func validateCalendarDay(field string, v *string) error {
	if v == nil {
		return nil
	}
	s := strings.TrimSpace(*v)
	if s == "" {
		return nil
	}
	if len(s) != 10 || s[4] != '-' || s[7] != '-' {
		return apperr.BadRequest(field + " 必须是 YYYY-MM-DD 格式的日历日")
	}
	if _, err := time.Parse("2006-01-02", s); err != nil {
		return apperr.BadRequest(field + " 必须是 YYYY-MM-DD 格式的日历日")
	}
	*v = s
	return nil
}

func validateDateRange(startAt, dueAt *string) error {
	if startAt == nil || dueAt == nil {
		return nil
	}
	if *startAt == "" || *dueAt == "" {
		return nil
	}
	if *dueAt < *startAt {
		return apperr.BadRequest("截止日期 due_at 不能早于开始日期 start_at")
	}
	return nil
}

func normalizeOptionalDay(v *string) *string {
	if v == nil {
		return nil
	}
	s := strings.TrimSpace(*v)
	if s == "" {
		return nil
	}
	return &s
}

type CreateIssueInput struct {
	ProjectID   string
	TeamID      string // 新建必填；须属本工作区
	Title       string
	Description string
	Status      string
	Priority    string
	Estimate    *int
	AssigneeID  *string
	CycleID     *string
	ParentID    *string
	StartAt     *string
	DueAt       *string
	BoardOrder  *float64
	LabelIDs    []string
	ActorID     string
	ActorKind   string
	// SkipNotify 为 true 时不扇出 Webhook（下发 Modal 未勾「立即推送」）
	SkipNotify bool
}

func (s *Service) CreateIssue(workspaceID string, in CreateIssueInput) (*models.Issue, error) {
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		return nil, apperr.BadRequest("标题不能为空")
	}
	if in.ProjectID == "" {
		return nil, apperr.BadRequest("缺少 project_id")
	}
	in.TeamID = strings.TrimSpace(in.TeamID)
	if in.TeamID == "" {
		return nil, apperr.BadRequest("创建 Issue 必须指定 team_id（当前团队）")
	}
	if err := s.ensureTeamInWorkspace(workspaceID, in.TeamID); err != nil {
		return nil, err
	}
	proj, err := s.GetProject(workspaceID, in.ProjectID)
	if err != nil {
		return nil, err
	}
	status := in.Status
	if status == "" {
		status = "backlog"
	}
	if !validStatuses[status] {
		return nil, apperr.BadRequest("无效的状态")
	}
	if in.ActorKind == "bot" && status == "done" {
		return nil, apperr.Forbidden("Bot 不能单独将 Issue 设为完成（done），请设为 in_review 并由人工确认")
	}
	// A7: Bot 不能创建无父节点的游离 Issue
	if in.ActorKind == "bot" && (in.ParentID == nil || strings.TrimSpace(*in.ParentID) == "") {
		return nil, apperr.Forbidden("Bot 不能创建无父节点的游离 Issue")
	}
	priority := in.Priority
	if priority == "" {
		priority = "none"
	}
	if !validPriorities[priority] {
		return nil, apperr.BadRequest("无效的优先级")
	}
	if in.CycleID != nil && *in.CycleID != "" {
		if err := s.ensureCycleInProject(workspaceID, *in.CycleID, in.ProjectID); err != nil {
			return nil, err
		}
	}
	if in.ParentID != nil && *in.ParentID != "" {
		if err := s.ensureIssueInWorkspace(workspaceID, *in.ParentID); err != nil {
			return nil, err
		}
	}
	if in.AssigneeID != nil && *in.AssigneeID != "" {
		if err := s.ensureMember(workspaceID, *in.AssigneeID); err != nil {
			return nil, err
		}
	}
	if err := validateCalendarDay("start_at", in.StartAt); err != nil {
		return nil, err
	}
	if err := validateCalendarDay("due_at", in.DueAt); err != nil {
		return nil, err
	}
	in.StartAt = normalizeOptionalDay(in.StartAt)
	in.DueAt = normalizeOptionalDay(in.DueAt)
	if err := validateDateRange(in.StartAt, in.DueAt); err != nil {
		return nil, err
	}

	boardOrder := 0.0
	if in.BoardOrder != nil {
		boardOrder = *in.BoardOrder
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var next int
	counterQ := `SELECT next_number FROM issue_counters WHERE project_id = ?`
	if s.postgres {
		counterQ += ` FOR UPDATE`
	}
	err = tx.QueryRow(counterQ, in.ProjectID).Scan(&next)
	if err == sql.ErrNoRows {
		next = 1
		if _, err = tx.Exec(`INSERT INTO issue_counters (project_id, next_number) VALUES (?, 2)`, in.ProjectID); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		if _, err = tx.Exec(`UPDATE issue_counters SET next_number = next_number + 1 WHERE project_id = ?`, in.ProjectID); err != nil {
			return nil, err
		}
	}

	id := idgen.New()
	identifier := fmt.Sprintf("%s-%d", proj.KeyPrefix, next)
	now := nowUTC()
	var reporter any
	if in.ActorID != "" {
		reporter = in.ActorID
	}
	_, err = tx.Exec(`INSERT INTO issues
		(id, project_id, workspace_id, team_id, number, identifier, title, description, status, priority,
		 estimate, assignee_id, reporter_id, cycle_id, parent_id, start_at, due_at, board_order, version, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,1,?,?)`,
		id, in.ProjectID, workspaceID, in.TeamID, next, identifier, in.Title, in.Description, status, priority,
		nullInt(in.Estimate), nullStr(in.AssigneeID), reporter, nullStr(in.CycleID), nullStr(in.ParentID),
		nullStr(in.StartAt), nullStr(in.DueAt), boardOrder, now, now)
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, apperr.Conflict("Issue 标识符冲突，请重试")
		}
		return nil, err
	}
	if err := s.replaceLabelsTx(tx, workspaceID, id, in.LabelIDs); err != nil {
		return nil, err
	}
	if err := appendActivity(tx, workspaceID, id, in.ActorID, in.ActorKind, "issue.created", map[string]any{
		"identifier": identifier, "title": in.Title, "status": status,
	}); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	iss, err := s.GetIssue(workspaceID, id)
	if err != nil {
		return nil, err
	}
	if !in.SkipNotify {
		s.emit(workspaceID, "issue.created", iss)
		if iss.AssigneeID != nil && strings.TrimSpace(*iss.AssigneeID) != "" {
			s.emit(workspaceID, "issue.assigned", iss)
		}
	}
	return iss, nil
}

type ListIssuesFilter struct {
	ProjectID      string
	TeamID         string // 非空：本队 + team_id IS NULL（全 WS 可见旧单）
	UnassignedOnly bool   // 仅 team_id IS NULL（Admin「未分队」）
	Status         string
	AssigneeID     string
	LabelID        string
	CycleID        string
	Q              string
	LeafOnly       bool
	Blocked        bool // PRODUCT_SUITE: has unfinished blocked_by relation
	Overdue        bool // PRODUCT_SUITE: due_at set and due_at < today (YYYY-MM-DD Asia/Shanghai)
}

func (s *Service) ListIssues(workspaceID string, f ListIssuesFilter) ([]models.Issue, error) {
	q := `SELECT DISTINCT i.id, i.project_id, i.workspace_id, i.team_id, i.number, i.identifier, i.title, i.description,
		i.status, i.priority, i.estimate, i.assignee_id, i.reporter_id, i.cycle_id, i.parent_id,
		i.start_at, i.due_at, i.board_order, i.version, i.created_at, i.updated_at
		FROM issues i`
	args := []any{}
	conds := []string{`i.workspace_id = ?`}
	args = append(args, workspaceID)
	if f.UnassignedOnly {
		conds = append(conds, `i.team_id IS NULL`)
	} else if f.TeamID != "" {
		// T6-I：旧单 NULL = 全工作区可见；队 A 的单不会出现在队 B 的 list
		conds = append(conds, `(i.team_id = ? OR i.team_id IS NULL)`)
		args = append(args, f.TeamID)
	}
	if f.LabelID != "" {
		q += ` JOIN issue_labels il ON il.issue_id = i.id`
		conds = append(conds, `il.label_id = ?`)
		args = append(args, f.LabelID)
	}
	if f.ProjectID != "" {
		conds = append(conds, `i.project_id = ?`)
		args = append(args, f.ProjectID)
	}
	if f.Status != "" {
		conds = append(conds, `i.status = ?`)
		args = append(args, f.Status)
	}
	if f.AssigneeID != "" {
		conds = append(conds, `i.assignee_id = ?`)
		args = append(args, f.AssigneeID)
	}
	if f.CycleID != "" {
		conds = append(conds, `i.cycle_id = ?`)
		args = append(args, f.CycleID)
	}
	if f.Q != "" {
		conds = append(conds, `(i.title LIKE ? OR i.identifier LIKE ? OR i.description LIKE ?)`)
		like := "%" + f.Q + "%"
		args = append(args, like, like, like)
	}
	if f.LeafOnly {
		// parent_id is tree authority — leaf = no children with this parent_id
		conds = append(conds, `NOT EXISTS (SELECT 1 FROM issues c WHERE c.parent_id = i.id AND c.workspace_id = i.workspace_id)`)
	}
	if f.Blocked {
		// Issue i is blocked if an unfinished blocker blocks it
		// (blocks: blocker→i, or blocked_by: i→blocker).
		conds = append(conds, `EXISTS (
			SELECT 1 FROM issue_relations r
			JOIN issues blocker ON (
				(r.relation_type = 'blocks' AND r.from_issue_id = blocker.id AND r.to_issue_id = i.id)
				OR (r.relation_type = 'blocked_by' AND r.to_issue_id = blocker.id AND r.from_issue_id = i.id)
			)
			WHERE r.workspace_id = i.workspace_id
			  AND blocker.status NOT IN ('done','canceled')
		)`)
	}
	if f.Overdue {
		loc, err := time.LoadLocation("Asia/Shanghai")
		if err != nil {
			loc = time.FixedZone("CST", 8*3600)
		}
		today := time.Now().In(loc).Format("2006-01-02")
		conds = append(conds, `i.due_at IS NOT NULL AND i.due_at != '' AND i.due_at < ?`)
		args = append(args, today)
	}
	q += ` WHERE ` + strings.Join(conds, " AND ") + ` ORDER BY i.board_order ASC, i.number ASC`
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.Issue{}
	for rows.Next() {
		iss, err := scanIssue(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, iss)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := s.attachChildCounts(workspaceID, list); err != nil {
		return nil, err
	}
	for i := range list {
		labels, err := s.labelsForIssue(list[i].ID)
		if err != nil {
			return nil, err
		}
		list[i].Labels = labels
		ids := make([]string, len(labels))
		for j, l := range labels {
			ids[j] = l.ID
		}
		list[i].LabelIDs = ids
	}
	if err := s.attachMemberRefs(workspaceID, list); err != nil {
		return nil, err
	}
	return list, nil
}

// attachChildCounts batches child_count / is_leaf for a list (no N+1).
// Tree authority is parent_id, not issue_relations.
func (s *Service) attachChildCounts(workspaceID string, list []models.Issue) error {
	if len(list) == 0 {
		return nil
	}
	ids := make([]string, len(list))
	idx := make(map[string]int, len(list))
	for i := range list {
		ids[i] = list[i].ID
		idx[list[i].ID] = i
		list[i].ChildCount = 0
		list[i].IsLeaf = true
	}
	placeholders := make([]string, len(ids))
	args := make([]any, 0, len(ids)+1)
	args = append(args, workspaceID)
	for i, id := range ids {
		placeholders[i] = "?"
		args = append(args, id)
	}
	q := `SELECT parent_id, COUNT(*) FROM issues
		WHERE workspace_id = ? AND parent_id IN (` + strings.Join(placeholders, ",") + `)
		GROUP BY parent_id`
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var parentID string
		var n int
		if err := rows.Scan(&parentID, &n); err != nil {
			return err
		}
		if i, ok := idx[parentID]; ok {
			list[i].ChildCount = n
			list[i].IsLeaf = n == 0
		}
	}
	return rows.Err()
}

func (s *Service) attachChildCountOne(workspaceID string, iss *models.Issue) error {
	var n int
	err := s.DB.QueryRow(`SELECT COUNT(*) FROM issues WHERE workspace_id = ? AND parent_id = ?`,
		workspaceID, iss.ID).Scan(&n)
	if err != nil {
		return err
	}
	iss.ChildCount = n
	iss.IsLeaf = n == 0
	return nil
}

// attachMemberRefs batch-loads members.display_name for assignee/reporter (no N+1).
// Missing member rows still get {id, display_name:""}; null id → null nested object.
func (s *Service) attachMemberRefs(workspaceID string, list []models.Issue) error {
	if len(list) == 0 {
		return nil
	}
	idSet := map[string]struct{}{}
	for i := range list {
		if list[i].AssigneeID != nil && strings.TrimSpace(*list[i].AssigneeID) != "" {
			idSet[*list[i].AssigneeID] = struct{}{}
		}
		if list[i].ReporterID != nil && strings.TrimSpace(*list[i].ReporterID) != "" {
			idSet[*list[i].ReporterID] = struct{}{}
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
		if list[i].AssigneeID != nil && strings.TrimSpace(*list[i].AssigneeID) != "" {
			id := *list[i].AssigneeID
			list[i].Assignee = &models.MemberRef{ID: id, DisplayName: names[id]}
		} else {
			list[i].Assignee = nil
		}
		if list[i].ReporterID != nil && strings.TrimSpace(*list[i].ReporterID) != "" {
			id := *list[i].ReporterID
			list[i].Reporter = &models.MemberRef{ID: id, DisplayName: names[id]}
		} else {
			list[i].Reporter = nil
		}
	}
	return nil
}

func (s *Service) GetIssue(workspaceID, issueID string) (*models.Issue, error) {
	return s.GetIssueInTeam(workspaceID, issueID, "")
}

// GetIssueInTeam 读取 Issue；teamID 非空时错队（已分队且非本队）→ 404。NULL team_id 全 WS 可见。
func (s *Service) GetIssueInTeam(workspaceID, issueID, teamID string) (*models.Issue, error) {
	iss, err := scanIssue(s.DB.QueryRow(`SELECT `+issueCols+` FROM issues WHERE id = ? AND workspace_id = ?`, issueID, workspaceID))
	if err == sql.ErrNoRows {
		return nil, apperr.NotFound("Issue 不存在")
	}
	if err != nil {
		return nil, err
	}
	if err := s.assertIssueTeamAccess(&iss, teamID); err != nil {
		return nil, err
	}
	if err := s.attachChildCountOne(workspaceID, &iss); err != nil {
		return nil, err
	}
	labels, err := s.labelsForIssue(iss.ID)
	if err != nil {
		return nil, err
	}
	iss.Labels = labels
	ids := make([]string, len(labels))
	for i, l := range labels {
		ids[i] = l.ID
	}
	iss.LabelIDs = ids
	one := []models.Issue{iss}
	if err := s.attachMemberRefs(workspaceID, one); err != nil {
		return nil, err
	}
	iss = one[0]
	return &iss, nil
}

// UpdateIssueInput — Version 必填；指针字段表示要更新。
// ClearAssignee/ClearCycle/ClearParent/ClearStartAt/ClearDueAt 用于显式清空。
type UpdateIssueInput struct {
	Version       int
	Title         *string
	Description   *string
	Status        *string
	Priority      *string
	Estimate      *int
	ClearEstimate bool
	AssigneeID    *string
	ClearAssignee bool
	CycleID       *string
	ClearCycle    bool
	ParentID      *string
	ClearParent   bool
	StartAt       *string
	ClearStartAt  bool
	DueAt         *string
	ClearDueAt    bool
	BoardOrder    *float64
	LabelIDs      *[]string
	TeamID        *string // 可写换队（须属本 WS）
	RequireTeamID string  // 非空：错队（issue.team_id 属其它队）→ 404
	ActorID       string
	ActorKind     string
	ForceAssignee bool
}

func (s *Service) UpdateIssue(workspaceID, issueID string, in UpdateIssueInput) (*models.Issue, error) {
	if in.Version < 1 {
		return nil, apperr.BadRequest("更新 Issue 必须提供 version（乐观锁）")
	}
	cur, err := s.GetIssue(workspaceID, issueID)
	if err != nil {
		return nil, err
	}
	if err := s.assertIssueTeamAccess(cur, in.RequireTeamID); err != nil {
		return nil, err
	}
	if cur.Version != in.Version {
		return nil, apperr.VersionConflict("乐观锁冲突：Issue 已被他人更新，请刷新后重试")
	}

	// Claim gate: assigning (non-empty assignee) only allowed on leaf issues.
	// ForceAssignee=false 时禁止抢占已有异人负责人。
	if in.AssigneeID != nil && strings.TrimSpace(*in.AssigneeID) != "" {
		if !cur.IsLeaf {
			return nil, apperr.Conflict("仅叶子任务可领取")
		}
		if cur.AssigneeID != nil && strings.TrimSpace(*cur.AssigneeID) != "" &&
			*cur.AssigneeID != strings.TrimSpace(*in.AssigneeID) && !in.ForceAssignee {
			return nil, apperr.Conflict("该 Issue 已被其他成员领取，禁止抢占")
		}
	}

	title, desc, status, priority := cur.Title, cur.Description, cur.Status, cur.Priority
	estimate := cur.Estimate
	assignee, cycle, parent := cur.AssigneeID, cur.CycleID, cur.ParentID
	teamID := cur.TeamID
	startAt, dueAt := cur.StartAt, cur.DueAt
	boardOrder := cur.BoardOrder
	changes := map[string]any{}

	if in.Title != nil {
		t := strings.TrimSpace(*in.Title)
		if t == "" {
			return nil, apperr.BadRequest("标题不能为空")
		}
		if t != title {
			changes["title"] = map[string]any{"from": title, "to": t}
			title = t
		}
	}
	if in.Description != nil && *in.Description != desc {
		changes["description"] = true
		desc = *in.Description
	}
	if in.Status != nil {
		if !validStatuses[*in.Status] {
			return nil, apperr.BadRequest("无效的状态")
		}
		if in.ActorKind == "bot" && *in.Status == "done" {
			return nil, apperr.Forbidden("Bot 不能单独将 Issue 设为完成（done），请设为 in_review 并由人工确认")
		}
		if *in.Status != status {
			changes["status"] = map[string]any{"from": status, "to": *in.Status}
			status = *in.Status
		}
	}
	if in.Priority != nil {
		if !validPriorities[*in.Priority] {
			return nil, apperr.BadRequest("无效的优先级")
		}
		if *in.Priority != priority {
			changes["priority"] = map[string]any{"from": priority, "to": *in.Priority}
			priority = *in.Priority
		}
	}
	if in.ClearEstimate {
		if estimate != nil {
			changes["estimate"] = map[string]any{"from": *estimate, "to": nil}
		}
		estimate = nil
	} else if in.Estimate != nil {
		if estimate == nil || *estimate != *in.Estimate {
			from := any(nil)
			if estimate != nil {
				from = *estimate
			}
			changes["estimate"] = map[string]any{"from": from, "to": *in.Estimate}
			estimate = in.Estimate
		}
	}
	if in.ClearAssignee {
		if assignee != nil {
			changes["assignee_id"] = map[string]any{"from": *assignee, "to": nil}
		}
		assignee = nil
	} else if in.AssigneeID != nil {
		if *in.AssigneeID != "" {
			if err := s.ensureMember(workspaceID, *in.AssigneeID); err != nil {
				return nil, err
			}
		}
		prev := ""
		if assignee != nil {
			prev = *assignee
		}
		if prev != *in.AssigneeID {
			changes["assignee_id"] = map[string]any{"from": assignee, "to": *in.AssigneeID}
			if *in.AssigneeID == "" {
				assignee = nil
			} else {
				assignee = in.AssigneeID
			}
		}
	}
	if in.ClearCycle {
		if cycle != nil {
			changes["cycle_id"] = map[string]any{"from": *cycle, "to": nil}
		}
		cycle = nil
	} else if in.CycleID != nil {
		if *in.CycleID != "" {
			if err := s.ensureCycleInProject(workspaceID, *in.CycleID, cur.ProjectID); err != nil {
				return nil, err
			}
		}
		prev := ""
		if cycle != nil {
			prev = *cycle
		}
		if prev != *in.CycleID {
			changes["cycle_id"] = map[string]any{"from": cycle, "to": *in.CycleID}
			if *in.CycleID == "" {
				cycle = nil
			} else {
				cycle = in.CycleID
			}
		}
	}
	if in.ClearParent {
		if parent != nil {
			changes["parent_id"] = map[string]any{"from": *parent, "to": nil}
		}
		parent = nil
	} else if in.ParentID != nil {
		if *in.ParentID != "" {
			if *in.ParentID == issueID {
				return nil, apperr.BadRequest("不能将自己设为父 Issue")
			}
			if err := s.ensureIssueInWorkspace(workspaceID, *in.ParentID); err != nil {
				return nil, err
			}
		}
		prev := ""
		if parent != nil {
			prev = *parent
		}
		if prev != *in.ParentID {
			changes["parent_id"] = map[string]any{"from": parent, "to": *in.ParentID}
			if *in.ParentID == "" {
				parent = nil
			} else {
				parent = in.ParentID
			}
		}
	}
	if in.ClearStartAt {
		if startAt != nil {
			changes["start_at"] = map[string]any{"from": *startAt, "to": nil}
		}
		startAt = nil
	} else if in.StartAt != nil {
		if err := validateCalendarDay("start_at", in.StartAt); err != nil {
			return nil, err
		}
		norm := normalizeOptionalDay(in.StartAt)
		prev := ""
		if startAt != nil {
			prev = *startAt
		}
		next := ""
		if norm != nil {
			next = *norm
		}
		if prev != next {
			changes["start_at"] = map[string]any{"from": startAt, "to": norm}
			startAt = norm
		}
	}
	if in.ClearDueAt {
		if dueAt != nil {
			changes["due_at"] = map[string]any{"from": *dueAt, "to": nil}
		}
		dueAt = nil
	} else if in.DueAt != nil {
		if err := validateCalendarDay("due_at", in.DueAt); err != nil {
			return nil, err
		}
		norm := normalizeOptionalDay(in.DueAt)
		prev := ""
		if dueAt != nil {
			prev = *dueAt
		}
		next := ""
		if norm != nil {
			next = *norm
		}
		if prev != next {
			changes["due_at"] = map[string]any{"from": dueAt, "to": norm}
			dueAt = norm
		}
	}
	if err := validateDateRange(startAt, dueAt); err != nil {
		return nil, err
	}
	if in.TeamID != nil {
		tid := strings.TrimSpace(*in.TeamID)
		if tid == "" {
			return nil, apperr.BadRequest("team_id 不能为空（清空请联系管理员处理未分队旧单）")
		}
		if err := s.ensureTeamInWorkspace(workspaceID, tid); err != nil {
			return nil, err
		}
		prev := ""
		if teamID != nil {
			prev = *teamID
		}
		if prev != tid {
			changes["team_id"] = map[string]any{"from": teamID, "to": tid}
			teamID = &tid
		}
	}
	if in.BoardOrder != nil && *in.BoardOrder != boardOrder {
		changes["board_order"] = map[string]any{"from": boardOrder, "to": *in.BoardOrder}
		boardOrder = *in.BoardOrder
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Parent cycle check inside the update tx when parent actually changes.
	if _, ok := changes["parent_id"]; ok && parent != nil && *parent != "" {
		if err := checkParentIDCycle(tx, workspaceID, issueID, *parent); err != nil {
			return nil, err
		}
	}

	now := nowUTC()
	res, err := tx.Exec(`UPDATE issues SET title=?, description=?, status=?, priority=?, estimate=?,
		assignee_id=?, cycle_id=?, parent_id=?, team_id=?, start_at=?, due_at=?, board_order=?, version=version+1, updated_at=?
		WHERE id=? AND workspace_id=? AND version=?`,
		title, desc, status, priority, nullInt(estimate),
		nullStr(assignee), nullStr(cycle), nullStr(parent), nullStr(teamID), nullStr(startAt), nullStr(dueAt), boardOrder, now,
		issueID, workspaceID, in.Version)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, apperr.VersionConflict("乐观锁冲突：Issue 已被他人更新，请刷新后重试")
	}
	if in.LabelIDs != nil {
		if err := s.replaceLabelsTx(tx, workspaceID, issueID, *in.LabelIDs); err != nil {
			return nil, err
		}
		changes["label_ids"] = *in.LabelIDs
	}
	if len(changes) > 0 {
		action := "issue.updated"
		if _, ok := changes["status"]; ok {
			action = "issue.status_changed"
		}
		if err := appendActivity(tx, workspaceID, issueID, in.ActorID, in.ActorKind, action, changes); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	iss, err := s.GetIssue(workspaceID, issueID)
	if err != nil {
		return nil, err
	}
	if len(changes) > 0 {
		s.emit(workspaceID, "issue.updated", map[string]any{"issue": iss, "changes": changes})
		if _, ok := changes["assignee_id"]; ok {
			s.emit(workspaceID, "issue.assigned", iss)
		}
	}
	return iss, nil
}

func (s *Service) DeleteIssue(workspaceID, issueID, actorID, actorKind string) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var ident string
	err = tx.QueryRow(`SELECT identifier FROM issues WHERE id=? AND workspace_id=?`, issueID, workspaceID).Scan(&ident)
	if err == sql.ErrNoRows {
		return apperr.NotFound("Issue 不存在")
	}
	if err != nil {
		return err
	}
	if err := appendActivity(tx, workspaceID, issueID, actorID, actorKind, "issue.deleted", map[string]any{"identifier": ident}); err != nil {
		return err
	}
	res, err := tx.Exec(`DELETE FROM issues WHERE id=? AND workspace_id=?`, issueID, workspaceID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperr.NotFound("Issue 不存在")
	}
	return tx.Commit()
}

// assertIssueTeamAccess：RequireTeamID 非空且 issue 已分到其它队 → 404（不泄露存在性）。
// team_id IS NULL（旧单）视为全工作区可见。
func (s *Service) assertIssueTeamAccess(iss *models.Issue, requireTeamID string) error {
	requireTeamID = strings.TrimSpace(requireTeamID)
	if requireTeamID == "" || iss == nil {
		return nil
	}
	if iss.TeamID == nil || strings.TrimSpace(*iss.TeamID) == "" {
		return nil
	}
	if *iss.TeamID != requireTeamID {
		return apperr.NotFound("Issue 不存在")
	}
	return nil
}

func (s *Service) ensureTeamInWorkspace(workspaceID, teamID string) error {
	var id string
	err := s.DB.QueryRow(`SELECT id FROM teams WHERE id=? AND workspace_id=?`, teamID, workspaceID).Scan(&id)
	if err == sql.ErrNoRows {
		return apperr.BadRequest("team_id 无效或不属于当前工作区")
	}
	return err
}

func (s *Service) ensureMember(workspaceID, memberID string) error {
	var id string
	err := s.DB.QueryRow(`SELECT id FROM members WHERE id=? AND workspace_id=?`, memberID, workspaceID).Scan(&id)
	if err == sql.ErrNoRows {
		return apperr.BadRequest("指派的成员不存在")
	}
	return err
}

func (s *Service) ensureIssueInWorkspace(workspaceID, issueID string) error {
	var id string
	err := s.DB.QueryRow(`SELECT id FROM issues WHERE id=? AND workspace_id=?`, issueID, workspaceID).Scan(&id)
	if err == sql.ErrNoRows {
		return apperr.NotFound("关联的 Issue 不存在")
	}
	return err
}

func (s *Service) ensureCycleInProject(workspaceID, cycleID, projectID string) error {
	var pid string
	err := s.DB.QueryRow(`SELECT c.project_id FROM cycles c
		JOIN projects p ON p.id = c.project_id
		WHERE c.id=? AND p.workspace_id=?`, cycleID, workspaceID).Scan(&pid)
	if err == sql.ErrNoRows {
		return apperr.NotFound("Cycle 不存在")
	}
	if err != nil {
		return err
	}
	if pid != projectID {
		return apperr.BadRequest("Cycle 不属于该 Issue 所在项目")
	}
	return nil
}

// checkParentIDCycle walks parent_id chain from newParent; if hits issueID → cycle.
// q should be the open *sql.Tx (or DB) so concurrent parent updates cannot race.
func checkParentIDCycle(q querier, workspaceID, issueID, newParent string) error {
	seen := map[string]bool{issueID: true}
	cur := newParent
	for cur != "" {
		if seen[cur] {
			return apperr.Conflict("依赖成环：父子关系形成环路")
		}
		seen[cur] = true
		var parent sql.NullString
		err := q.QueryRow(`SELECT parent_id FROM issues WHERE id=? AND workspace_id=?`, cur, workspaceID).Scan(&parent)
		if err == sql.ErrNoRows {
			return apperr.NotFound("父 Issue 不存在")
		}
		if err != nil {
			return err
		}
		if !parent.Valid {
			break
		}
		cur = parent.String
	}
	return nil
}

func (s *Service) labelsForIssue(issueID string) ([]models.Label, error) {
	rows, err := s.DB.Query(`SELECT l.id, l.workspace_id, l.name, l.color, l.created_at
		FROM labels l JOIN issue_labels il ON il.label_id = l.id
		WHERE il.issue_id = ? ORDER BY l.name`, issueID)
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

func (s *Service) replaceLabelsTx(tx *sql.Tx, workspaceID, issueID string, labelIDs []string) error {
	if _, err := tx.Exec(`DELETE FROM issue_labels WHERE issue_id = ?`, issueID); err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, lid := range labelIDs {
		if lid == "" || seen[lid] {
			continue
		}
		seen[lid] = true
		var id string
		err := tx.QueryRow(`SELECT id FROM labels WHERE id=? AND workspace_id=?`, lid, workspaceID).Scan(&id)
		if err == sql.ErrNoRows {
			return apperr.BadRequest("标签不存在: " + lid)
		}
		if err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO issue_labels (issue_id, label_id) VALUES (?,?)`, issueID, lid); err != nil {
			return err
		}
	}
	return nil
}
