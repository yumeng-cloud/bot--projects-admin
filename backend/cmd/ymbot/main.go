package main

import (
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
	fmt.Print(`ymbot — 看板 Bot CLI / MCP（通道 DX：docs/BOT_CHANNEL_DX_P0.md；续命：docs/BOT_HANDOFF_CONTINUITY_P0.md）

续工: 先 whoami → handoff get（无文件，服务端真相源）；进度写 memory/todo/comment，勿攒到最后 export。
export: 换机备份可选；日常续工不需要。import: 救场/迁入可选。

环境:
  YMBOT_API_URL           默认 http://127.0.0.1:8080
  YMBOT_API_KEY           必填，Bearer ymk_…
  YMBOT_WORKSPACE_ID      必填
  YMBOT_MEMBER_ID         可选（缺省用 Key 绑定 / bots/me）
  YMBOT_IDEMPOTENCY_KEY   可选；写操作未传 --idempotency-key 时使用
  YMBOT_AUTOSAVE          可选=1；claim/comment/memory set 成功后 stderr 提醒「已写服务端」（非 daemon）

命令:
  whoami [--full|--id|--json]
  issue list [--status] [--assignee me|id] [--project] [--q] [--cycle] [--label] [--leaf-only] [--full|--id|--json]
  issue get <YM-12|id> [--full|--id|--json]
  issue create --project ID --title T --parent YM-n|ID […]
  issue claim <YM-12|id> [--version N] [--idempotency-key] [--autosave]
  issue update <YM-12|id> --version N […]   （version 见 issue get 默认输出）
  comment list --issue YM-12 [--full|--id|--json]
  comment create --issue YM-12 -m|--body TEXT [--idempotency-key] [--autosave]
  memory list [--q] [--mode keyword|semantic|hybrid] [--limit N] [--full|--id|--json]
  memory set --key K --value V [--tags] [--idempotency-key] [--autosave]
  handoff get [--full|--id|--json]          续命权威：服务端摘要，无文件
  handoff export [--out file.json] […]      换机备份可选；日常续工不需要
  handoff import --file file.json|- […]     救场/迁入可选；非日常续工
  mcp-serve            stdio MCP（full/ids 同 --full/--id）

输出档位:
  默认     文本友好；Issue 用 YM-n；无 32 位 hex UUID；禁 REST 原样 dump
  --id     仍瘦，附加裸 id（MCP: ids=true）
  --full   接近 REST 全量（MCP: full=true）
  --json   强制瘦 JSON（可与 --full/--id 组合）

护栏: 只领叶子；禁无父游离 create；禁 status=done；幂等 Key 可透传。
鉴权: 只接受 API Key。错误为一行中文 + 退出码（401/403/404/409）。
`)
}
