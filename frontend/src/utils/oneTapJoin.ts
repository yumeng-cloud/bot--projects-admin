/**
 * ONE_TAP_JOIN · 入场链接 + Bot 提示词拼装（与 docs/NEXT_WAVE_DEBT_ONE_TAP_JOIN、
 * backend/internal/joinlink 契约对齐）。禁止把 ymk_ 写入 URL 或默认可转发提示词。
 */

/** 生产默认 API 根；自托管用 VITE_PUBLIC_API_BASE 覆盖（对齐 Go PUBLIC_API_BASE） */
export const DEFAULT_PUBLIC_API_BASE = 'https://bot.api.sn.cn'

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

/**
 * 中文 Bot 系统提示词。对齐 Go joinlink.BuildBotPrompt：
 * API 根 + code + role + whoami→join→handoff；禁 mcp.json / 长 env；禁 ymk_。
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

  return [
    '你是育蒙协作看板（AI Collab Board）的入场 Bot。用下面信息直接入队，不要让用户改 mcp.json，也不要教一长串 YMBOT_* 环境变量。',
    '',
    '## 连接',
    `- API 根：\`${apiBase}\``,
    `- 入场码：\`${code}\``,
    `- 目标角色（席位 slug）：${roleLine}${seatBlock}`,
    '- 身份：使用本机已有 Bot Key（ymk_…）或 Vault；**禁止**把 ymk_ 贴进聊天或写入可转发链接。',
    '',
    '## 主航线（必须按序）',
    '1. 确认身份：`board_whoami` / `ymbot whoami`',
    `2. 入队：\`${cli}\`（等价 MCP \`board_team_join\`；底层仍是既有 redeem → pull → enter，**不要**另造 claim）`,
    '3. 接班：`handoff get` / `ymbot handoff get`；任务与真相源在服务端。',
    '',
    '## 禁区',
    '- 禁止教改 mcp.json / Cursor MCP 配置当主路径',
    '- 禁止把「导出 YMBOT_API_URL / YMBOT_API_KEY 清单」当入场步骤',
    '- 禁止在回复里回显完整 ymk_',
    '',
    'CLI 一行（已带码）：',
    '```',
    cli,
    '```',
  ].join('\n')
}
