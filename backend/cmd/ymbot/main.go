package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/botclient"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		printHelp()
		return
	}
	if os.Args[1] == "mcp-serve" {
		if err := runMCP(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if err := runCLI(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		if ae, ok := err.(*botclient.APIError); ok {
			os.Exit(ae.ExitCode())
		}
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Print(`ymbot — 看板 Bot CLI / MCP（P0，见 docs/BOT_MCP_CLI.md）

环境:
  YMBOT_API_URL           默认 http://127.0.0.1:8080
  YMBOT_API_KEY           必填，Bearer ymk_…
  YMBOT_WORKSPACE_ID      必填
  YMBOT_MEMBER_ID         可选（缺省用 Key 绑定 / bots/me）
  YMBOT_IDEMPOTENCY_KEY   可选；写操作未传 --idempotency-key 时使用

命令:
  whoami
  issue list [--status] [--assignee me|id] [--project] [--q] [--cycle] [--label] [--leaf-only]
  issue get <id>
  issue create --project ID --title T --parent PARENT_ID […]  （须 --parent；禁无父游离）
  issue claim <id> [--version N] [--idempotency-key]   （只领叶子；≠ 角色 role claim）
  issue update <id> --version N [--title] [--status] [--priority] [--description] [--assignee] [--idempotency-key]
  comment list --issue ID
  comment create --issue ID -m|--body TEXT [--idempotency-key]
  memory list [--q] [--mode keyword|semantic|hybrid] [--limit N]
  memory set --key K --value V [--tags] [--idempotency-key]
  handoff get|export [--out file.json]
  handoff import --file file.json|- [--idempotency-key]
  mcp-serve            stdio MCP（含 board_issue_claim；写 tool 可传 idempotency_key）

护栏: 只领叶子（issue claim）；禁无父游离 create；禁 status=done；幂等 Key 可透传。
角色 claim（岗位，P1 board_role_claim）≠ Issue 叶子 claim（board_issue_claim）。
鉴权: 只接受 API Key。错误为一行中文 + 退出码（401/403/404/409）。
`)
}

func printJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func printRaw(b json.RawMessage) {
	var v any
	if json.Unmarshal(b, &v) == nil {
		printJSON(v)
		return
	}
	fmt.Println(string(b))
}
