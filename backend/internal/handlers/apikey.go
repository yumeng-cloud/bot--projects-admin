package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"
	"net/http"
	"strings"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
)

type APIKeyHandler struct {
	DB *sql.DB
}

type apiKeyDTO struct {
	ID          string  `json:"id"`
	WorkspaceID string  `json:"workspace_id"`
	MemberID    string  `json:"member_id"`
	Name        string  `json:"name"`
	KeyPrefix   string  `json:"key_prefix"`
	LastUsedAt  *string `json:"last_used_at"`
	CreatedAt   string  `json:"created_at"`
	RevokedAt   *string `json:"revoked_at"`
}

func (h *APIKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	if middleware.MemberRole(r.Context()) != "admin" {
		middleware.HandleError(w, apperr.Forbidden("需要管理员权限"))
		return
	}
	wsID := middleware.WorkspaceID(r.Context())
	rows, err := h.DB.Query(`
		SELECT id, workspace_id, member_id, name, key_prefix, last_used_at, created_at, revoked_at
		FROM api_keys WHERE workspace_id = ? ORDER BY created_at DESC`, wsID)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	defer rows.Close()
	list := []apiKeyDTO{}
	for rows.Next() {
		var k apiKeyDTO
		var last, rev sql.NullString
		if err := rows.Scan(&k.ID, &k.WorkspaceID, &k.MemberID, &k.Name, &k.KeyPrefix, &last, &k.CreatedAt, &rev); err != nil {
			middleware.HandleError(w, err)
			return
		}
		if last.Valid {
			k.LastUsedAt = &last.String
		}
		if rev.Valid {
			k.RevokedAt = &rev.String
		}
		list = append(list, k)
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"api_keys": list})
}

func (h *APIKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	if middleware.MemberRole(r.Context()) != "admin" {
		middleware.HandleError(w, apperr.Forbidden("需要管理员权限"))
		return
	}
	wsID := middleware.WorkspaceID(r.Context())
	var req struct {
		Name     string `json:"name"`
		MemberID string `json:"member_id"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		middleware.HandleError(w, apperr.BadRequest("名称不能为空"))
		return
	}
	memberID := strings.TrimSpace(req.MemberID)
	if memberID == "" {
		// default: first bot member in workspace
		err := h.DB.QueryRow(`SELECT id FROM members WHERE workspace_id=? AND is_bot=1 ORDER BY created_at LIMIT 1`, wsID).Scan(&memberID)
		if err == sql.ErrNoRows {
			middleware.HandleError(w, apperr.BadRequest("请指定 member_id（Bot 成员）"))
			return
		}
		if err != nil {
			middleware.HandleError(w, err)
			return
		}
	} else {
		var exists int
		err := h.DB.QueryRow(`SELECT COUNT(1) FROM members WHERE id=? AND workspace_id=?`, memberID, wsID).Scan(&exists)
		if err != nil || exists == 0 {
			middleware.HandleError(w, apperr.BadRequest("成员不存在或不属于该工作空间"))
			return
		}
	}

	raw := "ymk_" + randomHex(24)
	sum := sha256.Sum256([]byte(raw))
	hash := hex.EncodeToString(sum[:])
	prefix := raw[:12]
	id := idgen.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := h.DB.Exec(`INSERT INTO api_keys (id, workspace_id, member_id, name, key_prefix, key_hash, created_at)
		VALUES (?,?,?,?,?,?,?)`, id, wsID, memberID, req.Name, prefix, hash, now)
	if err != nil {
		if db.IsUniqueViolation(err) {
			middleware.HandleError(w, apperr.Conflict("API Key 哈希冲突，请重试"))
			return
		}
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusCreated, map[string]any{
		"api_key": apiKeyDTO{
			ID: id, WorkspaceID: wsID, MemberID: memberID, Name: req.Name,
			KeyPrefix: prefix, CreatedAt: now,
		},
		"token": raw, // 明文仅此一次
	})
}

func (h *APIKeyHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	keyID := r.PathValue("keyID")
	if keyID == "" {
		keyID = r.PathValue("id")
	}
	var wsID string
	var revoked sql.NullString
	err := h.DB.QueryRow(`SELECT workspace_id, revoked_at FROM api_keys WHERE id=?`, keyID).Scan(&wsID, &revoked)
	if err == sql.ErrNoRows {
		middleware.HandleError(w, apperr.NotFound("API Key 不存在"))
		return
	}
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	if !h.ensureAdminOf(w, r, wsID) {
		return
	}
	if revoked.Valid {
		middleware.JSON(w, http.StatusOK, map[string]string{"message": "API Key 已吊销"})
		return
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = h.DB.Exec(`UPDATE api_keys SET revoked_at=? WHERE id=?`, now, keyID)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "API Key 已吊销"})
}

func (h *APIKeyHandler) ensureAdminOf(w http.ResponseWriter, r *http.Request, wsID string) bool {
	if middleware.ActorKind(r.Context()) == "bot" {
		middleware.HandleError(w, apperr.Forbidden("Bot 不能管理 API Key"))
		return false
	}
	if middleware.WorkspaceID(r.Context()) == wsID && middleware.MemberRole(r.Context()) == "admin" {
		return true
	}
	uid := middleware.UserID(r.Context())
	var role string
	err := h.DB.QueryRow(`SELECT role FROM members WHERE workspace_id=? AND user_id=?`, wsID, uid).Scan(&role)
	if err == sql.ErrNoRows {
		middleware.HandleError(w, apperr.Forbidden("您不是该工作空间的成员"))
		return false
	}
	if err != nil {
		middleware.HandleError(w, err)
		return false
	}
	if role != "admin" {
		middleware.HandleError(w, apperr.Forbidden("需要管理员权限"))
		return false
	}
	return true
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
