import { api, type ApiRequestInit } from './client'

const ws = (id: string) => `/workspaces/${id}`

/** 对齐 Go apiKeyDTO（snake_case） */
export type ApiKey = {
  id: string
  workspace_id: string
  member_id: string
  name: string
  key_prefix: string
  last_used_at?: string | null
  created_at: string
  revoked_at?: string | null
}

export type ApiKeyCreateResult = {
  api_key: ApiKey
  /** 明文仅创建当次；之后不可再取 */
  token: string
}

export function listApiKeys(workspaceId: string) {
  return api<{ api_keys: ApiKey[] }>(`${ws(workspaceId)}/api-keys`)
}

export function createApiKey(
  workspaceId: string,
  body: { name: string; member_id?: string },
  idempotencyKey?: string,
) {
  return api<ApiKeyCreateResult>(`${ws(workspaceId)}/api-keys`, {
    method: 'POST',
    body: JSON.stringify(body),
    idempotencyKey,
  })
}

/** DELETE 或 POST …/revoke；幂等已吊销由后端处理 */
export function revokeApiKey(
  workspaceId: string,
  keyId: string,
  opts?: { method?: 'DELETE' | 'POST'; idempotencyKey?: string },
) {
  const method = opts?.method ?? 'POST'
  const path =
    method === 'DELETE'
      ? `${ws(workspaceId)}/api-keys/${keyId}`
      : `${ws(workspaceId)}/api-keys/${keyId}/revoke`
  const init: ApiRequestInit = {
    method,
    idempotencyKey: opts?.idempotencyKey,
    skipToast: true,
  }
  if (method === 'POST') init.body = JSON.stringify({})
  return api<{ api_key?: ApiKey; ok?: boolean } | ApiKey>(path, init)
}
