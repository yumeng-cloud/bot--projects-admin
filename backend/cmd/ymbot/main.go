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
	fmt.Print(`ymbot — 看板 Bot CLI / MCP（= bot-team-template 班底作业手册里的手；非另一套系统）

主航线（唯一）: team pull → seat enter → handoff get → 干活
  术语: 模板角色 slug = 席位（seat / role_slug）
  MCP:  board_team_pull → board_seat_enter → board_handoff_get
  说明书: bot-team-template/playbook/看板席位迁绑.md · templates/ymbot/
  契约:   docs/BOT_TEAM_SEAT_P0.md
whoami 可前置自检；migrate enter = seat enter（别名，非第二条主航线）。
进度写 memory/todo/comment（服务端真相源）。export/import 仅换机备份/救场。

环境:
  YMBOT_API_URL           默认 http://127.0.0.1:8080
  YMBOT_API_KEY           必填，Bearer ymk_…
  YMBOT_WORKSPACE_ID      必填
  YMBOT_MEMBER_ID         可选（缺省用 Key 绑定 / bots/me）
  YMBOT_IDEMPOTENCY_KEY   可选；写操作未传 --idempotency-key 时使用
  YMBOT_AUTOSAVE          可选=1；claim/comment/memory set 成功后 stderr 提醒「已写服务端」（非 daemon）

命令（主航线优先）:
  team pull --uuid <team_uuid>              拉班底（幂等 ensure roster；不清 memory）
  seat enter --role <slug> [--team] [--force]
                                            进席（slug=席位；有 --team 走 bind；409 禁瞎抢）
  handoff get [--full|--id|--json]          续命权威：服务端摘要，无文件
  whoami [--full|--id|--json]               可选自检（非主航线另立）
  migrate enter --role <slug> [--team] [--force]
                                            = seat enter（别名）
  seat leave --role <slug> [--team]         离席：有 --team 走 unbind，否则 release
  seat expand --team <uuid> [--project] [--limit N] [--dry-run]
                                            扩工补位：未领叶子→空闲已绑席；禁改 in-progress
  issue list [--status] [--assignee me|id] [--project] [--q] [--cycle] [--label] [--leaf-only] [--full|--id|--json]
  issue get <YM-12|id> [--full|--id|--json]
  issue create --project ID --title T --parent YM-n|ID […]
  issue claim <YM-12|id> [--version N] [--idempotency-key] [--autosave]
  issue update <YM-12|id> --version N […]   （version 见 issue get 默认输出）
  comment list --issue YM-12 [--full|--id|--json]
  comment create --issue YM-12 -m|--body TEXT [--idempotency-key] [--autosave]
  memory list [--q] [--mode keyword|semantic|hybrid] [--limit N] [--full|--id|--json]
  memory set --key K --value V [--tags] [--idempotency-key] [--autosave]
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
