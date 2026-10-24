package handlers

import (
	"github.com/yumeng-tech/ai-collab-board/backend/internal/db"

	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/auth"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/config"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/ratelimit"
)

type AuthHandler struct {
	DB      *sql.DB
	Cfg     config.Config
	Limiter *ratelimit.Limiter
}

type registerReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userDTO struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Name = strings.TrimSpace(req.Name)
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		middleware.HandleError(w, apperr.BadRequest("邮箱格式不正确"))
		return
	}
	if len(req.Password) < 8 {
		middleware.HandleError(w, apperr.BadRequest("密码至少 8 位"))
		return
	}
	if req.Name == "" {
		req.Name = strings.Split(req.Email, "@")[0]
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	id := idgen.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = h.DB.Exec(`INSERT INTO users (id, email, password_hash, name, created_at, updated_at) VALUES (?,?,?,?,?,?)`,
		id, req.Email, hash, req.Name, now, now)
	if err != nil {
		if db.IsUniqueViolation(err) {
			middleware.HandleError(w, apperr.Conflict("该邮箱已注册"))
			return
		}
		middleware.HandleError(w, err)
		return
	}
	token, err := auth.IssueToken(h.Cfg.JWTSecret, id, req.Email, h.Cfg.JWTTTL)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	setSessionCookie(w, token, h.Cfg)
	middleware.JSON(w, http.StatusCreated, map[string]any{
		"token": token,
		"user":  userDTO{ID: id, Email: req.Email, Name: req.Name, CreatedAt: now},
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	ip := r.RemoteAddr
	key := ip + "|" + req.Email
	if h.Limiter != nil && !h.Limiter.Allow(key) {
		middleware.HandleError(w, apperr.TooMany("登录尝试过于频繁，请稍后再试"))
		return
	}
	var id, hash, name, created string
	err := h.DB.QueryRow(`SELECT id, password_hash, name, created_at FROM users WHERE email = ?`, req.Email).
		Scan(&id, &hash, &name, &created)
	if err == sql.ErrNoRows {
		middleware.HandleError(w, apperr.Unauthorized("邮箱或密码错误"))
		return
	}
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	ok, err := auth.VerifyPassword(hash, req.Password)
	if err != nil || !ok {
		middleware.HandleError(w, apperr.Unauthorized("邮箱或密码错误"))
		return
	}
	token, err := auth.IssueToken(h.Cfg.JWTSecret, id, req.Email, h.Cfg.JWTTTL)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	setSessionCookie(w, token, h.Cfg)
	middleware.JSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user":  userDTO{ID: id, Email: req.Email, Name: name, CreatedAt: created},
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "已退出登录"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserID(r.Context())
	var email, name, created string
	err := h.DB.QueryRow(`SELECT email, name, created_at FROM users WHERE id = ?`, uid).Scan(&email, &name, &created)
	if err == sql.ErrNoRows {
		middleware.HandleError(w, apperr.Unauthorized("用户不存在"))
		return
	}
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{
		"user": userDTO{ID: uid, Email: email, Name: name, CreatedAt: created},
	})
}

func setSessionCookie(w http.ResponseWriter, token string, cfg config.Config) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(cfg.JWTTTL.Seconds()),
		Secure:   cfg.AppEnv == "production",
	})
}
