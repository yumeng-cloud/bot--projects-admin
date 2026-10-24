import { api } from './client'

export type WorkspaceMember = {
  id: string
  workspace_id?: string
  user_id?: string | null
  role: string
  display_name: string
  is_bot?: boolean | number
  email?: string | null
  created_at?: string
}

export function listMembers(workspaceId: string) {
  return api<{ members: WorkspaceMember[] }>(`/workspaces/${workspaceId}/members`)
}

/** B1：Admin 邀请已注册邮箱入席 */
export function inviteMember(
  workspaceId: string,
  body: { email: string; role: 'admin' | 'member'; display_name?: string },
) {
  return api<{ member: WorkspaceMember }>(`/workspaces/${workspaceId}/members`, {
    method: 'POST',
    body: JSON.stringify({
      email: body.email.trim().toLowerCase(),
      role: body.role,
      ...(body.display_name?.trim() ? { display_name: body.display_name.trim() } : {}),
    }),
    skipToast: true,
  })
}

/** B2：改角色 / 显示名 */
export function patchMember(
  workspaceId: string,
  memberId: string,
  body: { role?: string; display_name?: string },
) {
  return api<{ member: WorkspaceMember }>(
    `/workspaces/${workspaceId}/members/${memberId}`,
    {
      method: 'PATCH',
      body: JSON.stringify(body),
      skipToast: true,
    },
  )
}

/** B2：移除成员 */
export function removeMember(workspaceId: string, memberId: string) {
  return api<{ message: string }>(`/workspaces/${workspaceId}/members/${memberId}`, {
    method: 'DELETE',
    skipToast: true,
  })
}
