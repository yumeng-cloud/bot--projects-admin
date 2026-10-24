import { api } from './client'

export type Workspace = {
  id: string
  name: string
  slug: string
  /** listMine 项上常有；GET 详情时 role 在响应顶层 */
  role?: string
  member_id?: string
}

/** GET /workspaces/{id}：role / member_id 在顶层，workspace 仅 id/name/slug */
export type GetWorkspaceResp = {
  workspace: { id: string; name: string; slug: string; role?: string }
  role?: string
  member_id?: string
}

export type CreateWorkspaceResp = {
  workspace: { id: string; name: string; slug: string }
  member_id: string
  role: string
}

/** 名称 → slug 候选：小写字母数字与连字符；纯中文等会得到空串，此时勿强塞，交给后端。 */
export function slugifyName(name: string): string {
  const s = name.trim().toLowerCase()
  let out = ''
  let prevDash = false
  for (const ch of s) {
    if ((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9')) {
      out += ch
      prevDash = false
    } else if (!prevDash) {
      out += '-'
      prevDash = true
    }
  }
  return out.replace(/^-+|-+$/g, '')
}

export function listMyWorkspaces() {
  return api<{ workspaces: Workspace[] }>('/workspaces')
}

export function getWorkspace(workspaceId: string) {
  return api<GetWorkspaceResp>(`/workspaces/${workspaceId}`)
}

/** A1：POST /workspaces；创建者成为 admin */
export function createWorkspace(body: { name: string; slug?: string }) {
  const payload: { name: string; slug?: string } = { name: body.name.trim() }
  const slug = (body.slug || '').trim().toLowerCase()
  if (slug) payload.slug = slug
  return api<CreateWorkspaceResp>('/workspaces', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

/** A5：PATCH /workspaces/{id} 改名/slug（WS admin） */
export function patchWorkspace(
  workspaceId: string,
  body: { name?: string; slug?: string },
) {
  return api<{ workspace: { id: string; name: string; slug: string } }>(
    `/workspaces/${workspaceId}`,
    {
      method: 'PATCH',
      body: JSON.stringify(body),
      skipToast: true,
    },
  )
}
