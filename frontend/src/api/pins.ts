import { api, ApiError } from './client'

const ws = (id: string) => `/workspaces/${id}`

/** 团队共享记忆墙条目（与 bot_memories 分表；UNIQUE(team_id, key)） */
export type TeamPin = {
  id: string
  workspace_id: string
  team_id: string
  key: string
  value: string
  updated_by?: string | null
  updated_at: string
}

export function listPins(workspaceId: string, teamId: string) {
  return api<{ pins: TeamPin[] }>(`${ws(workspaceId)}/teams/${teamId}/pins`, {
    skipToast: true,
  })
}

export function upsertPin(
  workspaceId: string,
  teamId: string,
  body: { key: string; value: string },
  idempotencyKey?: string,
) {
  return api<{ pin: TeamPin }>(`${ws(workspaceId)}/teams/${teamId}/pins`, {
    method: 'PUT',
    body: JSON.stringify(body),
    idempotencyKey,
    skipToast: true,
  })
}

export function deletePin(
  workspaceId: string,
  teamId: string,
  key: string,
  idempotencyKey?: string,
) {
  const enc = encodeURIComponent(key)
  return api<{ ok: boolean }>(`${ws(workspaceId)}/teams/${teamId}/pins/${enc}`, {
    method: 'DELETE',
    idempotencyKey,
    skipToast: true,
  })
}

/** 接口尚未就绪（契约齐、实现待 SHA）或错队 → 404/501 */
export function isPinsServiceMissing(err: unknown): boolean {
  return err instanceof ApiError && (err.status === 404 || err.status === 501)
}
