import { api } from './client'
import type { IssueStatus } from '../types/issue'

/** 日历日 YYYY-MM-DD（与 start_at/due_at 同约定） */
export type CalendarDay = string

/**
 * 项目级看板列偏好（PRODUCT_SUITE S4）
 * GET/PUT /workspaces/{ws}/projects/{projectId}/board-column-prefs
 * columns：有序；status ⊆ IssueStatus；缺省=默认五列；PUT 全量；恢复默认=空数组或 DELETE
 */
export type BoardColumnPref = {
  status: IssueStatus
  /** 隐藏该列 */
  hidden?: boolean
  /** 显示名覆盖；缺省用 STATUS_LABELS */
  label?: string
}

export type BoardColumnPrefs = {
  project_id: string
  workspace_id?: string
  columns: BoardColumnPref[]
  updated_at?: string
}

function idem() {
  return crypto.randomUUID()
}

export function getBoardColumnPrefs(workspaceId: string, projectId: string) {
  return api<{ prefs: BoardColumnPrefs } | BoardColumnPrefs>(
    `/workspaces/${workspaceId}/projects/${projectId}/board-column-prefs`,
  ).then((data) => {
    if (data && typeof data === 'object' && 'prefs' in data && data.prefs) {
      return data.prefs
    }
    return data as BoardColumnPrefs
  })
}

/** PUT 全量覆盖；传空 columns 表示恢复默认（与后端删行/空行约定一致） */
export function putBoardColumnPrefs(
  workspaceId: string,
  projectId: string,
  columns: BoardColumnPref[],
) {
  return api<{ prefs: BoardColumnPrefs } | BoardColumnPrefs>(
    `/workspaces/${workspaceId}/projects/${projectId}/board-column-prefs`,
    {
      method: 'PUT',
      body: JSON.stringify({ columns }),
      idempotencyKey: idem(),
    },
  ).then((data) => {
    if (data && typeof data === 'object' && 'prefs' in data && data.prefs) {
      return data.prefs
    }
    return data as BoardColumnPrefs
  })
}

/** 恢复默认列（空 PUT；若后端提供 DELETE 也可） */
export function resetBoardColumnPrefs(workspaceId: string, projectId: string) {
  return putBoardColumnPrefs(workspaceId, projectId, [])
}
