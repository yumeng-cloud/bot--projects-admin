package botclient

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ViewOpts 通道展示档位（CLI --full/--id/--json；MCP full/ids）。
type ViewOpts struct {
	Full   bool // 接近 REST 全量（含裸 UUID）
	IDs    bool // 仍瘦，但附加裸 id
	JSON   bool // 强制 JSON（默认文本）
	Detail bool // issue get/写回结果：瘦输出含 version（list 仍不含）
}

// HexUUIDRe 匹配 8-4-4-4-12 或连续 32 hex（验收扫默认 stdout 用）。
var HexUUIDRe = regexp.MustCompile(`(?i)(?:\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b|\b[0-9a-f]{32}\b)`)

func ContainsHexUUID(s string) bool {
	return HexUUIDRe.MatchString(s)
}

func shortTime(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t.UTC().Format("2006-01-02T15:04:05Z")
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UTC().Format("2006-01-02T15:04:05Z")
	}
	if len(s) >= 19 {
		return s[:19] + "Z"
	}
	return s
}

func trunc(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

func dash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

func strField(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil {
			switch t := v.(type) {
			case string:
				return t
			default:
				s := strings.TrimSpace(fmt.Sprint(t))
				if s != "" && s != "<nil>" {
					return s
				}
			}
		}
	}
	return ""
}

func ptrStrField(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case map[string]any:
		return ""
	default:
		s := strings.TrimSpace(fmt.Sprint(t))
		if s == "<nil>" {
			return ""
		}
		return s
	}
}

func asObjectList(raw json.RawMessage, wrapKey string) ([]map[string]any, map[string]any, error) {
	var top any
	if err := json.Unmarshal(raw, &top); err != nil {
		return nil, nil, err
	}
	switch t := top.(type) {
	case []any:
		out := make([]map[string]any, 0, len(t))
		for _, it := range t {
			if m, ok := it.(map[string]any); ok {
				out = append(out, m)
			}
		}
		return out, nil, nil
	case map[string]any:
		if wrapKey != "" {
			if arr, ok := t[wrapKey].([]any); ok {
				out := make([]map[string]any, 0, len(arr))
				for _, it := range arr {
					if m, ok := it.(map[string]any); ok {
						out = append(out, m)
					}
				}
				return out, t, nil
			}
			if one, ok := t[wrapKey].(map[string]any); ok {
				return []map[string]any{one}, t, nil
			}
		}
		return []map[string]any{t}, t, nil
	default:
		return nil, nil, fmt.Errorf("unexpected JSON shape")
	}
}

// --- Issue ---

// nestedDisplayName 读嵌套对象的 display_name（或纯字符串）；无则空。
func nestedDisplayName(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case map[string]any:
		return strings.TrimSpace(strField(t, "display_name", "name"))
	default:
		return ""
	}
}

// resolveMemberLabel：嵌套名优先，再 MemberNames(id)；members 失败时默认 "-"，--id/--full 才回落裸 id。
func resolveMemberLabel(m map[string]any, nestKey, idKey string, names map[string]string, opts ViewOpts) string {
	if n := nestedDisplayName(m, nestKey); n != "" {
		return n
	}
	aid := ptrStrField(m, idKey)
	if aid == "" {
		return "-"
	}
	if n := names[aid]; n != "" {
		return n
	}
	if opts.IDs || opts.Full {
		return aid
	}
	return "-"
}

func (c *Client) projectIssueMap(m map[string]any, names map[string]string, opts ViewOpts) map[string]any {
	id := strField(m, "id")
	ident := strField(m, "identifier")
	if id != "" && ident != "" {
		c.rememberIssue(id, ident)
	}
	assignee := resolveMemberLabel(m, "assignee", "assignee_id", names, opts)
	if opts.Full {
		// near-REST：保留 REST 原字段；仅当无嵌套 display_name 时注入可读字符串 assignee
		out := map[string]any{}
		for k, v := range m {
			out[k] = v
		}
		if nestedDisplayName(m, "assignee") == "" {
			out["assignee"] = assignee
		}
		return out
	}
	out := map[string]any{
		"identifier": ident,
		"title":      strField(m, "title"),
		"status":     strField(m, "status"),
		"priority":   dash(strField(m, "priority")),
		"assignee":   assignee,
	}
	if ua := strField(m, "updated_at"); ua != "" {
		out["updated_at"] = shortTime(ua)
	}
	if opts.Detail {
		if v, ok := m["version"]; ok && v != nil {
			out["version"] = v
		}
	}
	if opts.IDs && id != "" {
		out["id"] = id
	}
	return out
}

func (c *Client) ProjectIssues(raw json.RawMessage, opts ViewOpts) (slim []map[string]any, err error) {
	var top map[string]any
	if err := json.Unmarshal(raw, &top); err != nil {
		return nil, err
	}
	var list []map[string]any
	if arr, ok := top["issues"].([]any); ok {
		for _, it := range arr {
			if m, ok := it.(map[string]any); ok {
				list = append(list, m)
			}
		}
	} else if one, ok := top["issue"].(map[string]any); ok {
		list = []map[string]any{one}
	} else if _, hasID := top["id"]; hasID || top["identifier"] != nil {
		list = []map[string]any{top}
	} else {
		return nil, fmt.Errorf("unexpected issue JSON shape")
	}
	names, _ := c.MemberNames()
	if names == nil {
		names = map[string]string{}
	}
	out := make([]map[string]any, 0, len(list))
	for _, m := range list {
		out = append(out, c.projectIssueMap(m, names, opts))
	}
	return out, nil
}

func FormatIssueListText(items []map[string]any) string {
	if len(items) == 0 {
		return "(无 Issue)\n"
	}
	var b strings.Builder
	for _, it := range items {
		fmt.Fprintf(&b, "%s\t%s\t%s\t%s\t%s",
			dash(strField(it, "identifier")),
			dash(strField(it, "title")),
			dash(strField(it, "status")),
			dash(strField(it, "priority")),
			dash(strField(it, "assignee")),
		)
		if ua := strField(it, "updated_at"); ua != "" {
			fmt.Fprintf(&b, "\t%s", ua)
		}
		if id := strField(it, "id"); id != "" {
			fmt.Fprintf(&b, "\tid=%s", id)
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func FormatIssueGetText(it map[string]any) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s  %s\n", dash(strField(it, "identifier")), dash(strField(it, "title")))
	fmt.Fprintf(&b, "status=%s  priority=%s  assignee=%s\n",
		dash(strField(it, "status")), dash(strField(it, "priority")), dash(strField(it, "assignee")))
	if v, ok := it["version"]; ok && v != nil {
		fmt.Fprintf(&b, "version=%v\n", v)
	}
	if ua := strField(it, "updated_at"); ua != "" {
		fmt.Fprintf(&b, "updated_at=%s\n", ua)
	}
	if id := strField(it, "id"); id != "" {
		fmt.Fprintf(&b, "id=%s\n", id)
	}
	return b.String()
}

// --- WhoAmI ---

func ProjectWhoAmI(raw json.RawMessage, opts ViewOpts) map[string]any {
	var m map[string]any
	_ = json.Unmarshal(raw, &m)
	if m == nil {
		m = map[string]any{}
	}
	if opts.Full {
		return m
	}
	out := map[string]any{
		"actor_kind":   strField(m, "actor_kind"),
		"display_name": strField(m, "display_name"),
	}
	if opts.IDs {
		if v := strField(m, "member_id"); v != "" {
			out["member_id"] = v
		}
		if v := strField(m, "workspace_id"); v != "" {
			out["workspace_id"] = v
		}
	}
	return out
}

func FormatWhoAmIText(m map[string]any) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s  %s\n", dash(strField(m, "actor_kind")), dash(strField(m, "display_name")))
	if v := strField(m, "member_id"); v != "" {
		fmt.Fprintf(&b, "member_id=%s\n", v)
	}
	if v := strField(m, "workspace_id"); v != "" {
		fmt.Fprintf(&b, "workspace_id=%s\n", v)
	}
	return b.String()
}

// --- Comments ---

func (c *Client) ProjectComments(raw json.RawMessage, issueIdent string, opts ViewOpts) ([]map[string]any, error) {
	var top map[string]any
	if err := json.Unmarshal(raw, &top); err != nil {
		return nil, err
	}
	var list []map[string]any
	if arr, ok := top["comments"].([]any); ok {
		for _, it := range arr {
			if m, ok := it.(map[string]any); ok {
				list = append(list, m)
			}
		}
	} else if one, ok := top["comment"].(map[string]any); ok {
		list = []map[string]any{one}
	} else {
		return nil, fmt.Errorf("unexpected comment JSON shape")
	}
	names, _ := c.MemberNames()
	if names == nil {
		names = map[string]string{}
	}
	out := make([]map[string]any, 0, len(list))
	for _, m := range list {
		if opts.Full {
			cp := map[string]any{}
			for k, v := range m {
				cp[k] = v
			}
			if issueIdent != "" {
				cp["issue"] = issueIdent
			}
			if nestedDisplayName(m, "author") == "" {
				aid := strField(m, "author_id")
				if n := names[aid]; n != "" {
					cp["author"] = n
				}
			}
			out = append(out, cp)
			continue
		}
		author := nestedDisplayName(m, "author")
		if author == "" {
			aid := strField(m, "author_id")
			if n := names[aid]; n != "" {
				author = n
			} else {
				author = "-"
			}
		}
		body := strField(m, "body")
		item := map[string]any{
			"author":     author,
			"created_at": shortTime(strField(m, "created_at")),
			"body":       trunc(body, 120),
		}
		if issueIdent != "" {
			item["issue"] = issueIdent
		}
		if opts.IDs {
			if id := strField(m, "id"); id != "" {
				item["id"] = id
			}
		}
		out = append(out, item)
	}
	return out, nil
}

func FormatCommentListText(items []map[string]any) string {
	if len(items) == 0 {
		return "(无评论)\n"
	}
	var b strings.Builder
	for _, it := range items {
		issue := strField(it, "issue")
		prefix := ""
		if issue != "" {
			prefix = issue + "  "
		}
		fmt.Fprintf(&b, "%s%s  %s  %s",
			prefix,
			dash(strField(it, "author")),
			dash(strField(it, "created_at")),
			dash(strField(it, "body")),
		)
		if id := strField(it, "id"); id != "" {
			fmt.Fprintf(&b, "  id=%s", id)
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// --- Memory ---

func ProjectMemories(raw json.RawMessage, opts ViewOpts) ([]map[string]any, error) {
	var top map[string]any
	if err := json.Unmarshal(raw, &top); err != nil {
		return nil, err
	}
	var list []map[string]any
	if arr, ok := top["memories"].([]any); ok {
		for _, it := range arr {
			if m, ok := it.(map[string]any); ok {
				list = append(list, m)
			}
		}
	} else if one, ok := top["memory"].(map[string]any); ok {
		list = []map[string]any{one}
	} else {
		return nil, fmt.Errorf("unexpected memory JSON shape")
	}
	out := make([]map[string]any, 0, len(list))
	for _, m := range list {
		if opts.Full {
			cp := map[string]any{}
			for k, v := range m {
				cp[k] = v
			}
			out = append(out, cp)
			continue
		}
		item := map[string]any{
			"key":        strField(m, "key_name", "key"),
			"value":      trunc(strField(m, "value_text", "value"), 80),
			"updated_at": shortTime(strField(m, "updated_at")),
		}
		if tags := strField(m, "tags"); tags != "" && tags != "[]" && tags != "null" {
			item["tags"] = trunc(tags, 40)
		}
		if opts.IDs {
			if id := strField(m, "id"); id != "" {
				item["id"] = id
			}
		}
		out = append(out, item)
	}
	return out, nil
}

func FormatMemoryListText(items []map[string]any) string {
	if len(items) == 0 {
		return "(无记忆)\n"
	}
	var b strings.Builder
	for _, it := range items {
		fmt.Fprintf(&b, "%s\t%s\t%s",
			dash(strField(it, "key")),
			dash(strField(it, "updated_at")),
			dash(strField(it, "value")),
		)
		if tags := strField(it, "tags"); tags != "" {
			fmt.Fprintf(&b, "\t%s", tags)
		}
		if id := strField(it, "id"); id != "" {
			fmt.Fprintf(&b, "\tid=%s", id)
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// --- Handoff ---

func (c *Client) ProjectHandoff(raw json.RawMessage, opts ViewOpts) (any, error) {
	var top map[string]any
	if err := json.Unmarshal(raw, &top); err != nil {
		return nil, err
	}
	hdo, _ := top["handoff"].(map[string]any)
	if hdo == nil {
		hdo = top
	}
	if opts.Full {
		return top, nil
	}
	names, _ := c.MemberNames()
	if names == nil {
		names = map[string]string{}
	}
	roleName := ""
	if role, ok := hdo["claimed_role"].(map[string]any); ok && role != nil {
		roleName = strField(role, "name", "slug")
	}
	memN, todoN := 0, 0
	if arr, ok := hdo["memories"].([]any); ok {
		memN = len(arr)
	}
	if arr, ok := hdo["todos"].([]any); ok {
		todoN = len(arr)
	}
	var issues []map[string]any
	if arr, ok := hdo["issues"].([]any); ok {
		for _, it := range arr {
			m, ok := it.(map[string]any)
			if !ok {
				continue
			}
			ident := strField(m, "identifier")
			item := map[string]any{
				"identifier": ident,
				"title":      strField(m, "title"),
				"status":     strField(m, "status"),
			}
			if opts.IDs {
				if id := strField(m, "id"); id != "" {
					item["id"] = id
				}
			}
			issues = append(issues, item)
			if id := strField(m, "id"); id != "" && ident != "" {
				c.rememberIssue(id, ident)
			}
		}
	}
	out := map[string]any{
		"claimed_role": dash(roleName),
		"memory_count": memN,
		"todo_count":   todoN,
		"issue_count":  len(issues),
		"issues":       issues,
	}
	AttachPlaybookRef(out, strField(hdo, "playbook_ref"))
	if opts.IDs {
		if mid := strField(hdo, "member_id"); mid != "" {
			out["member_id"] = mid
			if n := names[mid]; n != "" {
				out["member"] = n
			}
		}
	} else if mid := strField(hdo, "member_id"); mid != "" {
		if n := names[mid]; n != "" {
			out["member"] = n
		}
	}
	return out, nil
}

func FormatHandoffText(v any) string {
	m, ok := v.(map[string]any)
	if !ok {
		b, _ := json.MarshalIndent(v, "", "  ")
		return string(b) + "\n"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "role=%s  memories=%v  todos=%v  open_issues=%v\n",
		dash(strField(m, "claimed_role")), m["memory_count"], m["todo_count"], m["issue_count"])
	if member := strField(m, "member"); member != "" {
		fmt.Fprintf(&b, "member=%s\n", member)
	}
	if p := strField(m, "playbook_ref"); p != "" {
		fmt.Fprintf(&b, "手册 playbook_ref=%s\n", p)
	}
	if w := strField(m, "playbook_warn"); w != "" {
		fmt.Fprintf(&b, "%s\n", w)
	}
	if arr, ok := m["issues"].([]map[string]any); ok {
		for _, it := range arr {
			fmt.Fprintf(&b, "  %s  %s  %s\n",
				dash(strField(it, "identifier")),
				dash(strField(it, "title")),
				dash(strField(it, "status")))
		}
	} else if arr, ok := m["issues"].([]any); ok {
		for _, raw := range arr {
			it, _ := raw.(map[string]any)
			if it == nil {
				continue
			}
			fmt.Fprintf(&b, "  %s  %s  %s\n",
				dash(strField(it, "identifier")),
				dash(strField(it, "title")),
				dash(strField(it, "status")))
		}
	}
	return b.String()
}

// ProjectHandoffImport 导入结果默认只留计数/warnings。
func ProjectHandoffImport(raw json.RawMessage, opts ViewOpts) any {
	var m map[string]any
	_ = json.Unmarshal(raw, &m)
	if m == nil {
		return map[string]any{}
	}
	if opts.Full {
		return m
	}
	out := map[string]any{}
	for _, k := range []string{"imported_memories", "imported_todos", "claimed", "warnings"} {
		if v, ok := m[k]; ok {
			out[k] = v
		}
	}
	return out
}

// MarshalView JSON 缩进。
func MarshalView(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprint(v)
	}
	return string(b) + "\n"
}
