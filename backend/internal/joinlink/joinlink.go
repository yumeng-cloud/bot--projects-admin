// Package joinlink builds one-tap team join URLs and Bot system prompts.
// Contract: docs/NEXT_WAVE_DEBT_ONE_TAP_JOIN.md (ONE_TAP_JOIN · A · E).
// Does not redeem or claim; landing still uses existing ymt_ + team join/redeem.
package joinlink

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/installscript"
)

// DefaultAPIBase is the production API root embedded in join links/prompts when unset.
const DefaultAPIBase = "https://bot.api.sn.cn"

// InviteCodePrefix is the team invite plaintext prefix (ymt_).
const InviteCodePrefix = "ymt_"

// GoInstallCmd is the fallback CLI install (module path = go.mod).
const GoInstallCmd = installscript.GoInstallCmd

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

// InstallCurlLine returns `curl -fsSL {api}/install.sh | bash`.
func InstallCurlLine(apiBase string) string {
	return "curl -fsSL " + NormalizeAPIBase(apiBase) + "/install.sh | bash"
}

// BuildBotPrompt returns a Chinese system prompt for pasting into Cursor/Agent.
// Step 0 is always install (never assume ymbot is already on PATH).
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
	installCurl := InstallCurlLine(api)

	var b strings.Builder
	// First line: copy-paste curl|bash (product: 首行即 curl|bash)
	b.WriteString(installCurl + "\n\n")
	b.WriteString("你是育蒙协作看板（AI Collab Board）的入场 Bot。对方**未必已装** `ymbot`：先安装再入队。不要让用户改 mcp.json，也不要教一长串 YMBOT_* 环境变量。\n\n")
	b.WriteString("## 0. 一键安装（强制·若本机还没有 ymbot）\n")
	b.WriteString("主路径（复制执行）：\n```\n")
	b.WriteString(installCurl)
	b.WriteString("\n```\n备路径（已装 Go 时）：\n```\n")
	b.WriteString(GoInstallCmd)
	b.WriteString("\n```\n装完自检：`ymbot -h`（或 `ymbot version`）。若 `~/.local/bin` 不在 PATH，按安装脚本提示 export。\n\n")
	b.WriteString("## 连接\n")
	b.WriteString("- API 根：`" + api + "`（whoami / join / handoff 都打这个根；可用环境变量 `YMBOT_API_URL=" + api + "`，勿把 Key 清单当主教程）\n")
	b.WriteString("- 入场码：`" + code + "`\n")
	b.WriteString("- 目标角色（席位 slug）：" + roleLine + seatBlock + "\n")
	b.WriteString("- 身份两态：\n")
	b.WriteString("  - **空手**：无本机 ymk_ 也可直接 join；成功响应会**自动发一把**身份钥匙（仅此一次，私存；**禁止**贴进聊天或写入可转发链接）\n")
	b.WriteString("  - **已有钥**：续队/换机带上原 `ymk_`（Bearer / `YMBOT_API_KEY`）；**不重发**钥匙\n\n")
	b.WriteString("## 主航线（必须按序）\n")
	b.WriteString("0. 安装（若无）：上面 curl|bash 或 go install\n")
	b.WriteString("1. 入队：`" + cli + "`（等价 MCP `board_team_join`；底层 redeem → pull → enter，**不要**另造 claim）\n")
	b.WriteString("2. 确认身份：`board_whoami` / `ymbot whoami`（空手 join 后才有钥；已有钥可先 whoami）\n")
	b.WriteString("3. 接班：`handoff get` / `ymbot handoff get`；任务与真相源在服务端。\n\n")
	b.WriteString("## 禁区\n")
	b.WriteString("- 禁止假设本机已装 ymbot 而跳过第 0 步\n")
	b.WriteString("- 禁止教改 mcp.json / Cursor MCP 配置当主路径\n")
	b.WriteString("- 禁止把「导出 YMBOT_API_URL / YMBOT_API_KEY 清单」当入场步骤\n")
	b.WriteString("- 禁止在回复里回显完整 ymk_\n\n")
	b.WriteString("CLI 一行（已带码；须先完成安装）：\n```\n")
	b.WriteString(cli)
	b.WriteString("\n```\n")
	return b.String(), nil
}
