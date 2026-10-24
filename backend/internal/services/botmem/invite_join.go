package botmem

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"strings"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/models"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
)

// RedeemJoinInput HUMAN_ONBOARD redeem 自动身份入参。
// Provision=true：无 Bearer bootstrap（造 is_bot + mint ymk_ + 绑 empty）。
// ExistingMemberID：持有效 ymk_ 续队（幂等绑席、不重发 token）。
// 二者皆空且 Provision=false：人类 JWT 经典定位（不造 Bot）。
type RedeemJoinInput struct {
	Code                 string
	Role                 string
	SeatID               string
	DisplayName          string
	ExistingMemberID     string
	ExistingWorkspaceID  string
	Provision            bool
}

const v1ReuseMsg = "请粘贴已发的身份钥匙续队"

type inviteRow struct {
	ID            string
	WorkspaceID   string
	TeamID        string
	AllowExpand   bool
	SeatSlug      string
	RoleHint      string
	MaxUses       sql.NullInt64
	UseCount      int
	ExpiresAt     sql.NullString
	RevokedAt     sql.NullString
	Bootstrapped  sql.NullString
}

func (s *Service) loadInviteByCode(code string) (*inviteRow, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, apperr.BadRequest("入场码不能为空")
	}
	if !strings.HasPrefix(code, teamInvitePrefix) {
		return nil, apperr.BadRequest("入场码须以 ymt_ 开头")
	}
	hash := hashInviteToken(code)
	var row inviteRow
	var allow int
	var seatSlug, roleHint sql.NullString
	err := s.DB.QueryRow(`
		SELECT id, workspace_id, team_id, allow_expand_seat, seat_slug, role_hint,
			max_uses, use_count, expires_at, revoked_at, bootstrapped_member_id
		FROM team_invites WHERE token_hash = ?`, hash).Scan(
		&row.ID, &row.WorkspaceID, &row.TeamID, &allow, &seatSlug, &roleHint,
		&row.MaxUses, &row.UseCount, &row.ExpiresAt, &row.RevokedAt, &row.Bootstrapped)
	if err == sql.ErrNoRows {
		return nil, apperr.NotFound("入场码无效或已失效")
	}
	if err != nil {
		// 旧库无 bootstrapped_member_id 列时回落（极罕见；sqlite schema 已加 ALTER）
		err2 := s.DB.QueryRow(`
			SELECT id, workspace_id, team_id, allow_expand_seat, seat_slug, role_hint,
				max_uses, use_count, expires_at, revoked_at
			FROM team_invites WHERE token_hash = ?`, hash).Scan(
			&row.ID, &row.WorkspaceID, &row.TeamID, &allow, &seatSlug, &roleHint,
			&row.MaxUses, &row.UseCount, &row.ExpiresAt, &row.RevokedAt)
		if err2 == sql.ErrNoRows {
			return nil, apperr.NotFound("入场码无效或已失效")
		}
		if err2 != nil {
			return nil, err
		}
	}
	row.AllowExpand = allow != 0
	if seatSlug.Valid {
		row.SeatSlug = seatSlug.String
	}
	if roleHint.Valid {
		row.RoleHint = roleHint.String
	}
	if row.RevokedAt.Valid && row.RevokedAt.String != "" {
		return nil, apperr.Forbidden("入场码已吊销")
	}
	if row.ExpiresAt.Valid && row.ExpiresAt.String != "" {
		exp, perr := time.Parse(time.RFC3339Nano, row.ExpiresAt.String)
		if perr != nil {
			exp, perr = time.Parse(time.RFC3339, row.ExpiresAt.String)
		}
		if perr == nil && time.Now().UTC().After(exp) {
			return nil, apperr.Forbidden("入场码已过期")
		}
	}
	if row.MaxUses.Valid && int(row.MaxUses.Int64) > 0 && row.UseCount >= int(row.MaxUses.Int64) {
		return nil, apperr.Forbidden("入场码已达使用上限")
	}
	return &row, nil
}

func inviteRedeemBase(row *inviteRow) *models.TeamInviteRedeemResult {
	out := &models.TeamInviteRedeemResult{
		WorkspaceID:     row.WorkspaceID,
		TeamID:          row.TeamID,
		AllowExpandSeat: row.AllowExpand,
	}
	if row.SeatSlug != "" {
		ss := row.SeatSlug
		out.SeatSlug = &ss
	}
	if row.RoleHint != "" {
		rh := row.RoleHint
		out.RoleHint = &rh
	} else if out.SeatSlug != nil {
		out.RoleHint = out.SeatSlug
	}
	return out
}

func (s *Service) bumpInviteUse(inviteID string) error {
	_, err := s.DB.Exec(`UPDATE team_invites SET use_count = use_count + 1, updated_at = ? WHERE id = ?`,
		nowUTC(), inviteID)
	return err
}

// RedeemInviteAndJoin HUMAN_ONBOARD：无 Bearer bootstrap / ymk_ 幂等 / 人类只定位。
// V1：同码无钥第二次拒造人；V2：不自动补 left/stopped；V6：一席一码+事务选席。
func (s *Service) RedeemInviteAndJoin(in RedeemJoinInput) (*models.TeamInviteRedeemResult, error) {
	row, err := s.loadInviteByCode(in.Code)
	if err != nil {
		return nil, err
	}
	out := inviteRedeemBase(row)

	existing := strings.TrimSpace(in.ExistingMemberID)
	if existing != "" {
		return s.redeemWithExisting(row, out, in)
	}
	if in.Provision {
		return s.redeemBootstrap(row, out, in)
	}
	// 人类 JWT / 经典定位：只定位，不造 Bot
	if err := s.bumpInviteUse(row.ID); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Service) redeemWithExisting(row *inviteRow, out *models.TeamInviteRedeemResult, in RedeemJoinInput) (*models.TeamInviteRedeemResult, error) {
	mid := strings.TrimSpace(in.ExistingMemberID)
	var wsCheck string
	err := s.DB.QueryRow(`SELECT workspace_id FROM members WHERE id = ?`, mid).Scan(&wsCheck)
	if err == sql.ErrNoRows {
		return nil, apperr.Unauthorized("无效的 API Key")
	}
	if err != nil {
		return nil, err
	}
	if wsCheck != row.WorkspaceID {
		return nil, apperr.Forbidden("身份钥匙不属于该入场码所在工作区")
	}
	if ew := strings.TrimSpace(in.ExistingWorkspaceID); ew != "" && ew != row.WorkspaceID {
		return nil, apperr.Forbidden("身份钥匙不属于该入场码所在工作区")
	}

	seat, bindErr := s.pickAndBindEmptySeat(row, mid, in, false)
	if bindErr != nil {
		// 已绑本席 / 无可绑空席：仍算续队成功（幂等定位）
		if ae, ok := bindErr.(*apperr.AppError); ok && ae.Code == "conflict" {
			// active 顶替冲突要透传
			if strings.Contains(ae.Message, "占用") || strings.Contains(ae.Message, "顶替") || strings.Contains(ae.Message, "已绑定其他") {
				return nil, bindErr
			}
		}
		// 无空席可绑：返回定位，不造新身份
		roster, _ := s.GetRoster(row.WorkspaceID, row.TeamID)
		for _, se := range roster {
			if se.BoundMemberID != nil && *se.BoundMemberID == mid {
				out.SeatID = se.ID
				out.MemberID = mid
				out.RoleSlug = se.RoleSlug
				break
			}
		}
		if out.MemberID == "" {
			out.MemberID = mid
		}
		if err := s.bumpInviteUse(row.ID); err != nil {
			return nil, err
		}
		return out, nil
	}
	if seat != nil {
		out.SeatID = seat.ID
		out.RoleSlug = seat.RoleSlug
	}
	out.MemberID = mid
	if err := s.bumpInviteUse(row.ID); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Service) redeemBootstrap(row *inviteRow, out *models.TeamInviteRedeemResult, in RedeemJoinInput) (*models.TeamInviteRedeemResult, error) {
	// V1：该码已成功 bootstrap → 拒造第二人
	if row.Bootstrapped.Valid && strings.TrimSpace(row.Bootstrapped.String) != "" {
		return nil, apperr.Conflict(v1ReuseMsg)
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// 事务内再读一遍（V6 并发）
	var boot sql.NullString
	var useCount int
	var maxUses sql.NullInt64
	err = tx.QueryRow(`SELECT bootstrapped_member_id, use_count, max_uses FROM team_invites WHERE id = ?`, row.ID).
		Scan(&boot, &useCount, &maxUses)
	if err != nil {
		return nil, err
	}
	if boot.Valid && strings.TrimSpace(boot.String) != "" {
		return nil, apperr.Conflict(v1ReuseMsg)
	}
	if maxUses.Valid && int(maxUses.Int64) > 0 && useCount >= int(maxUses.Int64) {
		return nil, apperr.Forbidden("入场码已达使用上限")
	}

	display := strings.TrimSpace(in.DisplayName)
	if display == "" {
		hint := row.RoleHint
		if hint == "" {
			hint = row.SeatSlug
		}
		if hint == "" {
			hint = "Bot"
		}
		display = hint
	}
	memberID := idgen.New()
	now := nowUTC()
	_, err = tx.Exec(`INSERT INTO members (id, workspace_id, user_id, role, display_name, is_bot, bot_enabled, created_at)
		VALUES (?,?,NULL,'bot',?,?,?,?)`, memberID, row.WorkspaceID, display, 1, 1, now)
	if err != nil {
		return nil, err
	}

	// 预检席位：active 占用 → 409；left/stopped → V2 不绑；open 空席 → 绑
	seatID, roleSlug, skipBind, perr := s.resolveBootstrapSeat(tx, row, in)
	if perr != nil {
		return nil, perr
	}

	rawToken, prefix, keyHash, err := mintAPIKeyPlain()
	if err != nil {
		return nil, err
	}
	keyID := idgen.New()
	_, err = tx.Exec(`INSERT INTO api_keys (id, workspace_id, member_id, name, key_prefix, key_hash, created_at)
		VALUES (?,?,?,?,?,?,?)`, keyID, row.WorkspaceID, memberID, "join-bootstrap", prefix, keyHash, now)
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, apperr.Conflict("API Key 哈希冲突，请重试")
		}
		return nil, err
	}

	_, err = tx.Exec(`UPDATE team_invites SET use_count = use_count + 1, bootstrapped_member_id = ?, updated_at = ? WHERE id = ?`,
		memberID, now, row.ID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	out.MemberID = memberID
	out.Token = rawToken
	out.TokenIssued = true

	if !skipBind && seatID != "" {
		bound, berr := s.BindSeat(row.WorkspaceID, row.TeamID, seatID, memberID, false)
		if berr != nil {
			// 成员与钥已落库；绑席失败不回滚身份（调用方可再 bind）
			if ae, ok := berr.(*apperr.AppError); ok && ae.Code == "conflict" {
				return nil, berr
			}
			return nil, berr
		}
		if bound != nil {
			out.SeatID = bound.ID
			out.RoleSlug = bound.RoleSlug
		}
	} else if roleSlug != "" {
		out.RoleSlug = roleSlug
	}
	return out, nil
}

// resolveBootstrapSeat 选席：仅 open 空席可绑；active 占用 → 409；left/stopped → V2 跳过绑（skipBind）。
func (s *Service) resolveBootstrapSeat(tx *sql.Tx, row *inviteRow, in RedeemJoinInput) (seatID, roleSlug string, skipBind bool, err error) {
	wantSlug := strings.TrimSpace(strings.ToLower(in.Role))
	if wantSlug == "" {
		wantSlug = strings.TrimSpace(strings.ToLower(row.SeatSlug))
	}
	if wantSlug == "" {
		wantSlug = strings.TrimSpace(strings.ToLower(row.RoleHint))
	}
	if sid := strings.TrimSpace(in.SeatID); sid != "" {
		var slug, status string
		var bound sql.NullString
		qerr := tx.QueryRow(`SELECT role_slug, COALESCE(NULLIF(status,''),'open'), bound_member_id
			FROM team_seats WHERE id = ? AND team_id = ? AND workspace_id = ?`,
			sid, row.TeamID, row.WorkspaceID).Scan(&slug, &status, &bound)
		if qerr == sql.ErrNoRows {
			return "", "", false, apperr.NotFound("席位不存在")
		}
		if qerr != nil {
			return "", "", false, qerr
		}
		st := normalizeSeatStatus(status)
		if st == models.SeatStatusActive && bound.Valid && bound.String != "" {
			return "", "", false, apperr.Conflict("该席位占用中，禁止顶替")
		}
		if st == models.SeatStatusStopped || st == models.SeatStatusLeft {
			// V2：不自动补位
			return sid, slug, true, nil
		}
		if bound.Valid && bound.String != "" {
			return "", "", false, apperr.Conflict("该席位已绑定其他成员")
		}
		return sid, slug, false, nil
	}

	rows, qerr := tx.Query(`SELECT id, role_slug, COALESCE(NULLIF(status,''),'open'), bound_member_id
		FROM team_seats WHERE team_id = ? AND workspace_id = ? ORDER BY role_slug`, row.TeamID, row.WorkspaceID)
	if qerr != nil {
		return "", "", false, qerr
	}
	defer rows.Close()

	type cand struct {
		id, slug, status string
		bound            sql.NullString
	}
	var all []cand
	for rows.Next() {
		var c cand
		if err := rows.Scan(&c.id, &c.slug, &c.status, &c.bound); err != nil {
			return "", "", false, err
		}
		c.status = normalizeSeatStatus(c.status)
		all = append(all, c)
	}
	if err := rows.Err(); err != nil {
		return "", "", false, err
	}

	// 若码指定 slug：先看该席状态
	if wantSlug != "" {
		for _, c := range all {
			if strings.EqualFold(c.slug, wantSlug) {
				bound := c.bound.Valid && c.bound.String != ""
				if c.status == models.SeatStatusActive && bound {
					return "", "", false, apperr.Conflict("该席位占用中，禁止顶替")
				}
				if c.status == models.SeatStatusStopped || c.status == models.SeatStatusLeft {
					return c.id, c.slug, true, nil // V2
				}
				if !bound && c.status == models.SeatStatusOpen {
					return c.id, c.slug, false, nil
				}
				if bound {
					return "", "", false, apperr.Conflict("该席位已绑定其他成员")
				}
			}
		}
	}

	// 任意 open 空席
	for _, c := range all {
		bound := c.bound.Valid && c.bound.String != ""
		if !bound && c.status == models.SeatStatusOpen {
			if wantSlug != "" && !strings.EqualFold(c.slug, wantSlug) {
				continue
			}
			return c.id, c.slug, false, nil
		}
	}
	// 无指定 slug 时再扫全部 open
	if wantSlug != "" {
		for _, c := range all {
			bound := c.bound.Valid && c.bound.String != ""
			if !bound && c.status == models.SeatStatusOpen {
				return c.id, c.slug, false, nil
			}
		}
	}

	// 无可绑空席：仍允许造人发钥（不绑）；left/stopped 不静默补
	return "", wantSlug, true, nil
}

func (s *Service) pickAndBindEmptySeat(row *inviteRow, memberID string, in RedeemJoinInput, force bool) (*models.TeamSeat, error) {
	roster, err := s.GetRoster(row.WorkspaceID, row.TeamID)
	if err != nil {
		return nil, err
	}
	// 已绑本成员 → 幂等
	for _, se := range roster {
		if se.BoundMemberID != nil && *se.BoundMemberID == memberID {
			return &se, nil
		}
	}
	wantSlug := strings.TrimSpace(strings.ToLower(in.Role))
	if wantSlug == "" {
		wantSlug = strings.TrimSpace(strings.ToLower(row.SeatSlug))
	}
	if wantSlug == "" {
		wantSlug = strings.TrimSpace(strings.ToLower(row.RoleHint))
	}
	if sid := strings.TrimSpace(in.SeatID); sid != "" {
		for _, se := range roster {
			if se.ID == sid {
				st := normalizeSeatStatus(se.Status)
				if st == models.SeatStatusStopped || st == models.SeatStatusLeft {
					return nil, apperr.Conflict("补位须走显式绑席，redeem 不自动补位")
				}
				return s.BindSeat(row.WorkspaceID, row.TeamID, se.ID, memberID, force)
			}
		}
		return nil, apperr.NotFound("席位不存在")
	}
	var openEmpty *models.TeamSeat
	for i := range roster {
		se := &roster[i]
		st := normalizeSeatStatus(se.Status)
		bound := se.BoundMemberID != nil && *se.BoundMemberID != ""
		if wantSlug != "" && strings.EqualFold(se.RoleSlug, wantSlug) {
			if st == models.SeatStatusActive && bound {
				return nil, apperr.Conflict("该席位占用中，禁止顶替")
			}
			if st == models.SeatStatusStopped || st == models.SeatStatusLeft {
				return nil, apperr.Conflict("补位须走显式绑席，redeem 不自动补位")
			}
			if !bound && st == models.SeatStatusOpen {
				return s.BindSeat(row.WorkspaceID, row.TeamID, se.ID, memberID, force)
			}
		}
		if !bound && st == models.SeatStatusOpen && openEmpty == nil {
			openEmpty = se
		}
	}
	if openEmpty != nil {
		return s.BindSeat(row.WorkspaceID, row.TeamID, openEmpty.ID, memberID, force)
	}
	return nil, apperr.Conflict("暂无空席可入")
}

func mintAPIKeyPlain() (plain, prefix, hash string, err error) {
	b := make([]byte, 24)
	if _, err = rand.Read(b); err != nil {
		return "", "", "", err
	}
	plain = "ymk_" + hex.EncodeToString(b)
	prefix = plain
	if len(prefix) > 12 {
		prefix = prefix[:12]
	}
	sum := sha256.Sum256([]byte(plain))
	hash = hex.EncodeToString(sum[:])
	return plain, prefix, hash, nil
}
