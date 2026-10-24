package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/middleware"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/apperr"
	"github.com/yumeng-tech/ai-collab-board/backend/internal/pkg/idgen"
)

type WebhookHandler struct {
	DB     *sql.DB
	Client *http.Client
}

func (h *WebhookHandler) client() *http.Client {
	if h.Client != nil {
		return h.Client
	}
	return &http.Client{Timeout: 10 * time.Second}
}

type webhookDTO struct {
	ID          string          `json:"id"`
	WorkspaceID string          `json:"workspace_id"`
	URL         string          `json:"url"`
	Events      json.RawMessage `json:"events"`
	Enabled     bool            `json:"enabled"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

func (h *WebhookHandler) List(w http.ResponseWriter, r *http.Request) {
	if middleware.MemberRole(r.Context()) != "admin" {
		middleware.HandleError(w, apperr.Forbidden("需要管理员权限"))
		return
	}
	wsID := middleware.WorkspaceID(r.Context())
	rows, err := h.DB.Query(`
		SELECT id, workspace_id, url, events, enabled, created_at, updated_at
		FROM webhooks WHERE workspace_id=? ORDER BY created_at DESC`, wsID)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	defer rows.Close()
	list := []webhookDTO{}
	for rows.Next() {
		var wh webhookDTO
		var en int
		var events string
		if err := rows.Scan(&wh.ID, &wh.WorkspaceID, &wh.URL, &events, &en, &wh.CreatedAt, &wh.UpdatedAt); err != nil {
			middleware.HandleError(w, err)
			return
		}
		wh.Enabled = en == 1
		wh.Events = json.RawMessage(events)
		list = append(list, wh)
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"webhooks": list})
}

func (h *WebhookHandler) Create(w http.ResponseWriter, r *http.Request) {
	if middleware.MemberRole(r.Context()) != "admin" {
		middleware.HandleError(w, apperr.Forbidden("需要管理员权限"))
		return
	}
	wsID := middleware.WorkspaceID(r.Context())
	var req struct {
		URL    string          `json:"url"`
		Secret string          `json:"secret"`
		Events json.RawMessage `json:"events"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	req.URL = strings.TrimSpace(req.URL)
	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
		middleware.HandleError(w, apperr.BadRequest("Webhook URL 须以 http:// 或 https:// 开头"))
		return
	}
	if len(req.Events) == 0 {
		req.Events = json.RawMessage(`["*"]`)
	}
	secret := strings.TrimSpace(req.Secret)
	if secret == "" {
		b := make([]byte, 16)
		_, _ = rand.Read(b)
		secret = hex.EncodeToString(b)
	}
	id := idgen.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := h.DB.Exec(`INSERT INTO webhooks (id, workspace_id, url, secret, events, enabled, created_at, updated_at)
		VALUES (?,?,?,?,?,1,?,?)`, id, wsID, req.URL, secret, string(req.Events), now, now)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusCreated, map[string]any{
		"webhook": webhookDTO{
			ID: id, WorkspaceID: wsID, URL: req.URL, Events: req.Events,
			Enabled: true, CreatedAt: now, UpdatedAt: now,
		},
		"secret": secret, // 仅创建时返回
	})
}

func (h *WebhookHandler) Patch(w http.ResponseWriter, r *http.Request) {
	if middleware.MemberRole(r.Context()) != "admin" {
		middleware.HandleError(w, apperr.Forbidden("需要管理员权限"))
		return
	}
	wsID := middleware.WorkspaceID(r.Context())
	wid := r.PathValue("webhookID")
	var url, events string
	var en int
	var created, updated string
	err := h.DB.QueryRow(`SELECT url, events, enabled, created_at, updated_at FROM webhooks WHERE id=? AND workspace_id=?`,
		wid, wsID).Scan(&url, &events, &en, &created, &updated)
	if err == sql.ErrNoRows {
		middleware.HandleError(w, apperr.NotFound("Webhook 不存在"))
		return
	}
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	var req struct {
		URL     *string          `json:"url"`
		Events  *json.RawMessage `json:"events"`
		Enabled *bool            `json:"enabled"`
		Secret  *string          `json:"secret"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.HandleError(w, apperr.BadRequest("请求体无效"))
		return
	}
	if req.URL != nil {
		url = strings.TrimSpace(*req.URL)
	}
	if req.Events != nil {
		events = string(*req.Events)
	}
	if req.Enabled != nil {
		if *req.Enabled {
			en = 1
		} else {
			en = 0
		}
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if req.Secret != nil && strings.TrimSpace(*req.Secret) != "" {
		_, err = h.DB.Exec(`UPDATE webhooks SET url=?, events=?, enabled=?, secret=?, updated_at=? WHERE id=?`,
			url, events, en, strings.TrimSpace(*req.Secret), now, wid)
	} else {
		_, err = h.DB.Exec(`UPDATE webhooks SET url=?, events=?, enabled=?, updated_at=? WHERE id=?`,
			url, events, en, now, wid)
	}
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]any{
		"webhook": webhookDTO{
			ID: wid, WorkspaceID: wsID, URL: url, Events: json.RawMessage(events),
			Enabled: en == 1, CreatedAt: created, UpdatedAt: now,
		},
	})
}

func (h *WebhookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if middleware.MemberRole(r.Context()) != "admin" {
		middleware.HandleError(w, apperr.Forbidden("需要管理员权限"))
		return
	}
	wsID := middleware.WorkspaceID(r.Context())
	wid := r.PathValue("webhookID")
	res, err := h.DB.Exec(`DELETE FROM webhooks WHERE id=? AND workspace_id=?`, wid, wsID)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		middleware.HandleError(w, apperr.NotFound("Webhook 不存在"))
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "Webhook 已删除"})
}

func (h *WebhookHandler) ListDeliveries(w http.ResponseWriter, r *http.Request) {
	if middleware.MemberRole(r.Context()) != "admin" {
		middleware.HandleError(w, apperr.Forbidden("需要管理员权限"))
		return
	}
	wsID := middleware.WorkspaceID(r.Context())
	wid := r.PathValue("webhookID")
	var check int
	_ = h.DB.QueryRow(`SELECT COUNT(1) FROM webhooks WHERE id=? AND workspace_id=?`, wid, wsID).Scan(&check)
	if check == 0 {
		middleware.HandleError(w, apperr.NotFound("Webhook 不存在"))
		return
	}
	rows, err := h.DB.Query(`
		SELECT id, webhook_id, event, payload, status_code, response_body, success, created_at
		FROM webhook_deliveries WHERE webhook_id=? ORDER BY created_at DESC LIMIT 50`, wid)
	if err != nil {
		middleware.HandleError(w, err)
		return
	}
	defer rows.Close()
	type del struct {
		ID           string  `json:"id"`
		WebhookID    string  `json:"webhook_id"`
		Event        string  `json:"event"`
		Payload      string  `json:"payload"`
		StatusCode   *int    `json:"status_code"`
		ResponseBody *string `json:"response_body"`
		Success      bool    `json:"success"`
		CreatedAt    string  `json:"created_at"`
	}
	list := []del{}
	for rows.Next() {
		var d del
		var sc sql.NullInt64
		var rb sql.NullString
		var succ int
		if err := rows.Scan(&d.ID, &d.WebhookID, &d.Event, &d.Payload, &sc, &rb, &succ, &d.CreatedAt); err != nil {
			middleware.HandleError(w, err)
			return
		}
		if sc.Valid {
			v := int(sc.Int64)
			d.StatusCode = &v
		}
		if rb.Valid {
			d.ResponseBody = &rb.String
		}
		d.Success = succ == 1
		list = append(list, d)
	}
	middleware.JSON(w, http.StatusOK, map[string]any{"deliveries": list})
}

// SignPayload HMAC-SHA256 hex digest for X-Signature-256 header.
func SignPayload(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// DispatchEvent fans out an event to enabled webhooks (async).
func (h *WebhookHandler) DispatchEvent(workspaceID, event string, payload any) {
	go h.dispatchSync(workspaceID, event, payload)
}

func (h *WebhookHandler) dispatchSync(workspaceID, event string, payload any) {
	body, err := json.Marshal(map[string]any{
		"event":      event,
		"workspace_id": workspaceID,
		"occurred_at": time.Now().UTC().Format(time.RFC3339Nano),
		"data":       payload,
	})
	if err != nil {
		return
	}
	rows, err := h.DB.Query(`SELECT id, url, secret, events FROM webhooks WHERE workspace_id=? AND enabled=1`, workspaceID)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id, url, secret, eventsJSON string
		if err := rows.Scan(&id, &url, &secret, &eventsJSON); err != nil {
			continue
		}
		if !eventMatch(eventsJSON, event) {
			continue
		}
		h.deliverOne(id, url, secret, event, body)
	}
}

func eventMatch(eventsJSON, event string) bool {
	var events []string
	if err := json.Unmarshal([]byte(eventsJSON), &events); err != nil {
		return true
	}
	for _, e := range events {
		if e == "*" || e == event {
			return true
		}
	}
	return false
}

func (h *WebhookHandler) deliverOne(webhookID, url, secret, event string, body []byte) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		h.logDelivery(webhookID, event, string(body), nil, "", false)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature-256", SignPayload(secret, body))
	req.Header.Set("X-Webhook-Event", event)
	resp, err := h.client().Do(req)
	if err != nil {
		msg := err.Error()
		h.logDelivery(webhookID, event, string(body), nil, msg, false)
		return
	}
	defer resp.Body.Close()
	rb, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	code := resp.StatusCode
	ok := code >= 200 && code < 300
	h.logDelivery(webhookID, event, string(body), &code, string(rb), ok)
}

func (h *WebhookHandler) logDelivery(webhookID, event, payload string, status *int, respBody string, success bool) {
	succ := 0
	if success {
		succ = 1
	}
	var sc any
	if status != nil {
		sc = *status
	}
	_, _ = h.DB.Exec(`INSERT INTO webhook_deliveries
		(id, webhook_id, event, payload, status_code, response_body, success)
		VALUES (?,?,?,?,?,?,?)`,
		idgen.New(), webhookID, event, payload, sc, respBody, succ)
}
