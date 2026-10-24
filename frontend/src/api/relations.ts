import { api } from './client'
import type { IssueRelation, RelationType } from '../types/issue'

function idem() {
  return crypto.randomUUID()
}

export function listRelations(workspaceId: string, issueId: string) {
  return api<{ relations: IssueRelation[] }>(
    `/workspaces/${workspaceId}/issues/${issueId}/relations`,
  )
}

export function createRelation(
  workspaceId: string,
  issueId: string,
  body: { to_issue_id: string; relation_type: RelationType },
) {
  return api<{ relation: IssueRelation }>(
    `/workspaces/${workspaceId}/issues/${issueId}/relations`,
    {
      method: 'POST',
      body: JSON.stringify(body),
      idempotencyKey: idem(),
    },
  )
}

export function deleteRelationOnIssue(
  workspaceId: string,
  issueId: string,
  relationId: string,
) {
  return api<{ ok: boolean }>(
    `/workspaces/${workspaceId}/issues/${issueId}/relations/${relationId}`,
    { method: 'DELETE', idempotencyKey: idem() },
  )
}

/** DELETE .../issues/{id}/relations?relation_id= */
export function deleteRelationCollection(
  workspaceId: string,
  issueId: string,
  relationId: string,
) {
  const qs = new URLSearchParams({ relation_id: relationId })
  return api<{ ok: boolean }>(
    `/workspaces/${workspaceId}/issues/${issueId}/relations?${qs}`,
    { method: 'DELETE', idempotencyKey: idem() },
  )
}

export function deleteRelation(workspaceId: string, relationId: string) {
  return api<{ ok: boolean }>(
    `/workspaces/${workspaceId}/relations/${relationId}`,
    { method: 'DELETE', idempotencyKey: idem() },
  )
}
