package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/botclient"
)

func runCLI(args []string) error {
	c := botclient.New(botclient.LoadConfig())
	if err := c.Require(); err != nil {
		return err
	}
	switch args[0] {
	case "whoami":
		raw, err := c.WhoAmI()
		if err != nil {
			return err
		}
		printRaw(raw)
		return nil
	case "issue":
		return cmdIssue(c, args[1:])
	case "comment":
		return cmdComment(c, args[1:])
	case "memory":
		return cmdMemory(c, args[1:])
	case "handoff":
		return cmdHandoff(c, args[1:])
	default:
		return fmt.Errorf("未知命令: %s（ymbot --help）", args[0])
	}
}

func applyIDKey(c *botclient.Client, key string) {
	if key != "" {
		c.IdempotencyKey = key
	}
}

func cmdIssue(c *botclient.Client, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("用法: ymbot issue list|get|create|update|claim …")
	}
	switch args[0] {
	case "list":
		fs := flag.NewFlagSet("issue list", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		status := fs.String("status", "", "")
		assignee := fs.String("assignee", "", "")
		project := fs.String("project", "", "")
		q := fs.String("q", "", "")
		cycle := fs.String("cycle", "", "")
		label := fs.String("label", "", "")
		leafOnly := fs.Bool("leaf-only", false, "仅叶子（query leaf_only=1；依赖后端）")
		unassigned := fs.Bool("unassigned", false, "未分配（query unassigned=1；依赖后端）")
		_ = fs.Parse(args[1:])
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
		printRaw(raw)
		return nil
	case "get":
		if len(args) < 2 {
			return fmt.Errorf("用法: ymbot issue get <id>")
		}
		raw, err := c.IssueGet(args[1])
		if err != nil {
			return err
		}
		printRaw(raw)
		return nil
	case "create":
		fs := flag.NewFlagSet("issue create", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		project := fs.String("project", "", "")
		title := fs.String("title", "", "")
		desc := fs.String("description", "", "")
		status := fs.String("status", "", "")
		priority := fs.String("priority", "", "")
		assignee := fs.String("assignee", "", "")
		parent := fs.String("parent", "", "父 Issue id（Bot 通道必填，禁游离）")
		idKey := fs.String("idempotency-key", "", "")
		_ = fs.Parse(args[1:])
		if *project == "" || *title == "" {
			return fmt.Errorf("需要 --project 与 --title")
		}
		if *parent == "" {
			return fmt.Errorf("禁止创建无父游离 Issue，请在已有父节点下拆叶子或由人类编排（须 --parent）")
		}
		applyIDKey(c, *idKey)
		body := map[string]any{"project_id": *project, "title": *title, "parent_id": *parent}
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
		printRaw(raw)
		return nil
	case "update":
		if len(args) < 2 {
			return fmt.Errorf("用法: ymbot issue update <id> --version N …")
		}
		id := args[1]
		fs := flag.NewFlagSet("issue update", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		version := fs.Int("version", -1, "")
		title := fs.String("title", "", "")
		desc := fs.String("description", "", "")
		status := fs.String("status", "", "")
		priority := fs.String("priority", "", "")
		assignee := fs.String("assignee", "", "")
		idKey := fs.String("idempotency-key", "", "")
		_ = fs.Parse(args[2:])
		if *version < 0 {
			return fmt.Errorf("需要 --version（乐观锁）")
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
				return fmt.Errorf("%w；请重新 issue get 取当前 version 再试", err)
			}
			return err
		}
		printRaw(raw)
		return nil
	case "claim":
		if len(args) < 2 {
			return fmt.Errorf("用法: ymbot issue claim <id> [--version N] [--idempotency-key KEY]\n只领叶子：assignee=me；父节点不可领（服务端非叶子 409）")
		}
		id := args[1]
		fs := flag.NewFlagSet("issue claim", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		version := fs.Int("version", -1, "乐观锁；缺省先 GET 取当前 version")
		idKey := fs.String("idempotency-key", "", "")
		_ = fs.Parse(args[2:])
		applyIDKey(c, *idKey)
		raw, err := c.IssueClaim(id, *version)
		if err != nil {
			if ae, ok := err.(*botclient.APIError); ok && ae.Status == 409 {
				// 叶子闸 / 乐观锁：原样中文包装已由 APIError.Error()
				return err
			}
			return err
		}
		printRaw(raw)
		return nil
	default:
		return fmt.Errorf("未知 issue 子命令: %s（list|get|create|update|claim）", args[0])
	}
}

func cmdComment(c *botclient.Client, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("用法: ymbot comment list|create …")
	}
	switch args[0] {
	case "list":
		fs := flag.NewFlagSet("comment list", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		issue := fs.String("issue", "", "")
		_ = fs.Parse(args[1:])
		if *issue == "" {
			return fmt.Errorf("需要 --issue")
		}
		raw, err := c.CommentList(*issue)
		if err != nil {
			return err
		}
		printRaw(raw)
		return nil
	case "create":
		fs := flag.NewFlagSet("comment create", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		issue := fs.String("issue", "", "")
		body := fs.String("body", "", "")
		msg := fs.String("m", "", "")
		idKey := fs.String("idempotency-key", "", "")
		_ = fs.Parse(args[1:])
		text := *body
		if text == "" {
			text = *msg
		}
		if *issue == "" || text == "" {
			return fmt.Errorf("用法: ymbot comment create --issue ID -m|--body TEXT [--idempotency-key KEY]")
		}
		applyIDKey(c, *idKey)
		raw, err := c.CommentCreate(*issue, text)
		if err != nil {
			return err
		}
		printRaw(raw)
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
		fs := flag.NewFlagSet("memory list", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		q := fs.String("q", "", "")
		mode := fs.String("mode", "", "keyword|semantic|hybrid")
		limit := fs.Int("limit", 0, "")
		_ = fs.Parse(args[1:])
		raw, err := c.MemoryList(*q, *mode, *limit)
		if err != nil {
			return err
		}
		printRaw(raw)
		return nil
	case "set":
		fs := flag.NewFlagSet("memory set", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		key := fs.String("key", "", "")
		value := fs.String("value", "", "")
		tags := fs.String("tags", "", "")
		idKey := fs.String("idempotency-key", "", "")
		_ = fs.Parse(args[1:])
		if *key == "" || *value == "" {
			return fmt.Errorf("需要 --key 与 --value")
		}
		applyIDKey(c, *idKey)
		raw, err := c.MemorySet(*key, *value, *tags)
		if err != nil {
			return err
		}
		printRaw(raw)
		return nil
	default:
		return fmt.Errorf("未知 memory 子命令: %s", args[0])
	}
}

func cmdHandoff(c *botclient.Client, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("用法: ymbot handoff get|export|import …")
	}
	switch args[0] {
	case "get", "export":
		fs := flag.NewFlagSet("handoff get", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		out := fs.String("out", "", "")
		_ = fs.Parse(args[1:])
		raw, err := c.HandoffGet()
		if err != nil {
			return err
		}
		if *out != "" {
			var v any
			_ = json.Unmarshal(raw, &v)
			b, _ := json.MarshalIndent(v, "", "  ")
			if err := os.WriteFile(*out, append(b, '\n'), 0o600); err != nil {
				return err
			}
			fmt.Println("wrote", *out)
			return nil
		}
		printRaw(raw)
		return nil
	case "import":
		fs := flag.NewFlagSet("handoff import", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		file := fs.String("file", "", "")
		idKey := fs.String("idempotency-key", "", "")
		_ = fs.Parse(args[1:])
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
		printRaw(res)
		return nil
	default:
		return fmt.Errorf("未知 handoff 子命令: %s", args[0])
	}
}
