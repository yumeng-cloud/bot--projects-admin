package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/botclient"
)

func runCLI(args []string) error {
	c := botclient.New(botclient.LoadConfig())
	// team join：workspace 由入场码 redeem 写入，仅需 API Key
	if args[0] == "team" && len(args) > 1 && args[1] == "join" {
		if err := c.RequireAPIKey(); err != nil {
			return err
		}
	} else if err := c.Require(); err != nil {
		return err
	}
	switch args[0] {
	case "whoami":
		return cmdWhoAmI(c, args[1:])
	case "issue":
		return cmdIssue(c, args[1:])
	case "comment":
		return cmdComment(c, args[1:])
	case "memory":
		return cmdMemory(c, args[1:])
	case "handoff":
		return cmdHandoff(c, args[1:])
	case "team":
		return cmdTeam(c, args[1:])
	case "seat":
		return cmdSeat(c, args[1:])
	case "migrate":
		return cmdMigrate(c, args[1:])
	case "patrol":
		return cmdPatrol(c, args[1:])
	case "pin":
		return cmdPin(c, args[1:])
	default:
		return fmt.Errorf("未知命令: %s（ymbot --help）", args[0])
	}
}

func applyIDKey(c *botclient.Client, key string) {
	if key != "" {
		c.IdempotencyKey = key
	}
}

func autosaveOn(flag bool) bool {
	if flag {
		return true
	}
	v := os.Getenv("YMBOT_AUTOSAVE")
	return v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes")
}

func remindAutosave() {
	fmt.Fprintln(os.Stderr, "已写服务端（进度在服务端；主航线 team pull → seat enter → handoff get；勿攒到最后 export）")
}

func cmdWhoAmI(c *botclient.Client, args []string) error {
	fs := newFS("whoami")
	full, ids, asJSON := viewFlags(fs)
	if err := fs.Parse(reorderFlags(args)); err != nil {
		return err
	}
	opts := optsFrom(*full, *ids, *asJSON)
	raw, err := c.WhoAmI()
	if err != nil {
		return err
	}
	if opts.Full {
		printRawFull(raw)
		return nil
	}
	v := botclient.ProjectWhoAmI(raw, opts)
	printView(opts, botclient.FormatWhoAmIText(v), v)
	return nil
}

func cmdIssue(c *botclient.Client, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("用法: ymbot issue list|get|create|update|claim …")
	}
	switch args[0] {
	case "list":
		fs := newFS("issue list")
		status := fs.String("status", "", "")
		assignee := fs.String("assignee", "", "")
		project := fs.String("project", "", "")
		q := fs.String("q", "", "")
		cycle := fs.String("cycle", "", "")
		label := fs.String("label", "", "")
		leafOnly := fs.Bool("leaf-only", false, "仅叶子（query leaf_only=1）")
		unassigned := fs.Bool("unassigned", false, "未分配")
		full, ids, asJSON := viewFlags(fs)
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		opts := optsFrom(*full, *ids, *asJSON)
		qv := url.Values{}
		set := func(k, v string) {
			if v != "" {
				qv.Set(k, v)
			}
		}
		set("status", *status)
		set("assignee_id", *assignee)
		set("project_id", *project)
		set("q", *q)
		set("cycle_id", *cycle)
		set("label_id", *label)
		if *leafOnly {
			qv.Set("leaf_only", "1")
		}
		if *unassigned {
			qv.Set("unassigned", "1")
		}
		raw, err := c.IssueList(qv)
		if err != nil {
			return err
		}
		if opts.Full {
			printRawFull(raw)
			return nil
		}
		items, err := c.ProjectIssues(raw, opts)
		if err != nil {
			return err
		}
		printView(opts, botclient.FormatIssueListText(items), items)
		return nil
	case "get":
		fs := newFS("issue get")
		full, ids, asJSON := viewFlags(fs)
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		ref := fs.Arg(0)
		if ref == "" {
			return fmt.Errorf("用法: ymbot issue get <YM-12|id> [--full|--id|--json]")
		}
		opts := optsDetail(*full, *ids, *asJSON)
		id, err := c.ResolveIssueRef(ref)
		if err != nil {
			return err
		}
		raw, err := c.IssueGet(id)
		if err != nil {
			return err
		}
		if opts.Full {
			printRawFull(raw)
			return nil
		}
		items, err := c.ProjectIssues(raw, opts)
		if err != nil {
			return err
		}
		var one map[string]any
		if len(items) > 0 {
			one = items[0]
		} else {
			one = map[string]any{}
		}
		printView(opts, botclient.FormatIssueGetText(one), one)
		return nil
	case "create":
		fs := newFS("issue create")
		project := fs.String("project", "", "")
		title := fs.String("title", "", "")
		desc := fs.String("description", "", "")
		status := fs.String("status", "", "")
		priority := fs.String("priority", "", "")
		assignee := fs.String("assignee", "", "")
		parent := fs.String("parent", "", "父 Issue（YM-n 或 id；Bot 通道必填）")
		idKey := fs.String("idempotency-key", "", "")
		full, ids, asJSON := viewFlags(fs)
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if *project == "" || *title == "" {
			return fmt.Errorf("需要 --project 与 --title")
		}
		if *parent == "" {
			return fmt.Errorf("禁止创建无父游离 Issue，请在已有父节点下拆叶子或由人类编排（须 --parent）")
		}
		parentID, err := c.ResolveIssueRef(*parent)
		if err != nil {
			return err
		}
		applyIDKey(c, *idKey)
		body := map[string]any{"project_id": *project, "title": *title, "parent_id": parentID}
		if *desc != "" {
			body["description"] = *desc
		}
		if *status != "" {
			body["status"] = *status
		}
		if *priority != "" {
			body["priority"] = *priority
		}
		if *assignee != "" {
			a := *assignee
			if a == "me" {
				mid, err := c.ResolveMember()
				if err != nil {
					return err
				}
				a = mid
			}
			body["assignee_id"] = a
		}
		raw, err := c.IssueCreate(body)
		if err != nil {
			return err
		}
		return printIssueResult(c, raw, optsDetail(*full, *ids, *asJSON))
	case "update":
		fs := newFS("issue update")
		version := fs.Int("version", -1, "")
		title := fs.String("title", "", "")
		desc := fs.String("description", "", "")
		status := fs.String("status", "", "")
		priority := fs.String("priority", "", "")
		assignee := fs.String("assignee", "", "")
		idKey := fs.String("idempotency-key", "", "")
		full, ids, asJSON := viewFlags(fs)
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		ref := fs.Arg(0)
		if ref == "" {
			return fmt.Errorf("用法: ymbot issue update <YM-12|id> --version N …")
		}
		if *version < 0 {
			return fmt.Errorf("需要 --version（乐观锁；ymbot issue get YM-n 默认可见 version）")
		}
		id, err := c.ResolveIssueRef(ref)
		if err != nil {
			return err
		}
		applyIDKey(c, *idKey)
		body := map[string]any{"version": *version}
		if *title != "" {
			body["title"] = *title
		}
		if *desc != "" {
			body["description"] = *desc
		}
		if *status != "" {
			body["status"] = *status
		}
		if *priority != "" {
			body["priority"] = *priority
		}
		if *assignee != "" {
			a := *assignee
			if a == "me" {
				mid, err := c.ResolveMember()
				if err != nil {
					return err
				}
				a = mid
			}
			body["assignee_id"] = a
		}
		raw, err := c.IssueUpdate(id, body)
		if err != nil {
			if ae, ok := err.(*botclient.APIError); ok && ae.Status == 409 {
				return fmt.Errorf("%w；请重新 issue get 取当前 version 再试（默认输出含 version，无需 --full）", err)
			}
			return err
		}
		return printIssueResult(c, raw, optsDetail(*full, *ids, *asJSON))
	case "claim":
		fs := newFS("issue claim")
		version := fs.Int("version", -1, "乐观锁；缺省先 GET 取当前 version")
		idKey := fs.String("idempotency-key", "", "")
		asave := fs.Bool("autosave", false, "成功后 stderr 提醒已写服务端（或 YMBOT_AUTOSAVE=1）")
		full, ids, asJSON := viewFlags(fs)
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		ref := fs.Arg(0)
		if ref == "" {
			return fmt.Errorf("用法: ymbot issue claim <YM-12|id> [--version N]\n只领叶子；父节点不可领（服务端非叶子 409）")
		}
		id, err := c.ResolveIssueRef(ref)
		if err != nil {
			return err
		}
		applyIDKey(c, *idKey)
		raw, err := c.IssueClaim(id, *version)
		if err != nil {
			return err
		}
		if autosaveOn(*asave) {
			remindAutosave()
		}
		return printIssueResult(c, raw, optsDetail(*full, *ids, *asJSON))
	case "patrol":
		return cmdPatrol(c, args[1:])
	default:
		return fmt.Errorf("未知 issue 子命令: %s（list|get|create|update|claim|patrol）", args[0])
	}
}

func printIssueResult(c *botclient.Client, raw json.RawMessage, opts botclient.ViewOpts) error {
	if opts.Full {
		printRawFull(raw)
		return nil
	}
	items, err := c.ProjectIssues(raw, opts)
	if err != nil {
		return err
	}
	var one map[string]any
	if len(items) > 0 {
		one = items[0]
	} else {
		one = map[string]any{}
	}
	printView(opts, botclient.FormatIssueGetText(one), one)
	return nil
}

func cmdComment(c *botclient.Client, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("用法: ymbot comment list|create …")
	}
	switch args[0] {
	case "list":
		fs := newFS("comment list")
		issue := fs.String("issue", "", "YM-n 或 id")
		full, ids, asJSON := viewFlags(fs)
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if *issue == "" {
			return fmt.Errorf("需要 --issue YM-12")
		}
		opts := optsFrom(*full, *ids, *asJSON)
		id, err := c.ResolveIssueRef(*issue)
		if err != nil {
			return err
		}
		ident := *issue
		if !botclient.LooksLikeIdentifier(*issue) {
			ident = c.CachedIdent(id)
		}
		raw, err := c.CommentList(id)
		if err != nil {
			return err
		}
		if opts.Full {
			printRawFull(raw)
			return nil
		}
		items, err := c.ProjectComments(raw, ident, opts)
		if err != nil {
			return err
		}
		printView(opts, botclient.FormatCommentListText(items), items)
		return nil
	case "create":
		fs := newFS("comment create")
		issue := fs.String("issue", "", "")
		body := fs.String("body", "", "")
		msg := fs.String("m", "", "")
		idKey := fs.String("idempotency-key", "", "")
		asave := fs.Bool("autosave", false, "成功后 stderr 提醒已写服务端（或 YMBOT_AUTOSAVE=1）")
		full, ids, asJSON := viewFlags(fs)
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		text := *body
		if text == "" {
			text = *msg
		}
		if *issue == "" || text == "" {
			return fmt.Errorf("用法: ymbot comment create --issue YM-12 -m|--body TEXT")
		}
		id, err := c.ResolveIssueRef(*issue)
		if err != nil {
			return err
		}
		applyIDKey(c, *idKey)
		raw, err := c.CommentCreate(id, text)
		if err != nil {
			return err
		}
		if autosaveOn(*asave) {
			remindAutosave()
		}
		opts := optsFrom(*full, *ids, *asJSON)
		if opts.Full {
			printRawFull(raw)
			return nil
		}
		ident := *issue
		items, err := c.ProjectComments(raw, ident, opts)
		if err != nil {
			return err
		}
		printView(opts, botclient.FormatCommentListText(items), items)
		return nil
	default:
		return fmt.Errorf("未知 comment 子命令: %s", args[0])
	}
}

func cmdMemory(c *botclient.Client, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("用法: ymbot memory list|set …")
	}
	switch args[0] {
	case "list":
		fs := newFS("memory list")
		q := fs.String("q", "", "")
		mode := fs.String("mode", "", "keyword|semantic|hybrid")
		limit := fs.Int("limit", 0, "")
		full, ids, asJSON := viewFlags(fs)
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		opts := optsFrom(*full, *ids, *asJSON)
		raw, err := c.MemoryList(*q, *mode, *limit)
		if err != nil {
			return err
		}
		if opts.Full {
			printRawFull(raw)
			return nil
		}
		items, err := botclient.ProjectMemories(raw, opts)
		if err != nil {
			return err
		}
		printView(opts, botclient.FormatMemoryListText(items), items)
		return nil
	case "set":
		fs := newFS("memory set")
		key := fs.String("key", "", "")
		value := fs.String("value", "", "")
		tags := fs.String("tags", "", "")
		idKey := fs.String("idempotency-key", "", "")
		asave := fs.Bool("autosave", false, "成功后 stderr 提醒已写服务端（或 YMBOT_AUTOSAVE=1）")
		full, ids, asJSON := viewFlags(fs)
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if *key == "" || *value == "" {
			return fmt.Errorf("需要 --key 与 --value")
		}
		applyIDKey(c, *idKey)
		raw, err := c.MemorySet(*key, *value, *tags)
		if err != nil {
			return err
		}
		if autosaveOn(*asave) {
			remindAutosave()
		}
		opts := optsFrom(*full, *ids, *asJSON)
		if opts.Full {
			printRawFull(raw)
			return nil
		}
		items, err := botclient.ProjectMemories(raw, opts)
		if err != nil {
			return err
		}
		printView(opts, botclient.FormatMemoryListText(items), items)
		return nil
	default:
		return fmt.Errorf("未知 memory 子命令: %s", args[0])
	}
}

func cmdHandoff(c *botclient.Client, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("用法: ymbot handoff get|export|import …\n主航线③：handoff get（无文件；先 team pull → seat enter）\n进度写 memory/todo/comment；export/import 仅备份/救场")
	}
	switch args[0] {
	case "get", "export":
		sub := args[0]
		fs := newFS("handoff " + sub)
		outHelp := "换机备份文件（可含 id）；日常续工不需要"
		if sub == "get" {
			outHelp = "可选写文件；日常续工不需要——默认终端摘要即续命"
		}
		out := fs.String("out", "", outHelp)
		full, ids, asJSON := viewFlags(fs)
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if sub == "export" && *out == "" {
			fmt.Fprintln(os.Stderr, "提示: handoff export 为换机备份可选；日常续工用 handoff get（无文件）。未传 --out 时仅打印服务端摘要。")
		}
		opts := optsFrom(*full, *ids, *asJSON)
		raw, err := c.HandoffGet()
		if err != nil {
			return err
		}
		if *out != "" {
			// 导出文件：保留接近 REST 的包（换机备份用）
			var v any
			_ = json.Unmarshal(raw, &v)
			b, _ := json.MarshalIndent(v, "", "  ")
			if err := os.WriteFile(*out, append(b, '\n'), 0o600); err != nil {
				return err
			}
			fmt.Println("wrote", *out)
			return nil
		}
		if opts.Full {
			printRawFull(raw)
			return nil
		}
		v, err := c.ProjectHandoff(raw, opts)
		if err != nil {
			return err
		}
		printView(opts, botclient.FormatHandoffText(v), v)
		return nil
	case "import":
		fs := newFS("handoff import")
		file := fs.String("file", "", "")
		idKey := fs.String("idempotency-key", "", "")
		full, ids, asJSON := viewFlags(fs)
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if *file == "" {
			return fmt.Errorf("需要 --file path.json 或 --file -")
		}
		applyIDKey(c, *idKey)
		var raw []byte
		var err error
		if *file == "-" {
			raw, err = io.ReadAll(os.Stdin)
		} else {
			raw, err = os.ReadFile(*file)
		}
		if err != nil {
			return err
		}
		var body any
		if err := json.Unmarshal(raw, &body); err != nil {
			return fmt.Errorf("无效 JSON: %w", err)
		}
		res, err := c.HandoffImport(body)
		if err != nil {
			return err
		}
		opts := optsFrom(*full, *ids, *asJSON)
		if opts.Full {
			printRawFull(res)
			return nil
		}
		v := botclient.ProjectHandoffImport(res, opts)
		printView(opts, botclient.MarshalView(v), v)
		return nil
	default:
		return fmt.Errorf("未知 handoff 子命令: %s", args[0])
	}
}

func cmdTeam(c *botclient.Client, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("用法: ymbot team pull|join …\njoin = 空手上场（redeem→pull→enter→handoff get）；主航线仍可写 pull→enter→handoff\n见 bot-team-template/playbook/看板席位迁绑.md")
	}
	switch args[0] {
	case "pull":
		fs := newFS("team pull")
		uuid := fs.String("uuid", "", "团队 UUID")
		idKey := fs.String("idempotency-key", "", "")
		full, ids, asJSON := viewFlags(fs)
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if *uuid == "" {
			return fmt.Errorf("用法: ymbot team pull --uuid <team_uuid>")
		}
		applyIDKey(c, *idKey)
		raw, err := c.TeamPull(*uuid)
		if err != nil {
			return err
		}
		opts := optsFrom(*full, *ids, *asJSON)
		if opts.Full {
			printRawFull(raw)
			return nil
		}
		items, err := botclient.ProjectTeamRoster(raw, opts)
		if err != nil {
			return err
		}
		printView(opts, botclient.FormatTeamRosterText(items), items)
		return nil
	case "join":
		return cmdTeamJoin(c, args[1:])
	default:
		return fmt.Errorf("未知 team 子命令: %s（pull|join）", args[0])
	}
}

func cmdTeamJoin(c *botclient.Client, args []string) error {
	fs := newFS("team join")
	code := fs.String("code", "", "团队入场码 ymt_…")
	role := fs.String("role", "", "可选目标席位 slug；活跃占用 409 禁顶替")
	newSeat := fs.String("new-seat", "", "无空席且允许扩编时的新席 slug（默认 role_hint 或 seat-N）")
	force := fs.Bool("force", false, "仅 stopped|left 清脏绑；活跃席即使传也不顶替")
	idKey := fs.String("idempotency-key", "", "")
	full, ids, asJSON := viewFlags(fs)
	if err := fs.Parse(reorderFlags(args)); err != nil {
		return err
	}
	if strings.TrimSpace(*code) == "" {
		return fmt.Errorf("用法: ymbot team join --code <ymt_…> [--role <slug>] [--new-seat <slug>] [--force]\n空手上场：redeem→team pull→seat enter→handoff get；身份靠 YMBOT_API_KEY；活跃席禁 force")
	}
	applyIDKey(c, *idKey)
	v, err := c.TeamJoin(botclient.TeamJoinOpts{
		Code:    *code,
		Role:    *role,
		NewSeat: *newSeat,
		Force:   *force,
	})
	if err != nil {
		return err
	}
	opts := optsFrom(*full, *ids, *asJSON)
	if opts.Full || opts.JSON {
		printView(opts, botclient.MarshalView(v), v)
		return nil
	}
	printView(opts, botclient.FormatTeamJoinText(v), v)
	return nil
}

func cmdMigrate(c *botclient.Client, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("用法: ymbot migrate enter --role <slug> [--team <uuid>] [--force]\n= seat enter（别名）；主航线：team pull → seat enter → handoff get")
	}
	switch args[0] {
	case "enter":
		return cmdSeatEnter(c, args[1:])
	default:
		return fmt.Errorf("未知 migrate 子命令: %s（enter）", args[0])
	}
}

func cmdSeat(c *botclient.Client, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("用法: ymbot seat enter|leave|expand …\n主航线：team pull → seat enter → handoff get（slug=席位）\nenter：进席 bind|claim；leave：unbind|release；expand：扩工补位")
	}
	switch args[0] {
	case "enter":
		return cmdSeatEnter(c, args[1:])
	case "leave":
		return cmdSeatLeave(c, args[1:])
	case "expand":
		return cmdSeatExpand(c, args[1:])
	default:
		return fmt.Errorf("未知 seat 子命令: %s（enter|leave|expand）", args[0])
	}
}

func cmdSeatEnter(c *botclient.Client, args []string) error {
	fs := newFS("seat enter")
	role := fs.String("role", "", "席位 slug（=模板角色；frontend/pm/qa/backend/wiki…）")
	team := fs.String("team", "", "团队 UUID（有则走 seat bind；否则 role claim）")
	force := fs.Bool("force", false, "仅清无 active claim 的残留脏绑；活跃占用仍 409")
	idKey := fs.String("idempotency-key", "", "")
	full, ids, asJSON := viewFlags(fs)
	if err := fs.Parse(reorderFlags(args)); err != nil {
		return err
	}
	if *role == "" {
		return fmt.Errorf("用法: ymbot seat enter --role <slug> [--team <uuid>] [--force]\n主航线②：slug=席位；有 --team 走 bind；活跃占用 409 禁顶替；--force 仅清脏绑\nmigrate enter = 本命令别名")
	}
	applyIDKey(c, *idKey)
	v, err := c.SeatEnter(*role, botclient.SeatEnterOpts{TeamID: *team, Force: *force})
	if err != nil {
		return err
	}
	opts := optsFrom(*full, *ids, *asJSON)
	if opts.Full || opts.JSON {
		printView(opts, botclient.MarshalView(v), v)
		return nil
	}
	printView(opts, botclient.FormatSeatEnterText(v), v)
	return nil
}

func cmdSeatLeave(c *botclient.Client, args []string) error {
	fs := newFS("seat leave")
	role := fs.String("role", "", "角色 slug")
	team := fs.String("team", "", "团队 UUID（有则 unbind；否则 role release）")
	idKey := fs.String("idempotency-key", "", "")
	full, ids, asJSON := viewFlags(fs)
	if err := fs.Parse(reorderFlags(args)); err != nil {
		return err
	}
	if *role == "" {
		return fmt.Errorf("用法: ymbot seat leave --role <slug> [--team <uuid>]")
	}
	applyIDKey(c, *idKey)
	v, err := c.SeatLeave(*role, botclient.SeatLeaveOpts{TeamID: *team})
	if err != nil {
		return err
	}
	opts := optsFrom(*full, *ids, *asJSON)
	via, _ := v["via"].(string)
	if via == "" {
		via = "-"
	}
	msg := fmt.Sprintf("left role=%s via=%s\n", *role, via)
	printView(opts, msg, v)
	return nil
}

func cmdSeatExpand(c *botclient.Client, args []string) error {
	fs := newFS("seat expand")
	team := fs.String("team", "", "团队 UUID")
	project := fs.String("project", "", "可选项目 UUID，限制未领叶子范围")
	limit := fs.Int("limit", 0, "最多派发条数（服务端默认 50）")
	dry := fs.Bool("dry-run", false, "只列建议不改库（服务端缺 expand 时建议模式）")
	idKey := fs.String("idempotency-key", "", "")
	full, ids, asJSON := viewFlags(fs)
	if err := fs.Parse(reorderFlags(args)); err != nil {
		return err
	}
	if *team == "" {
		return fmt.Errorf("用法: ymbot seat expand --team <uuid> [--project id] [--limit N] [--dry-run]\n仅派未领叶子；禁改他人 in-progress")
	}
	applyIDKey(c, *idKey)
	dryRun := *dry
	raw, err := c.TeamSeatExpand(*team, botclient.TeamExpandOpts{
		ProjectID: *project,
		Limit:     *limit,
		DryRun:    dryRun,
	})
	if err != nil {
		if ae, ok := err.(*botclient.APIError); ok && ae.Status == 404 && !dryRun {
			return fmt.Errorf("%w；可加 --dry-run 只列建议", err)
		}
		return err
	}
	opts := optsFrom(*full, *ids, *asJSON)
	if opts.Full {
		printRawFull(raw)
		return nil
	}
	v, err := botclient.ProjectTeamExpand(raw, opts)
	if err != nil {
		return err
	}
	printView(opts, botclient.FormatTeamExpandText(v), v)
	return nil
}

func cmdPatrol(c *botclient.Client, args []string) error {
	fs := newFS("patrol")
	quiet := fs.Bool("quiet-if-unchanged", false, "无变化则无 stdout、exit 0（STATE_FILE fingerprint）")
	state := fs.String("state-file", "", "fingerprint 路径（默认 YMBOT_PATROL_STATE_FILE 或 ~/.ymbot/patrol-state）")
	full, ids, asJSON := viewFlags(fs)
	if err := fs.Parse(reorderFlags(args)); err != nil {
		return err
	}
	opts := botclient.PatrolOpts{
		Quiet:     *quiet,
		StateFile: *state,
		View:      optsFrom(*full, *ids, *asJSON),
	}
	res, err := c.Patrol(opts)
	if err != nil {
		return err
	}
	if res.Suppressed {
		return nil
	}
	printView(opts.View, res.Text, res.Items)
	return nil
}

func cmdPin(c *botclient.Client, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("用法: ymbot pin list|get|set|delete …\n队共享记忆墙（非私有 memory；切队换 --team）")
	}
	switch args[0] {
	case "list":
		fs := newFS("pin list")
		team := fs.String("team", "", "team uuid")
		full, ids, asJSON := viewFlags(fs)
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if *team == "" {
			return fmt.Errorf("需要 --team <uuid>")
		}
		opts := optsFrom(*full, *ids, *asJSON)
		raw, err := c.PinList(*team)
		if err != nil {
			return err
		}
		if opts.Full {
			printRawFull(raw)
			return nil
		}
		items, err := botclient.ProjectPins(raw, opts)
		if err != nil {
			return err
		}
		printView(opts, botclient.FormatPinListText(items), items)
		return nil
	case "get":
		fs := newFS("pin get")
		team := fs.String("team", "", "team uuid")
		key := fs.String("key", "", "")
		full, ids, asJSON := viewFlags(fs)
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if *team == "" || *key == "" {
			return fmt.Errorf("需要 --team <uuid> 与 --key")
		}
		opts := optsFrom(*full, *ids, *asJSON)
		raw, err := c.PinGet(*team, *key)
		if err != nil {
			return err
		}
		if opts.Full {
			printRawFull(raw)
			return nil
		}
		items, err := botclient.ProjectPins(raw, opts)
		if err != nil {
			return err
		}
		printView(opts, botclient.FormatPinListText(items), items)
		return nil
	case "set":
		fs := newFS("pin set")
		team := fs.String("team", "", "team uuid")
		key := fs.String("key", "", "")
		value := fs.String("value", "", "")
		idKey := fs.String("idempotency-key", "", "")
		full, ids, asJSON := viewFlags(fs)
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if *team == "" || *key == "" {
			return fmt.Errorf("需要 --team <uuid> 与 --key（--value 可空串）")
		}
		applyIDKey(c, *idKey)
		raw, err := c.PinSet(*team, *key, *value)
		if err != nil {
			return err
		}
		opts := optsFrom(*full, *ids, *asJSON)
		if opts.Full {
			printRawFull(raw)
			return nil
		}
		items, err := botclient.ProjectPins(raw, opts)
		if err != nil {
			return err
		}
		printView(opts, botclient.FormatPinListText(items), items)
		return nil
	case "delete":
		fs := newFS("pin delete")
		team := fs.String("team", "", "team uuid")
		key := fs.String("key", "", "")
		full, ids, asJSON := viewFlags(fs)
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if *team == "" || *key == "" {
			return fmt.Errorf("需要 --team <uuid> 与 --key")
		}
		opts := optsFrom(*full, *ids, *asJSON)
		raw, err := c.PinDelete(*team, *key)
		if err != nil {
			return err
		}
		if opts.Full {
			printRawFull(raw)
			return nil
		}
		items, err := botclient.ProjectPins(raw, opts)
		if err != nil {
			return err
		}
		printView(opts, botclient.FormatPinListText(items), items)
		return nil
	default:
		return fmt.Errorf("未知 pin 子命令: %s", args[0])
	}
}
