package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/botclient"
)

type rpcReq struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type rpcRes struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcErr         `json:"error,omitempty"`
}

type rpcErr struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func runMCP() error {
	c := botclient.New(botclient.LoadConfig())
	br := bufio.NewReader(os.Stdin)
	for {
		line, err := br.ReadBytes('\n')
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if len(line) == 0 || string(line) == "\n" {
			continue
		}
		var req rpcReq
		if err := json.Unmarshal(line, &req); err != nil {
			writeRPC(rpcRes{JSONRPC: "2.0", ID: nil, Error: &rpcErr{Code: -32700, Message: "parse error"}})
			continue
		}
		res := handleMCP(c, req)
		if req.ID == nil || string(req.ID) == "null" {
			continue
		}
		writeRPC(res)
	}
}

func writeRPC(r rpcRes) {
	r.JSONRPC = "2.0"
	b, _ := json.Marshal(r)
	os.Stdout.Write(append(b, '\n'))
}

func handleMCP(c *botclient.Client, req rpcReq) rpcRes {
	id := req.ID
	switch req.Method {
	case "initialize":
		return rpcRes{ID: id, Result: map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]any{"name": "ymbot", "version": "0.2.0"},
		}}
	case "notifications/initialized", "initialized":
		return rpcRes{ID: id, Result: map[string]any{}}
	case "ping":
		return rpcRes{ID: id, Result: map[string]any{}}
	case "tools/list":
		return rpcRes{ID: id, Result: map[string]any{"tools": mcpTools()}}
	case "tools/call":
		var p struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}
		_ = json.Unmarshal(req.Params, &p)
		if p.Arguments == nil {
			p.Arguments = map[string]any{}
		}
		text, err := callTool(c, p.Name, p.Arguments)
		if err != nil {
			return rpcRes{ID: id, Result: map[string]any{
				"content": []map[string]any{{"type": "text", "text": err.Error()}},
				"isError": true,
			}}
		}
		return rpcRes{ID: id, Result: map[string]any{
			"content": []map[string]any{{"type": "text", "text": text}},
		}}
	default:
		return rpcRes{ID: id, Error: &rpcErr{Code: -32601, Message: "method not found: " + req.Method}}
	}
}

func mcpTools() []map[string]any {
	obj := func(props map[string]any, req ...string) map[string]any {
		m := map[string]any{"type": "object", "properties": props}
		if len(req) > 0 {
			m["required"] = req
		}
		return m
	}
	str := map[string]any{"type": "string"}
	num := map[string]any{"type": "number"}
	bol := map[string]any{"type": "boolean"}
	idKey := map[string]any{"type": "string", "description": "幂等 Key"}
	full := map[string]any{"type": "boolean", "description": "全量含 UUID"}
	ids := map[string]any{"type": "boolean", "description": "附加裸 id"}
	issueRef := map[string]any{"type": "string", "description": "YM-n 或 id"}
	return []map[string]any{
		{"name": "board_whoami", "description": "可选自检：Bot 身份短文本；主航线仍是 team_pull→seat_enter→handoff_get", "inputSchema": obj(map[string]any{"full": full, "ids": ids})},
		{"name": "board_issue_list", "description": "Issue 瘦表 YM-n（无 UUID）；默认剔归档；leaf_only=可领池；team_id=按队过滤", "inputSchema": obj(map[string]any{
			"status": str, "assignee_id": str, "project_id": str, "q": str, "cycle_id": str, "label_id": str, "team_id": str,
			"leaf_only": bol, "include_archived": bol, "archived_only": bol, "full": full, "ids": ids,
		})},
		{"name": "board_issue_get", "description": "Issue 瘦读；默认含 version（乐观锁），无 UUID", "inputSchema": obj(map[string]any{"issue_id": issueRef, "full": full, "ids": ids}, "issue_id")},
		{"name": "board_issue_create", "description": "父下建叶子（须 parent_id）；禁无父游离", "inputSchema": obj(map[string]any{
			"project_id": str, "title": str, "parent_id": issueRef, "description": str, "status": str, "priority": str, "assignee_id": str,
			"idempotency_key": idKey, "full": full, "ids": ids,
		}, "project_id", "title", "parent_id")},
		{"name": "board_issue_update", "description": "更新 Issue；version 来自 get 默认输出", "inputSchema": obj(map[string]any{
			"issue_id": issueRef, "version": num, "title": str, "description": str, "status": str, "priority": str, "assignee_id": str,
			"idempotency_key": idKey, "full": full, "ids": ids,
		}, "issue_id", "version")},
		{"name": "board_issue_claim", "description": "领叶子；进度写 memory/comment，勿依赖 export", "inputSchema": obj(map[string]any{
			"issue_id": issueRef, "version": num, "idempotency_key": idKey, "full": full, "ids": ids,
		}, "issue_id")},
		{"name": "board_comment_list", "description": "评论瘦摘要（YM-n；无 UUID）", "inputSchema": obj(map[string]any{"issue_id": issueRef, "full": full, "ids": ids}, "issue_id")},
		{"name": "board_comment_create", "description": "写评论到服务端（续命进度）", "inputSchema": obj(map[string]any{
			"issue_id": issueRef, "body": str, "idempotency_key": idKey, "full": full, "ids": ids,
		}, "issue_id", "body")},
		{"name": "board_memory_list", "description": "记忆瘦摘要 key+截断值", "inputSchema": obj(map[string]any{"q": str, "mode": str, "limit": num, "full": full, "ids": ids})},
		{"name": "board_memory_set", "description": "Upsert 记忆到服务端；勿攒到 export", "inputSchema": obj(map[string]any{
			"key_name": str, "value_text": str, "tags": str, "idempotency_key": idKey, "full": full, "ids": ids,
		}, "key_name", "value_text")},
		{"name": "board_handoff_get", "description": "主航线③：续命权威（服务端摘要无文件）；先 team_pull→seat_enter", "inputSchema": obj(map[string]any{"full": full, "ids": ids})},
		{"name": "board_handoff_import", "description": "从备份/外来包救场；非日常续工", "inputSchema": obj(map[string]any{
			"handoff": map[string]any{"type": "object"}, "idempotency_key": idKey, "full": full, "ids": ids,
		}, "handoff")},
		{"name": "board_team_pull", "description": "主航线①：按团队 UUID 拉班底（slug=席位；幂等 ensure；不清 memory）", "inputSchema": obj(map[string]any{
			"team_uuid": str, "idempotency_key": idKey, "full": full, "ids": ids,
		}, "team_uuid")},
		{"name": "board_team_join", "description": "空手上场封装：redeem(ymt_)→team pull→seat enter→handoff get；活跃席禁顶替；空手自动发 ymk_，有钥续队不重发", "inputSchema": obj(map[string]any{
			"code": str, "role": str, "new_seat": str, "force": bol, "idempotency_key": idKey, "full": full, "ids": ids,
		}, "code")},
		{"name": "board_team_templates", "description": "内置班底模板清单（key/席位 slug；瘦输出）", "inputSchema": obj(map[string]any{
			"full": full, "ids": ids,
		})},
		{"name": "board_team_create_from_template", "description": "从模板建队+空席（template_key+name；未知模板 400 人话）", "inputSchema": obj(map[string]any{
			"template_key": str, "name": str, "slug": str, "idempotency_key": idKey, "full": full, "ids": ids,
		}, "template_key", "name")},
		{"name": "board_seat_enter", "description": "主航线②：进席（role slug=席位；有 team_uuid 走 bind；409 禁瞎抢）", "inputSchema": obj(map[string]any{
			"role": str, "team_uuid": str, "force": bol, "idempotency_key": idKey, "full": full, "ids": ids,
		}, "role")},
		{"name": "board_seat_leave", "description": "离席（slug=席位）：有 team_uuid 则 unbind，否则 release", "inputSchema": obj(map[string]any{
			"role": str, "team_uuid": str, "idempotency_key": idKey, "full": full, "ids": ids,
		}, "role")},
		{"name": "board_seat_expand", "description": "扩工补位（挂主航线后）：未领叶子→空闲已绑席；禁改 in-progress", "inputSchema": obj(map[string]any{
			"team_uuid": str, "project_id": str, "limit": map[string]any{"type": "number"}, "dry_run": bol, "idempotency_key": idKey, "full": full, "ids": ids,
		}, "team_uuid")},
		{"name": "board_patrol", "description": "辅路径巡检：当前席 open+assignee=me；quiet_if_unchanged 无变化则空内容 exit 等价安静", "inputSchema": obj(map[string]any{
			"quiet_if_unchanged": bol, "state_file": str, "full": full, "ids": ids,
		})},
		{"name": "board_pin_list", "description": "队共享记忆墙列表（非私有 memory）", "inputSchema": obj(map[string]any{
			"team_uuid": str, "full": full, "ids": ids,
		}, "team_uuid")},
		{"name": "board_pin_get", "description": "读一条队 pin；无则 404", "inputSchema": obj(map[string]any{
			"team_uuid": str, "key": str, "full": full, "ids": ids,
		}, "team_uuid", "key")},
		{"name": "board_pin_set", "description": "Upsert 队 pin（共享墙；非 handoff）", "inputSchema": obj(map[string]any{
			"team_uuid": str, "key": str, "value": str, "idempotency_key": idKey, "full": full, "ids": ids,
		}, "team_uuid", "key")},
		{"name": "board_pin_delete", "description": "删一条队 pin", "inputSchema": obj(map[string]any{
			"team_uuid": str, "key": str, "full": full, "ids": ids,
		}, "team_uuid", "key")},
	}
}

func strArg(a map[string]any, k string) string {
	if v, ok := a[k]; ok && v != nil {
		return fmt.Sprint(v)
	}
	return ""
}

func boolArg(a map[string]any, k string) bool {
	v, ok := a[k]
	if !ok || v == nil {
		return false
	}
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return t == "1" || strings.EqualFold(t, "true")
	default:
		s := fmt.Sprint(t)
		return s == "1" || s == "true"
	}
}

func viewFromArgs(a map[string]any) botclient.ViewOpts {
	return botclient.ViewOpts{Full: boolArg(a, "full"), IDs: boolArg(a, "ids")}
}

func viewDetailFromArgs(a map[string]any) botclient.ViewOpts {
	o := viewFromArgs(a)
	o.Detail = true
	return o
}

func callTool(c *botclient.Client, name string, args map[string]any) (string, error) {
	if name == "board_team_join" {
		// no RequireAPIKey — HUMAN_ONBOARD bootstrap（成功响应自动发 ymk_）
	} else if err := c.Require(); err != nil {
		return "", err
	}
	if k := strArg(args, "idempotency_key"); k != "" {
		c.IdempotencyKey = k
	} else {
		c.IdempotencyKey = ""
	}
	opts := viewFromArgs(args)
	switch name {
	case "board_issue_get", "board_issue_create", "board_issue_update", "board_issue_claim":
		opts.Detail = true
	}

	emitIssue := func(raw json.RawMessage) (string, error) {
		if opts.Full {
			var v any
			_ = json.Unmarshal(raw, &v)
			return botclient.MarshalView(v), nil
		}
		items, err := c.ProjectIssues(raw, opts)
		if err != nil {
			return "", err
		}
		if name == "board_issue_list" {
			return botclient.FormatIssueListText(items), nil
		}
		if len(items) == 0 {
			return botclient.FormatIssueGetText(map[string]any{}), nil
		}
		return botclient.FormatIssueGetText(items[0]), nil
	}

	switch name {
	case "board_whoami":
		raw, err := c.WhoAmI()
		if err != nil {
			return "", err
		}
		if opts.Full {
			var v any
			_ = json.Unmarshal(raw, &v)
			return botclient.MarshalView(v), nil
		}
		v := botclient.ProjectWhoAmI(raw, opts)
		return botclient.FormatWhoAmIText(v), nil
	case "board_issue_list":
		qv := url.Values{}
		for _, k := range []string{"status", "assignee_id", "project_id", "q", "cycle_id", "label_id", "team_id"} {
			if v := strArg(args, k); v != "" {
				qv.Set(k, v)
			}
		}
		if boolArg(args, "leaf_only") {
			qv.Set("leaf_only", "1")
		}
		// 默认不传 include_archived / archived；显式参数才传（archived=only 优先）
		if boolArg(args, "archived_only") {
			qv.Set("archived", "only")
			opts.IncludeArchived = true
		} else if boolArg(args, "include_archived") {
			qv.Set("include_archived", "1")
			opts.IncludeArchived = true
		}
		raw, err := c.IssueList(qv)
		if err != nil {
			return "", err
		}
		if !opts.IncludeArchived {
			raw = botclient.DropArchivedFromIssueListJSON(raw)
		}
		return emitIssue(raw)
	case "board_issue_get":
		ref := strArg(args, "issue_id")
		id, err := c.ResolveIssueRef(ref)
		if err != nil {
			return "", err
		}
		raw, err := c.IssueGet(id)
		if err != nil {
			return "", err
		}
		return emitIssue(raw)
	case "board_issue_create":
		parentRef := strArg(args, "parent_id")
		parentID, err := c.ResolveIssueRef(parentRef)
		if err != nil {
			return "", err
		}
		body := map[string]any{"project_id": strArg(args, "project_id"), "title": strArg(args, "title"), "parent_id": parentID}
		for _, k := range []string{"description", "status", "priority", "assignee_id"} {
			if v := strArg(args, k); v != "" {
				body[k] = v
			}
		}
		raw, err := c.IssueCreate(body)
		if err != nil {
			return "", err
		}
		return emitIssue(raw)
	case "board_issue_claim":
		ref := strArg(args, "issue_id")
		id, err := c.ResolveIssueRef(ref)
		if err != nil {
			return "", err
		}
		ver := -1
		if v, ok := args["version"]; ok && v != nil {
			switch n := v.(type) {
			case float64:
				ver = int(n)
			case int:
				ver = n
			}
		}
		raw, err := c.IssueClaim(id, ver)
		if err != nil {
			return "", err
		}
		return emitIssue(raw)
	case "board_issue_update":
		ref := strArg(args, "issue_id")
		id, err := c.ResolveIssueRef(ref)
		if err != nil {
			return "", err
		}
		body := map[string]any{}
		if v, ok := args["version"]; ok {
			body["version"] = v
		}
		for _, k := range []string{"title", "description", "status", "priority", "assignee_id"} {
			if v := strArg(args, k); v != "" {
				body[k] = v
			}
		}
		raw, err := c.IssueUpdate(id, body)
		if err != nil {
			return "", err
		}
		return emitIssue(raw)
	case "board_comment_list":
		ref := strArg(args, "issue_id")
		id, err := c.ResolveIssueRef(ref)
		if err != nil {
			return "", err
		}
		ident := ref
		if !botclient.LooksLikeIdentifier(ref) {
			ident = c.CachedIdent(id)
		}
		raw, err := c.CommentList(id)
		if err != nil {
			return "", err
		}
		if opts.Full {
			var v any
			_ = json.Unmarshal(raw, &v)
			return botclient.MarshalView(v), nil
		}
		items, err := c.ProjectComments(raw, ident, opts)
		if err != nil {
			return "", err
		}
		return botclient.FormatCommentListText(items), nil
	case "board_comment_create":
		ref := strArg(args, "issue_id")
		id, err := c.ResolveIssueRef(ref)
		if err != nil {
			return "", err
		}
		raw, err := c.CommentCreate(id, strArg(args, "body"))
		if err != nil {
			return "", err
		}
		if opts.Full {
			var v any
			_ = json.Unmarshal(raw, &v)
			return botclient.MarshalView(v), nil
		}
		items, err := c.ProjectComments(raw, ref, opts)
		if err != nil {
			return "", err
		}
		return botclient.FormatCommentListText(items), nil
	case "board_memory_list":
		limit := 0
		if v, ok := args["limit"]; ok && v != nil {
			switch n := v.(type) {
			case float64:
				limit = int(n)
			case int:
				limit = n
			}
		}
		raw, err := c.MemoryList(strArg(args, "q"), strArg(args, "mode"), limit)
		if err != nil {
			return "", err
		}
		if opts.Full {
			var v any
			_ = json.Unmarshal(raw, &v)
			return botclient.MarshalView(v), nil
		}
		items, err := botclient.ProjectMemories(raw, opts)
		if err != nil {
			return "", err
		}
		return botclient.FormatMemoryListText(items), nil
	case "board_memory_set":
		raw, err := c.MemorySet(strArg(args, "key_name"), strArg(args, "value_text"), strArg(args, "tags"))
		if err != nil {
			return "", err
		}
		if opts.Full {
			var v any
			_ = json.Unmarshal(raw, &v)
			return botclient.MarshalView(v), nil
		}
		items, err := botclient.ProjectMemories(raw, opts)
		if err != nil {
			return "", err
		}
		return botclient.FormatMemoryListText(items), nil
	case "board_handoff_get":
		raw, err := c.HandoffGet()
		if err != nil {
			return "", err
		}
		if opts.Full {
			var v any
			_ = json.Unmarshal(raw, &v)
			return botclient.MarshalView(v), nil
		}
		v, err := c.ProjectHandoff(raw, opts)
		if err != nil {
			return "", err
		}
		return botclient.FormatHandoffText(v), nil
	case "board_handoff_import":
		var body any
		if h, ok := args["handoff"]; ok {
			body = map[string]any{"handoff": h}
		} else {
			body = args
		}
		raw, err := c.HandoffImport(body)
		if err != nil {
			return "", err
		}
		if opts.Full {
			var v any
			_ = json.Unmarshal(raw, &v)
			return botclient.MarshalView(v), nil
		}
		return botclient.MarshalView(botclient.ProjectHandoffImport(raw, opts)), nil
	case "board_team_pull":
		uuid := strArg(args, "team_uuid")
		if uuid == "" {
			uuid = strArg(args, "uuid")
		}
		raw, err := c.TeamPull(uuid)
		if err != nil {
			return "", err
		}
		if opts.Full {
			var v any
			_ = json.Unmarshal(raw, &v)
			return botclient.MarshalView(v), nil
		}
		items, err := botclient.ProjectTeamRoster(raw, opts)
		if err != nil {
			return "", err
		}
		return botclient.FormatTeamRosterText(items), nil
	case "board_team_join":
		code := strArg(args, "code")
		if code == "" {
			return "", fmt.Errorf("需要 code（ymt_… 团队入场码）")
		}
		v, err := c.TeamJoin(botclient.TeamJoinOpts{
			Code:    code,
			Role:    strArg(args, "role"),
			NewSeat: strArg(args, "new_seat"),
			Force:   boolArg(args, "force"),
		})
		if err != nil {
			return "", err
		}
		if opts.Full {
			return botclient.MarshalView(v), nil
		}
		return botclient.FormatTeamJoinText(v), nil
	case "board_team_templates":
		raw, err := c.TeamTemplatesList()
		if err != nil {
			return "", err
		}
		if opts.Full {
			var v any
			_ = json.Unmarshal(raw, &v)
			return botclient.MarshalView(v), nil
		}
		items, err := botclient.ProjectTeamTemplates(raw, opts)
		if err != nil {
			return "", err
		}
		return botclient.FormatTeamTemplatesText(items), nil
	case "board_team_create_from_template":
		tk := strArg(args, "template_key")
		if tk == "" {
			tk = strArg(args, "template")
		}
		name := strArg(args, "name")
		if tk == "" || name == "" {
			return "", fmt.Errorf("需要 template_key 与 name")
		}
		raw, err := c.CreateTeamFromTemplate(botclient.CreateTeamFromTemplateOpts{
			Name: name, Slug: strArg(args, "slug"), TemplateKey: tk,
		})
		if err != nil {
			return "", err
		}
		if opts.Full {
			var v any
			_ = json.Unmarshal(raw, &v)
			return botclient.MarshalView(v), nil
		}
		v, err := botclient.ProjectCreateTeamFromTemplate(raw, opts)
		if err != nil {
			return "", err
		}
		return botclient.FormatCreateTeamFromTemplateText(v), nil
	case "board_seat_enter":
		team := strArg(args, "team_uuid")
		if team == "" {
			team = strArg(args, "uuid")
		}
		v, err := c.SeatEnter(strArg(args, "role"), botclient.SeatEnterOpts{
			TeamID: team,
			Force:  boolArg(args, "force"),
		})
		if err != nil {
			return "", err
		}
		if opts.Full {
			return botclient.MarshalView(v), nil
		}
		return botclient.FormatSeatEnterText(v), nil
	case "board_seat_leave":
		role := strArg(args, "role")
		team := strArg(args, "team_uuid")
		if team == "" {
			team = strArg(args, "uuid")
		}
		v, err := c.SeatLeave(role, botclient.SeatLeaveOpts{TeamID: team})
		if err != nil {
			return "", err
		}
		via, _ := v["via"].(string)
		if via == "" {
			via = "-"
		}
		return fmt.Sprintf("left role=%s via=%s\n", role, via), nil
	case "board_seat_expand":
		uuid := strArg(args, "team_uuid")
		if uuid == "" {
			uuid = strArg(args, "uuid")
		}
		dry := boolArg(args, "dry_run")
		limit := 0
		if v, ok := args["limit"]; ok && v != nil {
			switch n := v.(type) {
			case float64:
				limit = int(n)
			case int:
				limit = n
			default:
				fmt.Sscanf(fmt.Sprint(n), "%d", &limit)
			}
		}
		raw, err := c.TeamSeatExpand(uuid, botclient.TeamExpandOpts{
			ProjectID: strArg(args, "project_id"),
			Limit:     limit,
			DryRun:    dry,
		})
		if err != nil {
			if ae, ok := err.(*botclient.APIError); ok && ae.Status == 404 && !dry {
				return "", fmt.Errorf("%w；可设 dry_run=true 只列建议", err)
			}
			return "", err
		}
		if opts.Full {
			var v any
			_ = json.Unmarshal(raw, &v)
			return botclient.MarshalView(v), nil
		}
		v, err := botclient.ProjectTeamExpand(raw, opts)
		if err != nil {
			return "", err
		}
		return botclient.FormatTeamExpandText(v), nil
	case "board_patrol":
		res, err := c.Patrol(botclient.PatrolOpts{
			Quiet:     boolArg(args, "quiet_if_unchanged"),
			StateFile: strArg(args, "state_file"),
			View:      opts,
		})
		if err != nil {
			return "", err
		}
		if res.Suppressed {
			return "", nil
		}
		return res.Text, nil
	case "board_pin_list", "board_pin_get", "board_pin_set", "board_pin_delete":
		uuid := strArg(args, "team_uuid")
		if uuid == "" {
			uuid = strArg(args, "uuid")
		}
		if uuid == "" {
			return "", fmt.Errorf("需要 team_uuid")
		}
		key := strArg(args, "key")
		var raw json.RawMessage
		var err error
		switch name {
		case "board_pin_list":
			raw, err = c.PinList(uuid)
		case "board_pin_get":
			if key == "" {
				return "", fmt.Errorf("需要 key")
			}
			raw, err = c.PinGet(uuid, key)
		case "board_pin_set":
			if key == "" {
				return "", fmt.Errorf("需要 key")
			}
			raw, err = c.PinSet(uuid, key, strArg(args, "value"))
		case "board_pin_delete":
			if key == "" {
				return "", fmt.Errorf("需要 key")
			}
			raw, err = c.PinDelete(uuid, key)
		}
		if err != nil {
			return "", err
		}
		if opts.Full {
			var v any
			_ = json.Unmarshal(raw, &v)
			return botclient.MarshalView(v), nil
		}
		items, err := botclient.ProjectPins(raw, opts)
		if err != nil {
			return "", err
		}
		return botclient.FormatPinListText(items), nil
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}
