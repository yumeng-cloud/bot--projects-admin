export type User = {
  id: string
  email: string
  name: string
  avatar_url?: string | null
  created_at?: string
}

export type Workspace = {
  id: string
  name: string
  slug: string
  role?: string
  member_id?: string
}

export type Project = {
  id: string
  workspace_id: string
  name: string
  key_prefix: string
  description?: string
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
  number: number
  identifier: string
  title: string
  description?: string
  status: IssueStatus
  priority: IssuePriority
  estimate?: number | null
  assignee_id?: string | null
  cycle_id?: string | null
  board_order: number
  version: number
  created_at?: string
  updated_at?: string
}

export type ApiErrorBody = { code: string; message: string }
