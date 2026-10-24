package middleware

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/auth"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
)

type ctxKey string

const UserIDKey ctxKey = "userID"
const EmailKey ctxKey = "email"
const MemberIDKey ctxKey = "memberID"
const WorkspaceIDKey ctxKey = "workspaceID"
const ActorKindKey ctxKey = "actorKind" // human | bot
const MemberRoleKey ctxKey = "memberRole"

type TokenParser interface {
	Parse(token string) (*auth.Claims, error)
}

type JWTParser struct {
	Secret string
}

func (p JWTParser) Parse(token string) (*auth.Claims, error) {
	return auth.ParseToken(p.Secret, token)
}

// AuthJWTOrAPIKey authenticates via JWT (human) or API Key (bot, prefix ymk_).
func AuthJWTOrAPIKey(parser TokenParser, db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearer(r)
			if token == "" {
				if c, err := r.Cookie("session"); err == nil {
					token = c.Value
				}
			}
			if token == "" {
				writeErr(w, apperr.Unauthorized("请先登录"))
				return
			}
			if strings.HasPrefix(token, "ymk_") {
				ctx, err := authenticateAPIKey(r.Context(), db, token)
				if err != nil {
					if ae, ok := err.(*apperr.AppError); ok {
						writeErr(w, ae)
					} else {
						writeErr(w, apperr.Unauthorized("鉴权失败"))
					}
					return
				}
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			claims, err := parser.Parse(token)
			if err != nil {
				writeErr(w, apperr.Unauthorized("登录已失效，请重新登录"))
				return
			}
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, EmailKey, claims.Email)
			ctx = context.WithValue(ctx, ActorKindKey, "human")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthJWT is JWT-only (for routes that must not accept API keys).
func AuthJWT(parser TokenParser) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearer(r)
			if token == "" {
				if c, err := r.Cookie("session"); err == nil {
					token = c.Value
				}
			}
			if token == "" {
				writeErr(w, apperr.Unauthorized("请先登录"))
				return
			}
			claims, err := parser.Parse(token)
			if err != nil {
				writeErr(w, apperr.Unauthorized("登录已失效，请重新登录"))
				return
			}
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, EmailKey, claims.Email)
			ctx = context.WithValue(ctx, ActorKindKey, "human")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func authenticateAPIKey(ctx context.Context, db *sql.DB, rawKey string) (context.Context, error) {
	sum := sha256.Sum256([]byte(rawKey))
	hash := hex.EncodeToString(sum[:])
	var keyID, wsID, memberID string
	var revoked sql.NullString
	err := db.QueryRow(`
		SELECT id, workspace_id, member_id, revoked_at
		FROM api_keys WHERE key_hash = ?`, hash).Scan(&keyID, &wsID, &memberID, &revoked)
	if err == sql.ErrNoRows {
		return ctx, apperr.Unauthorized("无效的 API Key")
	}
	if err != nil {
		return ctx, apperr.Internal("鉴权失败")
	}
	if revoked.Valid {
		return ctx, apperr.Unauthorized("API Key 已吊销")
	}
	var role string
	var userID sql.NullString
	err = db.QueryRow(`SELECT role, user_id FROM members WHERE id = ?`, memberID).Scan(&role, &userID)
	if err != nil {
		return ctx, apperr.Unauthorized("API Key 绑定成员无效")
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, _ = db.Exec(`UPDATE api_keys SET last_used_at = ? WHERE id = ?`, now, keyID)

	ctx = context.WithValue(ctx, MemberIDKey, memberID)
	ctx = context.WithValue(ctx, WorkspaceIDKey, wsID)
	ctx = context.WithValue(ctx, ActorKindKey, "bot")
	ctx = context.WithValue(ctx, MemberRoleKey, role)
	if userID.Valid {
		ctx = context.WithValue(ctx, UserIDKey, userID.String)
	}
	return ctx, nil
}

func extractBearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(h), "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return ""
}

// BearerToken 读 Authorization: Bearer …（无则空串）。
func BearerToken(r *http.Request) string {
	return extractBearer(r)
}

// AuthenticateAPIKey 校验 ymk_ 并注入 bot 上下文（供可选鉴权路由）。
func AuthenticateAPIKey(ctx context.Context, db *sql.DB, rawKey string) (context.Context, error) {
	return authenticateAPIKey(ctx, db, rawKey)
}

func writeErr(w http.ResponseWriter, e *apperr.AppError) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(e.StatusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{"code": e.Code, "message": e.Message})
}

func UserID(ctx context.Context) string {
	v, _ := ctx.Value(UserIDKey).(string)
	return v
}

func MemberID(ctx context.Context) string {
	v, _ := ctx.Value(MemberIDKey).(string)
	return v
}

func ActorKind(ctx context.Context) string {
	v, _ := ctx.Value(ActorKindKey).(string)
	return v
}

func WorkspaceID(ctx context.Context) string {
	v, _ := ctx.Value(WorkspaceIDKey).(string)
	return v
}

func MemberRole(ctx context.Context) string {
	v, _ := ctx.Value(MemberRoleKey).(string)
	return v
}
