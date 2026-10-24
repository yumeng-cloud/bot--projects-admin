import { api } from './client'
import { mockCycles, mockIssues, preferMock } from './mockData'
import type { Cycle } from '../types/issue'

function applyMockCyclePatch(
  id: string,
  body: Partial<Pick<Cycle, 'name' | 'status' | 'capacity' | 'start_date' | 'end_date'>>,
): Cycle {
  const idx = mockCycles.findIndex((c) => c.id === id)
  if (idx < 0) throw Object.assign(new Error('Cycle 不存在'), { code: 'not_found', status: 404 })
  const next = { ...mockCycles[idx], ...body, updated_at: new Date().toISOString() }
  mockCycles[idx] = next
  if (body.status === 'completed') {
    for (const iss of mockIssues) {
      if (iss.cycle_id === id && iss.status !== 'done' && iss.status !== 'canceled') {
        iss.cycle_id = null
        iss.status = 'backlog'
        iss.version += 1
        iss.updated_at = new Date().toISOString()
      }
    }
  }
  return next
}

export async function listCycles(
  workspaceId: string,
  projectId?: string,
): Promise<{ cycles: Cycle[]; mock: boolean }> {
  if (preferMock()) {
    return {
      cycles: mockCycles.filter((c) => !projectId || c.project_id === projectId),
      mock: true,
    }
  }
  const path = projectId
    ? `/workspaces/${workspaceId}/projects/${projectId}/cycles`
    : `/workspaces/${workspaceId}/cycles`
  const data = await api<{ cycles: Cycle[] }>(path)
  return { cycles: data.cycles || [], mock: false }
}

export async function patchCycle(
  workspaceId: string,
  id: string,
  body: Partial<Pick<Cycle, 'name' | 'status' | 'capacity' | 'start_date' | 'end_date'>>,
): Promise<{ cycle: Cycle; mock: boolean }> {
  if (preferMock()) {
    return { cycle: applyMockCyclePatch(id, body), mock: true }
  }
  const data = await api<{ cycle: Cycle }>(`/workspaces/${workspaceId}/cycles/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(body),
  })
  return { cycle: data.cycle, mock: false }
}

export function cycleEstimateSum(
  cycleId: string,
  issues: { cycle_id: string | null; estimate: number | null }[],
): number {
  return issues
    .filter((i) => i.cycle_id === cycleId)
    .reduce((s, i) => s + (i.estimate ?? 0), 0)
}
