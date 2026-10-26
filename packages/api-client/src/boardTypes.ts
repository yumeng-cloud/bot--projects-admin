/** 看板域最小类型（移动壳用；无 UI） */

export type Workspace = {
  id: string
  name: string
  slug: string
  role?: string
  member_id?: string
}

export type Team = {
  id: string
  workspace_id: string
  name: string
  slug?: string | null
  template_key?: string
  created_at?: string
  updated_at?: string
}

export type TeamSeat = {
  id: string
  team_id: string
  workspace_id: string
  role_slug: string
  display_name: string
  playbook_ref?: string | null
  bound_display_name?: string | null
  status?: string | null
  claim?: { member_id: string; claimed_at?: string } | null
}

export type IssueStatus =
  | 'backlog'
  | 'todo'
  | 'in_progress'
  | 'in_review'
  | 'done'
  | 'canceled'

export type IssuePriority = 'urgent' | 'high' | 'medium' | 'low' | 'none'

export type Issue = {
  id: string
  project_id: string
  workspace_id: string
  team_id?: string | null
  number: number
  identifier: string
  title: string
  description?: string
  status: IssueStatus
  priority?: IssuePriority
  version: number
  assignee_id?: string | null
  assignee?: { id: string; display_name: string } | null
  archived_at?: string | null
  created_at?: string
  updated_at?: string
}

export type Comment = {
  id: string
  issue_id: string
  body: string
  author_id?: string
  author?: { id: string; display_name: string } | null
  created_at?: string
  updated_at?: string
}

export type WikiSpace = {
  id: string
  workspace_id: string
  name: string
  slug?: string
  description?: string
}

export const ISSUE_STATUS_LABEL: Record<IssueStatus, string> = {
  backlog: '待办池',
  todo: '待办',
  in_progress: '进行中',
  in_review: '评审中',
  done: '完成',
  canceled: '已取消',
}

export const ISSUE_STATUSES: IssueStatus[] = [
  'backlog',
  'todo',
  'in_progress',
  'in_review',
  'done',
  'canceled',
]
