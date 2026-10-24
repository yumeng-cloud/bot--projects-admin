/**
 * ONE_TAP_JOIN · 入场链接 + Bot 提示词拼装（与 docs/NEXT_WAVE_DEBT_ONE_TAP_JOIN、
 * backend/internal/joinlink 契约对齐）。禁止把 ymk_ 写入 URL 或默认可转发提示词。
 * E4：与 Go BuildBotPrompt（4cdaa04）同构——第 0 步安装；禁假设已装 ymbot。
 * FOOLPROOF：仅 buildFirstRunPrompt（创建成功私贴）可嵌当次 ymk_；席页仍禁。
 */

/** 生产默认 API 根；自托管用 VITE_PUBLIC_API_BASE 覆盖（对齐 Go PUBLIC_API_BASE） */
export const DEFAULT_PUBLIC_API_BASE = 'https://bot.api.sn.cn'

/** 备路径：对齐 installscript.GoInstallCmd / go.mod module */
export const GO_INSTALL_CMD =
  'go install github.com/yumeng-tech/ai-collab-board/backend/cmd/ymbot@latest'

export const INVITE_CODE_PREFIX = 'ymt_'

export function getPublicApiBase(): string {
  const fromEnv =
    (typeof import.meta !== 'undefined' &&
      (import.meta.env?.VITE_PUBLIC_API_BASE as string | undefined)?.trim()) ||
    ''
  const raw = fromEnv || DEFAULT_PUBLIC_API_BASE
  return raw.replace(/\/+$/, '')
}

/** 对齐 Go ValidateInviteCode；非法抛中文 Error */
export function validateInviteCode(code: string): void {
  const c = (code || '').trim()
  if (!c) throw new Error('缺少入场码：请带上完整的 ymt_… 码')
  if (!c.startsWith(INVITE_CODE_PREFIX)) {
    throw new Error(
      `入场码无效：须以 ${INVITE_CODE_PREFIX} 开头（团队入场码，不是工作区邀请或 Bot Key）`,
    )
  }
  if (c.length <= INVITE_CODE_PREFIX.length) {
    throw new Error('入场码不完整：请确认复制了完整的 ymt_… 码')
  }
}

export type JoinUrlParams = {
  code: string
  role?: string | null
  seat?: string | null
  api?: string | null
  base?: string | null
}

/** 拼规范 join URL：{base}/join?code=ymt_…&role=… */
export function buildJoinUrl(params: JoinUrlParams): string {
  validateInviteCode(params.code)
  const code = params.code.trim()
  const base = (params.base || getPublicApiBase()).replace(/\/+$/, '')
  const q = new URLSearchParams()
  q.set('code', code)
  const role = (params.role || '').trim()
  if (role) q.set('role', role)
  const seat = (params.seat || '').trim()
  if (seat) q.set('seat', seat)
  const apiOverride = (params.api || '').trim().replace(/\/+$/, '')
  if (apiOverride && apiOverride !== DEFAULT_PUBLIC_API_BASE) {
    q.set('api', apiOverride)
  }
  return `${base}/join?${q.toString()}`
}

export type BotJoinPromptParams = {
  code: string
  role?: string | null
  roleDisplayName?: string | null
  seat?: string | null
  apiBase?: string | null
  api?: string | null
}

/** 码末四（去 ymt_ 前缀），供 Toast */
export function inviteCodeTail(code: string): string {
  const clean = (code || '').replace(/^ymt_/i, '').replace(/[^a-zA-Z0-9]/g, '')
  return clean.slice(-4) || '····'
}

/** `ymbot team join --code … [--role …]`（对齐 Go CLIJoinLine） */
export function cliJoinLine(params: { code: string; role?: string | null }): string {
  validateInviteCode(params.code)
  const code = params.code.trim()
  const role = (params.role || '').trim()
  return role
    ? `ymbot team join --code ${code} --role ${role}`
    : `ymbot team join --code ${code}`
}

/** 对齐 Go InstallCurlLine：`curl -fsSL {api}/install.sh | bash` */
export function installCurlLine(apiBase?: string | null): string {
  const api = (apiBase || getPublicApiBase()).replace(/\/+$/, '')
  return `curl -fsSL ${api}/install.sh | bash`
}

/**
 * 中文 Bot 系统提示词。与 Go joinlink.BuildBotPrompt 同构（E1/E4）：
 * 首行 curl|bash；§0 安装（主 curl|bash + 备 go install）；API 根 + code + role；
 * 主航线 0 安装→join→whoami→handoff；身份两态（空手自动得钥 / 已有钥带 Bearer）；禁假设已装 ymbot / mcp.json / 长 env。
 */
export function buildBotJoinPrompt(params: BotJoinPromptParams): string {
  validateInviteCode(params.code)
  const code = params.code.trim()
  const apiOverride = (params.api || '').trim().replace(/\/+$/, '')
  const apiBase = (
    apiOverride ||
    (params.apiBase || '').trim() ||
    getPublicApiBase()
  ).replace(/\/+$/, '')
  const role = (params.role || '').trim()
  const roleName = (params.roleDisplayName || '').trim()
  const seat = (params.seat || '').trim()

  // Go 默认只写 `slug`；FE 有显示名时附带（席页传入），结构同构
  let roleLine = '（未指定；join 时按新人自动选席）'
  if (role) {
    roleLine =
      roleName && roleName !== role
        ? `${roleName}（slug: \`${role}\`）`
        : `\`${role}\``
  }
  const seatBlock = seat
    ? `\n- 指定席位 ID（高级，优先于 role）：\`${seat}\``
    : ''
  const cli = cliJoinLine({ code, role: role || null })
  const installCurl = installCurlLine(apiBase)

  return [
    installCurl,
    '',
    '你是育蒙协作看板（AI Collab Board）的入场 Bot。对方**未必已装** `ymbot`：先安装再入队。不要让用户改 mcp.json，也不要教一长串 YMBOT_* 环境变量。',
    '',
    '## 0. 一键安装（强制·若本机还没有 ymbot）',
    '主路径（复制执行）：',
    '```',
    installCurl,
    '```',
    '备路径（已装 Go 时）：',
    '```',
    GO_INSTALL_CMD,
    '```',
    '装完自检：`ymbot -h`（或 `ymbot version`）。若 `~/.local/bin` 不在 PATH，按安装脚本提示 export。',
    '',
    '## 连接',
    `- API 根：\`${apiBase}\`（whoami / join / handoff 都打这个根；可用环境变量 \`YMBOT_API_URL=${apiBase}\`，勿把 Key 清单当主教程）`,
    `- 入场码：\`${code}\``,
    `- 目标角色（席位 slug）：${roleLine}${seatBlock}`,
    '- 身份两态：',
    '  - **空手**：无本机 ymk_ 也可直接 join；成功响应会**自动发一把**身份钥匙（仅此一次，私存；**禁止**贴进聊天或写入可转发链接）',
    '  - **已有钥**：续队/换机带上原 `ymk_`（Bearer / `YMBOT_API_KEY`）；**不重发**钥匙',
    '',
    '## 主航线（必须按序）',
    '0. 安装（若无）：上面 curl|bash 或 go install',
    `1. 入队：\`${cli}\`（等价 MCP \`board_team_join\`；底层 redeem → pull → enter，**不要**另造 claim）`,
    '2. 确认身份：`board_whoami` / `ymbot whoami`（空手 join 后才有钥；已有钥可先 whoami）',
    '3. 接班：`handoff get` / `ymbot handoff get`；任务与真相源在服务端。',
    '',
    '## 禁区',
    '- 禁止假设本机已装 ymbot 而跳过第 0 步',
    '- 禁止教改 mcp.json / Cursor MCP 配置当主路径',
    '- 禁止把「导出 YMBOT_API_URL / YMBOT_API_KEY 清单」当入场步骤',
    '- 禁止在回复里回显完整 ymk_',
    '',
    'CLI 一行（已带码；须先完成安装）：',
    '```',
    cli,
    '```',
  ].join('\n')
}

export type FirstRunPromptParams = {
  /** 当次明文身份钥匙 ymk_…（仅首跑启动提示词可嵌） */
  token: string
  apiBase?: string | null
  /** 可选：已签发入场码；无则提示另发入场链接 */
  inviteCode?: string | null
  role?: string | null
}

/**
 * 首跑启动提示词（FOOLPROOF F3）：可嵌当次 ymk_，仅 Admin→本机 Agent 私贴。
 * 与席页 buildBotJoinPrompt 分家——后者禁嵌 ymk_。
 */
export function buildFirstRunPrompt(params: FirstRunPromptParams): string {
  const token = (params.token || '').trim()
  if (!token) throw new Error('缺少身份钥匙')
  if (!token.startsWith('ymk_')) {
    throw new Error('身份钥匙无效：须以 ymk_ 开头')
  }
  const apiBase = ((params.apiBase || '').trim() || getPublicApiBase()).replace(/\/+$/, '')
  const installCurl = installCurlLine(apiBase)
  const invite = (params.inviteCode || '').trim()
  const role = (params.role || '').trim()

  const lines: string[] = [
    installCurl,
    '',
    '你是育蒙协作看板（AI Collab Board）的首跑启动助手。本提示含**身份钥匙**，仅私贴本机 Agent；**禁止转发公开群**。',
    '',
    '## 0. 一键安装（若本机还没有 ymbot）',
    '主路径：',
    '```',
    installCurl,
    '```',
    '备路径（已装 Go）：',
    '```',
    GO_INSTALL_CMD,
    '```',
    '装完自检：`ymbot -h`（或 `ymbot version`）。',
    '',
    '## 1. API 根',
    `\`${apiBase}\``,
    `可用：\`export YMBOT_API_URL=${apiBase}\``,
    '',
    '## 2. 身份钥匙（仅此一次·勿转发）',
    '```',
    `export YMBOT_API_KEY=${token}`,
    '```',
    '',
    '## 3. 确认身份',
    '`ymbot whoami`（或 MCP `board_whoami`）',
    '',
  ]

  if (invite && invite.startsWith(INVITE_CODE_PREFIX)) {
    const cli = cliJoinLine({ code: invite, role: role || null })
    lines.push(
      '## 4. 入队（可选·本提示已附入场码）',
      `- 入场码：\`${invite}\``,
      '```',
      cli,
      '```',
      '',
    )
  } else {
    lines.push(
      '## 4. 入队',
      '本提示**未**附入场码。入队另发「入场链接 / 入场提示词」（含 `ymt_`）；身份钥匙与入场码两套，勿混用。',
      '',
    )
  }

  lines.push(
    '## 警告',
    '- **含密钥，勿发公开群 / 勿写入可转发入场 URL**',
    '- 日常席页「复制入场链接/提示词」只含 `ymt_`，不含本钥',
    '- 列表此后只留末四；丢钥请重新签发身份钥匙',
  )

  return lines.join('\n')
}

export type ParsedJoinQuery = {
  code: string
  role: string
  seat: string
  api: string
}

/** 从 location.search / URLSearchParams 解析 join 深链参（C1） */
export function parseJoinSearchParams(search: string | URLSearchParams): ParsedJoinQuery {
  const q =
    typeof search === 'string'
      ? new URLSearchParams(search.startsWith('?') ? search.slice(1) : search)
      : search
  return {
    code: (q.get('code') || '').trim(),
    role: (q.get('role') || '').trim(),
    seat: (q.get('seat') || '').trim(),
    api: (q.get('api') || '').trim(),
  }
}

/** 同源 FE 深链路径：/join?code=ymt_…&role=…（对齐 Go PUBLIC_BOARD_BASE 提示） */
export function buildBoardJoinPath(params: JoinUrlParams): string {
  validateInviteCode(params.code)
  const q = new URLSearchParams()
  q.set('code', params.code.trim())
  const role = (params.role || '').trim()
  if (role) q.set('role', role)
  const seat = (params.seat || '').trim()
  if (seat) q.set('seat', seat)
  const apiOverride = (params.api || '').trim().replace(/\/+$/, '')
  if (apiOverride && apiOverride !== DEFAULT_PUBLIC_API_BASE) {
    q.set('api', apiOverride)
  }
  return `/join?${q.toString()}`
}

