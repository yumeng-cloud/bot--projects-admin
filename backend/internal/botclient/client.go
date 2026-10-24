package botclient

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Env: YMBOT_API_URL, YMBOT_API_KEY, YMBOT_WORKSPACE_ID, optional YMBOT_MEMBER_ID,
// optional YMBOT_IDEMPOTENCY_KEY（写操作未显式传 Key 时使用）.
type Config struct {
	BaseURL     string
	APIKey      string
	WorkspaceID string
	MemberID    string
}

func LoadConfig() Config {
	base := envOr("YMBOT_API_URL", "http://127.0.0.1:8080")
	return Config{
		BaseURL:     strings.TrimRight(base, "/"),
		APIKey:      os.Getenv("YMBOT_API_KEY"),
		WorkspaceID: os.Getenv("YMBOT_WORKSPACE_ID"),
		MemberID:    os.Getenv("YMBOT_MEMBER_ID"),
	}
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

type APIError struct {
	Status  int
	Code    string
	Message string
}

// Error 返回一行中文包装，供 CLI stderr / MCP isError content 直接使用。
func (e *APIError) Error() string {
	msg := strings.TrimSpace(e.Message)
	switch e.Status {
	case 401:
		if msg == "" {
			msg = "API Key 无效或已吊销"
		} else if !containsAny(msg, "已吊销", "无效") {
			msg = msg + "（无效或已吊销）"
		}
		return fmt.Sprintf("认证失败：%s（HTTP 401）", msg)
	case 403:
		if msg == "" {
			msg = "无权限执行此操作"
		}
		return fmt.Sprintf("无权限：%s（HTTP 403）", msg)
	case 404:
		if msg == "" {
			msg = "资源不存在"
		}
		return fmt.Sprintf("未找到：%s（HTTP 404）", msg)
	case 409:
		if msg == "" {
			msg = "资源冲突"
		}
		return fmt.Sprintf("冲突：%s（HTTP 409）", msg)
	default:
		if msg == "" {
			msg = e.Code
		}
		if msg == "" {
			msg = "请求失败"
		}
		if e.Status > 0 {
			return fmt.Sprintf("%s（HTTP %d）", msg, e.Status)
		}
		return msg
	}
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func (e *APIError) ExitCode() int {
	switch e.Status {
	case 401:
		return 10
	case 403:
		return 11
	case 404:
		return 12
	case 409:
		return 13
	default:
		if e.Status >= 400 && e.Status < 500 {
			return 20
		}
		return 1
	}
}

type Client struct {
	Cfg  Config
	HTTP *http.Client
	// IdempotencyKey 可选：写操作优先用此值，其次 YMBOT_IDEMPOTENCY_KEY，否则自动生成 UUID。
	IdempotencyKey string

	// 通道 DX 缓存（identifier↔id、成员显示名）；不落库、不改 REST。
	cacheOnce     sync.Once
	cacheMu       sync.Mutex
	identToID     map[string]string
	idToIdent     map[string]string
	memberNames   map[string]string
	membersLoaded bool
}

func New(cfg Config) *Client {
	return &Client{Cfg: cfg, HTTP: &http.Client{Timeout: 30 * time.Second}}
}

func (c *Client) Require() error {
	if c.Cfg.APIKey == "" {
		return fmt.Errorf("缺少 YMBOT_API_KEY（仅接受 ymk_… API Key，不主推密码登录）")
	}
	if c.Cfg.WorkspaceID == "" {
		return fmt.Errorf("缺少 YMBOT_WORKSPACE_ID")
	}
	return nil
}

func NewIDKey() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// ResolveIDKey 优先级：显式参数 > 环境变量 YMBOT_IDEMPOTENCY_KEY > 自动 UUID。
func ResolveIDKey(explicit string) string {
	if v := strings.TrimSpace(explicit); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("YMBOT_IDEMPOTENCY_KEY")); v != "" {
		return v
	}
	return NewIDKey()
}

func (c *Client) Do(method, path string, query url.Values, body any, idempotent bool) (json.RawMessage, error) {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		rdr = bytes.NewReader(b)
	}
	u := c.Cfg.BaseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequest(method, u, rdr)
	if err != nil {
		return nil, err
	}
	if k := strings.TrimSpace(c.Cfg.APIKey); k != "" {
		req.Header.Set("Authorization", "Bearer "+k)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if idempotent {
		req.Header.Set("Idempotency-Key", ResolveIDKey(c.IdempotencyKey))
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		if len(b) == 0 {
			return json.RawMessage(`{}`), nil
		}
		return json.RawMessage(b), nil
	}
	var ae struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(b, &ae)
	msg := ae.Message
	if msg == "" {
		msg = strings.TrimSpace(string(b))
	}
	return nil, &APIError{Status: res.StatusCode, Code: ae.Code, Message: msg}
}

func (c *Client) ws(suffix string) string {
	return "/api/v1/workspaces/" + c.Cfg.WorkspaceID + suffix
}

func GuardBotDone(status string) error {
	if strings.EqualFold(status, "done") {
		return &APIError{Status: 403, Code: "forbidden", Message: "Bot 不能将 Issue 设为 done，需人类确认（可改为 in_review）"}
	}
	return nil
}

// GuardBotOrphanCreate 通道产品闸：Bot 禁止创建无 parent_id 的游离 Issue（A7）。
// 须在已有父节点下拆叶子，或由人类编排。
func GuardBotOrphanCreate(parentID string) error {
	if strings.TrimSpace(parentID) == "" {
		return &APIError{
			Status:  403,
			Code:    "forbidden",
			Message: "禁止创建无父游离 Issue，请在已有父节点下拆叶子或由人类编排",
		}
	}
	return nil
}

// parentIDFromBody 从 create body 取 parent_id（string 或可 Sprint）。
func parentIDFromBody(body map[string]any) string {
	if body == nil {
		return ""
	}
	v, ok := body["parent_id"]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		s := strings.TrimSpace(fmt.Sprint(t))
		if s == "" || s == "<nil>" {
			return ""
		}
		return s
	}
}

// WhoAmI uses GET …/bots/me；若无此路由则回退 roster。
func (c *Client) WhoAmI() (json.RawMessage, error) {
	raw, err := c.Do(http.MethodGet, c.ws("/bots/me"), nil, nil, false)
	if err == nil {
		return raw, nil
	}
	if ae, ok := err.(*APIError); ok && ae.Status == 404 {
		raw2, err2 := c.Do(http.MethodGet, c.ws("/bots"), nil, nil, false)
		if err2 != nil {
			return nil, err2
		}
		var wrap struct {
			Bots []struct {
				MemberID    string `json:"member_id"`
				DisplayName string `json:"display_name"`
			} `json:"bots"`
		}
		_ = json.Unmarshal(raw2, &wrap)
		mid := c.Cfg.MemberID
		display := ""
		if mid == "" && len(wrap.Bots) == 1 {
			mid = wrap.Bots[0].MemberID
			display = wrap.Bots[0].DisplayName
			c.Cfg.MemberID = mid
		}
		out := map[string]any{
			"workspace_id": c.Cfg.WorkspaceID,
			"member_id":    mid,
			"actor_kind":   "bot",
			"display_name": display,
			"bots":         wrap.Bots,
			"note":         "bots/me 不可用；已用 roster 推断（多 Bot 时请设 YMBOT_MEMBER_ID）",
		}
		b, _ := json.Marshal(out)
		return b, nil
	}
	return nil, err
}

func (c *Client) ResolveMember() (string, error) {
	if c.Cfg.MemberID != "" {
		return c.Cfg.MemberID, nil
	}
	raw, err := c.Do(http.MethodGet, c.ws("/bots/me"), nil, nil, false)
	if err == nil {
		var me struct {
			MemberID string `json:"member_id"`
		}
		_ = json.Unmarshal(raw, &me)
		if me.MemberID != "" {
			c.Cfg.MemberID = me.MemberID
			return me.MemberID, nil
		}
	}
	// fallback: single bot in roster
	raw2, err2 := c.Do(http.MethodGet, c.ws("/bots"), nil, nil, false)
	if err2 != nil {
		return "", fmt.Errorf("无法解析 member：%w；请设 YMBOT_MEMBER_ID", err)
	}
	var wrap struct {
		Bots []struct {
			MemberID string `json:"member_id"`
		} `json:"bots"`
	}
	_ = json.Unmarshal(raw2, &wrap)
	if len(wrap.Bots) == 1 && wrap.Bots[0].MemberID != "" {
		c.Cfg.MemberID = wrap.Bots[0].MemberID
		return c.Cfg.MemberID, nil
	}
	return "", fmt.Errorf("无法解析 member（bots=%d）；请设 YMBOT_MEMBER_ID", len(wrap.Bots))
}

func (c *Client) IssueList(q url.Values) (json.RawMessage, error) {
	if q == nil {
		q = url.Values{}
	}
	if q.Get("assignee_id") == "me" {
		mid, err := c.ResolveMember()
		if err != nil {
			return nil, err
		}
		q.Set("assignee_id", mid)
	}
	return c.Do(http.MethodGet, c.ws("/issues"), q, nil, false)
}

func (c *Client) IssueGet(id string) (json.RawMessage, error) {
	return c.Do(http.MethodGet, c.ws("/issues/"+id), nil, nil, false)
}

func (c *Client) IssueCreate(body map[string]any) (json.RawMessage, error) {
	if err := GuardBotOrphanCreate(parentIDFromBody(body)); err != nil {
		return nil, err
	}
	if s, ok := body["status"].(string); ok {
		if err := GuardBotDone(s); err != nil {
			return nil, err
		}
	}
	return c.Do(http.MethodPost, c.ws("/issues"), nil, body, true)
}

// IssueClaim 领取叶子 Issue：assignee_id=自己 member_id + 乐观锁 version。
// version < 0 时先 GET 取当前 version。服务端非叶子 409（含「叶子」文案）由 APIError 原样中文包装透出。
// 与角色 bot-roles/.../claim / board_role_claim（岗位）命名分离。
func (c *Client) IssueClaim(id string, version int) (json.RawMessage, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("需要 issue id")
	}
	mid, err := c.ResolveMember()
	if err != nil {
		return nil, err
	}
	ver := version
	if ver < 0 {
		raw, err := c.IssueGet(id)
		if err != nil {
			return nil, err
		}
		var wrap struct {
			Issue   *struct{ Version int `json:"version"` } `json:"issue"`
			Version int `json:"version"`
		}
		if err := json.Unmarshal(raw, &wrap); err != nil {
			return nil, fmt.Errorf("解析 Issue version 失败: %w", err)
		}
		if wrap.Issue != nil {
			ver = wrap.Issue.Version
		} else {
			ver = wrap.Version
		}
	}
	body := map[string]any{
		"version":     ver,
		"assignee_id": mid,
	}
	return c.IssueUpdate(id, body)
}

func (c *Client) IssueUpdate(id string, body map[string]any) (json.RawMessage, error) {
	if s, ok := body["status"].(string); ok {
		if err := GuardBotDone(s); err != nil {
			return nil, err
		}
	}
	return c.Do(http.MethodPatch, c.ws("/issues/"+id), nil, body, true)
}

func (c *Client) CommentList(issueID string) (json.RawMessage, error) {
	return c.Do(http.MethodGet, c.ws("/issues/"+issueID+"/comments"), nil, nil, false)
}

func (c *Client) CommentCreate(issueID, body string) (json.RawMessage, error) {
	return c.Do(http.MethodPost, c.ws("/issues/"+issueID+"/comments"), nil, map[string]string{"body": body}, true)
}

func (c *Client) MemoryList(q, mode string, limit int) (json.RawMessage, error) {
	mid, err := c.ResolveMember()
	if err != nil {
		return nil, err
	}
	qv := url.Values{}
	if q != "" {
		qv.Set("q", q)
	}
	if mode != "" {
		qv.Set("mode", mode)
	}
	if limit > 0 {
		qv.Set("limit", strconv.Itoa(limit))
	}
	return c.Do(http.MethodGet, c.ws("/bots/"+mid+"/memories"), qv, nil, false)
}

func (c *Client) MemorySet(key, value, tags string) (json.RawMessage, error) {
	mid, err := c.ResolveMember()
	if err != nil {
		return nil, err
	}
	body := map[string]string{"key_name": key, "value_text": value}
	if tags != "" {
		body["tags"] = tags
	}
	return c.Do(http.MethodPost, c.ws("/bots/"+mid+"/memories"), nil, body, true)
}

func (c *Client) HandoffGet() (json.RawMessage, error) {
	mid, err := c.ResolveMember()
	if err != nil {
		return nil, err
	}
	return c.Do(http.MethodGet, c.ws("/bots/"+mid+"/handoff"), nil, nil, false)
}

func (c *Client) HandoffImport(body any) (json.RawMessage, error) {
	mid, err := c.ResolveMember()
	if err != nil {
		return nil, err
	}
	return c.Do(http.MethodPost, c.ws("/bots/"+mid+"/handoff"), nil, body, true)
}
