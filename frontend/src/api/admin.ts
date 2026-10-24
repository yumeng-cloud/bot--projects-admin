import { api } from './client'

/** C3 / C8：全站工作区（超管） */
export type AdminWorkspace = {
  id: string
  name: string
  slug: string
  member_count?: number
  disabled_at?: string | null
  status?: string
  created_at?: string
  owner_email?: string
}

export type AdminUserWorkspace = {
  id: string
  name: string
  slug: string
}

export type AdminUser = {
  id: string
  email: string
  name: string
  is_platform_admin?: boolean
  created_at?: string
  /** D2：所属工作区（members）；无成员时为空数组 */
  workspaces?: AdminUserWorkspace[]
}

export function listAdminWorkspaces() {
  return api<{ workspaces: AdminWorkspace[] }>('/admin/workspaces', { skipToast: true })
}

/** PATCH body 仅 {disabled:boolean}（API.md） */
export function patchAdminWorkspace(id: string, body: { disabled: boolean }) {
  return api<{ workspace: AdminWorkspace }>(`/admin/workspaces/${id}`, {
    method: 'PATCH',
    body: JSON.stringify({ disabled: body.disabled }),
    skipToast: true,
  })
}

export function createAdminWorkspace(body: {
  name: string
  slug?: string
  owner_email?: string
  owner_user_id?: string
}) {
  return api<{ workspace: AdminWorkspace }>('/admin/workspaces', {
    method: 'POST',
    body: JSON.stringify(body),
    skipToast: true,
  })
}

/**
 * D1：可选 q 按邮箱/名过滤 → GET /admin/users?q=
 * D3：API 暂无 workspace_id；所属 WS 筛由 FE Soft†（row.workspaces[].id）完成。
 * 预留 workspace_id 参数：若后端日后支持则透传；当前调用方勿依赖服务端滤。
 */
export function listAdminUsers(opts?: { q?: string; workspace_id?: string }) {
  const params = new URLSearchParams()
  const q = opts?.q?.trim()
  if (q) params.set('q', q)
  // Soft†：不向 API 发 workspace_id（后端未实现）；保留签名供日后切换
  void opts?.workspace_id
  const qs = params.toString()
  const path = qs ? `/admin/users?${qs}` : '/admin/users'
  return api<{ users: AdminUser[] }>(path, { skipToast: true })
}

export function patchAdminUser(id: string, body: { is_platform_admin: boolean }) {
  return api<{ user: AdminUser }>(`/admin/users/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(body),
    skipToast: true,
  })
}
