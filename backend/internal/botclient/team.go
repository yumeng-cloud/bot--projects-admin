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
// body: force?；Bot 侧 member 由服务端取 actor。二绑无 force → 409。
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
	msg = fmt.Sprintf("%s；禁止瞎抢，请 seat leave 或显式接管", msg)
	return &APIError{Status: 409, Code: ae.Code, Message: msg}
}

// SeatEnterOpts 进场：有 TeamID 走 bind（服务端顺带 claim）；否则仅 role claim。
type SeatEnterOpts struct {
	TeamID string
	Force  bool
}

// SeatEnter 迁绑进场：whoami → handoff get → bind|role claim。
// 冲突 409 不吞；有 team 时 bind 含单活接管（须 --force）。
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
		seatID := strField(seat, "id")
		bindRaw, err := c.TeamSeatBind(teamID, seatID, opts.Force)
		if err != nil {
			if ae, ok := err.(*APIError); ok && ae.Status == 409 {
				return nil, c.enrichBindConflict(roleSlug, seat, ae)
			}
			return nil, err
		}
		out["bind"] = projectSeatSlim(bindRaw, roleSlug)
		out["claim"] = map[string]any{"ok": true, "via": "bind"}
		return out, nil
	}

	out["step"] = "whoami→handoff get→role claim"
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
	msg = fmt.Sprintf("%s；禁止瞎抢，请 seat leave 或显式 --force 接管", msg)
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

func projectSeatSlim(raw json.RawMessage, slug string) map[string]any {
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
		if pref := strField(m, "playbook_ref"); pref != "" {
			item["playbook"] = trunc(pref, 48)
		}
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
		if p := strField(it, "playbook"); p != "" {
			fmt.Fprintf(&b, "\t%s", p)
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
