package botclient

import (
	"encoding/json"
	"strings"
)

// 作业手册唯一入口 = 席位 playbook_ref。空/非法只警示，绝不回落硬编码别的 playbook。
const (
	PlaybookRefEmptyWarn = "警示：本席 playbook_ref 为空，无作业手册入口（勿回落硬编码流程）"
	PlaybookRefBadWarn   = "警示：本席 playbook_ref 非法，无法作为手册入口（勿回落硬编码流程）"
	PlaybookRefNoSeatWarn = "警示：无团队席位上下文，playbook_ref 不可用（请 seat enter --team）"
)

// ProjectPlaybookRef 投影手册路径；空/非法 → warn 一行中文；绝不替换成硬编码手册。
// 返回的 ref 可截断展示；warn 非空时调用方应原样打印。
func ProjectPlaybookRef(raw string) (ref string, warn string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", PlaybookRefEmptyWarn
	}
	if strings.Contains(raw, "..") || strings.ContainsRune(raw, 0) || strings.ContainsAny(raw, "\n\r") {
		return trunc(raw, 72), PlaybookRefBadWarn
	}
	if strings.HasPrefix(raw, "{") {
		var p struct {
			Repo string `json:"repo"`
			Ref  string `json:"ref"`
			Path string `json:"path"`
		}
		if err := json.Unmarshal([]byte(raw), &p); err != nil {
			return trunc(raw, 72), PlaybookRefBadWarn
		}
		if strings.TrimSpace(p.Repo) == "" || strings.TrimSpace(p.Ref) == "" || strings.TrimSpace(p.Path) == "" {
			return trunc(raw, 72), PlaybookRefBadWarn
		}
		if strings.Contains(p.Path, "..") {
			return trunc(raw, 72), PlaybookRefBadWarn
		}
		formatted := strings.TrimSpace(p.Repo) + "@" + strings.TrimSpace(p.Ref) + ":" + strings.TrimSpace(p.Path)
		return trunc(formatted, 96), ""
	}
	// repo@ref:path — 若形似但缺 path 段则非法
	if at := strings.Index(raw, "@"); at > 0 {
		rest := raw[at+1:]
		colon := strings.Index(rest, ":")
		if colon <= 0 || strings.TrimSpace(rest[colon+1:]) == "" {
			return trunc(raw, 72), PlaybookRefBadWarn
		}
		if strings.Contains(rest[colon+1:], "..") {
			return trunc(raw, 72), PlaybookRefBadWarn
		}
		return trunc(raw, 96), ""
	}
	// 纯路径：须含 / 或 .md（与 Go ParsePlaybookRef 对齐）；否则非法、不回落
	if !strings.Contains(raw, "/") && !strings.HasSuffix(raw, ".md") {
		return trunc(raw, 72), PlaybookRefBadWarn
	}
	return trunc(raw, 96), ""
}

// AttachPlaybookRef 写入瘦投影字段 playbook_ref（及可选 playbook_warn）；无第二套「流程说明」字段。
func AttachPlaybookRef(dst map[string]any, raw string) {
	if dst == nil {
		return
	}
	ref, warn := ProjectPlaybookRef(raw)
	if ref != "" {
		dst["playbook_ref"] = ref
	}
	if warn != "" {
		dst["playbook_warn"] = warn
	}
}
