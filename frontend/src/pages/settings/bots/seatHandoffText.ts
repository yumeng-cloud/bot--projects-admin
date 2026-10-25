import type { BotHandoff, BotMemory, BotRole, BotTodo, HandoffIssue } from '../../../api/bots'
import { cleanSeatSlug, isDirtySeatKey } from '../../../utils/seatLabels'
import { formatShanghai } from '../../../utils/time'

/** 从 memories 取出人话焦点：优先 slug:last_focus，再 last_focus */
export function pickLastFocus(memories: BotMemory[] | undefined, roleSlug?: string | null): string | null {
  if (!memories?.length) return null
  const slugKey = roleSlug ? `${roleSlug}:last_focus` : null
  const bySlug = slugKey ? memories.find((m) => m.key_name === slugKey) : undefined
  if (bySlug?.value_text?.trim()) return bySlug.value_text.trim()
  const plain = memories.find((m) => m.key_name === 'last_focus' || m.key_name.endsWith(':last_focus'))
  return plain?.value_text?.trim() || null
}

export function openTodos(todos: BotTodo[] | undefined): BotTodo[] {
  return (todos || []).filter((t) => t.status === 'todo' || t.status === 'doing')
}

const STATUS_CN: Record<string, string> = {
  todo: '待办',
  doing: '进行中',
  done: '已完成',
  canceled: '已取消',
  in_progress: '进行中',
  in_review: '评审中',
  backlog: '待排期',
  done_status: '已完成',
  canceled_status: '已取消',
  triage: '分诊',
}

export function statusCn(s: string): string {
  return STATUS_CN[s] || s
}

export type SeatCardMeta = {
  roleName: string
  roleSlug?: string
  boundLabel: string /** 已绑显示名或「未进席」/「可补位」 */
  keyHint?: string /** 仅末四 / 前缀尾，勿完整 Key */
  claimedAt?: string | null
}

/** 只读纯文本交接卡；禁止完整 API Key */
export function buildHandoffCardText(opts: {
  seat: SeatCardMeta
  handoff: BotHandoff | null
  focus: string | null
  todos: BotTodo[]
  issues: HandoffIssue[]
}): string {
  const { seat, handoff, focus, todos, issues } = opts
  const lines: string[] = []
  lines.push(`【席位交接卡】${seat.roleName}`)
  lines.push(`绑定：${seat.boundLabel}${seat.keyHint ? ` · Key ${seat.keyHint}` : ''}`)
  if (seat.claimedAt) lines.push(`认领时间：${formatShanghai(seat.claimedAt)}`)
  if (handoff?.generated_at) lines.push(`摘要生成：${formatShanghai(handoff.generated_at)}`)
  lines.push('')
  lines.push(`当前焦点：${focus || '（暂无）'}`)
  lines.push('')
  lines.push(`进行中 Todos（${todos.length}）`)
  if (todos.length) {
    for (const t of todos) {
      lines.push(`  · [${statusCn(t.status)}] ${t.title}`)
    }
  } else {
    lines.push('  （无）')
  }
  lines.push('')
  lines.push(`开放 Issues（${issues.length}）`)
  if (issues.length) {
    for (const i of issues) {
      lines.push(`  · ${i.identifier} ${i.title}（${statusCn(i.status)}）`)
    }
  } else {
    lines.push('  （无）')
  }
  lines.push('')
  lines.push('— 本卡不含 API Key；换机请：绑同一席位 → ymbot handoff get')
  return lines.join('\n')
}

export function roleDisplay(role: BotRole): string {
  const name = (role.name || '').trim()
  if (name && !isDirtySeatKey(name)) return name
  const clean = cleanSeatSlug(role.slug)
  // 人话优先；短 slug 仅作无 name 时的兜底，永不渲染 team:uuid:…
  if (clean) return clean
  return '未命名席位'
}

/** SEAT_PANEL_UX P-R1：席卡三态（只认本队 roster；禁可补≈空闲） */
export function seatOccupancyTagLabel(
  kind: 'active' | 'empty' | 'left',
  occupantName?: string | null,
): string {
  if (kind === 'active') {
    const name = (occupantName || '').trim()
    return name ? `在席 · ${name}` : '在席'
  }
  if (kind === 'left') return '停工可补'
  return '空闲（从未占用）'
}
