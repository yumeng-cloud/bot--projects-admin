import { api } from './client'
import type { Comment } from './boardTypes'

export function listComments(workspaceId: string, issueId: string) {
  return api<{ comments: Comment[] }>(
    `/workspaces/${workspaceId}/issues/${issueId}/comments`,
  )
}

export function createComment(workspaceId: string, issueId: string, body: string) {
  const key =
    typeof crypto !== 'undefined' && 'randomUUID' in crypto
      ? crypto.randomUUID()
      : `m-${Date.now()}`
  return api<{ comment: Comment }>(
    `/workspaces/${workspaceId}/issues/${issueId}/comments`,
    {
      method: 'POST',
      body: JSON.stringify({ body }),
      idempotencyKey: key,
    },
  )
}
