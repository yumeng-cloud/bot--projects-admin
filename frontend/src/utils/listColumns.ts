/** PRO_ADMIN C3：列表次要列显隐，localStorage 按 ws 持久 */

export type SecondaryColumnKey = 'labels' | 'start_at' | 'due_at' | 'estimate'

export type ListColumnVisibility = Record<SecondaryColumnKey, boolean>

/** 默认专业子集：次要列全隐（对标 Linear 密度） */
export const DEFAULT_LIST_COLUMN_VISIBILITY: ListColumnVisibility = {
  labels: false,
  start_at: false,
  due_at: false,
  estimate: false,
}

export const SECONDARY_COLUMN_OPTIONS: {
  key: SecondaryColumnKey
  label: string
}[] = [
  { key: 'labels', label: '标签' },
  { key: 'start_at', label: '开始' },
  { key: 'due_at', label: '截止' },
  { key: 'estimate', label: '估时' },
]

export function listColumnsStorageKey(workspaceId: string): string {
  return `acb.list.columns.${workspaceId}`
}

export function readListColumnVisibility(workspaceId: string): ListColumnVisibility {
  const base = { ...DEFAULT_LIST_COLUMN_VISIBILITY }
  if (!workspaceId) return base
  try {
    const raw = localStorage.getItem(listColumnsStorageKey(workspaceId))
    if (!raw) return base
    const parsed = JSON.parse(raw) as Partial<ListColumnVisibility>
    for (const opt of SECONDARY_COLUMN_OPTIONS) {
      if (typeof parsed[opt.key] === 'boolean') {
        base[opt.key] = parsed[opt.key] as boolean
      }
    }
    return base
  } catch {
    return base
  }
}

export function writeListColumnVisibility(
  workspaceId: string,
  visibility: ListColumnVisibility,
): void {
  if (!workspaceId) return
  try {
    localStorage.setItem(listColumnsStorageKey(workspaceId), JSON.stringify(visibility))
  } catch {
    /* ignore quota / private mode */
  }
}
