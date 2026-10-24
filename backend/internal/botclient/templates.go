package botclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// TeamTemplatesList GET …/team-templates
func (c *Client) TeamTemplatesList() (json.RawMessage, error) {
	return c.Do(http.MethodGet, c.ws("/team-templates"), nil, nil, false)
}

// CreateTeamFromTemplateOpts POST …/teams/from-template（或 CreateTeam + template_key）。
type CreateTeamFromTemplateOpts struct {
	Name        string
	Slug        string
	TemplateKey string
	ID          string // 可选客户端指定 uuid
}

// CreateTeamFromTemplate POST …/teams/from-template → {team, roster}；未知模板 → 400 人话。
func (c *Client) CreateTeamFromTemplate(opts CreateTeamFromTemplateOpts) (json.RawMessage, error) {
	name := strings.TrimSpace(opts.Name)
	tk := strings.TrimSpace(opts.TemplateKey)
	if name == "" {
		return nil, fmt.Errorf("需要 --name <团队名>")
	}
	if tk == "" {
		return nil, fmt.Errorf("需要 --template <template_key>（如 default）")
	}
	body := map[string]any{
		"name":         name,
		"template_key": tk,
	}
	if s := strings.TrimSpace(opts.Slug); s != "" {
		body["slug"] = s
	}
	if id := strings.TrimSpace(opts.ID); id != "" {
		body["id"] = id
	}
	return c.Do(http.MethodPost, c.ws("/teams/from-template"), nil, body, true)
}

// CreateTeam POST …/teams；可带 template_key（与 from-template 同形 roster）。
func (c *Client) CreateTeam(opts CreateTeamFromTemplateOpts) (json.RawMessage, error) {
	name := strings.TrimSpace(opts.Name)
	if name == "" {
		return nil, fmt.Errorf("需要 --name <团队名>")
	}
	body := map[string]any{"name": name}
	if s := strings.TrimSpace(opts.Slug); s != "" {
		body["slug"] = s
	}
	if tk := strings.TrimSpace(opts.TemplateKey); tk != "" {
		body["template_key"] = tk
	}
	if id := strings.TrimSpace(opts.ID); id != "" {
		body["id"] = id
	}
	return c.Do(http.MethodPost, c.ws("/teams"), nil, body, true)
}

// ProjectTeamTemplates 瘦投影 templates[] → [{key,name,seats_n,seats:[slug…]}]；--id 附加无。
func ProjectTeamTemplates(raw json.RawMessage, opts ViewOpts) ([]map[string]any, error) {
	var top map[string]any
	if err := json.Unmarshal(raw, &top); err != nil {
		return nil, err
	}
	var list []map[string]any
	if arr, ok := top["templates"].([]any); ok {
		for _, it := range arr {
			if m, ok := it.(map[string]any); ok {
				list = append(list, m)
			}
		}
	} else {
		return nil, fmt.Errorf("unexpected templates JSON shape")
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
		key := strField(m, "key", "id")
		name := strField(m, "name")
		item := map[string]any{
			"key":  key,
			"name": dash(name),
		}
		if d := strField(m, "description"); d != "" {
			item["description"] = trunc(d, 80)
		}
		var slugs []string
		if arr, ok := m["seats"].([]any); ok {
			for _, it := range arr {
				sm, ok := it.(map[string]any)
				if !ok {
					continue
				}
				slug := strField(sm, "slug", "role_slug")
				if slug == "" {
					continue
				}
				slugs = append(slugs, slug)
			}
		}
		item["seats_n"] = len(slugs)
		item["seats"] = slugs
		out = append(out, item)
	}
	return out, nil
}

// FormatTeamTemplatesText 默认瘦文本：key\tname\tseats=N\tslug,…
func FormatTeamTemplatesText(items []map[string]any) string {
	if len(items) == 0 {
		return "(无模板)\n"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "templates=%d\n", len(items))
	for _, it := range items {
		fmt.Fprintf(&b, "%s\t%s\tseats=%v",
			dash(strField(it, "key")),
			dash(strField(it, "name")),
			it["seats_n"],
		)
		if arr, ok := it["seats"].([]string); ok && len(arr) > 0 {
			fmt.Fprintf(&b, "\t%s", strings.Join(arr, ","))
		}
		if d := strField(it, "description"); d != "" {
			fmt.Fprintf(&b, "\t%s", d)
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// ProjectCreateTeamFromTemplate 瘦投影 {team,roster}；默认无裸 UUID（--id 才带）。
func ProjectCreateTeamFromTemplate(raw json.RawMessage, opts ViewOpts) (map[string]any, error) {
	var top map[string]any
	if err := json.Unmarshal(raw, &top); err != nil {
		return nil, err
	}
	if opts.Full {
		return top, nil
	}
	out := map[string]any{}
	team, _ := top["team"].(map[string]any)
	if team == nil {
		team = top
	}
	tSlim := map[string]any{
		"name": dash(strField(team, "name")),
		"slug": dash(strField(team, "slug")),
	}
	if tk := strField(team, "template_key"); tk != "" {
		tSlim["template"] = tk
	}
	if opts.IDs {
		if id := strField(team, "id"); id != "" {
			tSlim["id"] = id
		}
	}
	out["team"] = tSlim

	// roster 可能在 top.roster 或 top.roster 包一层
	var rosterRaw json.RawMessage
	if r, ok := top["roster"]; ok && r != nil {
		b, _ := json.Marshal(r)
		// 若已是 {roster:[…]} 或裸数组，统一包一层
		var probe any
		_ = json.Unmarshal(b, &probe)
		switch probe.(type) {
		case []any:
			rosterRaw = json.RawMessage(`{"roster":` + string(b) + `}`)
		default:
			rosterRaw = b
			// 若是对象但没有 roster/seats key，再包
			if m, ok := probe.(map[string]any); ok {
				if _, hasR := m["roster"]; !hasR {
					if _, hasS := m["seats"]; !hasS {
						rosterRaw = json.RawMessage(`{"roster":` + string(b) + `}`)
					}
				}
			}
		}
		items, err := ProjectTeamRoster(rosterRaw, opts)
		if err == nil {
			out["roster"] = items
			out["seats_n"] = len(items)
		}
	}
	return out, nil
}

// FormatCreateTeamFromTemplateText 创建结果瘦文本。
func FormatCreateTeamFromTemplateText(v map[string]any) string {
	var b strings.Builder
	team, _ := v["team"].(map[string]any)
	if team == nil {
		team = map[string]any{}
	}
	fmt.Fprintf(&b, "created name=%s  slug=%s  template=%s",
		dash(strField(team, "name")),
		dash(strField(team, "slug")),
		dash(strField(team, "template")),
	)
	if id := strField(team, "id"); id != "" {
		fmt.Fprintf(&b, "  id=%s", id)
	}
	b.WriteByte('\n')
	if items, ok := v["roster"].([]map[string]any); ok {
		b.WriteString(FormatTeamRosterText(items))
	} else if n, ok := v["seats_n"]; ok {
		fmt.Fprintf(&b, "seats=%v\n", n)
	}
	return b.String()
}
