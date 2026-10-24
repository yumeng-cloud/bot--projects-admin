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

type CreateInviteInput struct {
	AllowExpandSeat bool
	SeatSlug        string
	RoleHint        string
	MaxUses         *int
	ExpiresAt       string // RFC3339 optional
	CreatedBy       string
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
func (s *Service) CreateInvite(workspaceID, teamID string, in CreateInviteInput) (*models.TeamInvite, error) {
	if _, err := s.GetTeam(workspaceID, teamID); err != nil {
		return nil, err
	}
	plain, prefix, hash, err := newInviteCode()
	if err != nil {
		return nil, err
	}
	now := nowUTC()
	id := idgen.New()
	var seatSlug, roleHint, expires, createdBy any
	if v := strings.TrimSpace(strings.ToLower(in.SeatSlug)); v != "" {
		seatSlug = v
	}
	if v := strings.TrimSpace(in.RoleHint); v != "" {
		roleHint = v
	} else if seatSlug != nil {
		roleHint = seatSlug
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
		id, workspaceID, teamID, hash, prefix, allow, seatSlug, roleHint,
		maxUses, expires, createdBy, now, now)
	if err != nil {
		return nil, err
	}
	inv, err := s.getInviteByID(workspaceID, id)
	if err != nil {
		return nil, err
	}
	inv.Code = plain
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

// RevokeInvite 吊销入场码（Admin）。
func (s *Service) RevokeInvite(workspaceID, teamID, inviteID string) error {
	now := nowUTC()
	res, err := s.DB.Exec(`UPDATE team_invites SET revoked_at = ?, updated_at = ?
		WHERE id = ? AND workspace_id = ? AND team_id = ? AND revoked_at IS NULL`,
		now, now, inviteID, workspaceID, teamID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperr.NotFound("入场码不存在或已吊销")
	}
	return nil
}
