import { api } from './client'
import type { Activity } from '../types/issue'

export function listIssueActivities(workspaceId: string, issueId: string) {
  return api<{ activities: Activity[] }>(
    `/workspaces/${workspaceId}/issues/${issueId}/activities`,
  )
}

export function listActivities(
  workspaceId: string,
  query: Record<string, string> = {},
) {
  const qs = new URLSearchParams(query).toString()
  return api<{ activities: Activity[] }>(
    `/workspaces/${workspaceId}/activities${qs ? `?${qs}` : ''}`,
  )
}
