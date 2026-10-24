import { api } from './client'

const ws = (id: string) => `/workspaces/${id}`

export type BotRoleClaim = {
  id: string
  workspace_id: string
  role_id: string
  member_id: string
  claimed_at: string
  released_at?: string | null
}

export type BotRole = {
  id: string
  workspace_id: string
  name: string
  slug: string
  responsibilities: string
  skills: string
  default_project_ids: string
  created_at: string
  updated_at: string
  claim?: BotRoleClaim | null
}

export type BotMemory = {
  id: string
  workspace_id: string
  member_id: string
  role_id?: string | null
  key_name: string
  value_text: string
  tags: string
  created_at: string
  updated_at: string
}

export type BotTodo = {
  id: string
  workspace_id: string
  member_id: string
  role_id?: string | null
  title: string
  status: 'todo' | 'doing' | 'done' | 'canceled'
  issue_id?: string | null
  sort_order: number
  created_at: string
  updated_at: string
}

export type BotRosterItem = {
  member_id: string
  display_name: string
  bot_enabled: boolean
  role: string
  claimed_role?: BotRole | null
}

export type HandoffIssue = {
  id: string
  identifier: string
  title: string
  status: string
  priority: string
  project_id: string
  cycle_id?: string | null
  updated_at: string
}

export type BotHandoff = {
  workspace_id: string
  member_id: string
  generated_at: string
  claimed_role?: BotRole | null
  memories: BotMemory[]
  todos: BotTodo[]
  issues: HandoffIssue[]
}

export function listBotRoles(workspaceId: string) {
  return api<{ bot_roles: BotRole[] }>(`${ws(workspaceId)}/bot-roles`)
}

export function createBotRole(
  workspaceId: string,
  body: { name: string; slug: string; responsibilities?: string; skills?: string },
  idempotencyKey?: string,
) {
  return api<{ bot_role: BotRole }>(`${ws(workspaceId)}/bot-roles`, {
    method: 'POST',
    body: JSON.stringify(body),
    idempotencyKey,
  })
}

export function claimBotRole(
  workspaceId: string,
  roleId: string,
  memberId?: string,
  idempotencyKey?: string,
  opts?: { skipToast?: boolean },
) {
  return api<{ claim: BotRoleClaim }>(`${ws(workspaceId)}/bot-roles/${roleId}/claim`, {
    method: 'POST',
    body: JSON.stringify(memberId ? { member_id: memberId } : {}),
    idempotencyKey,
    skipToast: opts?.skipToast,
  })
}

export function releaseBotRole(workspaceId: string, roleId: string, idempotencyKey?: string) {
  return api<{ message: string }>(`${ws(workspaceId)}/bot-roles/${roleId}/release`, {
    method: 'POST',
    body: '{}',
    idempotencyKey,
  })
}

export function listBots(workspaceId: string) {
  return api<{ bots: BotRosterItem[] }>(`${ws(workspaceId)}/bots`)
}

export function listMemories(workspaceId: string, memberId: string, q?: string) {
  const qs = q ? `?q=${encodeURIComponent(q)}` : ''
  return api<{ memories: BotMemory[] }>(`${ws(workspaceId)}/bots/${memberId}/memories${qs}`)
}

export function upsertMemory(
  workspaceId: string,
  memberId: string,
  body: { key_name: string; value_text: string; tags?: string; role_id?: string },
  idempotencyKey?: string,
) {
  return api<{ memory: BotMemory }>(`${ws(workspaceId)}/bots/${memberId}/memories`, {
    method: 'POST',
    body: JSON.stringify(body),
    idempotencyKey,
  })
}

export function patchMemory(
  workspaceId: string,
  memberId: string,
  memoryId: string,
  body: Partial<{ key_name: string; value_text: string; tags: string; role_id: string }>,
  idempotencyKey?: string,
) {
  return api<{ memory: BotMemory }>(`${ws(workspaceId)}/bots/${memberId}/memories/${memoryId}`, {
    method: 'PATCH',
    body: JSON.stringify(body),
    idempotencyKey,
  })
}

export function deleteMemory(workspaceId: string, memberId: string, memoryId: string) {
  return api<{ message: string }>(`${ws(workspaceId)}/bots/${memberId}/memories/${memoryId}`, {
    method: 'DELETE',
  })
}

export function listTodos(workspaceId: string, memberId: string, status?: string) {
  const qs = status ? `?status=${encodeURIComponent(status)}` : ''
  return api<{ todos: BotTodo[] }>(`${ws(workspaceId)}/bots/${memberId}/todos${qs}`)
}

export function createTodo(
  workspaceId: string,
  memberId: string,
  body: { title: string; status?: string; role_id?: string; issue_id?: string; sort_order?: number },
  idempotencyKey?: string,
) {
  return api<{ todo: BotTodo }>(`${ws(workspaceId)}/bots/${memberId}/todos`, {
    method: 'POST',
    body: JSON.stringify(body),
    idempotencyKey,
  })
}

export function patchTodo(
  workspaceId: string,
  memberId: string,
  todoId: string,
  body: Partial<{ title: string; status: string; role_id: string; issue_id: string; sort_order: number }>,
  idempotencyKey?: string,
) {
  return api<{ todo: BotTodo }>(`${ws(workspaceId)}/bots/${memberId}/todos/${todoId}`, {
    method: 'PATCH',
    body: JSON.stringify(body),
    idempotencyKey,
  })
}

export function getHandoff(workspaceId: string, memberId: string) {
  return api<{ handoff: BotHandoff }>(`${ws(workspaceId)}/bots/${memberId}/handoff`)
}

export type ImportHandoffResult = {
  imported_memories: number
  imported_todos: number
  claimed: boolean
  warnings: string[]
  issue_ids?: string[]
  handoff: BotHandoff
}

export function importHandoff(
  workspaceId: string,
  memberId: string,
  body: { handoff?: BotHandoff } & Partial<Pick<BotHandoff, 'memories' | 'todos' | 'issues'>> & {
    claimed_role_id?: string
    claimed_role_slug?: string
    claimed_role?: BotRole
  },
  idempotencyKey?: string,
) {
  return api<ImportHandoffResult>(`${ws(workspaceId)}/bots/${memberId}/handoff`, {
    method: 'POST',
    body: JSON.stringify(body),
    idempotencyKey,
  })
}
