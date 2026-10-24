import { api } from './client'
import type { Comment } from '../types/issue'

function idem() {
  return crypto.randomUUID()
}

export function listComments(workspaceId: string, issueId: string) {
  return api<{ comments: Comment[] }>(
    `/workspaces/${workspaceId}/issues/${issueId}/comments`,
  )
}

export function createComment(workspaceId: string, issueId: string, body: string) {
  return api<{ comment: Comment }>(
    `/workspaces/${workspaceId}/issues/${issueId}/comments`,
    {
      method: 'POST',
      body: JSON.stringify({ body }),
      idempotencyKey: idem(),
    },
  )
}

export function patchComment(
  workspaceId: string,
  issueId: string,
  commentId: string,
  body: string,
) {
  return api<{ comment: Comment }>(
    `/workspaces/${workspaceId}/issues/${issueId}/comments/${commentId}`,
    {
      method: 'PATCH',
      body: JSON.stringify({ body }),
      idempotencyKey: idem(),
    },
  )
}

export function deleteComment(workspaceId: string, issueId: string, commentId: string) {
  return api<{ ok: boolean }>(
    `/workspaces/${workspaceId}/issues/${issueId}/comments/${commentId}`,
    { method: 'DELETE', idempotencyKey: idem() },
  )
}
