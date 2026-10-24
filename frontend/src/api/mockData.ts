import type { Cycle, Issue, Label, MemberRef } from '../types/issue'

export const MOCK_WS = 'ws-demo'
export const MOCK_PROJECT = 'proj-demo'

export const mockMembers: MemberRef[] = [
  { id: 'm-1', display_name: '陈思', email: 'chen@example.com', name: '陈思' },
  { id: 'm-2', display_name: '李航', email: 'li@example.com', name: '李航' },
  { id: 'm-3', display_name: '王敏', email: 'wang@example.com', name: '王敏' },
]

export const mockLabels: Label[] = [
  { id: 'l-1', workspace_id: MOCK_WS, name: '前端', color: '#1677ff' },
  { id: 'l-2', workspace_id: MOCK_WS, name: '后端', color: '#52c41a' },
  { id: 'l-3', workspace_id: MOCK_WS, name: '阻塞', color: '#ff4d4f' },
]

const now = () => new Date().toISOString()

function issue(
  n: number,
  partial: Partial<Issue> & Pick<Issue, 'title' | 'status' | 'priority'>,
): Issue {
  return {
    id: `iss-${n}`,
    project_id: MOCK_PROJECT,
    workspace_id: MOCK_WS,
    number: n,
    identifier: `DEMO-${n}`,
    description: partial.description ?? '',
    estimate: partial.estimate ?? null,
    assignee_id: partial.assignee_id ?? null,
    reporter_id: 'm-1',
    cycle_id: partial.cycle_id ?? null,
    parent_id: null,
    board_order: partial.board_order ?? n * 100,
    version: 1,
    labels: partial.labels ?? [],
    created_at: now(),
    updated_at: now(),
    ...partial,
  }
}

export const mockCycles: Cycle[] = [
  {
    id: 'cyc-1',
    project_id: MOCK_PROJECT,
    name: 'Sprint 12',
    start_date: '2026-10-20',
    end_date: '2026-11-02',
    capacity: 20,
    status: 'active',
  },
  {
    id: 'cyc-2',
    project_id: MOCK_PROJECT,
    name: 'Sprint 13',
    start_date: '2026-11-03',
    end_date: '2026-11-16',
    capacity: 18,
    status: 'planned',
  },
]

export let mockIssues: Issue[] = [
  issue(1, {
    title: '排期能力（父）',
    status: 'in_progress',
    priority: 'high',
    estimate: 8,
    assignee_id: 'm-1',
    cycle_id: 'cyc-1',
    labels: [mockLabels[0]],
    board_order: 100,
    start_at: '2026-10-20',
    due_at: '2026-10-31',
  }),
  issue(2, {
    title: '看板拖拽改状态',
    status: 'in_progress',
    priority: 'high',
    estimate: 5,
    assignee_id: 'm-1',
    cycle_id: 'cyc-1',
    labels: [mockLabels[0]],
    board_order: 100,
    parent_id: 'iss-1',
    start_at: '2026-10-20',
    due_at: '2026-10-24',
  }),
  issue(3, {
    title: '列表筛选与空状态',
    status: 'todo',
    priority: 'medium',
    estimate: 3,
    assignee_id: 'm-2',
    cycle_id: 'cyc-1',
    labels: [mockLabels[0]],
    board_order: 200,
    parent_id: 'iss-1',
    start_at: '2026-10-23',
    due_at: '2026-10-28',
  }),
  issue(10, {
    title: '拖拽手柄细节（叶）',
    status: 'todo',
    priority: 'medium',
    estimate: 1,
    assignee_id: 'm-2',
    cycle_id: 'cyc-1',
    board_order: 100,
    parent_id: 'iss-2',
    start_at: '2026-10-21',
    due_at: '2026-10-22',
  }),
  issue(11, {
    title: '乐观锁回滚提示（叶）',
    status: 'todo',
    priority: 'low',
    estimate: 1,
    board_order: 200,
    parent_id: 'iss-2',
    start_at: '2026-10-22',
    due_at: '2026-10-23',
  }),
  issue(4, {
    title: 'Cycle 容量过载展示',
    status: 'backlog',
    priority: 'urgent',
    estimate: 8,
    assignee_id: 'm-3',
    cycle_id: 'cyc-1',
    labels: [mockLabels[1]],
    board_order: 200,
    start_at: '2026-10-27',
    due_at: '2026-11-02',
  }),
  issue(5, {
    title: '登录页中文文案',
    status: 'done',
    priority: 'low',
    estimate: 1,
    assignee_id: 'm-2',
    cycle_id: 'cyc-1',
    board_order: 100,
    start_at: '2026-10-15',
    due_at: '2026-10-16',
  }),
  issue(6, {
    title: 'Wiki 搜索 API',
    status: 'todo',
    priority: 'medium',
    estimate: 5,
    assignee_id: 'm-3',
    cycle_id: 'cyc-2',
    labels: [mockLabels[1]],
    board_order: 200,
    start_at: '2026-11-03',
    due_at: '2026-11-10',
  }),
  issue(7, {
    title: '工时报表导出',
    status: 'backlog',
    priority: 'none',
    estimate: 3,
    board_order: 300,
  }),
  issue(8, {
    title: 'Bot handoff 包',
    status: 'in_progress',
    priority: 'high',
    estimate: 5,
    assignee_id: 'm-1',
    labels: [mockLabels[1], mockLabels[2]],
    board_order: 200,
  }),
  issue(9, {
    title: 'Webhook 重试策略',
    status: 'canceled',
    priority: 'low',
    estimate: 2,
    board_order: 100,
  }),
]

const MOCK_FLAG = 'ai-collab-board:force-mock'

/** 默认先打真实 API；localStorage 设 force-mock=1 或 VITE_USE_MOCK=1 强制演示数据 */
export function preferMock(): boolean {
  if (import.meta.env.VITE_USE_MOCK === '1') return true
  return localStorage.getItem(MOCK_FLAG) === '1'
}

export function setForceMock(on: boolean) {
  if (on) localStorage.setItem(MOCK_FLAG, '1')
  else localStorage.removeItem(MOCK_FLAG)
}

export function isMockActive(): boolean {
  return preferMock() || (typeof window !== 'undefined' && !!(window as unknown as { __boardMockFallback?: boolean }).__boardMockFallback)
}

export function markMockFallback(active: boolean) {
  ;(window as unknown as { __boardMockFallback?: boolean }).__boardMockFallback = active
}
