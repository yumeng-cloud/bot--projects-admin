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
			continue // notification
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
			"serverInfo":      map[string]any{"name": "ymbot", "version": "0.1.0"},
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
	idKey := map[string]any{"type": "string", "description": "可选 Idempotency-Key；缺省读 YMBOT_IDEMPOTENCY_KEY 或自动生成"}
	return []map[string]any{
		{"name": "board_whoami", "description": "当前 Bot 身份（workspace/member）", "inputSchema": obj(map[string]any{})},
		{"name": "board_issue_list", "description": "列出 Issues（可 leaf_only=1 拉可领叶子池；依赖后端）", "inputSchema": obj(map[string]any{
			"status": str, "assignee_id": str, "project_id": str, "q": str, "cycle_id": str, "label_id": str,
			"leaf_only": map[string]any{"type": "boolean", "description": "仅叶子 Issue（query leaf_only=1）"},
		})},
		{"name": "board_issue_get", "description": "读取 Issue", "inputSchema": obj(map[string]any{"issue_id": str}, "issue_id")},
		{"name": "board_issue_create", "description": "在已有父节点下创建叶子 Issue（须 parent_id；禁无父游离；禁 done）", "inputSchema": obj(map[string]any{
			"project_id": str, "title": str, "parent_id": str, "description": str, "status": str, "priority": str, "assignee_id": str,
			"idempotency_key": idKey,
		}, "project_id", "title", "parent_id")},
		{"name": "board_issue_update", "description": "更新 Issue（需 version；Bot 禁 done）", "inputSchema": obj(map[string]any{
			"issue_id": str, "version": num, "title": str, "description": str, "status": str, "priority": str, "assignee_id": str,
			"idempotency_key": idKey,
		}, "issue_id", "version")},
		{"name": "board_issue_claim", "description": "领取叶子 Issue（assignee=me；≠ 角色 board_role_claim；父节点不可领，非叶子 409）", "inputSchema": obj(map[string]any{
			"issue_id": str, "version": num, "idempotency_key": idKey,
		}, "issue_id")},
		{"name": "board_comment_list", "description": "列出评论", "inputSchema": obj(map[string]any{"issue_id": str}, "issue_id")},
		{"name": "board_comment_create", "description": "写评论", "inputSchema": obj(map[string]any{
			"issue_id": str, "body": str, "idempotency_key": idKey,
		}, "issue_id", "body")},
		{"name": "board_memory_list", "description": "列出/检索记忆（mode=keyword|semantic|hybrid）", "inputSchema": obj(map[string]any{"q": str, "mode": str, "limit": num})},
		{"name": "board_memory_set", "description": "Upsert 记忆", "inputSchema": obj(map[string]any{
			"key_name": str, "value_text": str, "tags": str, "idempotency_key": idKey,
		}, "key_name", "value_text")},
		{"name": "board_handoff_get", "description": "导出 handoff", "inputSchema": obj(map[string]any{})},
		{"name": "board_handoff_import", "description": "导入 handoff JSON", "inputSchema": obj(map[string]any{
			"handoff": map[string]any{"type": "object"}, "idempotency_key": idKey,
		}, "handoff")},
	}
}

func strArg(a map[string]any, k string) string {
	if v, ok := a[k]; ok && v != nil {
		return fmt.Sprint(v)
	}
	return ""
}

func callTool(c *botclient.Client, name string, args map[string]any) (string, error) {
	if err := c.Require(); err != nil {
		return "", err
	}
	// 写工具：透传 idempotency_key（否则保留 env / 自动生成）
	if k := strArg(args, "idempotency_key"); k != "" {
		c.IdempotencyKey = k
	} else {
		c.IdempotencyKey = "" // 让 ResolveIDKey 走 env / UUID
	}
	var raw json.RawMessage
	var err error
	switch name {
	case "board_whoami":
		raw, err = c.WhoAmI()
	case "board_issue_list":
		qv := url.Values{}
		for _, k := range []string{"status", "assignee_id", "project_id", "q", "cycle_id", "label_id"} {
			if v := strArg(args, k); v != "" {
				qv.Set(k, v)
			}
		}
		if v, ok := args["leaf_only"]; ok && v != nil {
			switch t := v.(type) {
			case bool:
				if t {
					qv.Set("leaf_only", "1")
				}
			case string:
				if t == "1" || strings.EqualFold(t, "true") {
					qv.Set("leaf_only", "1")
				}
			default:
				if fmt.Sprint(t) == "1" || fmt.Sprint(t) == "true" {
					qv.Set("leaf_only", "1")
				}
			}
		}
		raw, err = c.IssueList(qv)
	case "board_issue_get":
		raw, err = c.IssueGet(strArg(args, "issue_id"))
	case "board_issue_create":
		body := map[string]any{"project_id": strArg(args, "project_id"), "title": strArg(args, "title")}
		if pid := strArg(args, "parent_id"); pid != "" {
			body["parent_id"] = pid
		}
		for _, k := range []string{"description", "status", "priority", "assignee_id"} {
			if v := strArg(args, k); v != "" {
				body[k] = v
			}
		}
		raw, err = c.IssueCreate(body)
	case "board_issue_claim":
		ver := -1
		if v, ok := args["version"]; ok && v != nil {
			switch n := v.(type) {
			case float64:
				ver = int(n)
			case int:
				ver = n
			}
		}
		raw, err = c.IssueClaim(strArg(args, "issue_id"), ver)
	case "board_issue_update":
		body := map[string]any{}
		if v, ok := args["version"]; ok {
			body["version"] = v
		}
		for _, k := range []string{"title", "description", "status", "priority", "assignee_id"} {
			if v := strArg(args, k); v != "" {
				body[k] = v
			}
		}
		raw, err = c.IssueUpdate(strArg(args, "issue_id"), body)
	case "board_comment_list":
		raw, err = c.CommentList(strArg(args, "issue_id"))
	case "board_comment_create":
		raw, err = c.CommentCreate(strArg(args, "issue_id"), strArg(args, "body"))
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
		raw, err = c.MemoryList(strArg(args, "q"), strArg(args, "mode"), limit)
	case "board_memory_set":
		raw, err = c.MemorySet(strArg(args, "key_name"), strArg(args, "value_text"), strArg(args, "tags"))
	case "board_handoff_get":
		raw, err = c.HandoffGet()
	case "board_handoff_import":
		var body any
		if h, ok := args["handoff"]; ok {
			body = map[string]any{"handoff": h}
		} else {
			body = args
		}
		raw, err = c.HandoffImport(body)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
	if err != nil {
		return "", err
	}
	var v any
	_ = json.Unmarshal(raw, &v)
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b), nil
}
