package botmem

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/models"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/services/issueboard"
)

type CreateTeamInput struct {
	ID          string
	Name        string
	Slug        string
	TemplateKey string // 同 template_id；有值则按模板幂等建空席（不绑 Key）
}

// CreateTeamResult 创建团队；挂模板时附带空席 roster。
type CreateTeamResult struct {
	Team   *models.Team
	Roster []models.TeamSeat
}

func (s *Service) CreateTeam(workspaceID string, in CreateTeamInput) (*models.Team, error) {
	res, err := s.CreateTeamFromTemplate(workspaceID, in)
	if err != nil {
		return nil, err
	}
	return res.Team, nil
}

// CreateTeamFromTemplate 创建团队；template_key/id 非空时按内置模板建空席（T7-R R1）。
func (s *Service) CreateTeamFromTemplate(workspaceID string, in CreateTeamInput) (*CreateTeamResult, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return nil, apperr.BadRequest("团队名称不能为空")
	}
	tk := strings.TrimSpace(in.TemplateKey)
	var tmpl *TeamTemplate
	if tk != "" {
		var err error
		tmpl, err = GetTeamTemplate(tk)
		if err != nil {
			return nil, err
		}
		tk = tmpl.ID
	}
	id := strings.TrimSpace(in.ID)
	if id == "" {
		id = idgen.New()
	}
	slug := strings.TrimSpace(in.Slug)
	var slugVal any
	if slug == "" {
		slugVal = nil
	} else {
		slugVal = slug
	}
	now := nowUTC()

	var existingWS string
	err := s.DB.QueryRow(`SELECT workspace_id FROM teams WHERE id = ?`, id).Scan(&existingWS)
	if err == nil {
		if existingWS != workspaceID {
			return nil, apperr.Conflict("该团队 ID 已属于其他工作区")
		}
		team, gerr := s.GetTeam(workspaceID, id)
		if gerr != nil {
			return nil, gerr
		}
		roster, rerr := s.GetRoster(workspaceID, id)
		if rerr != nil {
			return nil, rerr
		}
		return &CreateTeamResult{Team: team, Roster: roster}, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	_, err = s.DB.Exec(`INSERT INTO teams (id, workspace_id, name, slug, template_key, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?)`, id, workspaceID, in.Name, slugVal, tk, now, now)
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, apperr.Conflict("该团队 slug 已存在")
		}
		return nil, err
	}
	team, err := s.GetTeam(workspaceID, id)
	if err != nil {
		return nil, err
	}
	res := &CreateTeamResult{Team: team, Roster: []models.TeamSeat{}}
	if tmpl != nil {
		roster, err := s.EnsureRoster(workspaceID, id, EnsureRosterInput{
			TemplateKey: tmpl.ID,
			Seats:       templateSeatsToEnsure(tmpl),
		})
		if err != nil {
			return nil, err
		}
		res.Roster = roster
		// 刷新 template_key（EnsureRoster 会写）
		if team2, gerr := s.GetTeam(workspaceID, id); gerr == nil {
			res.Team = team2
		}
	}
	return res, nil
}

func (s *Service) ListTeams(workspaceID string) ([]models.Team, error) {
	rows, err := s.DB.Query(`SELECT id, workspace_id, name, slug, COALESCE(template_key,''), created_at, updated_at
		FROM teams WHERE workspace_id = ? ORDER BY created_at`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Team
	for rows.Next() {
		t, err := scanTeam(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if out == nil {
		out = []models.Team{}
	}
	return out, nil
}

func (s *Service) GetTeam(workspaceID, teamID string) (*models.Team, error) {
	t, err := scanTeam(s.DB.QueryRow(`SELECT id, workspace_id, name, slug, COALESCE(template_key,''), created_at, updated_at
		FROM teams WHERE id = ? AND workspace_id = ?`, teamID, workspaceID))
	if err == sql.ErrNoRows {
		return nil, apperr.NotFound("团队不存在")
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func scanTeam(row interface{ Scan(dest ...any) error }) (models.Team, error) {
	var t models.Team
	var slug sql.NullString
	err := row.Scan(&t.ID, &t.WorkspaceID, &t.Name, &slug, &t.TemplateKey, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return t, err
	}
	if slug.Valid && slug.String != "" {
		t.Slug = &slug.String
	}
	return t, nil
}

type EnsureSeatInput struct {
	Slug        string
	Name        string
	PlaybookRef string
}

type EnsureRosterInput struct {
	Seats       []EnsureSeatInput
	TemplateKey string
}

func (s *Service) EnsureRoster(workspaceID, teamID string, in EnsureRosterInput) ([]models.TeamSeat, error) {
	if _, err := s.GetTeam(workspaceID, teamID); err != nil {
		return nil, err
	}
	seats := in.Seats
	tk := strings.TrimSpace(strings.ToLower(in.TemplateKey))
	if len(seats) == 0 {
		if tk == "" {
			// 无模板、无显式席位 → 用内置 default（含 playbook_ref）
			tmpl, err := GetTeamTemplate("default")
			if err != nil {
				return nil, err
			}
			seats = templateSeatsToEnsure(tmpl)
			tk = tmpl.ID
			in.TemplateKey = tmpl.ID
		} else {
			tmpl, err := GetTeamTemplate(tk)
			if err != nil {
				return nil, err
			}
			seats = templateSeatsToEnsure(tmpl)
			in.TemplateKey = tmpl.ID
		}
	} else if tk != "" {
		// 显式席位 + 模板 key：校验模板存在；席位 playbook 仍以入参为准
		if _, err := GetTeamTemplate(tk); err != nil {
			return nil, err
		}
	}
	// T7-3：写入席位须带合法 playbook_ref；空/非法 → 4xx，禁止默默换硬编码手册
	for i := range seats {
		canon, err := CanonicalPlaybookRefString(seats[i].PlaybookRef)
		if err != nil {
			return nil, err
		}
		seats[i].PlaybookRef = canon
	}
	if len(seats) == 0 {
		return s.GetRoster(workspaceID, teamID)
	}
	now := nowUTC()
	for _, seat := range seats {
		slug := strings.TrimSpace(strings.ToLower(seat.Slug))
		name := strings.TrimSpace(seat.Name)
		if slug == "" {
			return nil, apperr.BadRequest("席位 slug 不能为空")
		}
		if name == "" {
			name = slug
		}
		role, err := s.ensureTeamRole(workspaceID, teamID, slug, name)
		if err != nil {
			return nil, err
		}
		playbook := strings.TrimSpace(seat.PlaybookRef)
		var existingID string
		err = s.DB.QueryRow(`SELECT id FROM team_seats WHERE team_id = ? AND role_slug = ?`, teamID, slug).Scan(&existingID)
		if err == nil {
			_, err = s.DB.Exec(`UPDATE team_seats SET display_name = ?, playbook_ref = ?,
				bot_role_id = ?, updated_at = ? WHERE id = ?`,
				name, playbook, role.ID, now, existingID)
			if err != nil {
				return nil, err
			}
			continue
		}
		if err != sql.ErrNoRows {
			return nil, err
		}
		sid := idgen.New()
		_, err = s.DB.Exec(`INSERT INTO team_seats
			(id, team_id, workspace_id, role_slug, display_name, playbook_ref, bot_role_id,
			 bound_member_id, bound_at, status, created_at, updated_at)
			VALUES (?,?,?,?,?,?,?,NULL,NULL,?,?,?)`,
			sid, teamID, workspaceID, slug, name, playbook, role.ID, models.SeatStatusOpen, now, now)
		if err != nil {
			if db.IsUniqueViolation(err) {
				continue
			}
			return nil, err
		}
	}
	if tk != "" {
		_, _ = s.DB.Exec(`UPDATE teams SET template_key = ?, updated_at = ? WHERE id = ? AND workspace_id = ?`,
			in.TemplateKey, now, teamID, workspaceID)
	}
	return s.GetRoster(workspaceID, teamID)
}

// teamScopedRoleSlug 将席位 slug 映射为按 team 唯一的 bot_roles.slug，
// 避免同 WS 多队共用 workspace 级角色导致 claim / syncSeats 串队（T6-R）。
func teamScopedRoleSlug(teamID, slug string) string {
	return "team:" + teamID + ":" + slug
}

// ensureTeamRole 为 (team, role_slug) 确保独立 bot_role；幂等。
// 席位对外仍用短 slug（team_seats.role_slug）；claim 单活落在本队 role 上。
func (s *Service) ensureTeamRole(workspaceID, teamID, slug, name string) (*models.BotRole, error) {
	roleSlug := teamScopedRoleSlug(teamID, slug)
	var id string
	err := s.DB.QueryRow(`SELECT id FROM bot_roles WHERE workspace_id = ? AND slug = ?`, workspaceID, roleSlug).Scan(&id)
	if err == nil {
		return s.GetRole(workspaceID, id)
	}
	if err != sql.ErrNoRows {
		return nil, err
	}
	return s.CreateRole(workspaceID, CreateRoleInput{Name: name, Slug: roleSlug})
}

func (s *Service) GetRoster(workspaceID, teamID string) ([]models.TeamSeat, error) {
	if _, err := s.GetTeam(workspaceID, teamID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(`
		SELECT ts.id, ts.team_id, ts.workspace_id, ts.role_slug, ts.display_name, ts.playbook_ref,
			ts.bot_role_id, ts.bound_member_id, ts.bound_at, COALESCE(NULLIF(ts.status,''), 'open'),
			ts.created_at, ts.updated_at,
			m.display_name,
			c.member_id, c.claimed_at
		FROM team_seats ts
		LEFT JOIN members m ON m.id = ts.bound_member_id
		LEFT JOIN bot_role_claims c ON c.role_id = ts.bot_role_id AND c.released_at IS NULL
		WHERE ts.team_id = ? AND ts.workspace_id = ?
		ORDER BY ts.role_slug`, teamID, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.TeamSeat
	for rows.Next() {
		seat, err := scanRosterRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, seat)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if out == nil {
		out = []models.TeamSeat{}
	}
	return out, nil
}

func scanRosterRow(row interface{ Scan(dest ...any) error }) (models.TeamSeat, error) {
	var seat models.TeamSeat
	var playbook, botRoleID, boundMember, boundAt, boundDN, claimMember, claimAt sql.NullString
	err := row.Scan(
		&seat.ID, &seat.TeamID, &seat.WorkspaceID, &seat.RoleSlug, &seat.DisplayName,
		&playbook, &botRoleID, &boundMember, &boundAt, &seat.Status, &seat.CreatedAt, &seat.UpdatedAt,
		&boundDN, &claimMember, &claimAt,
	)
	if err != nil {
		return seat, err
	}
	seat.Status = normalizeSeatStatus(seat.Status)
	if playbook.Valid {
		seat.PlaybookRef = &playbook.String
	}
	if botRoleID.Valid {
		seat.BotRoleID = &botRoleID.String
	}
	if boundMember.Valid {
		seat.BoundMemberID = &boundMember.String
		seat.MemberID = &boundMember.String
		// 迁移前已绑席可能仍落 open；有 bound 且非停工/离职 → 读出按 active
		if seat.Status == models.SeatStatusOpen {
			seat.Status = models.SeatStatusActive
		}
	}
	if boundAt.Valid {
		seat.BoundAt = &boundAt.String
	}
	if boundDN.Valid {
		seat.BoundDisplayName = &boundDN.String
	}
	if claimMember.Valid {
		seat.Claim = &models.TeamSeatClaim{MemberID: claimMember.String, ClaimedAt: claimAt.String}
	}
	return seat, nil
}

// seatPlaybookRefOrErr 读出席位可读 playbook_ref；空/非法 → 4xx（T7-3）。
func seatPlaybookRefOrErr(seat *models.TeamSeat) (string, error) {
	if seat == nil {
		return "", apperr.BadRequest("playbook_ref 不能为空")
	}
	raw := ""
	if seat.PlaybookRef != nil {
		raw = *seat.PlaybookRef
	}
	return CanonicalPlaybookRefString(raw)
}

func normalizeSeatStatus(st string) string {
	st = strings.ToLower(strings.TrimSpace(st))
	switch st {
	case models.SeatStatusActive, models.SeatStatusStopped, models.SeatStatusLeft, models.SeatStatusOpen:
		return st
	case "empty", "departed":
		if st == "departed" {
			return models.SeatStatusLeft
		}
		return models.SeatStatusOpen
	case "":
		return models.SeatStatusOpen
	default:
		return models.SeatStatusOpen
	}
}

func (s *Service) getSeat(workspaceID, teamID, seatID string) (*models.TeamSeat, error) {
	row := s.DB.QueryRow(`
		SELECT ts.id, ts.team_id, ts.workspace_id, ts.role_slug, ts.display_name, ts.playbook_ref,
			ts.bot_role_id, ts.bound_member_id, ts.bound_at, COALESCE(NULLIF(ts.status,''), 'open'),
			ts.created_at, ts.updated_at,
			m.display_name,
			c.member_id, c.claimed_at
		FROM team_seats ts
		LEFT JOIN members m ON m.id = ts.bound_member_id
		LEFT JOIN bot_role_claims c ON c.role_id = ts.bot_role_id AND c.released_at IS NULL
		WHERE ts.id = ? AND ts.team_id = ? AND ts.workspace_id = ?`, seatID, teamID, workspaceID)
	seat, err := scanRosterRow(row)
	if err == sql.ErrNoRows {
		return nil, apperr.NotFound("席位不存在")
	}
	if err != nil {
		return nil, err
	}
	return &seat, nil
}

func (s *Service) BindSeat(workspaceID, teamID, seatID, memberID string, force bool) (*models.TeamSeat, error) {
	memberID = strings.TrimSpace(memberID)
	if memberID == "" {
		return nil, apperr.BadRequest("member_id 不能为空")
	}
	var mid string
	err := s.DB.QueryRow(`SELECT id FROM members WHERE id = ? AND workspace_id = ?`, memberID, workspaceID).Scan(&mid)
	if err == sql.ErrNoRows {
		return nil, apperr.NotFound("成员不存在")
	}
	if err != nil {
		return nil, err
	}
	seat, err := s.getSeat(workspaceID, teamID, seatID)
	if err != nil {
		return nil, err
	}
	// T7-3 enter/bind：席位须有可读 playbook_ref；空/非法 → 4xx，不回落硬编码
	if _, err := seatPlaybookRefOrErr(seat); err != nil {
		return nil, err
	}
	prevStatus := normalizeSeatStatus(seat.Status)
	stoppedOrLeft := prevStatus == models.SeatStatusStopped || prevStatus == models.SeatStatusLeft
	writeForceAudit := false
	var fromMember string
	if seat.BoundMemberID != nil && *seat.BoundMemberID != "" && *seat.BoundMemberID != memberID {
		// 异人 active claim → 即使 force 也 409（活跃禁止顶替）
		if seat.BotRoleID != nil && *seat.BotRoleID != "" {
			active, aerr := s.activeClaim(*seat.BotRoleID)
			if aerr != nil {
				return nil, aerr
			}
			if active != nil && active.MemberID != memberID {
				return nil, apperr.Conflict("该席位占用中，禁止顶替")
			}
		}
		// 无 active claim 的残留 bound = 脏绑；stopped/left 可补位
		if !force && !stoppedOrLeft {
			msg := "该席位已绑定其他成员"
			if seat.BoundDisplayName != nil {
				msg = fmt.Sprintf("该席位已绑定其他成员（%s）", *seat.BoundDisplayName)
			}
			return nil, apperr.Conflict(msg)
		}
		fromMember = *seat.BoundMemberID
		if force {
			writeForceAudit = true
		}
		_, err = s.DB.Exec(`UPDATE team_seats SET bound_member_id = NULL, bound_at = NULL, updated_at = ? WHERE id = ?`,
			nowUTC(), seatID)
		if err != nil {
			return nil, err
		}
	} else if force && stoppedOrLeft {
		// 停工/离职席已清绑：force 成功仍须人话审计（N3 / J4 P0.1）
		writeForceAudit = true
		fromMember = s.lastClaimMemberID(seat)
	}
	now := nowUTC()
	_, err = s.DB.Exec(`UPDATE team_seats SET bound_member_id = ?, bound_at = ?, status = ?, updated_at = ? WHERE id = ?`,
		memberID, now, models.SeatStatusActive, now, seatID)
	if err != nil {
		return nil, err
	}
	if seat.BotRoleID != nil && *seat.BotRoleID != "" {
		if _, err := s.Claim(workspaceID, *seat.BotRoleID, memberID, false); err != nil {
			return nil, err
		}
	}
	if writeForceAudit {
		s.writeSeatForceAudit(workspaceID, memberID, seat, fromMember, memberID, prevStatus)
	}
	return s.getSeat(workspaceID, teamID, seatID)
}

// lastClaimMemberID 停工/离职清绑后仍可从 claim 历史还原原占用者（审计用）。
func (s *Service) lastClaimMemberID(seat *models.TeamSeat) string {
	if seat == nil || seat.BotRoleID == nil || *seat.BotRoleID == "" {
		return ""
	}
	var mid string
	err := s.DB.QueryRow(`SELECT member_id FROM bot_role_claims WHERE role_id = ? ORDER BY claimed_at DESC LIMIT 1`,
		*seat.BotRoleID).Scan(&mid)
	if err != nil {
		return ""
	}
	return mid
}

func seatStatusZh(st string) string {
	switch normalizeSeatStatus(st) {
	case models.SeatStatusStopped:
		return "停工"
	case models.SeatStatusLeft:
		return "离职"
	case models.SeatStatusActive:
		return "占用"
	case models.SeatStatusOpen:
		return "空席"
	default:
		if st == "" {
			return "未知"
		}
		return st
	}
}

func (s *Service) memberDisplayName(workspaceID, memberID string) string {
	if memberID == "" {
		return ""
	}
	var name string
	err := s.DB.QueryRow(`SELECT COALESCE(NULLIF(display_name,''), id) FROM members WHERE id = ? AND workspace_id = ?`,
		memberID, workspaceID).Scan(&name)
	if err != nil || name == "" {
		return memberID
	}
	return name
}

// writeSeatForceAudit force 成功后的人对人可见审计：workspace activities 人话句 + 关联开放 Issue comment。
// 仅 stopped/left / 脏绑 force 路径调用；禁止只 UpsertMemory。
func (s *Service) writeSeatForceAudit(workspaceID, actorID string, seat *models.TeamSeat, fromMember, toMember, fromStatus string) {
	toName := s.memberDisplayName(workspaceID, toMember)
	fromName := s.memberDisplayName(workspaceID, fromMember)
	fromZh := seatStatusZh(fromStatus)
	// active+force 成功 ⇒ 无 active claim 的脏绑（真活跃已在 BindSeat 409）
	if normalizeSeatStatus(fromStatus) == models.SeatStatusActive {
		fromZh = "脏绑"
	}
	var comment string
	if fromMember != "" {
		comment = fmt.Sprintf("%s 强制补位席位「%s」（%s）：原成员 %s（%s）→ 占用（force）",
			toName, seat.DisplayName, seat.RoleSlug, fromName, fromZh)
	} else {
		comment = fmt.Sprintf("%s 强制补位席位「%s」（%s）：%s → 占用（force）",
			toName, seat.DisplayName, seat.RoleSlug, fromZh)
	}
	payload, _ := json.Marshal(map[string]any{
		"comment":        comment,
		"summary":        comment,
		"seat_id":        seat.ID,
		"team_id":        seat.TeamID,
		"role_slug":      seat.RoleSlug,
		"from_member_id": fromMember,
		"to_member_id":   toMember,
		"from_status":    fromStatus,
		"to_status":      models.SeatStatusActive,
		"force":          true,
	})
	_, _ = s.DB.Exec(`INSERT INTO activities
		(id, workspace_id, issue_id, actor_id, actor_kind, action, payload, created_at)
		VALUES (?,?,NULL,?,?,?,?,?)`,
		idgen.New(), workspaceID, actorID, "bot", "seat.force_bind", string(payload), nowUTC())

	// 人对人可见：挂到原占用者（或补位者）名下开放 Issue 评论区
	s.writeSeatForceIssueComments(workspaceID, actorID, fromMember, toMember, comment)
}

func (s *Service) writeSeatForceIssueComments(workspaceID, authorID, fromMember, toMember, body string) {
	ids := map[string]struct{}{}
	for _, mid := range []string{fromMember, toMember} {
		if mid == "" {
			continue
		}
		rows, err := s.DB.Query(`SELECT id FROM issues
			WHERE workspace_id = ? AND assignee_id = ? AND status IN ('todo','in_progress','in_review')`,
			workspaceID, mid)
		if err != nil {
			continue
		}
		for rows.Next() {
			var id string
			if rows.Scan(&id) == nil && id != "" {
				ids[id] = struct{}{}
			}
		}
		rows.Close()
	}
	if len(ids) == 0 {
		return
	}
	ib := issueboard.New(s.DB)
	for issueID := range ids {
		_, _ = ib.CreateComment(workspaceID, issueID, issueboard.CreateCommentInput{
			Body:      body,
			AuthorID:  authorID,
			ActorKind: "bot",
		})
	}
}

func (s *Service) UnbindSeat(workspaceID, teamID, seatID, memberID string, asAdmin bool) (*models.TeamSeat, error) {
	seat, err := s.getSeat(workspaceID, teamID, seatID)
	if err != nil {
		return nil, err
	}
	if seat.BoundMemberID == nil || *seat.BoundMemberID == "" {
		return seat, nil
	}
	if !asAdmin && memberID != "" && *seat.BoundMemberID != memberID {
		return nil, apperr.Forbidden("只能解绑自己绑定的席位")
	}
	bound := *seat.BoundMemberID
	_, err = s.DB.Exec(`UPDATE team_seats SET bound_member_id = NULL, bound_at = NULL, status = ?, updated_at = ? WHERE id = ?`,
		models.SeatStatusLeft, nowUTC(), seatID)
	if err != nil {
		return nil, err
	}
	if seat.BotRoleID != nil && *seat.BotRoleID != "" {
		if active, aerr := s.activeClaim(*seat.BotRoleID); aerr == nil && active != nil && active.MemberID == bound {
			_ = s.Release(workspaceID, *seat.BotRoleID)
		}
	}
	return s.getSeat(workspaceID, teamID, seatID)
}

// CreateSeat 加新席（G2）；同 team+slug 幂等返回原席。
func (s *Service) CreateSeat(workspaceID, teamID string, in EnsureSeatInput) (*models.TeamSeat, error) {
	if _, err := s.GetTeam(workspaceID, teamID); err != nil {
		return nil, err
	}
	slug := strings.TrimSpace(strings.ToLower(in.Slug))
	name := strings.TrimSpace(in.Name)
	if slug == "" {
		return nil, apperr.BadRequest("席位 slug 不能为空")
	}
	if name == "" {
		name = slug
	}
	var existingID string
	err := s.DB.QueryRow(`SELECT id FROM team_seats WHERE team_id = ? AND role_slug = ?`, teamID, slug).Scan(&existingID)
	if err == nil {
		return s.getSeat(workspaceID, teamID, existingID)
	}
	if err != sql.ErrNoRows {
		return nil, err
	}
	playbook, err := CanonicalPlaybookRefString(in.PlaybookRef)
	if err != nil {
		return nil, err
	}
	role, err := s.ensureTeamRole(workspaceID, teamID, slug, name)
	if err != nil {
		return nil, err
	}
	now := nowUTC()
	sid := idgen.New()
	_, err = s.DB.Exec(`INSERT INTO team_seats
		(id, team_id, workspace_id, role_slug, display_name, playbook_ref, bot_role_id,
		 bound_member_id, bound_at, status, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,NULL,NULL,?,?,?)`,
		sid, teamID, workspaceID, slug, name, playbook, role.ID, models.SeatStatusOpen, now, now)
	if err != nil {
		if db.IsUniqueViolation(err) {
			var id2 string
			if e2 := s.DB.QueryRow(`SELECT id FROM team_seats WHERE team_id = ? AND role_slug = ?`, teamID, slug).Scan(&id2); e2 == nil {
				return s.getSeat(workspaceID, teamID, id2)
			}
		}
		return nil, err
	}
	return s.getSeat(workspaceID, teamID, sid)
}

// StopSeat 标记停工：清绑+释 claim，status=stopped（可 force 补位）。
func (s *Service) StopSeat(workspaceID, teamID, seatID string) (*models.TeamSeat, error) {
	seat, err := s.getSeat(workspaceID, teamID, seatID)
	if err != nil {
		return nil, err
	}
	if seat.BotRoleID != nil && *seat.BotRoleID != "" {
		if active, aerr := s.activeClaim(*seat.BotRoleID); aerr == nil && active != nil {
			_ = s.Release(workspaceID, *seat.BotRoleID)
		}
	}
	_, err = s.DB.Exec(`UPDATE team_seats SET bound_member_id = NULL, bound_at = NULL, status = ?, updated_at = ? WHERE id = ?`,
		models.SeatStatusStopped, nowUTC(), seatID)
	if err != nil {
		return nil, err
	}
	return s.getSeat(workspaceID, teamID, seatID)
}

// ExpandInput 补位扩工：把未领叶子派给空闲已绑席位。
type ExpandInput struct {
	ProjectID string
	Limit     int
}

type ExpandAssigned struct {
	IssueID    string `json:"issue_id"`
	Identifier string `json:"identifier"`
	SeatID     string `json:"seat_id"`
	MemberID   string `json:"member_id"`
}

type ExpandSkipped struct {
	IssueID string `json:"issue_id"`
	Reason  string `json:"reason"`
}

type ExpandResult struct {
	Assigned []ExpandAssigned `json:"assigned"`
	Skipped  []ExpandSkipped  `json:"skipped"`
}

func (s *Service) ExpandSeats(workspaceID, teamID string, in ExpandInput) (*ExpandResult, error) {
	if _, err := s.GetTeam(workspaceID, teamID); err != nil {
		return nil, err
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 50
	}
	roster, err := s.GetRoster(workspaceID, teamID)
	if err != nil {
		return nil, err
	}
	type idleSeat struct {
		seatID, memberID string
	}
	var idle []idleSeat
	for _, seat := range roster {
		if seat.BoundMemberID == nil || *seat.BoundMemberID == "" {
			continue
		}
		// claim 空闲：无 active，或 active 就是绑定成员
		if seat.Claim != nil && seat.Claim.MemberID != *seat.BoundMemberID {
			continue
		}
		idle = append(idle, idleSeat{seatID: seat.ID, memberID: *seat.BoundMemberID})
	}
	res := &ExpandResult{Assigned: []ExpandAssigned{}, Skipped: []ExpandSkipped{}}
	if len(idle) == 0 {
		return res, nil
	}

	args := []any{workspaceID}
	q := `
		SELECT i.id, i.identifier, i.version, i.assignee_id, i.status
		FROM issues i
		WHERE i.workspace_id = ?
		  AND i.assignee_id IS NULL
		  AND i.status IN ('backlog','todo')
		  AND NOT EXISTS (SELECT 1 FROM issues c WHERE c.parent_id = i.id)`
	if strings.TrimSpace(in.ProjectID) != "" {
		q += ` AND i.project_id = ?`
		args = append(args, strings.TrimSpace(in.ProjectID))
	}
	q += ` ORDER BY i.created_at, i.number LIMIT ?`
	args = append(args, limit)

	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type cand struct {
		id, ident string
		version   int
	}
	var candidates []cand
	for rows.Next() {
		var c cand
		var assignee sql.NullString
		var status string
		if err := rows.Scan(&c.id, &c.ident, &c.version, &assignee, &status); err != nil {
			return nil, err
		}
		if assignee.Valid && assignee.String != "" {
			res.Skipped = append(res.Skipped, ExpandSkipped{IssueID: c.id, Reason: "已有负责人"})
			continue
		}
		if status == "in_progress" || status == "done" || status == "canceled" {
			res.Skipped = append(res.Skipped, ExpandSkipped{IssueID: c.id, Reason: "状态不可派发"})
			continue
		}
		candidates = append(candidates, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	ib := issueboard.New(s.DB)
	ib.DispatchEvent = s.IssueDispatch
	si := 0
	for _, c := range candidates {
		if si >= len(idle) {
			res.Skipped = append(res.Skipped, ExpandSkipped{IssueID: c.id, Reason: "无更多空闲席位"})
			continue
		}
		seat := idle[si]
		mid := seat.memberID
		_, err := ib.UpdateIssue(workspaceID, c.id, issueboard.UpdateIssueInput{
			Version:    c.version,
			AssigneeID: &mid,
			ActorID:    mid,
			ActorKind:  "bot",
		})
		if err != nil {
			reason := err.Error()
			if ae, ok := err.(*apperr.AppError); ok {
				reason = ae.Message
			}
			res.Skipped = append(res.Skipped, ExpandSkipped{IssueID: c.id, Reason: reason})
			continue
		}
		res.Assigned = append(res.Assigned, ExpandAssigned{
			IssueID: c.id, Identifier: c.ident, SeatID: seat.seatID, MemberID: seat.memberID,
		})
		si++
	}
	return res, nil
}
