package botmem

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"strings"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/models"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
)

const teamInvitePrefix = "ymt_"

// CreateInviteInput 签发入场码；须 seat_slug 或 role_hint（role_slug 由 handler 归一到 RoleHint）。
type CreateInviteInput struct {
	AllowExpandSeat bool
	SeatSlug        string
	RoleHint        string
	MaxUses         *int
	ExpiresAt       string // RFC3339 optional
	CreatedBy       string
	// Replace=true：先幂等吊销该席/角色下 active ymt_ 再发 1；缺省/false 可不吊旧直接再发 1。
	Replace bool
}

// ListInvitesFilter 可选过滤；空字段忽略。
type ListInvitesFilter struct {
	SeatSlug string
	Status   string
}

func hashInviteToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func newInviteCode() (plain, prefix, hash string, err error) {
	b := make([]byte, 24)
	if _, err = rand.Read(b); err != nil {
		return "", "", "", err
	}
	plain = teamInvitePrefix + hex.EncodeToString(b)
	prefix = plain
	if len(prefix) > 12 {
		prefix = prefix[:12]
	}
	hash = hashInviteToken(plain)
	return plain, prefix, hash, nil
}

// CreateInvite 签发 ymt_ 入场码；明文仅返回一次。
// I-G1：须 seat_slug 或 role_hint；活跃席 → 409；replace 先吊旧。
func (s *Service) CreateInvite(workspaceID, teamID string, in CreateInviteInput) (*models.TeamInvite, error) {
	if _, err := s.GetTeam(workspaceID, teamID); err != nil {
		return nil, err
	}
	seatSlug := strings.TrimSpace(strings.ToLower(in.SeatSlug))
	roleHint := strings.TrimSpace(in.RoleHint)
	if seatSlug == "" && roleHint == "" {
		return nil, apperr.BadRequest("须指定 seat_slug 或 role_hint/role_slug")
	}
	if roleHint == "" && seatSlug != "" {
		roleHint = seatSlug
	}

	// 活跃席禁顶替：目标席 occupancy=active → 409
	if seat, _ := s.findSeatForInvite(workspaceID, teamID, seatSlug, roleHint); seat != nil {
		if isSeatActiveOccupancy(seat) {
			return nil, apperr.Conflict("席位已有人在席，禁止签发顶替入场码")
		}
	}

	if in.Replace {
		if err := s.revokeActiveInvitesForSeatRole(workspaceID, teamID, seatSlug, roleHint); err != nil {
			return nil, err
		}
	}

	plain, prefix, hash, err := newInviteCode()
	if err != nil {
		return nil, err
	}
	now := nowUTC()
	id := idgen.New()
	var seatSlugVal, roleHintVal, expires, createdBy any
	if seatSlug != "" {
		seatSlugVal = seatSlug
	}
	if roleHint != "" {
		roleHintVal = roleHint
	}
	if v := strings.TrimSpace(in.ExpiresAt); v != "" {
		if _, err := time.Parse(time.RFC3339, v); err != nil {
			if _, err2 := time.Parse(time.RFC3339Nano, v); err2 != nil {
				return nil, apperr.BadRequest("expires_at 须为 RFC3339")
			}
		}
		expires = v
	}
	if v := strings.TrimSpace(in.CreatedBy); v != "" {
		createdBy = v
	}
	var maxUses any
	if in.MaxUses != nil {
		if *in.MaxUses < 0 {
			return nil, apperr.BadRequest("max_uses 不能为负")
		}
		maxUses = *in.MaxUses
	}
	allow := 0
	if in.AllowExpandSeat {
		allow = 1
	}
	_, err = s.DB.Exec(`INSERT INTO team_invites
		(id, workspace_id, team_id, token_hash, token_prefix, allow_expand_seat, seat_slug, role_hint,
		 max_uses, use_count, expires_at, revoked_at, created_by, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,0,?,NULL,?,?,?)`,
		id, workspaceID, teamID, hash, prefix, allow, seatSlugVal, roleHintVal,
		maxUses, expires, createdBy, now, now)
	if err != nil {
		return nil, err
	}
	inv, err := s.getInviteByID(workspaceID, id)
	if err != nil {
		return nil, err
	}
	inv.Code = plain
	s.attachInviteLabel(workspaceID, teamID, inv)
	return inv, nil
}

func (s *Service) getInviteByID(workspaceID, id string) (*models.TeamInvite, error) {
	row := s.DB.QueryRow(`
		SELECT id, workspace_id, team_id, token_prefix, allow_expand_seat, seat_slug, role_hint,
			max_uses, use_count, expires_at, revoked_at, created_by, created_at, updated_at
		FROM team_invites WHERE id = ? AND workspace_id = ?`, id, workspaceID)
	inv, err := scanInvite(row)
	if err == sql.ErrNoRows {
		return nil, apperr.NotFound("入场码不存在")
	}
	if err != nil {
		return nil, err
	}
	enrichInvite(&inv)
	return &inv, nil
}

func scanInvite(row interface{ Scan(dest ...any) error }) (models.TeamInvite, error) {
	var inv models.TeamInvite
	var allow int
	var seatSlug, roleHint, expires, revoked, createdBy sql.NullString
	var maxUses sql.NullInt64
	err := row.Scan(
		&inv.ID, &inv.WorkspaceID, &inv.TeamID, &inv.TokenPrefix, &allow,
		&seatSlug, &roleHint, &maxUses, &inv.UseCount, &expires, &revoked, &createdBy,
		&inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		return inv, err
	}
	inv.AllowExpandSeat = allow != 0
	if seatSlug.Valid && seatSlug.String != "" {
		inv.SeatSlug = &seatSlug.String
	}
	if roleHint.Valid && roleHint.String != "" {
		inv.RoleHint = &roleHint.String
	}
	if maxUses.Valid {
		v := int(maxUses.Int64)
		inv.MaxUses = &v
	}
	if expires.Valid && expires.String != "" {
		inv.ExpiresAt = &expires.String
	}
	if revoked.Valid && revoked.String != "" {
		inv.RevokedAt = &revoked.String
	}
	if createdBy.Valid && createdBy.String != "" {
		inv.CreatedBy = &createdBy.String
	}
	return inv, nil
}

// RedeemInvite 兑换 ymt_；递增 use_count；不返回明文。
func (s *Service) RedeemInvite(code string) (*models.TeamInviteRedeemResult, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, apperr.BadRequest("入场码不能为空")
	}
	if !strings.HasPrefix(code, teamInvitePrefix) {
		return nil, apperr.BadRequest("入场码须以 ymt_ 开头")
	}
	hash := hashInviteToken(code)
	var id, ws, teamID string
	var allow int
	var seatSlug, roleHint, expires, revoked sql.NullString
	var maxUses sql.NullInt64
	var useCount int
	err := s.DB.QueryRow(`
		SELECT id, workspace_id, team_id, allow_expand_seat, seat_slug, role_hint,
			max_uses, use_count, expires_at, revoked_at
		FROM team_invites WHERE token_hash = ?`, hash).Scan(
		&id, &ws, &teamID, &allow, &seatSlug, &roleHint, &maxUses, &useCount, &expires, &revoked)
	if err == sql.ErrNoRows {
		return nil, apperr.NotFound("入场码无效或已失效")
	}
	if err != nil {
		return nil, err
	}
	if revoked.Valid && revoked.String != "" {
		return nil, apperr.Forbidden("入场码已吊销")
	}
	if expires.Valid && expires.String != "" {
		exp, perr := time.Parse(time.RFC3339Nano, expires.String)
		if perr != nil {
			exp, perr = time.Parse(time.RFC3339, expires.String)
		}
		if perr == nil && time.Now().UTC().After(exp) {
			return nil, apperr.Forbidden("入场码已过期")
		}
	}
	if maxUses.Valid && int(maxUses.Int64) > 0 && useCount >= int(maxUses.Int64) {
		return nil, apperr.Forbidden("入场码已达使用上限")
	}
	_, err = s.DB.Exec(`UPDATE team_invites SET use_count = use_count + 1, updated_at = ? WHERE id = ?`,
		nowUTC(), id)
	if err != nil {
		return nil, err
	}
	out := &models.TeamInviteRedeemResult{
		WorkspaceID:     ws,
		TeamID:          teamID,
		AllowExpandSeat: allow != 0,
	}
	if seatSlug.Valid && seatSlug.String != "" {
		out.SeatSlug = &seatSlug.String
	}
	if roleHint.Valid && roleHint.String != "" {
		out.RoleHint = &roleHint.String
	} else if out.SeatSlug != nil {
		out.RoleHint = out.SeatSlug
	}
	return out, nil
}

// RevokeInvite 吊销入场码（Admin）。已吊销再调仍成功，返回 status=revoked（幂等）。
func (s *Service) RevokeInvite(workspaceID, teamID, inviteID string) (*models.TeamInvite, error) {
	inv, err := s.getInviteByID(workspaceID, inviteID)
	if err != nil {
		return nil, err
	}
	if inv.TeamID != teamID {
		return nil, apperr.NotFound("入场码不存在")
	}
	if inv.RevokedAt != nil && strings.TrimSpace(*inv.RevokedAt) != "" {
		inv.Code = ""
		s.attachInviteLabel(workspaceID, teamID, inv)
		return inv, nil
	}
	now := nowUTC()
	_, err = s.DB.Exec(`UPDATE team_invites SET revoked_at = ?, updated_at = ?
		WHERE id = ? AND workspace_id = ? AND team_id = ? AND revoked_at IS NULL`,
		now, now, inviteID, workspaceID, teamID)
	if err != nil {
		return nil, err
	}
	inv, err = s.getInviteByID(workspaceID, inviteID)
	if err != nil {
		return nil, err
	}
	inv.Code = ""
	s.attachInviteLabel(workspaceID, teamID, inv)
	return inv, nil
}

// enrichInvite 填充 status / code_tail；不设明文 code（Create 在调用后自行赋值）。
func enrichInvite(inv *models.TeamInvite) {
	if inv == nil {
		return
	}
	inv.CodeTail = inviteCodeTail(inv.TokenPrefix)
	inv.Status = computeInviteStatus(inv)
}

func inviteCodeTail(prefix string) string {
	clean := strings.TrimPrefix(prefix, teamInvitePrefix)
	if clean == "" {
		return ""
	}
	if len(clean) <= 4 {
		return clean
	}
	return clean[len(clean)-4:]
}

func computeInviteStatus(inv *models.TeamInvite) string {
	if inv.RevokedAt != nil && strings.TrimSpace(*inv.RevokedAt) != "" {
		return models.InviteStatusRevoked
	}
	if inv.ExpiresAt != nil && strings.TrimSpace(*inv.ExpiresAt) != "" {
		exp, perr := time.Parse(time.RFC3339Nano, *inv.ExpiresAt)
		if perr != nil {
			exp, perr = time.Parse(time.RFC3339, *inv.ExpiresAt)
		}
		if perr == nil && time.Now().UTC().After(exp) {
			return models.InviteStatusExpired
		}
	}
	if inv.MaxUses != nil && *inv.MaxUses > 0 && inv.UseCount >= *inv.MaxUses {
		return models.InviteStatusExhausted
	}
	return models.InviteStatusActive
}

// ListInvites Admin 列团队已签发入场码；不含明文 code / token_hash。
// 可选 filter：seat_slug、status（如 active）。每项附 label。
func (s *Service) ListInvites(workspaceID, teamID string, filter ...ListInvitesFilter) ([]models.TeamInvite, error) {
	if _, err := s.GetTeam(workspaceID, teamID); err != nil {
		return nil, err
	}
	var f ListInvitesFilter
	if len(filter) > 0 {
		f = filter[0]
	}
	f.SeatSlug = strings.TrimSpace(strings.ToLower(f.SeatSlug))
	f.Status = strings.TrimSpace(strings.ToLower(f.Status))

	// roster 须在打开 invites rows 之前取：sqlite MaxOpenConns=1，持 rows 再 Query 会死锁。
	roster, _ := s.GetRoster(workspaceID, teamID)
	labelBySlug := inviteSeatLabelMap(roster)

	rows, err := s.DB.Query(`
		SELECT id, workspace_id, team_id, token_prefix, allow_expand_seat, seat_slug, role_hint,
			max_uses, use_count, expires_at, revoked_at, created_by, created_at, updated_at
		FROM team_invites
		WHERE workspace_id = ? AND team_id = ?
		ORDER BY created_at DESC`, workspaceID, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.TeamInvite, 0)
	for rows.Next() {
		inv, err := scanInvite(rows)
		if err != nil {
			return nil, err
		}
		enrichInvite(&inv)
		inv.Code = ""
		if f.SeatSlug != "" {
			ss := ""
			if inv.SeatSlug != nil {
				ss = strings.ToLower(*inv.SeatSlug)
			}
			rh := ""
			if inv.RoleHint != nil {
				rh = strings.ToLower(strings.TrimSpace(*inv.RoleHint))
			}
			if ss != f.SeatSlug && rh != f.SeatSlug {
				continue
			}
		}
		if f.Status != "" && inv.Status != f.Status {
			continue
		}
		inv.Label = resolveInviteLabel(&inv, labelBySlug)
		out = append(out, inv)
	}
	return out, rows.Err()
}

// LookupEntry I-G1：按席/角色查可用入场凭证（优先未吊销 ymk_，否则 active ymt_）。
// seat_slug 与 role_slug 二选一必填；都给时优先 seat_slug。都无 → 400；都无可用 → 404。
func (s *Service) LookupEntry(workspaceID, teamID, seatSlug, roleSlug string) (*models.TeamEntryLookup, error) {
	if _, err := s.GetTeam(workspaceID, teamID); err != nil {
		return nil, err
	}
	seatSlug = strings.TrimSpace(strings.ToLower(seatSlug))
	roleSlug = strings.TrimSpace(strings.ToLower(roleSlug))
	if seatSlug == "" && roleSlug == "" {
		return nil, apperr.BadRequest("须指定 seat_slug 或 role_slug")
	}
	lookupKey := seatSlug
	if lookupKey == "" {
		lookupKey = roleSlug
	}

	seat, _ := s.findSeatForInvite(workspaceID, teamID, seatSlug, roleSlug)
	if seat != nil && isSeatActiveOccupancy(seat) && seat.BoundMemberID != nil && *seat.BoundMemberID != "" {
		memberID := *seat.BoundMemberID
		keyID, prefix, ok := s.findUnrevokedAPIKey(workspaceID, memberID)
		if ok {
			dn := ""
			if seat.BoundDisplayName != nil {
				dn = *seat.BoundDisplayName
			}
			if dn == "" {
				dn = seat.DisplayName
			}
			return &models.TeamEntryLookup{
				Kind:        "ymk_",
				KeyID:       keyID,
				TokenPrefix: prefix,
				TokenTail:   apiKeyTail(prefix),
				MemberID:    memberID,
				DisplayName: dn,
			}, nil
		}
	}

	list, err := s.ListInvites(workspaceID, teamID, ListInvitesFilter{SeatSlug: lookupKey, Status: models.InviteStatusActive})
	if err != nil {
		return nil, err
	}
	if len(list) == 0 && seatSlug != "" && roleSlug != "" && seatSlug != roleSlug {
		list, err = s.ListInvites(workspaceID, teamID, ListInvitesFilter{SeatSlug: roleSlug, Status: models.InviteStatusActive})
		if err != nil {
			return nil, err
		}
	}
	if len(list) > 0 {
		inv := list[0]
		ss := lookupKey
		if inv.SeatSlug != nil && *inv.SeatSlug != "" {
			ss = *inv.SeatSlug
		}
		return &models.TeamEntryLookup{
			Kind:     "ymt_",
			InviteID: inv.ID,
			CodeTail: inv.CodeTail,
			Status:   inv.Status,
			SeatSlug: ss,
		}, nil
	}
	return nil, apperr.NotFound("该席位暂无可用入场凭证")
}

func (s *Service) findUnrevokedAPIKey(workspaceID, memberID string) (id, prefix string, ok bool) {
	err := s.DB.QueryRow(`
		SELECT id, key_prefix FROM api_keys
		WHERE workspace_id = ? AND member_id = ?
		  AND (revoked_at IS NULL OR revoked_at = '')
		ORDER BY created_at DESC
		LIMIT 1`, workspaceID, memberID).Scan(&id, &prefix)
	if err != nil {
		return "", "", false
	}
	return id, prefix, true
}

func apiKeyTail(prefix string) string {
	clean := strings.TrimPrefix(prefix, "ymk_")
	if clean == "" {
		return ""
	}
	if len(clean) <= 4 {
		return clean
	}
	return clean[len(clean)-4:]
}

func (s *Service) findSeatForInvite(workspaceID, teamID, seatSlug, roleHint string) (*models.TeamSeat, error) {
	roster, err := s.GetRoster(workspaceID, teamID)
	if err != nil {
		return nil, err
	}
	seatSlug = strings.TrimSpace(strings.ToLower(seatSlug))
	roleHint = strings.TrimSpace(strings.ToLower(roleHint))
	if seatSlug != "" {
		for i := range roster {
			if strings.ToLower(roster[i].RoleSlug) == seatSlug {
				return &roster[i], nil
			}
		}
	}
	if roleHint != "" {
		for i := range roster {
			if strings.ToLower(roster[i].RoleSlug) == roleHint {
				return &roster[i], nil
			}
		}
	}
	return nil, nil
}

func isSeatActiveOccupancy(seat *models.TeamSeat) bool {
	if seat == nil {
		return false
	}
	st := normalizeSeatStatus(seat.Status)
	if st != models.SeatStatusActive {
		return false
	}
	return seat.BoundMemberID != nil && strings.TrimSpace(*seat.BoundMemberID) != ""
}

func (s *Service) revokeActiveInvitesForSeatRole(workspaceID, teamID, seatSlug, roleHint string) error {
	list, err := s.ListInvites(workspaceID, teamID)
	if err != nil {
		return err
	}
	seatSlug = strings.TrimSpace(strings.ToLower(seatSlug))
	roleHint = strings.TrimSpace(strings.ToLower(roleHint))
	for _, inv := range list {
		if inv.Status != models.InviteStatusActive {
			continue
		}
		if !inviteMatchesSeatRole(&inv, seatSlug, roleHint) {
			continue
		}
		if _, err := s.RevokeInvite(workspaceID, teamID, inv.ID); err != nil {
			return err
		}
	}
	return nil
}

func inviteMatchesSeatRole(inv *models.TeamInvite, seatSlug, roleHint string) bool {
	if inv == nil {
		return false
	}
	ss, rh := "", ""
	if inv.SeatSlug != nil {
		ss = strings.ToLower(strings.TrimSpace(*inv.SeatSlug))
	}
	if inv.RoleHint != nil {
		rh = strings.ToLower(strings.TrimSpace(*inv.RoleHint))
	}
	if seatSlug != "" && (ss == seatSlug || rh == seatSlug) {
		return true
	}
	if roleHint != "" && (ss == roleHint || rh == roleHint) {
		return true
	}
	return false
}

func inviteSeatLabelMap(roster []models.TeamSeat) map[string]string {
	m := make(map[string]string, len(roster))
	for _, seat := range roster {
		slug := strings.ToLower(strings.TrimSpace(seat.RoleSlug))
		if slug == "" {
			continue
		}
		name := strings.TrimSpace(seat.DisplayName)
		if name == "" || looksLikeDirtySeatKey(name) {
			name = "未命名席"
		}
		m[slug] = name
	}
	return m
}

func resolveInviteLabel(inv *models.TeamInvite, labelBySlug map[string]string) string {
	keys := make([]string, 0, 2)
	if inv.SeatSlug != nil && strings.TrimSpace(*inv.SeatSlug) != "" {
		keys = append(keys, strings.ToLower(strings.TrimSpace(*inv.SeatSlug)))
	}
	if inv.RoleHint != nil && strings.TrimSpace(*inv.RoleHint) != "" {
		keys = append(keys, strings.ToLower(strings.TrimSpace(*inv.RoleHint)))
	}
	for _, k := range keys {
		if lab, ok := labelBySlug[k]; ok && lab != "" {
			return lab
		}
	}
	return "未命名席"
}

func looksLikeDirtySeatKey(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return true
	}
	if strings.HasPrefix(s, "team:") && strings.Count(s, ":") >= 2 {
		return true
	}
	return false
}

func (s *Service) attachInviteLabel(workspaceID, teamID string, inv *models.TeamInvite) {
	if inv == nil {
		return
	}
	roster, err := s.GetRoster(workspaceID, teamID)
	if err != nil {
		inv.Label = "未命名席"
		return
	}
	inv.Label = resolveInviteLabel(inv, inviteSeatLabelMap(roster))
}
