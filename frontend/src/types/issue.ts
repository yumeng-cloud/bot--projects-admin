/** Issue / Cycle / Label — 对齐 backend schema（snake_case） */

export type IssueStatus =
  | 'backlog'
  | 'todo'
  | 'in_progress'
  | 'in_review'
  | 'done'
  | 'canceled'

export type IssuePriority = 'urgent' | 'high' | 'medium' | 'low' | 'none'

export type CycleStatus = 'planned' | 'active' | 'completed'

export interface Label {
  id: string
  workspace_id: string
  name: string
  color: string
}

export interface MemberRef {
  id: string
  display_name: string
  email?: string | null
  /** @deprecated mock / legacy — prefer display_name */
  name?: string
}

/** Human-readable member label — never raw id/UUID. */
export function memberLabel(
  m: Pick<MemberRef, 'display_name' | 'email' | 'name'> | null | undefined,
): string {
  if (!m) return '成员'
  return (m.display_name || m.email || m.name || '成员').trim() || '成员'
}

export interface Cycle {
  id: string
  project_id: string
  name: string
  start_date: string | null
  end_date: string | null
  capacity: number
  status: CycleStatus
  created_at?: string
  updated_at?: string
}

export interface Issue {
  id: string
  project_id: string
  workspace_id: string
  number: number
  identifier: string
  title: string
  description: string
  status: IssueStatus
  priority: IssuePriority
  estimate: number | null
  assignee_id: string | null
  /** 嵌套负责人；与 assignee_id 并存，可空 */
  assignee?: MemberRef | null
  reporter_id: string | null
  /** 嵌套报告人；与 reporter_id 并存，可空 */
  reporter?: MemberRef | null
  cycle_id: string | null
  parent_id: string | null
  board_order: number
  version: number
  /** 排期起日：日历日 YYYY-MM-DD（Asia/Shanghai 显）；可空 */
  start_at?: string | null
  /** 排期止日：YYYY-MM-DD；与 start_at 同有时须 >= start_at */
  due_at?: string | null
  /** S6：无子节点则为叶子 */
  is_leaf?: boolean
  /** S6：直接子节点数 */
  child_count?: number
  labels?: Label[]
  created_at: string
  updated_at: string
}

export const ISSUE_STATUSES: IssueStatus[] = [
  'backlog',
  'todo',
  'in_progress',
  'in_review',
  'done',
  'canceled',
]

export const STATUS_LABELS: Record<IssueStatus, string> = {
  backlog: '待办池',
  todo: '待处理',
  in_progress: '进行中',
  in_review: '评审中',
  done: '已完成',
  canceled: '已取消',
}

export const PRIORITY_LABELS: Record<IssuePriority, string> = {
  urgent: '紧急',
  high: '高',
  medium: '中',
  low: '低',
  none: '无',
}

export const BOARD_COLUMNS: IssueStatus[] = [
  'backlog',
  'todo',
  'in_progress',
  'in_review',
  'done',
]


export interface Comment {
  id: string
  issue_id: string
  author_id: string
  /** 嵌套作者（N3）；优先于 memberName(author_id) */
  author?: { id?: string; display_name?: string } | null
  body: string
  created_at: string
  updated_at: string
}

export type RelationType = 'blocks' | 'blocked_by' | 'related' | 'parent'

export const RELATION_TYPE_LABELS: Record<RelationType, string> = {
  blocks: '阻塞',
  blocked_by: '被阻塞',
  related: '相关',
  parent: '父任务',
}

export interface IssueRelation {
  id: string
  workspace_id: string
  from_issue_id: string
  to_issue_id: string
  relation_type: RelationType
  created_at: string
}

export interface Activity {
  id: string
  workspace_id: string
  issue_id?: string | null
  actor_id?: string | null
  actor_kind: string
  action: string
  payload: string
  created_at: string
}

/**
 * Issue / Comment 附件 — 对齐 Go 后端契约 v1（snake_case）。
 * content 相对 url 须带 Bearer，禁止无头 window.open。
 */
export interface IssueAttachment {
  id: string
  workspace_id: string
  issue_id: string
  /** 评论附件时有值；Issue 级为 null/undefined */
  comment_id?: string | null
  uploader_id: string
  filename: string
  content_type: string
  size_bytes: number
  /** @deprecated legacy alias — prefer size_bytes; list UI maps via attachmentSizeBytes() */
  size?: number
  created_at: string
  /** 相对路径，如 /api/v1/workspaces/{ws}/attachments/{id}/content；请求须带 Bearer */
  url: string
}

/** Normalize byte size from API (`size_bytes`) or legacy `size`. */
export function attachmentSizeBytes(
  a: Pick<IssueAttachment, 'size_bytes'> & { size?: number },
): number {
  const n = a.size_bytes
  if (typeof n === 'number' && Number.isFinite(n) && n >= 0) return n
  const legacy = a.size
  if (typeof legacy === 'number' && Number.isFinite(legacy) && legacy >= 0) return legacy
  return 0
}
