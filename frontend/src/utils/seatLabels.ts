/**
 * INVITE IQ1-3：席位人话 label / 脏串清洗。
 * bot_roles.slug 可能是 team:{teamId}:{short}；team_seats.role_slug 才是短 slug。
 * 用户可见文案（Toast / Tooltip / 列表分组）永不渲染 team:… / 裸 UUID。
 */
import type { TeamSeat } from '../api/teams'

const UUID_RE =
  /[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/i

/** 脏串：team:uuid:slug、裸 UUID、含 UUID 的键 */
export function isDirtySeatKey(key: string): boolean {
  const k = (key || '').trim()
  if (!k || k === '__none__') return false
  if (/^team:/i.test(k)) return true
  if (UUID_RE.test(k)) return true
  return false
}

/**
 * 抽出可签发的短 slug：
 * - `team:uuid:wiki` → `wiki`
 * - 已是短 slug → 原样
 * - 裸 UUID / 空 / 无法解析 → `''`（勿把 UUID 当 slug）
 */
export function cleanSeatSlug(raw: string | null | undefined): string {
  const s = (raw || '').trim()
  if (!s) return ''
  if (/^team:/i.test(s)) {
    const parts = s.split(':')
    // team : uuid : slug… → 取末段
    const last = (parts[parts.length - 1] || '').trim()
    if (!last || UUID_RE.test(last)) return ''
    return last.toLowerCase()
  }
  if (UUID_RE.test(s)) return ''
  return s.toLowerCase()
}

/** 席表上解析真正的短 role_slug（优先 seat.role_slug） */
export function resolveSeatCleanSlug(
  seat: TeamSeat | null | undefined,
  fallbackSlug?: string | null,
): string {
  const fromSeat = cleanSeatSlug(seat?.role_slug)
  if (fromSeat) return fromSeat
  return cleanSeatSlug(fallbackSlug)
}

/**
 * 人话席名：display_name → 席表匹配 →「未命名席」。
 * 永不返回 team:… / UUID。
 */
export function humanSeatLabel(opts: {
  displayName?: string | null
  seat?: TeamSeat | null
  slug?: string | null
  seats?: TeamSeat[]
}): string {
  const tryName = (v: string | null | undefined): string | null => {
    const t = (v || '').trim()
    if (!t || isDirtySeatKey(t)) return null
    return t
  }

  const fromDn = tryName(opts.displayName)
  if (fromDn) return fromDn

  const fromSeat = tryName(opts.seat?.display_name)
  if (fromSeat) return fromSeat

  const clean = resolveSeatCleanSlug(opts.seat, opts.slug)
  if (clean && opts.seats?.length) {
    const hit = opts.seats.find(
      (s) => cleanSeatSlug(s.role_slug) === clean || (s.role_slug || '').trim() === clean,
    )
    const n = tryName(hit?.display_name)
    if (n) return n
  }

  // 脏串 / 对不上 → 未命名；短 slug 也不直接出（产品债：人话优先）
  if (opts.slug && isDirtySeatKey(opts.slug)) return '未命名席'
  if (opts.seat?.role_slug && isDirtySeatKey(opts.seat.role_slug) && !fromSeat) {
    return '未命名席'
  }
  return '未命名席'
}

/** 已签发列表分组标题：脏串 / 对不上 →「未命名席」 */
export function seatGroupLabel(key: string, seats: TeamSeat[]): string {
  if (key === '__none__' || !key) return '未指定席'
  const clean = cleanSeatSlug(key)
  const hit =
    seats.find((s) => (s.role_slug || '').trim() === key) ||
    (clean ? seats.find((s) => cleanSeatSlug(s.role_slug) === clean) : undefined) ||
    seats.find((s) => s.id === key) ||
    undefined
  if (hit) {
    const n = (hit.display_name || '').trim()
    if (n && !isDirtySeatKey(n)) return n
    const slug = cleanSeatSlug(hit.role_slug)
    // 有席但无显示名：仍「未命名席」，勿出脏串
    if (slug && !isDirtySeatKey(slug)) {
      // 对得上席但无 display_name —— 产品要求对不上才未命名；有席无 name 也算未命名展示
      return '未命名席'
    }
    return '未命名席'
  }
  if (isDirtySeatKey(key)) return '未命名席'
  // 干净 key 但不在席表
  return '未命名席'
}
