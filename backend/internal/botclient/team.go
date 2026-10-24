package botclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// DefaultTeamSeatSlugs 班底缺省角色（与 seed / 模板对齐；ensure 可不传 body）。
var DefaultTeamSeatSlugs = []string{"frontend", "pm", "qa", "backend", "wiki"}

// SeatMemoryKey 兼岗记忆隔离：seat_slug:key（如 frontend:last_focus）。
func SeatMemoryKey(seatSlug, key string) string {
	seatSlug = strings.TrimSpace(seatSlug)
	key = strings.TrimSpace(key)
	if seatSlug == "" {
		return key
	}
	if key == "" {
		return seatSlug + ":"
	}
	return seatSlug + ":" + key
}

// TeamRosterGet GET …/teams/{teamID}/roster
func (c *Client) TeamRosterGet(teamID string) (json.RawMessage, error) {
	teamID = strings.TrimSpace(teamID)
	if teamID == "" {
		return nil, fmt.Errorf("需要 team uuid（--uuid）")
	}
	return c.Do(http.MethodGet, c.ws("/teams/"+teamID+"/roster"), nil, nil, false)
}

// TeamRosterEnsure POST …/teams/{teamID}/roster/ensure
// seats 可选；nil/空 → 服务端用 seed 缺省角色。幂等：不双建、不清 memory。
func (c *Client) TeamRosterEnsure(teamID string, seats []map[string]any) (json.RawMessage, error) {
	teamID = strings.TrimSpace(teamID)
	if teamID == "" {
		return nil, fmt.Errorf("需要 team uuid（--uuid）")
	}
	var body any
	if len(seats) > 0 {
		body = map[string]any{"seats": seats}
	} else {
		body = map[string]any{}
	}
	return c.Do(http.MethodPost, c.ws("/teams/"+teamID+"/roster/ensure"), nil, body, true)
}

// TeamSeatExpand POST …/teams/{teamID}/expand（别名 …/seats/expand）
// Go ExpandInput：project_id?、limit? → {assigned,skipped}；不碰 in-progress。
// DryRun=true 且 404 时只读建议，不静默改库。
func (c *Client) TeamSeatExpand(teamID string, opts TeamExpandOpts) (json.RawMessage, error) {
	teamID = strings.TrimSpace(teamID)
	if teamID == "" {
		return nil, fmt.Errorf("需要 --team <uuid>（或当前上下文）")
	}
	body := map[string]any{}
	if opts.ProjectID != "" {
		body["project_id"] = opts.ProjectID
	}
	if opts.Limit > 0 {
		body["limit"] = opts.Limit
	}
	if opts.DryRun {
		body["dry_run"] = true
	}
	raw, err := c.Do(http.MethodPost, c.ws("/teams/"+teamID+"/expand"), nil, body, true)
	if err == nil {
		return raw, nil
	}
	ae, ok := err.(*APIError)
	if !ok || ae.Status != 404 || !opts.DryRun {
		return nil, err
	}
	return c.teamExpandDryRunLocal(teamID, opts)
}

// TeamExpandOpts 对齐 Go ExpandInput + 通道 dry-run。
type TeamExpandOpts struct {
	ProjectID string
	Limit     int
	DryRun    bool
}

func (c *Client) teamExpandDryRunLocal(teamID string, opts TeamExpandOpts) (json.RawMessage, error) {
	_ = teamID
	_ = opts
	out := map[string]any{
		"dry_run":  true,
		"note":     "expand API 未就绪；dry-run 仅列建议，未改库、未 claim、未改派 in-progress",
		"assigned": []any{},
		"skipped": []map[string]any{
			{"reason": "须服务端 POST …/teams/{id}/expand；仅派未领叶子给空闲已绑席"},
		},
	}
	b, _ := json.Marshal(out)
	return b, nil
}

// TeamSeatBind POST …/teams/{teamID}/seats/{seatID}/bind
// body: force?；Bot 侧 member 由服务端取 actor。活跃占用一律 409；force 仅清无 active claim 的脏绑。
func (c *Client) TeamSeatBind(teamID, seatID string, force bool) (json.RawMessage, error) {
	teamID = strings.TrimSpace(teamID)
	seatID = strings.TrimSpace(seatID)
	if teamID == "" || seatID == "" {
		return nil, fmt.Errorf("需要 team uuid 与 seat id")
	}
	body := map[string]any{}
	if force {
		body["force"] = true
	}
	return c.Do(http.MethodPost, c.ws("/teams/"+teamID+"/seats/"+seatID+"/bind"), nil, body, true)
}

// TeamSeatUnbind POST …/teams/{teamID}/seats/{seatID}/unbind
func (c *Client) TeamSeatUnbind(teamID, seatID string) (json.RawMessage, error) {
	teamID = strings.TrimSpace(teamID)
	seatID = strings.TrimSpace(seatID)
	if teamID == "" || seatID == "" {
		return nil, fmt.Errorf("需要 team uuid 与 seat id")
	}
	return c.Do(http.MethodPost, c.ws("/teams/"+teamID+"/seats/"+seatID+"/unbind"), nil, map[string]any{}, true)
}

// SeatBySlug 从 roster 找 role_slug；返回 seat 原始 map（含 id）。
func SeatBySlug(rosterRaw json.RawMessage, slug string) (map[string]any, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil, fmt.Errorf("需要 --role <slug>")
	}
	seats, _, err := asObjectListPrefer(rosterRaw, "roster", "seats")
	if err != nil {
		return nil, err
	}
	for _, s := range seats {
		if strField(s, "role_slug", "slug") == slug {
			return s, nil
		}
	}
	return nil, &APIError{Status: 404, Code: "not_found", Message: fmt.Sprintf("班底无席位 slug=%s", slug)}
}

// ListBotRoles GET …/bot-roles
func (c *Client) ListBotRoles() (json.RawMessage, error) {
	return c.Do(http.MethodGet, c.ws("/bot-roles"), nil, nil, false)
}

// RoleBySlug 从 bot-roles 列表解析 slug → role（含 claim）。
func (c *Client) RoleBySlug(slug string) (map[string]any, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil, fmt.Errorf("需要 --role <slug>")
	}
	raw, err := c.ListBotRoles()
	if err != nil {
		return nil, err
	}
	roles, _, err := asObjectList(raw, "bot_roles")
	if err != nil {
		return nil, err
	}
	for _, r := range roles {
		if strField(r, "slug") == slug {
			return r, nil
		}
	}
	return nil, &APIError{Status: 404, Code: "not_found", Message: fmt.Sprintf("找不到角色 slug=%s", slug)}
}

// RoleClaim 按 role id claim（现有 REST）。
func (c *Client) RoleClaim(roleID string) (json.RawMessage, error) {
	roleID = strings.TrimSpace(roleID)
	if roleID == "" {
		return nil, fmt.Errorf("需要 role id")
	}
	mid, err := c.ResolveMember()
	if err != nil {
		return nil, err
	}
	body := map[string]string{"member_id": mid}
	return c.Do(http.MethodPost, c.ws("/bot-roles/"+roleID+"/claim"), nil, body, true)
}

// RoleRelease 释放角色 claim。
func (c *Client) RoleRelease(roleID string) (json.RawMessage, error) {
	roleID = strings.TrimSpace(roleID)
	if roleID == "" {
		return nil, fmt.Errorf("需要 role id")
	}
	return c.Do(http.MethodPost, c.ws("/bot-roles/"+roleID+"/release"), nil, map[string]any{}, true)
}

// RoleClaimBySlug slug→role→claim；409 时附占用者一行中文（禁瞎抢）。
func (c *Client) RoleClaimBySlug(slug string) (json.RawMessage, error) {
	role, err := c.RoleBySlug(slug)
	if err != nil {
		return nil, err
	}
	roleID := strField(role, "id")
	raw, err := c.RoleClaim(roleID)
	if err == nil {
		return raw, nil
	}
	ae, ok := err.(*APIError)
	if !ok || ae.Status != 409 {
		return nil, err
	}
	return nil, c.enrichRoleClaimConflict(slug, role, ae)
}

// RoleReleaseBySlug slug→release。
func (c *Client) RoleReleaseBySlug(slug string) (json.RawMessage, error) {
	role, err := c.RoleBySlug(slug)
	if err != nil {
		return nil, err
	}
	return c.RoleRelease(strField(role, "id"))
}

func (c *Client) enrichRoleClaimConflict(slug string, role map[string]any, ae *APIError) error {
	holder := ""
	if claim, ok := role["claim"].(map[string]any); ok && claim != nil {
		mid := strField(claim, "member_id")
		if mid != "" {
			names, _ := c.MemberNames()
			if names != nil {
				holder = names[mid]
			}
			if holder == "" {
				holder = mid
			}
		}
	}
	// 再拉一次列表，claim 可能刚写入
	if holder == "" {
		if r2, err := c.RoleBySlug(slug); err == nil {
			if claim, ok := r2["claim"].(map[string]any); ok && claim != nil {
				mid := strField(claim, "member_id")
				names, _ := c.MemberNames()
				if names != nil && names[mid] != "" {
					holder = names[mid]
				} else if mid != "" {
					holder = mid
				}
			}
		}
	}
	msg := strings.TrimSpace(ae.Message)
	if msg == "" {
		msg = "该角色已被其他成员认领"
	}
	if holder != "" && !strings.Contains(msg, holder) {
		msg = fmt.Sprintf("%s（占用者：%s）", msg, holder)
	}
	msg = fmt.Sprintf("%s；禁止瞎抢；活跃席不可 force 顶替，请对方 seat leave", msg)
	return &APIError{Status: 409, Code: ae.Code, Message: msg}
}

// SeatEnterOpts 进场：有 TeamID 走 bind（服务端顺带 claim）；否则仅 role claim。
type SeatEnterOpts struct {
	TeamID string
	Force  bool
}

// SeatEnter 迁绑进场：whoami → handoff get → bind|role claim。
// 冲突 409 不吞；活跃席不可 force 顶替；--force 仅清脏绑。
func (c *Client) SeatEnter(roleSlug string, opts SeatEnterOpts) (map[string]any, error) {
	roleSlug = strings.TrimSpace(roleSlug)
	if roleSlug == "" {
		return nil, fmt.Errorf("用法: ymbot seat enter --role <slug> [--team <uuid>] [--force]")
	}
	whoRaw, err := c.WhoAmI()
	if err != nil {
		return nil, err
	}
	who := ProjectWhoAmI(whoRaw, ViewOpts{})

	handRaw, err := c.HandoffGet()
	if err != nil {
		return nil, err
	}
	hand, err := c.ProjectHandoff(handRaw, ViewOpts{})
	if err != nil {
		return nil, err
	}

	out := map[string]any{
		"role":            roleSlug,
		"whoami":          who,
		"handoff":         hand,
		"memory_key_hint": SeatMemoryKey(roleSlug, "last_focus"),
	}

	if teamID := strings.TrimSpace(opts.TeamID); teamID != "" {
		out["step"] = "whoami→handoff get→seat bind"
		rosterRaw, err := c.TeamRosterGet(teamID)
		if err != nil {
			// roster 空/团队新建：先 ensure 再 get
			if ae, ok := err.(*APIError); ok && ae.Status == 404 {
				rosterRaw, err = c.TeamPull(teamID)
			}
			if err != nil {
				return nil, err
			}
		}
		seat, err := SeatBySlug(rosterRaw, roleSlug)
		if err != nil {
			// ensure 后再试一次（席位可能尚未 seed）
			rosterRaw, err2 := c.TeamPull(teamID)
			if err2 != nil {
				return nil, err
			}
			seat, err = SeatBySlug(rosterRaw, roleSlug)
			if err != nil {
				return nil, err
			}
		}
		AttachPlaybookRef(out, strField(seat, "playbook_ref"))
		seatID := strField(seat, "id")
		bindRaw, err := c.TeamSeatBind(teamID, seatID, opts.Force)
		if err != nil {
			if ae, ok := err.(*APIError); ok && ae.Status == 409 {
				return nil, c.enrichBindConflict(roleSlug, seat, ae)
			}
			return nil, err
		}
		out["bind"] = projectSeatSlim(bindRaw, roleSlug, strField(seat, "playbook_ref"))
		out["claim"] = map[string]any{"ok": true, "via": "bind"}
		return out, nil
	}

	out["step"] = "whoami→handoff get→role claim"
	out["playbook_warn"] = PlaybookRefNoSeatWarn
	claimRaw, err := c.RoleClaimBySlug(roleSlug)
	if err != nil {
		return nil, err
	}
	out["claim"] = projectClaimSlim(claimRaw, roleSlug)
	return out, nil
}

func (c *Client) enrichBindConflict(slug string, seat map[string]any, ae *APIError) error {
	holder := strField(seat, "bound_display_name")
	if holder == "" {
		if claim, ok := seat["claim"].(map[string]any); ok && claim != nil {
			mid := strField(claim, "member_id")
			names, _ := c.MemberNames()
			if names != nil && names[mid] != "" {
				holder = names[mid]
			} else {
				holder = mid
			}
		}
	}
	msg := strings.TrimSpace(ae.Message)
	if msg == "" {
		msg = "该席位已绑定其他成员"
	}
	if holder != "" && !strings.Contains(msg, holder) {
		msg = fmt.Sprintf("%s（占用者：%s）", msg, holder)
	}
	msg = fmt.Sprintf("%s；禁止瞎抢；活跃席不可 --force 顶替，请对方 seat leave", msg)
	return &APIError{Status: 409, Code: ae.Code, Message: msg}
}

// SeatLeaveOpts 离席：有 TeamID 走 unbind（服务端顺带 release）；否则仅 role release。
type SeatLeaveOpts struct {
	TeamID string
}

// SeatLeave 离席：unbind 或 role release。
func (c *Client) SeatLeave(roleSlug string, opts SeatLeaveOpts) (map[string]any, error) {
	roleSlug = strings.TrimSpace(roleSlug)
	if roleSlug == "" {
		return nil, fmt.Errorf("用法: ymbot seat leave --role <slug> [--team <uuid>]")
	}
	out := map[string]any{"role": roleSlug, "released": true}
	if teamID := strings.TrimSpace(opts.TeamID); teamID != "" {
		rosterRaw, err := c.TeamRosterGet(teamID)
		if err != nil {
			return nil, err
		}
		seat, err := SeatBySlug(rosterRaw, roleSlug)
		if err != nil {
			return nil, err
		}
		raw, err := c.TeamSeatUnbind(teamID, strField(seat, "id"))
		if err != nil {
			return nil, err
		}
		out["via"] = "unbind"
		out["seat"] = projectSeatSlim(raw, roleSlug)
		return out, nil
	}
	if _, err := c.RoleReleaseBySlug(roleSlug); err != nil {
		return nil, err
	}
	out["via"] = "role_release"
	return out, nil
}

func projectSeatSlim(raw json.RawMessage, slug string, playbookRaw ...string) map[string]any {
	var top map[string]any
	_ = json.Unmarshal(raw, &top)
	out := map[string]any{"slug": slug, "ok": true}
	seat, _ := top["seat"].(map[string]any)
	if seat == nil {
		seat = top
	}
	if v := strField(seat, "bound_at"); v != "" {
		out["bound_at"] = shortTime(v)
	}
	if v := strField(seat, "bound_display_name"); v != "" {
		out["holder"] = v
	}
	if claim, ok := seat["claim"].(map[string]any); ok && claim != nil {
		if v := strField(claim, "claimed_at"); v != "" {
			out["claimed_at"] = shortTime(v)
		}
	}
	pref := strField(seat, "playbook_ref")
	if pref == "" && len(playbookRaw) > 0 {
		pref = playbookRaw[0]
	}
	AttachPlaybookRef(out, pref)
	return out
}

func projectClaimSlim(raw json.RawMessage, slug string) map[string]any {
	var top map[string]any
	_ = json.Unmarshal(raw, &top)
	out := map[string]any{"slug": slug, "ok": true}
	if top == nil {
		return out
	}
	claim, _ := top["claim"].(map[string]any)
	if claim == nil {
		claim = top
	}
	if v := strField(claim, "claimed_at"); v != "" {
		out["claimed_at"] = shortTime(v)
	}
	return out
}

// FormatSeatEnterText 进场瘦文本（无默认 UUID）。
func FormatSeatEnterText(v map[string]any) string {
	var b strings.Builder
	fmt.Fprintf(&b, "enter role=%s  %s\n", dash(strField(v, "role")), dash(strField(v, "step")))
	if who, ok := v["whoami"].(map[string]any); ok {
		fmt.Fprintf(&b, "whoami %s", FormatWhoAmIText(who))
	}
	if hand, ok := v["handoff"]; ok {
		fmt.Fprintf(&b, "handoff %s", FormatHandoffText(hand))
	}
	if bind, ok := v["bind"].(map[string]any); ok {
		fmt.Fprintf(&b, "bind ok=%v  at=%s  holder=%s\n", bind["ok"], dash(strField(bind, "bound_at")), dash(strField(bind, "holder")))
	}
	if claim, ok := v["claim"].(map[string]any); ok {
		fmt.Fprintf(&b, "claim ok=%v  at=%s  via=%s\n", claim["ok"], dash(strField(claim, "claimed_at")), dash(strField(claim, "via")))
	}
	if p := strField(v, "playbook_ref"); p != "" {
		fmt.Fprintf(&b, "手册 playbook_ref=%s\n", p)
	}
	if w := strField(v, "playbook_warn"); w != "" {
		fmt.Fprintf(&b, "%s\n", w)
	}
	if hint := strField(v, "memory_key_hint"); hint != "" {
		fmt.Fprintf(&b, "memory_key=%s（席位隔离）\n", hint)
	}
	return b.String()
}

// asObjectListPrefer 按候选 wrapKey 取数组；避免 asObjectList 在缺 key 时把整包当成一行。
func asObjectListPrefer(raw json.RawMessage, keys ...string) ([]map[string]any, map[string]any, error) {
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
		for _, k := range keys {
			if arr, ok := t[k].([]any); ok {
				out := make([]map[string]any, 0, len(arr))
				for _, it := range arr {
					if m, ok := it.(map[string]any); ok {
						out = append(out, m)
					}
				}
				return out, t, nil
			}
		}
		return nil, t, nil
	default:
		return nil, nil, fmt.Errorf("unexpected JSON shape")
	}
}

// ProjectTeamRoster 班底瘦投影：slug / name / claim 态；默认无 UUID。
func ProjectTeamRoster(raw json.RawMessage, opts ViewOpts) ([]map[string]any, error) {
	if opts.Full {
		var v any
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil, err
		}
		if m, ok := v.(map[string]any); ok {
			if arr, ok := m["seats"].([]any); ok {
				out := make([]map[string]any, 0, len(arr))
				for _, it := range arr {
					if x, ok := it.(map[string]any); ok {
						out = append(out, x)
					}
				}
				return out, nil
			}
			if arr, ok := m["roster"].([]any); ok {
				out := make([]map[string]any, 0, len(arr))
				for _, it := range arr {
					if x, ok := it.(map[string]any); ok {
						out = append(out, x)
					}
				}
				return out, nil
			}
		}
	}
	// Go 契约 key=roster；兼容 seats / 裸数组
	seats, top, err := asObjectListPrefer(raw, "roster", "seats", "bot_roles")
	if err != nil {
		return nil, err
	}
	_ = top
	out := make([]map[string]any, 0, len(seats))
	for _, m := range seats {
		if opts.Full {
			out = append(out, m)
			continue
		}
		slug := strField(m, "role_slug", "slug")
		name := strField(m, "display_name", "name")
		item := map[string]any{
			"slug": slug,
			"name": dash(name),
		}
		claimState := "empty"
		holder := ""
		if claim, ok := m["claim"].(map[string]any); ok && claim != nil {
			if strField(claim, "released_at") == "" && strField(claim, "member_id") != "" {
				claimState = "claimed"
				holder = strField(m, "bound_display_name")
				if holder == "" {
					holder = strField(claim, "member_display", "display_name")
				}
				if holder == "" {
					holder = strField(claim, "member_id")
					if !opts.IDs && len(holder) >= 32 {
						holder = "claimed"
					}
				}
			}
		} else if bound := strField(m, "bound_member_id", "member_id"); bound != "" {
			claimState = "bound"
			holder = strField(m, "bound_display_name", "bound_member_display", "member_display")
			if holder == "" {
				if opts.IDs {
					holder = bound
				} else {
					holder = "bound"
				}
			}
		}
		item["claim"] = claimState
		if holder != "" {
			item["holder"] = holder
		}
		AttachPlaybookRef(item, strField(m, "playbook_ref"))
		if opts.IDs {
			if id := strField(m, "id"); id != "" {
				item["id"] = id
			}
			if brid := strField(m, "bot_role_id", "role_id"); brid != "" {
				item["bot_role_id"] = brid
			}
		}
		out = append(out, item)
	}
	return out, nil
}

// FormatTeamRosterText 班底表（YM 风格瘦输出，无默认 UUID）。
func FormatTeamRosterText(items []map[string]any) string {
	if len(items) == 0 {
		return "(空班底)\n"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "seats=%d\n", len(items))
	for _, it := range items {
		fmt.Fprintf(&b, "%s\t%s\t%s",
			dash(strField(it, "slug")),
			dash(strField(it, "name")),
			dash(strField(it, "claim")),
		)
		if h := strField(it, "holder"); h != "" {
			fmt.Fprintf(&b, "\t%s", h)
		}
		if p := strField(it, "playbook_ref"); p != "" {
			fmt.Fprintf(&b, "\t%s", p)
		} else if w := strField(it, "playbook_warn"); w != "" {
			fmt.Fprintf(&b, "\t%s", w)
		}
		if id := strField(it, "id"); id != "" {
			fmt.Fprintf(&b, "\tid=%s", id)
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// ProjectTeamExpand 扩工结果瘦投影（对齐 Go ExpandResult：assigned/skipped）。
func ProjectTeamExpand(raw json.RawMessage, opts ViewOpts) (map[string]any, error) {
	var top map[string]any
	if err := json.Unmarshal(raw, &top); err != nil {
		return nil, err
	}
	if opts.Full {
		return top, nil
	}
	out := map[string]any{}
	if v, ok := top["dry_run"]; ok {
		out["dry_run"] = v
	}
	if n := strField(top, "note"); n != "" {
		out["note"] = trunc(n, 120)
	}
	var assigned []map[string]any
	if arr, ok := top["assigned"].([]any); ok {
		for _, it := range arr {
			m, ok := it.(map[string]any)
			if !ok {
				continue
			}
			item := map[string]any{
				"identifier": dash(strField(m, "identifier")),
			}
			if opts.IDs {
				if id := strField(m, "issue_id", "id"); id != "" {
					item["id"] = id
				}
				if sid := strField(m, "seat_id"); sid != "" {
					item["seat_id"] = sid
				}
			}
			assigned = append(assigned, item)
		}
	}
	out["assigned"] = assigned
	var skipped []map[string]any
	if arr, ok := top["skipped"].([]any); ok {
		for _, it := range arr {
			m, ok := it.(map[string]any)
			if !ok {
				continue
			}
			item := map[string]any{"reason": trunc(strField(m, "reason"), 80)}
			if opts.IDs {
				if id := strField(m, "issue_id"); id != "" {
					item["id"] = id
				}
			}
			skipped = append(skipped, item)
		}
	}
	// dry-run suggestions → skipped 摘要
	if len(skipped) == 0 {
		if arr, ok := top["suggestions"].([]any); ok {
			for _, it := range arr {
				m, ok := it.(map[string]any)
				if !ok {
					continue
				}
				skipped = append(skipped, map[string]any{
					"reason": trunc(strField(m, "rule", "action", "blocked"), 80),
				})
			}
		}
	}
	out["skipped"] = skipped
	out["assigned_n"] = len(assigned)
	out["skipped_n"] = len(skipped)
	return out, nil
}

func FormatTeamExpandText(v map[string]any) string {
	var b strings.Builder
	if dry, ok := v["dry_run"].(bool); ok && dry {
		b.WriteString("expand dry-run\n")
	} else {
		b.WriteString("expand\n")
	}
	if n := strField(v, "note"); n != "" {
		fmt.Fprintf(&b, "%s\n", n)
	}
	fmt.Fprintf(&b, "assigned=%v  skipped=%v\n", v["assigned_n"], v["skipped_n"])
	if arr, ok := v["assigned"].([]map[string]any); ok {
		for _, it := range arr {
			fmt.Fprintf(&b, "  + %s\n", dash(strField(it, "identifier")))
		}
	}
	if arr, ok := v["skipped"].([]map[string]any); ok {
		for _, it := range arr {
			fmt.Fprintf(&b, "  - %s\n", dash(strField(it, "reason")))
		}
	}
	return b.String()
}

// GuardExpandNoStealInProgress 通道闸：扩工不得把他人 in-progress Issue 改派给自己。
// status 为 in_progress/doing 且 assignee 非空且非 me → 拒绝。
func GuardExpandNoStealInProgress(status, assigneeID, myMemberID string) error {
	st := strings.ToLower(strings.TrimSpace(status))
	if st != "in_progress" && st != "doing" && st != "in-progress" {
		return nil
	}
	assigneeID = strings.TrimSpace(assigneeID)
	if assigneeID == "" {
		return nil
	}
	if myMemberID != "" && assigneeID == myMemberID {
		return nil
	}
	return &APIError{
		Status:  409,
		Code:    "conflict",
		Message: "禁止抢走他人进行中的 Issue（in-progress）；仅空闲资源可接手",
	}
}

// ResolveTeamExpandIssueRef 可选：把 YM-n 解析成 id 供 expand body。
func (c *Client) ResolveTeamExpandIssueRef(ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", nil
	}
	return c.ResolveIssueRef(ref)
}

// TeamPull = ensure roster（幂等拉班底）。
func (c *Client) TeamPull(teamID string) (json.RawMessage, error) {
	return c.TeamRosterEnsure(teamID, nil)
}

// ---- team join（空手上场封装：redeem → pull → enter → handoff get）----

// TeamInviteRedeemResult G3 兑换结果；404 时通道提示等服务端。
type TeamInviteRedeemResult struct {
	WorkspaceID     string
	TeamID          string
	AllowExpandSeat bool
	RoleHint        string
	TemplateKey     string
	SeatID          string
	MemberID        string
	RoleSlug        string
	Token           string // bootstrap 当次明文；续队空
	TokenIssued     bool
	Raw             json.RawMessage
}

// TeamJoinOpts 通道 join 选项。
type TeamJoinOpts struct {
	Code    string // ymt_…
	Role    string // 可选目标 slug
	NewSeat string // 扩编时新席 slug（优先于 role_hint / seat-N）
	Force   bool   // 仅 stopped|left 且非活跃时可清脏绑；活跃绝不 force
}

// TeamSeatCreateOpts G2 加新席。
type TeamSeatCreateOpts struct {
	RoleSlug    string
	DisplayName string
	PlaybookRef string
}

// errWaitG2G3 把 404 收成等人话。
func errWaitG2G3(ae *APIError, which string) error {
	msg := strings.TrimSpace(ae.Message)
	switch which {
	case "redeem":
		if msg == "" || strings.EqualFold(msg, "not found") || strings.Contains(msg, "404") {
			msg = "入场码兑换未就绪，等服务端 G3（POST /api/v1/team-invites/redeem）"
		} else if !strings.Contains(msg, "G3") && !strings.Contains(msg, "等服务端") {
			msg = msg + "；等服务端 G3（team-invites/redeem）"
		}
	case "create_seat":
		if msg == "" || strings.EqualFold(msg, "not found") || strings.Contains(msg, "404") {
			msg = "加新席未就绪，等服务端 G2（POST …/teams/{id}/seats）"
		} else if !strings.Contains(msg, "G2") && !strings.Contains(msg, "等服务端") {
			msg = msg + "；等服务端 G2（POST …/seats）"
		}
	default:
		if msg == "" {
			msg = "相关 API 未就绪，等服务端 G2/G3"
		}
	}
	return &APIError{Status: 404, Code: ae.Code, Message: msg}
}

// TeamInviteRedeem POST /api/v1/team-invites/redeem {code,role?}。
// 无 Bearer → 服务端 bootstrap 发 ymk_；有钥续队不重发。成功后写 WorkspaceID；
// 若响应含 token 则写入 c.Cfg.APIKey（空手得钥），含 member_id 则写入 MemberID。
func (c *Client) TeamInviteRedeem(code string, role string) (*TeamInviteRedeemResult, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, fmt.Errorf("需要 --code <ymt_…> 团队入场码")
	}
	body := map[string]any{"code": code}
	if r := strings.TrimSpace(role); r != "" {
		body["role"] = r
	}
	raw, err := c.Do(http.MethodPost, "/api/v1/team-invites/redeem", nil, body, true)
	if err != nil {
		if ae, ok := err.(*APIError); ok && ae.Status == 404 {
			return nil, errWaitG2G3(ae, "redeem")
		}
		return nil, err
	}
	var top map[string]any
	if err := json.Unmarshal(raw, &top); err != nil {
		return nil, err
	}
	out := &TeamInviteRedeemResult{
		WorkspaceID: strField(top, "workspace_id"),
		TeamID:      strField(top, "team_id"),
		RoleHint:    strField(top, "role_hint", "role_slug", "slug"),
		TemplateKey: strField(top, "template_key", "template_ref"),
		SeatID:      strField(top, "seat_id"),
		MemberID:    strField(top, "member_id"),
		RoleSlug:    strField(top, "role_slug"),
		Token:       strField(top, "token"),
		Raw:         raw,
	}
	if v, ok := top["allow_expand_seat"].(bool); ok {
		out.AllowExpandSeat = v
	}
	if v, ok := top["token_issued"].(bool); ok {
		out.TokenIssued = v
	} else if out.Token != "" {
		out.TokenIssued = true
	}
	if out.WorkspaceID == "" || out.TeamID == "" {
		return nil, &APIError{Status: 502, Code: "bad_gateway", Message: "兑换结果缺少 workspace_id 或 team_id"}
	}
	c.Cfg.WorkspaceID = out.WorkspaceID
	if tok := strings.TrimSpace(out.Token); tok != "" {
		c.Cfg.APIKey = tok
	}
	if mid := strings.TrimSpace(out.MemberID); mid != "" {
		c.Cfg.MemberID = mid
	}
	return out, nil
}

// TeamSeatCreate POST …/teams/{teamID}/seats（G2；幂等按 slug）。
func (c *Client) TeamSeatCreate(teamID string, opts TeamSeatCreateOpts) (json.RawMessage, error) {
	teamID = strings.TrimSpace(teamID)
	slug := strings.TrimSpace(opts.RoleSlug)
	if teamID == "" || slug == "" {
		return nil, fmt.Errorf("需要 team uuid 与 role_slug")
	}
	body := map[string]any{
		"role_slug": slug,
		"slug":      slug, // 兼容契约 slug|role_slug
	}
	if n := strings.TrimSpace(opts.DisplayName); n != "" {
		body["display_name"] = n
		body["name"] = n
	}
	if p := strings.TrimSpace(opts.PlaybookRef); p != "" {
		body["playbook_ref"] = p
	}
	raw, err := c.Do(http.MethodPost, c.ws("/teams/"+teamID+"/seats"), nil, body, true)
	if err != nil {
		if ae, ok := err.(*APIError); ok && ae.Status == 404 {
			return nil, errWaitG2G3(ae, "create_seat")
		}
		return nil, err
	}
	return raw, nil
}

// ClassifySeatOccupancy 选席推断：empty|self|active；若 JSON 带 status=stopped|left 则返回该字面量。
// 未落库前勿写死 open|active 枚举——有 bound/claim 即活跃，无则空。
func ClassifySeatOccupancy(seat map[string]any, myMemberID string) string {
	if seat == nil {
		return "empty"
	}
	st := strings.ToLower(strings.TrimSpace(strField(seat, "status")))
	if st == "stopped" || st == "left" {
		return st
	}
	myMemberID = strings.TrimSpace(myMemberID)
	if claim, ok := seat["claim"].(map[string]any); ok && claim != nil {
		mid := strField(claim, "member_id")
		if mid != "" && strField(claim, "released_at") == "" {
			if myMemberID != "" && mid == myMemberID {
				return "self"
			}
			return "active"
		}
	}
	bound := strField(seat, "bound_member_id", "member_id")
	if bound != "" {
		if myMemberID != "" && bound == myMemberID {
			return "self"
		}
		return "active"
	}
	return "empty"
}

// SeatIsActiveOccupancy 活跃占用（有人 claim/bound 且非 self、非 stopped|left）。
func SeatIsActiveOccupancy(kind string) bool {
	return kind == "active"
}

func nextAutoSeatSlug(seats []map[string]any) string {
	used := map[string]bool{}
	maxN := 0
	for _, s := range seats {
		slug := strField(s, "role_slug", "slug")
		used[slug] = true
		if strings.HasPrefix(slug, "seat-") {
			var n int
			if _, err := fmt.Sscanf(slug, "seat-%d", &n); err == nil && n > maxN {
				maxN = n
			}
		}
	}
	for i := maxN + 1; i < maxN+1000; i++ {
		cand := fmt.Sprintf("seat-%d", i)
		if !used[cand] {
			return cand
		}
	}
	return fmt.Sprintf("seat-%d", len(seats)+1)
}

func listRosterSeats(rosterRaw json.RawMessage) ([]map[string]any, error) {
	seats, _, err := asObjectListPrefer(rosterRaw, "roster", "seats")
	return seats, err
}

func findEmptySeat(seats []map[string]any, myMemberID string) map[string]any {
	for _, s := range seats {
		if ClassifySeatOccupancy(s, myMemberID) == "empty" {
			return s
		}
	}
	return nil
}

// pickJoinSeat 按 --role / 空席 / 扩编规则选席；forceActive 禁止。
// 返回 seat map（可无 id）、slug、是否已绑本 member（跳过 enter）、是否需 CreateSeat。
func pickJoinSeat(seats []map[string]any, myMemberID string, opts TeamJoinOpts, allowExpand bool, roleHint string) (seat map[string]any, slug string, alreadySelf bool, needCreate bool, err error) {
	role := strings.TrimSpace(opts.Role)
	if role != "" {
		var found map[string]any
		for _, s := range seats {
			if strField(s, "role_slug", "slug") == role {
				found = s
				break
			}
		}
		if found == nil {
			if allowExpand {
				return nil, role, false, true, nil
			}
			return nil, "", false, false, &APIError{
				Status:  404,
				Code:    "not_found",
				Message: fmt.Sprintf("班底无席位 slug=%s，且入场码未允许加新席", role),
			}
		}
		kind := ClassifySeatOccupancy(found, myMemberID)
		switch kind {
		case "self":
			return found, role, true, false, nil
		case "empty":
			return found, role, false, false, nil
		case "stopped", "left":
			// 可进；活跃顶替禁止——stopped|left 非活跃
			return found, role, false, false, nil
		case "active":
			holder := strField(found, "bound_display_name")
			msg := fmt.Sprintf("席位 %s 占用中，禁止顶替", role)
			if holder != "" {
				msg = fmt.Sprintf("席位 %s 占用中（%s），禁止顶替；请对方 seat leave 或等停工后再进", role, holder)
			} else {
				msg = msg + "；请对方 seat leave 或等停工后再进"
			}
			return nil, "", false, false, &APIError{Status: 409, Code: "conflict", Message: msg}
		default:
			return found, role, false, false, nil
		}
	}

	// 无 --role：优先空席
	if empty := findEmptySeat(seats, myMemberID); empty != nil {
		return empty, strField(empty, "role_slug", "slug"), false, false, nil
	}
	// 已绑本 member 任一席 → 幂等跳过
	for _, s := range seats {
		if ClassifySeatOccupancy(s, myMemberID) == "self" {
			return s, strField(s, "role_slug", "slug"), true, false, nil
		}
	}
	if !allowExpand {
		return nil, "", false, false, &APIError{
			Status:  409,
			Code:    "conflict",
			Message: "无空席且入场码未允许加新席（allow_expand_seat=false）；请总指挥扩编或指定 --role 停工席",
		}
	}
	slug = strings.TrimSpace(opts.NewSeat)
	if slug == "" {
		slug = strings.TrimSpace(roleHint)
	}
	if slug == "" {
		slug = nextAutoSeatSlug(seats)
	}
	return nil, slug, false, true, nil
}

// TeamJoin 空手上场：redeem →（写 workspace）→ team pull → 选席/加席 → seat enter（禁活跃 force）→ handoff get。
func (c *Client) TeamJoin(opts TeamJoinOpts) (map[string]any, error) {
	redeem, err := c.TeamInviteRedeem(opts.Code, opts.Role)
	if err != nil {
		return nil, err
	}
	teamID := redeem.TeamID

	rosterRaw, err := c.TeamPull(teamID)
	if err != nil {
		return nil, err
	}
	seats, err := listRosterSeats(rosterRaw)
	if err != nil {
		return nil, err
	}

	myMemberID := strings.TrimSpace(c.Cfg.MemberID)
	if myMemberID == "" && strings.TrimSpace(redeem.MemberID) != "" {
		myMemberID = strings.TrimSpace(redeem.MemberID)
		c.Cfg.MemberID = myMemberID
	}
	if myMemberID == "" {
		if mid, err := c.ResolveMember(); err == nil {
			myMemberID = mid
			c.Cfg.MemberID = mid
		}
	}

	seat, slug, alreadySelf, needCreate, err := pickJoinSeat(seats, myMemberID, opts, redeem.AllowExpandSeat, redeem.RoleHint)
	if err != nil {
		return nil, err
	}

	created := false
	if needCreate {
		if _, err := c.TeamSeatCreate(teamID, TeamSeatCreateOpts{RoleSlug: slug}); err != nil {
			return nil, err
		}
		created = true
		rosterRaw, err = c.TeamPull(teamID)
		if err != nil {
			return nil, err
		}
		seats, err = listRosterSeats(rosterRaw)
		if err != nil {
			return nil, err
		}
		seat, err = SeatBySlug(rosterRaw, slug)
		if err != nil {
			return nil, err
		}
	}

	out := map[string]any{
		"step":              "redeem→team pull→seat enter→handoff get",
		"team_id":           teamID,
		"workspace_id":      redeem.WorkspaceID,
		"role":              slug,
		"allow_expand_seat": redeem.AllowExpandSeat,
		"seat_created":      created,
		"skipped_enter":     alreadySelf,
	}
	if mid := strings.TrimSpace(redeem.MemberID); mid != "" {
		out["member_id"] = mid
	}
	if redeem.TokenIssued || strings.TrimSpace(redeem.Token) != "" {
		out["token_issued"] = true
		if tok := strings.TrimSpace(redeem.Token); tok != "" {
			out["token"] = tok
		}
	}
	if redeem.RoleHint != "" {
		out["role_hint"] = redeem.RoleHint
	}
	if redeem.TemplateKey != "" {
		out["template_key"] = redeem.TemplateKey
	}

	enterForce := false
	if !alreadySelf {
		kind := ClassifySeatOccupancy(seat, myMemberID)
		if SeatIsActiveOccupancy(kind) {
			// 双保险：选席已拒，enter 前再拒且绝不 force
			return nil, &APIError{
				Status:  409,
				Code:    "conflict",
				Message: fmt.Sprintf("席位 %s 占用中，禁止顶替（join 不对活跃席传 force）", slug),
			}
		}
		// --force 仅 stopped|left；活跃路径上面已 return
		if opts.Force && (kind == "stopped" || kind == "left") {
			enterForce = true
		}
		// join 默认不 force；empty/self 不需要
		enterView, err := c.SeatEnter(slug, SeatEnterOpts{TeamID: teamID, Force: enterForce})
		if err != nil {
			return nil, err
		}
		out["enter"] = enterView
		if hand, ok := enterView["handoff"]; ok {
			out["handoff"] = hand
		}
	} else {
		handRaw, err := c.HandoffGet()
		if err != nil {
			return nil, err
		}
		hand, err := c.ProjectHandoff(handRaw, ViewOpts{})
		if err != nil {
			return nil, err
		}
		out["handoff"] = hand
		out["note"] = "已绑本席，跳过 enter，直接 handoff get"
	}

	// 权威续命再拉一次（与封装契约一致；幂等）
	if _, has := out["handoff"]; !has || alreadySelf {
		// already handled
	} else {
		handRaw, err := c.HandoffGet()
		if err == nil {
			if hand, err2 := c.ProjectHandoff(handRaw, ViewOpts{}); err2 == nil {
				out["handoff"] = hand
			}
		}
	}
	out["force_used"] = enterForce
	return out, nil
}

// FormatTeamJoinText 瘦文本（无默认 UUID）。
func FormatTeamJoinText(v map[string]any) string {
	var b strings.Builder
	fmt.Fprintf(&b, "join role=%s  %s\n", dash(strField(v, "role")), dash(strField(v, "step")))
	if mid := strField(v, "member_id"); mid != "" {
		fmt.Fprintf(&b, "member_id=%s\n", mid)
	}
	if tok := strField(v, "token"); tok != "" {
		fmt.Fprintf(&b, "token=%s\n", tok)
		b.WriteString("token_issued=true（仅此一次，私存；禁止贴进聊天）\n")
	}
	if v["seat_created"] == true {
		b.WriteString("seat_created=true\n")
	}
	if v["skipped_enter"] == true {
		b.WriteString("skipped_enter=true（已绑本席）\n")
	}
	if n := strField(v, "note"); n != "" {
		fmt.Fprintf(&b, "%s\n", n)
	}
	if enter, ok := v["enter"].(map[string]any); ok {
		fmt.Fprintf(&b, "%s", FormatSeatEnterText(enter))
	} else if hand, ok := v["handoff"]; ok {
		fmt.Fprintf(&b, "handoff %s", FormatHandoffText(hand))
	}
	return b.String()
}

// RequireAPIKey 仅校验身份钥匙（join 的 workspace 由 redeem 写入）。
func (c *Client) RequireAPIKey() error {
	if c.Cfg.APIKey == "" {
		return fmt.Errorf("缺少 YMBOT_API_KEY（仅接受 ymk_… API Key，不主推密码登录）")
	}
	return nil
}
