// Package joinlink builds one-tap team join URLs and Bot system prompts.
// Contract: docs/NEXT_WAVE_DEBT_ONE_TAP_JOIN.md (ONE_TAP_JOIN · A).
// Does not redeem or claim; landing still uses existing ymt_ + team join/redeem.
package joinlink

import (
	"fmt"
	"net/url"
	"strings"
)

// DefaultAPIBase is the production API root embedded in join links/prompts when unset.
const DefaultAPIBase = "https://bot.api.sn.cn"

// InviteCodePrefix is the team invite plaintext prefix (ymt_).
const InviteCodePrefix = "ymt_"

// Params for BuildJoinURL / BuildBotPrompt.
type Params struct {
	Code string // required plaintext ymt_…
	Role string // optional seat slug (e.g. frontend)
	Seat string // optional seat_id; product: prefer over role when both set
	API  string // optional API-root override for the api= query; omit when empty or production
}

// NormalizeAPIBase trims trailing slashes; empty → DefaultAPIBase.
func NormalizeAPIBase(base string) string {
	b := strings.TrimSpace(base)
	b = strings.TrimRight(b, "/")
	if b == "" {
		return DefaultAPIBase
	}
	return b
}

// ValidateInviteCode returns a Chinese error if code is missing or not ymt_.
func ValidateInviteCode(code string) error {
	c := strings.TrimSpace(code)
	if c == "" {
		return fmt.Errorf("缺少入场码：请在链接中带上 code=ymt_… 参数")
	}
	if !strings.HasPrefix(c, InviteCodePrefix) {
		return fmt.Errorf("入场码无效：须以 %s 开头（团队入场码，不是工作区邀请或 Bot Key）", InviteCodePrefix)
	}
	if len(c) <= len(InviteCodePrefix) {
		return fmt.Errorf("入场码不完整：请确认复制了完整的 ymt_… 码")
	}
	return nil
}

// BuildJoinURL builds {hostBase}/join?code=…&role=…[&seat=][&api=].
// hostBase is the join-page origin (usually PUBLIC_API_BASE). api= query is only
// added when Params.API is set and differs from DefaultAPIBase.
func BuildJoinURL(hostBase string, p Params) (string, error) {
	if err := ValidateInviteCode(p.Code); err != nil {
		return "", err
	}
	base := NormalizeAPIBase(hostBase)
	u, err := url.Parse(base + "/join")
	if err != nil {
		return "", fmt.Errorf("API 根无效：%w", err)
	}
	q := url.Values{}
	q.Set("code", strings.TrimSpace(p.Code))
	if role := strings.TrimSpace(p.Role); role != "" {
		q.Set("role", role)
	}
	if seat := strings.TrimSpace(p.Seat); seat != "" {
		q.Set("seat", seat)
	}
	if api := strings.TrimSpace(p.API); api != "" {
		api = NormalizeAPIBase(api)
		if api != DefaultAPIBase {
			q.Set("api", api)
		}
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// CLIJoinLine returns `ymbot team join --code … [--role …]`.
func CLIJoinLine(p Params) (string, error) {
	if err := ValidateInviteCode(p.Code); err != nil {
		return "", err
	}
	line := "ymbot team join --code " + strings.TrimSpace(p.Code)
	if role := strings.TrimSpace(p.Role); role != "" {
		line += " --role " + role
	}
	return line, nil
}

// EffectiveAPIBase resolves the API root for prompts: Params.API override, else apiBase, else default.
func EffectiveAPIBase(apiBase string, p Params) string {
	if override := strings.TrimSpace(p.API); override != "" {
		return NormalizeAPIBase(override)
	}
	return NormalizeAPIBase(apiBase)
}

// BuildBotPrompt returns a Chinese system prompt for pasting into Cursor/Agent.
// Never includes ymk_ secrets or mcp.json / YMBOT_* teaching as the main path.
func BuildBotPrompt(apiBase string, p Params) (string, error) {
	if err := ValidateInviteCode(p.Code); err != nil {
		return "", err
	}
	api := EffectiveAPIBase(apiBase, p)
	code := strings.TrimSpace(p.Code)
	role := strings.TrimSpace(p.Role)
	seat := strings.TrimSpace(p.Seat)

	roleLine := "（未指定；join 时按新人自动选席）"
	if role != "" {
		roleLine = "`" + role + "`"
	}
	seatBlock := ""
	if seat != "" {
		seatBlock = "\n- 指定席位 ID（高级，优先于 role）：`" + seat + "`"
	}
	cli, err := CLIJoinLine(p)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("你是育蒙协作看板（AI Collab Board）的入场 Bot。用下面信息直接入队，不要让用户改 mcp.json，也不要教一长串 YMBOT_* 环境变量。\n\n")
	b.WriteString("## 连接\n")
	b.WriteString("- API 根：`" + api + "`\n")
	b.WriteString("- 入场码：`" + code + "`\n")
	b.WriteString("- 目标角色（席位 slug）：" + roleLine + seatBlock + "\n")
	b.WriteString("- 身份：使用本机已有 Bot Key（ymk_…）或 Vault；**禁止**把 ymk_ 贴进聊天或写入可转发链接。\n\n")
	b.WriteString("## 主航线（必须按序）\n")
	b.WriteString("1. 确认身份：`board_whoami` / `ymbot whoami`\n")
	b.WriteString("2. 入队：`" + cli + "`（等价 MCP `board_team_join`；底层仍是既有 redeem → pull → enter，**不要**另造 claim）\n")
	b.WriteString("3. 接班：`handoff get` / `ymbot handoff get`；任务与真相源在服务端。\n\n")
	b.WriteString("## 禁区\n")
	b.WriteString("- 禁止教改 mcp.json / Cursor MCP 配置当主路径\n")
	b.WriteString("- 禁止把「导出 YMBOT_API_URL / YMBOT_API_KEY 清单」当入场步骤\n")
	b.WriteString("- 禁止在回复里回显完整 ymk_\n\n")
	b.WriteString("CLI 一行（已带码）：\n```\n")
	b.WriteString(cli)
	b.WriteString("\n```\n")
	return b.String(), nil
}
