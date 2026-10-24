/** UI copy helpers — never leak raw slugs / hex ids to humans. */

const ROLE_LABELS: Record<string, string> = {
  admin: '管理员',
  owner: '所有者',
  member: '成员',
  bot: '机器人',
  viewer: '访客',
}

const CYCLE_STATUS_LABELS: Record<string, string> = {
  planned: '计划中',
  active: '进行中',
  completed: '已结束',
}

export function roleLabel(role: string | null | undefined): string {
  const key = (role || '').trim().toLowerCase()
  if (!key) return '成员'
  return ROLE_LABELS[key] || role || '成员'
}

export function cycleStatusLabel(status: string | null | undefined): string {
  const key = (status || '').trim().toLowerCase()
  if (!key) return '—'
  return CYCLE_STATUS_LABELS[key] || status || '—'
}

/** Prefer YM-n (+ title); never show bare uuid/hex. */
export function issueReadable(
  ident: string | null | undefined,
  title?: string | null,
  fallbackId?: string | null,
): string {
  const idPart = (ident || '').trim()
  const titlePart = (title || '').trim()
  if (idPart && titlePart) return `${idPart} ${titlePart}`
  if (idPart) return idPart
  if (titlePart) return titlePart
  const raw = (fallbackId || '').trim()
  if (raw && /^[0-9a-f-]{16,}$/i.test(raw)) return '未命名 Issue'
  return raw || '未命名 Issue'
}
