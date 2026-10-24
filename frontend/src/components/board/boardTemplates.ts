/**
 * Board column templates + layout helpers.
 * Persistence: project board-column-prefs API (see api/boardColumnPrefs.ts).
 */

import type { IssueStatus } from '../../types/issue'
import { BOARD_COLUMNS, STATUS_LABELS } from '../../types/issue'
import type { BoardColumnPref } from '../../api/boardColumnPrefs'

export type BoardColumnDef = {
  status: IssueStatus
  /** Display label override; falls back to STATUS_LABELS */
  label?: string
  hidden?: boolean
}

/** Local UI layout (templateId is client-only; API has no templateId). */
export type BoardColumnLayout = {
  /** Built-in template id, or `custom` */
  templateId: string
  columns: BoardColumnDef[]
}

export type BuiltInTemplate = {
  id: string
  name: string
  description: string
  system: true
  columns: BoardColumnDef[]
}

/** Built-in packs — status ⊆ existing IssueStatus; no new enums. */
export const BUILTIN_TEMPLATES: BuiltInTemplate[] = [
  {
    id: 'default-5',
    name: '默认五列',
    description: '待办池 → 待处理 → 进行中 → 评审中 → 已完成',
    system: true,
    columns: BOARD_COLUMNS.map((status) => ({ status })),
  },
  {
    id: 'simple-3',
    name: '精简三列',
    description: '待处理 → 进行中 → 已完成',
    system: true,
    columns: [
      { status: 'todo' },
      { status: 'in_progress' },
      { status: 'done' },
    ],
  },
  {
    id: 'bug-triage',
    name: '缺陷分诊',
    description: '待办池 → 待处理 → 进行中 → 已完成',
    system: true,
    columns: [
      { status: 'backlog', label: '待分诊' },
      { status: 'todo', label: '确认' },
      { status: 'in_progress', label: '修复中' },
      { status: 'done', label: '已关闭' },
    ],
  },
  {
    id: 'sprint-flow',
    name: 'Sprint 流',
    description: '待办池 → 待处理 → 进行中 → 评审中 → 已完成',
    system: true,
    columns: [
      { status: 'backlog', label: 'Backlog' },
      { status: 'todo', label: 'Todo' },
      { status: 'in_progress', label: 'In Progress' },
      { status: 'in_review', label: 'In Review' },
      { status: 'done', label: 'Done' },
    ],
  },
]

export const DEFAULT_TEMPLATE_ID = 'default-5'

export function defaultPrefs(): BoardColumnLayout {
  const t = BUILTIN_TEMPLATES.find((x) => x.id === DEFAULT_TEMPLATE_ID)!
  return { templateId: t.id, columns: t.columns.map((c) => ({ ...c })) }
}

export function visibleColumns(prefs: BoardColumnLayout): BoardColumnDef[] {
  return prefs.columns.filter((c) => !c.hidden)
}

export function columnLabel(def: BoardColumnDef): string {
  return (def.label && def.label.trim()) || STATUS_LABELS[def.status]
}

export function prefsFromTemplate(templateId: string): BoardColumnLayout {
  const t = BUILTIN_TEMPLATES.find((x) => x.id === templateId)
  if (!t) return defaultPrefs()
  return { templateId: t.id, columns: t.columns.map((c) => ({ ...c })) }
}

function normLabel(label?: string): string {
  return (label ?? '').trim()
}

function columnsEqual(a: BoardColumnDef[], b: BoardColumnDef[]): boolean {
  if (a.length !== b.length) return false
  return a.every((c, i) => {
    const o = b[i]
    return (
      c.status === o.status &&
      !!c.hidden === !!o.hidden &&
      normLabel(c.label) === normLabel(o.label)
    )
  })
}

/** Infer built-in template id from column list (status order + labels); else `custom`. */
export function inferTemplateId(columns: BoardColumnDef[]): string {
  for (const t of BUILTIN_TEMPLATES) {
    if (columnsEqual(columns, t.columns)) return t.id
  }
  return 'custom'
}

export function isDefaultLayout(layout: BoardColumnLayout): boolean {
  const d = defaultPrefs()
  return layout.templateId === d.templateId && columnsEqual(layout.columns, d.columns)
}

/** Sanitize API columns → layout; empty/missing → default five. */
export function layoutFromApiColumns(
  columns: BoardColumnPref[] | null | undefined,
): BoardColumnLayout {
  if (!columns || columns.length === 0) return defaultPrefs()
  const allowed = new Set<string>(BOARD_COLUMNS)
  const sanitized: BoardColumnDef[] = []
  for (const c of columns) {
    if (!c || !allowed.has(c.status)) continue
    const def: BoardColumnDef = { status: c.status }
    if (c.hidden) def.hidden = true
    const label = c.label?.trim()
    if (label) def.label = label
    sanitized.push(def)
  }
  if (sanitized.length === 0) return defaultPrefs()
  return {
    templateId: inferTemplateId(sanitized),
    columns: sanitized,
  }
}

/** Layout → API columns payload (status, optional hidden/label only). */
export function layoutToApiColumns(layout: BoardColumnLayout): BoardColumnPref[] {
  return layout.columns.map((c) => {
    const out: BoardColumnPref = { status: c.status }
    if (c.hidden) out.hidden = true
    const label = c.label?.trim()
    if (label) out.label = label
    return out
  })
}
