package botclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// identifierRe matches key_prefix-number (e.g. YM-12, SMK-1).
var identifierRe = regexp.MustCompile(`(?i)^[A-Z][A-Z0-9]*-\d+$`)

// uuidishRe: 32 hex (with optional dashes) — REST issue ids are 32 hex.
var uuidishRe = regexp.MustCompile(`(?i)^[0-9a-f]{32}$|^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func LooksLikeIdentifier(s string) bool {
	return looksLikeIdentifier(s)
}

func looksLikeIdentifier(s string) bool {
	return identifierRe.MatchString(strings.TrimSpace(s))
}

func looksLikeUUID(s string) bool {
	t := strings.ReplaceAll(strings.TrimSpace(s), "-", "")
	return uuidishRe.MatchString(strings.TrimSpace(s)) || (len(t) == 32 && uuidishRe.MatchString(t))
}

// ResolveIssueRef 将 YM-12 或 UUID 解析为 REST id；失败返回一行中文可用错误。
func (c *Client) ResolveIssueRef(ref string) (id string, err error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", fmt.Errorf("需要 Issue 标识（如 YM-12）或 id")
	}
	if !looksLikeIdentifier(ref) {
		// UUID or opaque id — pass through
		return ref, nil
	}
	c.ensureCaches()
	c.cacheMu.Lock()
	if id, ok := c.identToID[strings.ToUpper(ref)]; ok {
		c.cacheMu.Unlock()
		return id, nil
	}
	c.cacheMu.Unlock()

	q := url.Values{}
	q.Set("q", ref)
	raw, err := c.Do(http.MethodGet, c.ws("/issues"), q, nil, false)
	if err != nil {
		return "", err
	}
	var wrap struct {
		Issues []struct {
			ID         string `json:"id"`
			Identifier string `json:"identifier"`
		} `json:"issues"`
	}
	if err := json.Unmarshal(raw, &wrap); err != nil {
		return "", fmt.Errorf("解析 Issue 列表失败: %w", err)
	}
	want := strings.ToUpper(ref)
	var matchID string
	for _, iss := range wrap.Issues {
		if strings.EqualFold(iss.Identifier, want) {
			matchID = iss.ID
			c.cacheMu.Lock()
			c.identToID[strings.ToUpper(iss.Identifier)] = iss.ID
			c.idToIdent[iss.ID] = iss.Identifier
			c.cacheMu.Unlock()
			break
		}
	}
	if matchID == "" {
		return "", &APIError{Status: 404, Code: "not_found", Message: fmt.Sprintf("找不到 Issue %s", ref)}
	}
	return matchID, nil
}

func (c *Client) ensureCaches() {
	c.cacheOnce.Do(func() {
		c.identToID = map[string]string{}
		c.idToIdent = map[string]string{}
		c.memberNames = map[string]string{}
	})
}

// MemberNames 懒加载工作区成员 id→display_name（通道展示用，不改 REST）。
func (c *Client) MemberNames() (map[string]string, error) {
	c.ensureCaches()
	c.cacheMu.Lock()
	if c.membersLoaded {
		out := copyStrMap(c.memberNames)
		c.cacheMu.Unlock()
		return out, nil
	}
	c.cacheMu.Unlock()

	if c.HTTP == nil {
		c.cacheMu.Lock()
		c.membersLoaded = true
		out := copyStrMap(c.memberNames)
		c.cacheMu.Unlock()
		return out, nil
	}
	raw, err := c.Do(http.MethodGet, c.ws("/members"), nil, nil, false)
	if err != nil {
		// 展示层降级：无成员名时仍出瘦字段
		return map[string]string{}, nil
	}
	var wrap struct {
		Members []struct {
			ID          string `json:"id"`
			DisplayName string `json:"display_name"`
		} `json:"members"`
	}
	_ = json.Unmarshal(raw, &wrap)
	c.cacheMu.Lock()
	for _, m := range wrap.Members {
		c.memberNames[m.ID] = m.DisplayName
	}
	c.membersLoaded = true
	out := copyStrMap(c.memberNames)
	c.cacheMu.Unlock()
	return out, nil
}

func (c *Client) memberName(id string) string {
	if id == "" {
		return ""
	}
	names, err := c.MemberNames()
	if err != nil {
		return ""
	}
	return names[id]
}

func (c *Client) rememberIssue(id, identifier string) {
	if id == "" || identifier == "" {
		return
	}
	c.ensureCaches()
	c.cacheMu.Lock()
	c.identToID[strings.ToUpper(identifier)] = id
	c.idToIdent[id] = identifier
	c.cacheMu.Unlock()
}

func (c *Client) cachedIdent(id string) string {
	c.ensureCaches()
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	return c.idToIdent[id]
}

func copyStrMap(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}


// CachedIdent 返回已缓存的 identifier（可能为空）。
func (c *Client) CachedIdent(id string) string {
	return c.cachedIdent(id)
}
