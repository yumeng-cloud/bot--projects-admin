package botmem

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/models"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/embedder"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
)

type Service struct {
	DB    *sql.DB
	Embed embedder.Embedder // optional; set via WithEmbedder
	Store *embedder.Store   // optional; Postgres+pgvector
	// IssueDispatch 可选：ExpandSeats 改派时转发给 issueboard（避免 handlers↔services 循环依赖）。
	IssueDispatch func(workspaceID, event string, payload any)
}

func New(db *sql.DB) *Service { return &Service{DB: db} }

func nowUTC() string { return time.Now().UTC().Format(time.RFC3339Nano) }

func nullStr(p *string) any {
	if p == nil || *p == "" {
		return nil
	}
	return *p
}

func scanRole(row interface{ Scan(dest ...any) error }) (models.BotRole, error) {
	var r models.BotRole
	err := row.Scan(&r.ID, &r.WorkspaceID, &r.Name, &r.Slug, &r.Responsibilities, &r.Skills, &r.DefaultProjectIDs, &r.CreatedAt, &r.UpdatedAt)
	return r, err
}

func scanClaim(row interface{ Scan(dest ...any) error }) (models.BotRoleClaim, error) {
	var c models.BotRoleClaim
	var released sql.NullString
	err := row.Scan(&c.ID, &c.WorkspaceID, &c.RoleID, &c.MemberID, &c.ClaimedAt, &released)
	if err != nil {
		return c, err
	}
	if released.Valid {
		c.ReleasedAt = &released.String
	}
	return c, nil
}

func scanMemory(row interface{ Scan(dest ...any) error }) (models.BotMemory, error) {
	var m models.BotMemory
	var roleID sql.NullString
	err := row.Scan(&m.ID, &m.WorkspaceID, &m.MemberID, &roleID, &m.KeyName, &m.ValueText, &m.Tags, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return m, err
	}
	if roleID.Valid {
		m.RoleID = &roleID.String
	}
	return m, nil
}

func scanTodo(row interface{ Scan(dest ...any) error }) (models.BotTodo, error) {
	var t models.BotTodo
	var roleID, issueID sql.NullString
	err := row.Scan(&t.ID, &t.WorkspaceID, &t.MemberID, &roleID, &t.Title, &t.Status, &issueID, &t.SortOrder, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return t, err
	}
	if roleID.Valid {
		t.RoleID = &roleID.String
	}
	if issueID.Valid {
		t.IssueID = &issueID.String
	}
	return t, nil
}

const roleCols = `id, workspace_id, name, slug, responsibilities, skills, default_project_ids, created_at, updated_at`
const claimCols = `id, workspace_id, role_id, member_id, claimed_at, released_at`
const memCols = `id, workspace_id, member_id, role_id, key_name, value_text, tags, created_at, updated_at`
const todoCols = `id, workspace_id, member_id, role_id, title, status, issue_id, sort_order, created_at, updated_at`

// --- Roles ---

type CreateRoleInput struct {
	Name              string
	Slug              string
	Responsibilities  string
	Skills            string
	DefaultProjectIDs string
}

func (s *Service) ListRoles(workspaceID string) ([]models.BotRole, error) {
	rows, err := s.DB.Query(`SELECT `+roleCols+` FROM bot_roles WHERE workspace_id = ? ORDER BY slug`, workspaceID)
	if err != nil {
		return nil, err
	}
	var out []models.BotRole
	for rows.Next() {
		r, err := scanRole(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		out = append(out, r)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	for i := range out {
		if c, err := s.activeClaim(out[i].ID); err == nil && c != nil {
			out[i].Claim = c
		}
	}
	if out == nil {
		out = []models.BotRole{}
	}
	return out, nil
}

func (s *Service) GetRole(workspaceID, roleID string) (*models.BotRole, error) {
	r, err := scanRole(s.DB.QueryRow(`SELECT `+roleCols+` FROM bot_roles WHERE id = ? AND workspace_id = ?`, roleID, workspaceID))
	if err == sql.ErrNoRows {
		return nil, apperr.NotFound("角色不存在")
	}
	if err != nil {
		return nil, err
	}
	if c, err := s.activeClaim(r.ID); err == nil && c != nil {
		r.Claim = c
	}
	return &r, nil
}

func (s *Service) CreateRole(workspaceID string, in CreateRoleInput) (*models.BotRole, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Slug = strings.TrimSpace(strings.ToLower(in.Slug))
	if in.Name == "" {
		return nil, apperr.BadRequest("角色名称不能为空")
	}
	if in.Slug == "" {
		return nil, apperr.BadRequest("slug 不能为空")
	}
	if in.Skills == "" {
		in.Skills = "[]"
	}
	if in.DefaultProjectIDs == "" {
		in.DefaultProjectIDs = "[]"
	}
	id := idgen.New()
	now := nowUTC()
	_, err := s.DB.Exec(`INSERT INTO bot_roles (id, workspace_id, name, slug, responsibilities, skills, default_project_ids, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?)`, id, workspaceID, in.Name, in.Slug, in.Responsibilities, in.Skills, in.DefaultProjectIDs, now, now)
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, apperr.Conflict("该 slug 已存在")
		}
		return nil, err
	}
	return s.GetRole(workspaceID, id)
}

func (s *Service) activeClaim(roleID string) (*models.BotRoleClaim, error) {
	c, err := scanClaim(s.DB.QueryRow(`SELECT `+claimCols+` FROM bot_role_claims WHERE role_id = ? AND released_at IS NULL`, roleID))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// Claim 认领角色。异人存在 active claim（released_at IS NULL）时一律 409，即使 force=true 也禁止顶替。
// force 保留兼容：无 active 时与普通认领等价（可清「已释放」历史行后重认领）。
// schema UNIQUE(role_id) 仅允许一行：已释放则 UPDATE 重认领。
func (s *Service) Claim(workspaceID, roleID, memberID string, force bool) (*models.BotRoleClaim, error) {
	_ = force // 活跃不可 force 顶替；参数仅兼容旧客户端
	if memberID == "" {
		return nil, apperr.BadRequest("member_id 不能为空")
	}
	if _, err := s.GetRole(workspaceID, roleID); err != nil {
		return nil, err
	}
	var mid string
	err := s.DB.QueryRow(`SELECT id FROM members WHERE id = ? AND workspace_id = ?`, memberID, workspaceID).Scan(&mid)
	if err == sql.ErrNoRows {
		return nil, apperr.NotFound("成员不存在")
	}
	if err != nil {
		return nil, err
	}

	if active, err := s.activeClaim(roleID); err != nil {
		return nil, err
	} else if active != nil {
		if active.MemberID == memberID {
			s.syncSeatsActiveOnClaim(workspaceID, roleID, memberID)
			return active, nil
		}
		return nil, apperr.Conflict("该席位占用中，禁止顶替")
	}

	now := nowUTC()
	// 已有历史行（已释放）→ 重认领；无行 → 插入
	var existingID string
	err = s.DB.QueryRow(`SELECT id FROM bot_role_claims WHERE role_id = ?`, roleID).Scan(&existingID)
	var claim *models.BotRoleClaim
	if err == nil {
		_, err = s.DB.Exec(`UPDATE bot_role_claims SET member_id = ?, claimed_at = ?, released_at = NULL, workspace_id = ? WHERE id = ?`,
			memberID, now, workspaceID, existingID)
		if err != nil {
			return nil, err
		}
		c, err := scanClaim(s.DB.QueryRow(`SELECT `+claimCols+` FROM bot_role_claims WHERE id = ?`, existingID))
		if err != nil {
			return nil, err
		}
		claim = &c
	} else if err != sql.ErrNoRows {
		return nil, err
	} else {
		id := idgen.New()
		_, err = s.DB.Exec(`INSERT INTO bot_role_claims (id, workspace_id, role_id, member_id, claimed_at, released_at)
			VALUES (?,?,?,?,?,NULL)`, id, workspaceID, roleID, memberID, now)
		if err != nil {
			if db.IsUniqueViolation(err) {
				return nil, apperr.Conflict("该席位占用中，禁止顶替")
			}
			return nil, err
		}
		c, err := scanClaim(s.DB.QueryRow(`SELECT `+claimCols+` FROM bot_role_claims WHERE id = ?`, id))
		if err != nil {
			return nil, err
		}
		claim = &c
	}
	s.syncSeatsActiveOnClaim(workspaceID, roleID, memberID)
	return claim, nil
}

// syncSeatsActiveOnClaim claim/enter 成功后：关联 team_seats 写 bound + status=active。
// 不碰 stopped/left（补位须走 BindSeat/force）。
func (s *Service) syncSeatsActiveOnClaim(workspaceID, roleID, memberID string) {
	if roleID == "" || memberID == "" {
		return
	}
	now := nowUTC()
	_, _ = s.DB.Exec(`UPDATE team_seats SET bound_member_id = ?, bound_at = ?, status = ?, updated_at = ?
		WHERE workspace_id = ? AND bot_role_id = ?
		  AND COALESCE(NULLIF(status, ''), 'open') NOT IN ('stopped', 'left')`,
		memberID, now, models.SeatStatusActive, now, workspaceID, roleID)
}

func (s *Service) Release(workspaceID, roleID string) error {
	if _, err := s.GetRole(workspaceID, roleID); err != nil {
		return err
	}
	now := nowUTC()
	res, err := s.DB.Exec(`UPDATE bot_role_claims SET released_at = ? WHERE role_id = ? AND workspace_id = ? AND released_at IS NULL`,
		now, roleID, workspaceID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperr.NotFound("该角色当前无有效认领")
	}
	return nil
}

// --- Memories ---

type UpsertMemoryInput struct {
	KeyName   string
	ValueText string
	Tags      string
	RoleID    *string
}


// assertMemberInWorkspace 硬门：path member 必须属于 path workspace；否则 404「成员不存在」（与 Handoff 齐，禁 200 空 / 201 脏写）。
func (s *Service) assertMemberInWorkspace(workspaceID, memberID string) error {
	if !s.memberExists(workspaceID, memberID) {
		return apperr.NotFound("成员不存在")
	}
	return nil
}

// ListMemories：q 非空时用 LIKE 关键词检索；语义/混合见 SearchMemories。
func (s *Service) ListMemories(workspaceID, memberID, q string) ([]models.BotMemory, error) {
	if err := s.assertMemberInWorkspace(workspaceID, memberID); err != nil {
		return nil, err
	}
	q = strings.TrimSpace(q)
	var rows *sql.Rows
	var err error
	if q == "" {
		rows, err = s.DB.Query(`SELECT `+memCols+` FROM bot_memories WHERE workspace_id = ? AND member_id = ? ORDER BY key_name`,
			workspaceID, memberID)
	} else {
		like := "%" + q + "%"
		rows, err = s.DB.Query(`SELECT `+memCols+` FROM bot_memories
			WHERE workspace_id = ? AND member_id = ? AND (key_name LIKE ? OR value_text LIKE ? OR tags LIKE ?)
			ORDER BY key_name`, workspaceID, memberID, like, like, like)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.BotMemory
	for rows.Next() {
		m, err := scanMemory(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if out == nil {
		out = []models.BotMemory{}
	}
	return out, rows.Err()
}

func (s *Service) GetMemory(workspaceID, memberID, memoryID string) (*models.BotMemory, error) {
	m, err := scanMemory(s.DB.QueryRow(`SELECT `+memCols+` FROM bot_memories WHERE id = ? AND workspace_id = ? AND member_id = ?`,
		memoryID, workspaceID, memberID))
	if err == sql.ErrNoRows {
		return nil, apperr.NotFound("记忆不存在")
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *Service) UpsertMemory(workspaceID, memberID string, in UpsertMemoryInput) (*models.BotMemory, error) {
	if err := s.assertMemberInWorkspace(workspaceID, memberID); err != nil {
		return nil, err
	}
	in.KeyName = strings.TrimSpace(in.KeyName)
	if in.KeyName == "" {
		return nil, apperr.BadRequest("key_name 不能为空")
	}
	if in.Tags == "" {
		in.Tags = "[]"
	}
	now := nowUTC()
	var existingID string
	err := s.DB.QueryRow(`SELECT id FROM bot_memories WHERE workspace_id = ? AND member_id = ? AND key_name = ?`,
		workspaceID, memberID, in.KeyName).Scan(&existingID)
	if err == nil {
		_, err = s.DB.Exec(`UPDATE bot_memories SET value_text = ?, tags = ?, role_id = ?, updated_at = ? WHERE id = ?`,
			in.ValueText, in.Tags, nullStr(in.RoleID), now, existingID)
		if err != nil {
			return nil, err
		}
		m, err := s.GetMemory(workspaceID, memberID, existingID)
		if err != nil {
			return nil, err
		}
		s.indexMemory(context.Background(), workspaceID, m.ID, m.KeyName, m.ValueText)
		return m, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}
	id := idgen.New()
	_, err = s.DB.Exec(`INSERT INTO bot_memories (id, workspace_id, member_id, role_id, key_name, value_text, tags, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?)`, id, workspaceID, memberID, nullStr(in.RoleID), in.KeyName, in.ValueText, in.Tags, now, now)
	if err != nil {
		return nil, err
	}
	m, err := s.GetMemory(workspaceID, memberID, id)
	if err != nil {
		return nil, err
	}
	s.indexMemory(context.Background(), workspaceID, m.ID, m.KeyName, m.ValueText)
	return m, nil
}

type PatchMemoryInput struct {
	KeyName   *string
	ValueText *string
	Tags      *string
	RoleID    *string
}

func (s *Service) PatchMemory(workspaceID, memberID, memoryID string, in PatchMemoryInput) (*models.BotMemory, error) {
	if err := s.assertMemberInWorkspace(workspaceID, memberID); err != nil {
		return nil, err
	}
	cur, err := s.GetMemory(workspaceID, memberID, memoryID)
	if err != nil {
		return nil, err
	}
	if in.KeyName != nil {
		cur.KeyName = strings.TrimSpace(*in.KeyName)
		if cur.KeyName == "" {
			return nil, apperr.BadRequest("key_name 不能为空")
		}
	}
	if in.ValueText != nil {
		cur.ValueText = *in.ValueText
	}
	if in.Tags != nil {
		cur.Tags = *in.Tags
	}
	role := cur.RoleID
	if in.RoleID != nil {
		if *in.RoleID == "" {
			role = nil
		} else {
			role = in.RoleID
		}
	}
	now := nowUTC()
	_, err = s.DB.Exec(`UPDATE bot_memories SET key_name = ?, value_text = ?, tags = ?, role_id = ?, updated_at = ? WHERE id = ?`,
		cur.KeyName, cur.ValueText, cur.Tags, nullStr(role), now, memoryID)
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, apperr.Conflict("key_name 已存在")
		}
		return nil, err
	}
	m, err := s.GetMemory(workspaceID, memberID, memoryID)
	if err != nil {
		return nil, err
	}
	s.indexMemory(context.Background(), workspaceID, m.ID, m.KeyName, m.ValueText)
	return m, nil
}

func (s *Service) DeleteMemory(workspaceID, memberID, memoryID string) error {
	if err := s.assertMemberInWorkspace(workspaceID, memberID); err != nil {
		return err
	}
	res, err := s.DB.Exec(`DELETE FROM bot_memories WHERE id = ? AND workspace_id = ? AND member_id = ?`, memoryID, workspaceID, memberID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperr.NotFound("记忆不存在")
	}
	s.dropMemoryIndex(context.Background(), workspaceID, memoryID)
	return nil
}

// --- Todos ---

var todoStatuses = map[string]bool{"todo": true, "doing": true, "done": true, "canceled": true}

type CreateTodoInput struct {
	Title     string
	Status    string
	RoleID    *string
	IssueID   *string
	SortOrder float64
}

func (s *Service) ListTodos(workspaceID, memberID, status string) ([]models.BotTodo, error) {
	if err := s.assertMemberInWorkspace(workspaceID, memberID); err != nil {
		return nil, err
	}
	var rows *sql.Rows
	var err error
	if status != "" {
		rows, err = s.DB.Query(`SELECT `+todoCols+` FROM bot_todos WHERE workspace_id = ? AND member_id = ? AND status = ? ORDER BY sort_order, created_at`,
			workspaceID, memberID, status)
	} else {
		rows, err = s.DB.Query(`SELECT `+todoCols+` FROM bot_todos WHERE workspace_id = ? AND member_id = ? ORDER BY sort_order, created_at`,
			workspaceID, memberID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.BotTodo
	for rows.Next() {
		t, err := scanTodo(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	if out == nil {
		out = []models.BotTodo{}
	}
	return out, rows.Err()
}

func (s *Service) GetTodo(workspaceID, memberID, todoID string) (*models.BotTodo, error) {
	t, err := scanTodo(s.DB.QueryRow(`SELECT `+todoCols+` FROM bot_todos WHERE id = ? AND workspace_id = ? AND member_id = ?`,
		todoID, workspaceID, memberID))
	if err == sql.ErrNoRows {
		return nil, apperr.NotFound("待办不存在")
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *Service) CreateTodo(workspaceID, memberID string, in CreateTodoInput) (*models.BotTodo, error) {
	if err := s.assertMemberInWorkspace(workspaceID, memberID); err != nil {
		return nil, err
	}
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		return nil, apperr.BadRequest("title 不能为空")
	}
	if in.Status == "" {
		in.Status = "todo"
	}
	if !todoStatuses[in.Status] {
		return nil, apperr.BadRequest("status 须为 todo|doing|done|canceled")
	}
	id := idgen.New()
	now := nowUTC()
	_, err := s.DB.Exec(`INSERT INTO bot_todos (id, workspace_id, member_id, role_id, title, status, issue_id, sort_order, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?)`, id, workspaceID, memberID, nullStr(in.RoleID), in.Title, in.Status, nullStr(in.IssueID), in.SortOrder, now, now)
	if err != nil {
		return nil, err
	}
	return s.GetTodo(workspaceID, memberID, id)
}

type PatchTodoInput struct {
	Title     *string
	Status    *string
	RoleID    *string
	IssueID   *string
	SortOrder *float64
}

func (s *Service) PatchTodo(workspaceID, memberID, todoID string, in PatchTodoInput) (*models.BotTodo, error) {
	if err := s.assertMemberInWorkspace(workspaceID, memberID); err != nil {
		return nil, err
	}
	cur, err := s.GetTodo(workspaceID, memberID, todoID)
	if err != nil {
		return nil, err
	}
	if in.Title != nil {
		cur.Title = strings.TrimSpace(*in.Title)
		if cur.Title == "" {
			return nil, apperr.BadRequest("title 不能为空")
		}
	}
	if in.Status != nil {
		if !todoStatuses[*in.Status] {
			return nil, apperr.BadRequest("status 须为 todo|doing|done|canceled")
		}
		cur.Status = *in.Status
	}
	role := cur.RoleID
	if in.RoleID != nil {
		if *in.RoleID == "" {
			role = nil
		} else {
			role = in.RoleID
		}
	}
	issue := cur.IssueID
	if in.IssueID != nil {
		if *in.IssueID == "" {
			issue = nil
		} else {
			issue = in.IssueID
		}
	}
	if in.SortOrder != nil {
		cur.SortOrder = *in.SortOrder
	}
	now := nowUTC()
	_, err = s.DB.Exec(`UPDATE bot_todos SET title = ?, status = ?, role_id = ?, issue_id = ?, sort_order = ?, updated_at = ? WHERE id = ?`,
		cur.Title, cur.Status, nullStr(role), nullStr(issue), cur.SortOrder, now, todoID)
	if err != nil {
		return nil, err
	}
	return s.GetTodo(workspaceID, memberID, todoID)
}

// --- Roster ---

func (s *Service) Roster(workspaceID string) ([]models.BotRosterItem, error) {
	rows, err := s.DB.Query(`SELECT id, display_name, bot_enabled, role FROM members WHERE workspace_id = ? AND is_bot = 1 ORDER BY display_name`, workspaceID)
	if err != nil {
		return nil, err
	}
	var out []models.BotRosterItem
	for rows.Next() {
		var item models.BotRosterItem
		var enabled int
		if err := rows.Scan(&item.MemberID, &item.DisplayName, &enabled, &item.Role); err != nil {
			rows.Close()
			return nil, err
		}
		item.BotEnabled = enabled != 0
		out = append(out, item)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	// SQLite MaxOpenConns(1)：必须先关掉 rows 再嵌套查询，否则会死锁
	for i := range out {
		var roleID string
		err := s.DB.QueryRow(`SELECT role_id FROM bot_role_claims WHERE member_id = ? AND workspace_id = ? AND released_at IS NULL LIMIT 1`,
			out[i].MemberID, workspaceID).Scan(&roleID)
		if err == nil {
			if r, err := s.GetRole(workspaceID, roleID); err == nil {
				out[i].ClaimedRole = r
			}
		}
	}
	if out == nil {
		out = []models.BotRosterItem{}
	}
	return out, nil
}

// --- Handoff ---

func (s *Service) Handoff(workspaceID, memberID string) (*models.BotHandoff, error) {
	if err := s.assertMemberInWorkspace(workspaceID, memberID); err != nil {
		return nil, err
	}

	h := &models.BotHandoff{
		WorkspaceID: workspaceID,
		MemberID:    memberID,
		GeneratedAt: nowUTC(),
		Memories:    []models.BotMemory{},
		Todos:       []models.BotTodo{},
		Issues:      []models.HandoffIssue{},
	}

	var roleID string
	err := s.DB.QueryRow(`SELECT role_id FROM bot_role_claims WHERE member_id = ? AND workspace_id = ? AND released_at IS NULL LIMIT 1`,
		memberID, workspaceID).Scan(&roleID)
	if err == nil {
		if r, err := s.GetRole(workspaceID, roleID); err == nil {
			h.ClaimedRole = r
		}
	}

	// T7-3：绑定席位的 playbook_ref 为唯一手册入口（优先匹配 claimed bot_role_id）
	var pref sql.NullString
	errPref := s.DB.QueryRow(`SELECT ts.playbook_ref FROM team_seats ts
		WHERE ts.workspace_id = ? AND ts.bound_member_id = ?
		ORDER BY CASE WHEN ? != '' AND ts.bot_role_id = ? THEN 0 ELSE 1 END, ts.bound_at DESC
		LIMIT 1`, workspaceID, memberID, roleID, roleID).Scan(&pref)
	if errPref == nil && pref.Valid {
		p := strings.TrimSpace(pref.String)
		if p != "" {
			h.PlaybookRef = &p
		}
	}

	mems, err := s.ListMemories(workspaceID, memberID, "")
	if err != nil {
		return nil, err
	}
	h.Memories = mems

	rows, err := s.DB.Query(`SELECT `+todoCols+` FROM bot_todos
		WHERE workspace_id = ? AND member_id = ? AND status IN ('todo','doing')
		ORDER BY sort_order, created_at`, workspaceID, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		t, err := scanTodo(rows)
		if err != nil {
			return nil, err
		}
		h.Todos = append(h.Todos, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	irows, err := s.DB.Query(`SELECT id, identifier, title, status, priority, project_id, cycle_id, updated_at
		FROM issues WHERE workspace_id = ? AND assignee_id = ? AND status IN ('todo','in_progress','in_review')
		ORDER BY updated_at DESC`, workspaceID, memberID)
	if err != nil {
		return nil, err
	}
	defer irows.Close()
	for irows.Next() {
		var iss models.HandoffIssue
		var cycle sql.NullString
		if err := irows.Scan(&iss.ID, &iss.Identifier, &iss.Title, &iss.Status, &iss.Priority, &iss.ProjectID, &cycle, &iss.UpdatedAt); err != nil {
			return nil, err
		}
		if cycle.Valid {
			iss.CycleID = &cycle.String
		}
		h.Issues = append(h.Issues, iss)
	}
	return h, irows.Err()
}

// --- ImportHandoff ---

// ImportHandoffInput 与 GET handoff 同形：可嵌套 Handoff，或扁平 memories/todos + claimed_role_id|slug。
type ImportHandoffInput struct {
	Handoff         *models.BotHandoff    `json:"handoff"`
	Memories        []models.BotMemory    `json:"memories"`
	Todos           []models.BotTodo      `json:"todos"`
	Issues          []models.HandoffIssue `json:"issues"`
	ClaimedRoleID   *string               `json:"claimed_role_id"`
	ClaimedRoleSlug *string               `json:"claimed_role_slug"`
	ClaimedRole     *models.BotRole       `json:"claimed_role"`
}

// ImportHandoffResult 导入计数、claim 结果、warnings，以及刷新后的 Handoff。
type ImportHandoffResult struct {
	ImportedMemories int                `json:"imported_memories"`
	ImportedTodos    int                `json:"imported_todos"`
	Claimed          bool               `json:"claimed"`
	Warnings         []string           `json:"warnings"`
	IssueIDs         []string           `json:"issue_ids"`
	Handoff          *models.BotHandoff `json:"handoff"`
}

func (s *Service) roleIDBySlug(workspaceID, slug string) (string, error) {
	slug = strings.TrimSpace(strings.ToLower(slug))
	if slug == "" {
		return "", apperr.BadRequest("slug 不能为空")
	}
	var id string
	err := s.DB.QueryRow(`SELECT id FROM bot_roles WHERE workspace_id = ? AND slug = ?`, workspaceID, slug).Scan(&id)
	if err == sql.ErrNoRows {
		return "", apperr.NotFound("角色不存在")
	}
	return id, err
}

func issueKey(issueID *string) string {
	if issueID == nil {
		return ""
	}
	return *issueID
}

func (s *Service) findOpenTodoByTitleIssue(workspaceID, memberID, title string, issueID *string) (*models.BotTodo, error) {
	rows, err := s.DB.Query(`SELECT `+todoCols+` FROM bot_todos
		WHERE workspace_id = ? AND member_id = ? AND title = ? AND status IN ('todo','doing')`,
		workspaceID, memberID, title)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	want := issueKey(issueID)
	for rows.Next() {
		t, err := scanTodo(rows)
		if err != nil {
			return nil, err
		}
		if issueKey(t.IssueID) == want {
			return &t, nil
		}
	}
	return nil, rows.Err()
}

// ImportHandoff 将续工包写入目标 member：memories upsert、未完成 todos 导入、尝试 claim；issues 只列 id。
func (s *Service) ImportHandoff(workspaceID, memberID string, in ImportHandoffInput) (*ImportHandoffResult, error) {
	if err := s.assertMemberInWorkspace(workspaceID, memberID); err != nil {
		return nil, err
	}

	mems := in.Memories
	todos := in.Todos
	issues := in.Issues
	roleID := in.ClaimedRoleID
	roleSlug := in.ClaimedRoleSlug
	if in.ClaimedRole != nil {
		if roleID == nil && in.ClaimedRole.ID != "" {
			id := in.ClaimedRole.ID
			roleID = &id
		}
		if roleSlug == nil && in.ClaimedRole.Slug != "" {
			sl := in.ClaimedRole.Slug
			roleSlug = &sl
		}
	}
	if in.Handoff != nil {
		if len(mems) == 0 {
			mems = in.Handoff.Memories
		}
		if len(todos) == 0 {
			todos = in.Handoff.Todos
		}
		if len(issues) == 0 {
			issues = in.Handoff.Issues
		}
		if in.Handoff.ClaimedRole != nil {
			if roleID == nil && in.Handoff.ClaimedRole.ID != "" {
				id := in.Handoff.ClaimedRole.ID
				roleID = &id
			}
			if roleSlug == nil && in.Handoff.ClaimedRole.Slug != "" {
				sl := in.Handoff.ClaimedRole.Slug
				roleSlug = &sl
			}
		}
	}

	res := &ImportHandoffResult{
		Warnings: []string{},
		IssueIDs: []string{},
	}

	for _, m := range mems {
		if strings.TrimSpace(m.KeyName) == "" {
			res.Warnings = append(res.Warnings, "skip memory: empty key_name")
			continue
		}
		tags := m.Tags
		if tags == "" {
			tags = "[]"
		}
		if _, err := s.UpsertMemory(workspaceID, memberID, UpsertMemoryInput{
			KeyName: m.KeyName, ValueText: m.ValueText, Tags: tags, RoleID: m.RoleID,
		}); err != nil {
			res.Warnings = append(res.Warnings, "memory "+m.KeyName+": "+err.Error())
			continue
		}
		res.ImportedMemories++
	}

	for _, t := range todos {
		st := t.Status
		if st == "" {
			st = "todo"
		}
		if st != "todo" && st != "doing" {
			continue
		}
		title := strings.TrimSpace(t.Title)
		if title == "" {
			res.Warnings = append(res.Warnings, "skip todo: empty title")
			continue
		}
		issueID := t.IssueID
		existing, err := s.findOpenTodoByTitleIssue(workspaceID, memberID, title, issueID)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			if _, err := s.PatchTodo(workspaceID, memberID, existing.ID, PatchTodoInput{
				Status: &st, RoleID: t.RoleID, IssueID: issueID, SortOrder: &t.SortOrder,
			}); err != nil && issueID != nil {
				// FK 等失败时去掉 issue_id 重试
				empty := ""
				_, err = s.PatchTodo(workspaceID, memberID, existing.ID, PatchTodoInput{
					Status: &st, RoleID: t.RoleID, IssueID: &empty, SortOrder: &t.SortOrder,
				})
				if err == nil {
					res.Warnings = append(res.Warnings, "todo "+title+": dropped invalid issue_id")
				}
			}
			if err != nil {
				res.Warnings = append(res.Warnings, "todo "+title+": "+err.Error())
				continue
			}
			res.ImportedTodos++
			continue
		}
		if _, err := s.CreateTodo(workspaceID, memberID, CreateTodoInput{
			Title: title, Status: st, RoleID: t.RoleID, IssueID: issueID, SortOrder: t.SortOrder,
		}); err != nil && issueID != nil {
			_, err = s.CreateTodo(workspaceID, memberID, CreateTodoInput{
				Title: title, Status: st, RoleID: t.RoleID, SortOrder: t.SortOrder,
			})
			if err == nil {
				res.Warnings = append(res.Warnings, "todo "+title+": dropped invalid issue_id")
			}
		}
		if err != nil {
			res.Warnings = append(res.Warnings, "todo "+title+": "+err.Error())
			continue
		}
		res.ImportedTodos++
	}

	claimRoleID := ""
	if roleID != nil && strings.TrimSpace(*roleID) != "" {
		claimRoleID = strings.TrimSpace(*roleID)
	} else if roleSlug != nil && strings.TrimSpace(*roleSlug) != "" {
		id, err := s.roleIDBySlug(workspaceID, *roleSlug)
		if err != nil {
			res.Warnings = append(res.Warnings, "claim: "+err.Error())
		} else {
			claimRoleID = id
		}
	}
	if claimRoleID != "" {
		if _, err := s.Claim(workspaceID, claimRoleID, memberID, false); err != nil {
			if ae, ok := err.(*apperr.AppError); ok && ae.Code == "conflict" {
				res.Warnings = append(res.Warnings, "claim conflict: "+ae.Message)
			} else {
				res.Warnings = append(res.Warnings, "claim: "+err.Error())
			}
		} else {
			res.Claimed = true
		}
	}

	for _, iss := range issues {
		if iss.ID != "" {
			res.IssueIDs = append(res.IssueIDs, iss.ID)
		}
	}

	h, err := s.Handoff(workspaceID, memberID)
	if err != nil {
		return nil, err
	}
	res.Handoff = h
	return res, nil
}
