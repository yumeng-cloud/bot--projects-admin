import { api, API_BASE, getToken, resolveUrl } from './client'
import type {
  BillingRate,
  BillingRateUpsert,
  BillingSummary,
  TimeEntry,
  TimeEntryCreate,
  TimeEntryPatch,
} from '../types/time'

function idem() {
  return crypto.randomUUID()
}

export function listTimeEntries(workspaceId: string, query: Record<string, string> = {}) {
  const qs = new URLSearchParams(query).toString()
  return api<{ time_entries: TimeEntry[] }>(
    `/workspaces/${workspaceId}/time-entries${qs ? `?${qs}` : ''}`,
  )
}

export function getTimeEntry(workspaceId: string, entryId: string) {
  return api<{ time_entry: TimeEntry }>(`/workspaces/${workspaceId}/time-entries/${entryId}`)
}

export function createTimeEntry(workspaceId: string, body: TimeEntryCreate) {
  return api<{ time_entry: TimeEntry }>(`/workspaces/${workspaceId}/time-entries`, {
    method: 'POST',
    body: JSON.stringify(body),
    idempotencyKey: idem(),
  })
}

export function updateTimeEntry(workspaceId: string, entryId: string, body: TimeEntryPatch) {
  return api<{ time_entry: TimeEntry }>(`/workspaces/${workspaceId}/time-entries/${entryId}`, {
    method: 'PATCH',
    body: JSON.stringify(body),
    idempotencyKey: idem(),
  })
}

export function deleteTimeEntry(workspaceId: string, entryId: string) {
  return api<{ message: string }>(`/workspaces/${workspaceId}/time-entries/${entryId}`, {
    method: 'DELETE',
    idempotencyKey: idem(),
  })
}

export function listBillingRates(workspaceId: string) {
  return api<{ billing_rates: BillingRate[] }>(`/workspaces/${workspaceId}/billing-rates`)
}

export function upsertBillingRate(workspaceId: string, body: BillingRateUpsert) {
  return api<{ billing_rate: BillingRate }>(`/workspaces/${workspaceId}/billing-rates`, {
    method: 'PUT',
    body: JSON.stringify(body),
    idempotencyKey: idem(),
  })
}

export function deleteBillingRate(workspaceId: string, rateId: string) {
  return api<{ message: string }>(`/workspaces/${workspaceId}/billing-rates/${rateId}`, {
    method: 'DELETE',
    idempotencyKey: idem(),
  })
}

export function billingSummary(
  workspaceId: string,
  query: { from?: string; to?: string; group_by?: 'project' | 'member' | 'date' } = {},
) {
  const qs = new URLSearchParams()
  if (query.from) qs.set('from', query.from)
  if (query.to) qs.set('to', query.to)
  if (query.group_by) qs.set('group_by', query.group_by)
  const s = qs.toString()
  return api<{ summary: BillingSummary }>(
    `/workspaces/${workspaceId}/billing/summary${s ? `?${s}` : ''}`,
  )
}

/** CSV 导出：带 Bearer 拉 blob 再下载（window.open 无法带 token） */
export async function exportBillingCsv(
  workspaceId: string,
  query: { from?: string; to?: string } = {},
) {
  const qs = new URLSearchParams()
  if (query.from) qs.set('from', query.from)
  if (query.to) qs.set('to', query.to)
  const s = qs.toString()
  const path = `/workspaces/${workspaceId}/billing/export.csv${s ? `?${s}` : ''}`
  const headers = new Headers()
  const token = getToken()
  if (token) headers.set('Authorization', `Bearer ${token}`)
  const res = await fetch(resolveUrl(path), { headers, credentials: 'include' })
  if (!res.ok) {
    const text = await res.text()
    throw new Error(text || `导出失败 HTTP ${res.status}`)
  }
  const blob = await res.blob()
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'time-entries.csv'
  a.click()
  URL.revokeObjectURL(url)
}

export { API_BASE }


export function listIssueTimeEntries(workspaceId: string, issueId: string) {
  return api<{ time_entries: TimeEntry[] }>(
    `/workspaces/${workspaceId}/issues/${issueId}/time-entries`,
  )
}

export function createIssueTimeEntry(
  workspaceId: string,
  issueId: string,
  body: Omit<TimeEntryCreate, 'issue_id'> & { issue_id?: string },
) {
  return api<{ time_entry: TimeEntry }>(
    `/workspaces/${workspaceId}/issues/${issueId}/time-entries`,
    {
      method: 'POST',
      body: JSON.stringify({ ...body, issue_id: issueId }),
      idempotencyKey: idem(),
    },
  )
}
