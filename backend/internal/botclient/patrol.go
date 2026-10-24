package botclient

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// OpenIssueStatuses：巡检「未完成」口径（与 BOT_PATROL_P05 一致）。
var OpenIssueStatuses = map[string]bool{
	"todo":        true,
	"in_progress": true,
	"in_review":   true,
}

// IsOpenIssueStatus 是否 open（todo/in_progress/in_review）。
func IsOpenIssueStatus(status string) bool {
	return OpenIssueStatuses[strings.TrimSpace(strings.ToLower(status))]
}

// FilterOpenIssues 仅保留 open 状态项（就地过滤 slim ProjectIssues 行）。
func FilterOpenIssues(items []map[string]any) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, it := range items {
		if IsOpenIssueStatus(strField(it, "status")) {
			out = append(out, it)
		}
	}
	return out
}

// PatrolFingerprint 稳定指纹：按 identifier 排序后 identifier|status|updated_at。
// 无变化 → 相同 hex；空列表 → 固定空指纹。
func PatrolFingerprint(items []map[string]any) string {
	type row struct{ ident, status, updated string }
	rows := make([]row, 0, len(items))
	for _, it := range items {
		rows = append(rows, row{
			ident:   strField(it, "identifier"),
			status:  strField(it, "status"),
			updated: strField(it, "updated_at"),
		})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].ident != rows[j].ident {
			return rows[i].ident < rows[j].ident
		}
		if rows[i].status != rows[j].status {
			return rows[i].status < rows[j].status
		}
		return rows[i].updated < rows[j].updated
	})
	var b strings.Builder
	for _, r := range rows {
		fmt.Fprintf(&b, "%s|%s|%s\n", r.ident, r.status, r.updated)
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

// PatrolQuietDecision：quiet 模式下是否应抑制 stdout。
// prevFP 空（无状态文件）且当前非空 → 不静默（要刷出新指派）。
// prevFP 空且当前空 → 静默（零单无刷屏）并应写入空指纹。
// prevFP == curFP → 静默；否则刷出。
func PatrolQuietDecision(quiet bool, prevFP, curFP string) (suppress bool) {
	if !quiet {
		return false
	}
	if prevFP == "" {
		// 首次：仅空 inbox 安静；有单则输出一次并落盘
		return curFP == PatrolFingerprint(nil)
	}
	return prevFP == curFP
}

// ReadPatrolStateFP 读 STATE_FILE 首行指纹（忽略空白/#注释行）。
func ReadPatrolStateFP(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		return line, nil
	}
	return "", nil
}

// WritePatrolStateFP 写指纹到 STATE_FILE（自动 mkdir）。
func WritePatrolStateFP(path, fp string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("需要 state file 路径")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body := "# ymbot patrol fingerprint\n" + strings.TrimSpace(fp) + "\n"
	return os.WriteFile(path, []byte(body), 0o600)
}

// DefaultPatrolStateFile：YMBOT_PATROL_STATE_FILE 或 $HOME/.ymbot/patrol-state。
func DefaultPatrolStateFile() string {
	if v := strings.TrimSpace(os.Getenv("YMBOT_PATROL_STATE_FILE")); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ".ymbot-patrol-state"
	}
	return filepath.Join(home, ".ymbot", "patrol-state")
}

// PatrolResult 一次巡检结果（CLI/MCP 共用）。
type PatrolResult struct {
	Items     []map[string]any
	Fingerprint string
	Suppressed  bool // quiet 且无变化：无 stdout
	Text        string
}

// PatrolOpts 巡检选项。
type PatrolOpts struct {
	Quiet     bool
	StateFile string
	View      ViewOpts
}

// Patrol 拉 assignee=me 的 Issue，过滤 open，可选 quiet fingerprint。
func (c *Client) Patrol(opts PatrolOpts) (*PatrolResult, error) {
	qv := url.Values{}
	qv.Set("assignee_id", "me")
	raw, err := c.IssueList(qv)
	if err != nil {
		return nil, err
	}
	var items []map[string]any
	if opts.View.Full {
		// Full：仍过滤 open，但保留接近 REST 的行
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
		}
		names, _ := c.MemberNames()
		if names == nil {
			names = map[string]string{}
		}
		for _, m := range list {
			if !IsOpenIssueStatus(strField(m, "status")) {
				continue
			}
			items = append(items, c.projectIssueMap(m, names, opts.View))
		}
	} else {
		slim, err := c.ProjectIssues(raw, opts.View)
		if err != nil {
			return nil, err
		}
		items = FilterOpenIssues(slim)
	}
	fp := PatrolFingerprint(items)
	statePath := strings.TrimSpace(opts.StateFile)
	if statePath == "" {
		statePath = DefaultPatrolStateFile()
	}
	prev, err := ReadPatrolStateFP(statePath)
	if err != nil {
		return nil, err
	}
	suppress := PatrolQuietDecision(opts.Quiet, prev, fp)
	res := &PatrolResult{Items: items, Fingerprint: fp, Suppressed: suppress}
	if !suppress {
		if opts.View.Full || opts.View.JSON {
			res.Text = MarshalView(items)
		} else {
			res.Text = FormatIssueListText(items)
		}
	}
	// 落盘 fingerprint，供下次 --quiet-if-unchanged 比对
	if err := WritePatrolStateFP(statePath, fp); err != nil {
		return nil, err
	}
	return res, nil
}
