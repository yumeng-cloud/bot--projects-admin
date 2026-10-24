import { useCallback, useEffect, useMemo, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import {
  Table,
  Space,
  Select,
  Input,
  Button,
  Tag,
  Typography,
  Empty,
  Spin,
  message,
  Checkbox,
} from 'antd'
import type { Key } from 'react'
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons'
import {
  createIssue,
  listIssues,
  listProjectLabels,
  listProjectMembers,
} from '../../api/issues'
import { IssueDrawer } from '../../components/issue'
import { MockBanner } from '../../components/common/MockBanner'
import { useCurrentTeam } from '../../hooks/useCurrentTeam'
import type { Issue, IssuePriority, IssueStatus, Label, MemberRef } from '../../types/issue'
import '../../components/board/board.css'
import './list.css'
import {
  ISSUE_STATUSES,
  PRIORITY_LABELS,
  STATUS_LABELS,
  isIssueArchived,
  memberLabel,
} from '../../types/issue'
import {
  buildIssueTree,
  hasActiveFilters,
  isOrphanSubitem,
  selectIssuesWithAncestors,
  type IssueTreeRow,
} from '../../utils/issueTree'

export type IssueListPageProps = {
  workspaceId: string
  projectId: string
}

export default function IssueListPage({ workspaceId, projectId }: IssueListPageProps) {
  const [searchParams, setSearchParams] = useSearchParams()
  const blocked = searchParams.get('blocked') === '1'
  const overdue = searchParams.get('overdue') === '1'
  const { teamId, loading: teamLoading } = useCurrentTeam(workspaceId)

  const [loading, setLoading] = useState(true)
  /** Full project list (unfiltered) so tree filters can keep ancestor context. */
  const [issues, setIssues] = useState<Issue[]>([])
  const [members, setMembers] = useState<MemberRef[]>([])
  const [labels, setLabels] = useState<Label[]>([])
  const [status, setStatus] = useState<IssueStatus | undefined>()
  const [priority, setPriority] = useState<IssuePriority | undefined>()
  const [assigneeId, setAssigneeId] = useState<string | undefined>()
  const [labelId, setLabelId] = useState<string | undefined>()
  const [q, setQ] = useState('')
  const [open, setOpen] = useState(false)
  const [current, setCurrent] = useState<Issue | null>(null)
  const [newTitle, setNewTitle] = useState('')
  const [creating, setCreating] = useState(false)
  const [expandedRowKeys, setExpandedRowKeys] = useState<Key[]>([])

  const setFlag = (key: 'blocked' | 'overdue', on: boolean) => {
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev)
        if (on) next.set(key, '1')
        else next.delete(key)
        return next
      },
      { replace: true },
    )
  }

  const loadMeta = useCallback(async () => {
    const [m, l] = await Promise.all([
      listProjectMembers(workspaceId),
      listProjectLabels(workspaceId),
    ])
    setMembers(m.members)
    setLabels(l.labels)
  }, [workspaceId])

  const load = useCallback(async () => {
    setLoading(true)
    setIssues([]) // I4: clear before fetch — no cross-team residue
    try {
      // Fetch team-scoped project issues; other filters applied client-side for tree ancestors.
      // `blocked` stays API-side (needs relations).
      const { issues: list } = await listIssues(workspaceId, {
        project_id: projectId,
        blocked: blocked || undefined,
        team_id: teamId || undefined,
      })
      // C2：默认不传 include_archived；防御性剔归档，避免主列表虚显
      setIssues(list.filter((i) => !isIssueArchived(i)))
    } catch {
      setIssues([])
      // 4xx toast from client; empty pool, no white screen
    } finally {
      setLoading(false)
    }
  }, [workspaceId, projectId, blocked, teamId])

  useEffect(() => {
    loadMeta()
  }, [loadMeta])

  useEffect(() => {
    if (teamLoading) return
    void load()
  }, [load, teamLoading])

  const memberName = useMemo(() => {
    const map = new Map(members.map((m) => [m.id, memberLabel(m)]))
    return (id: string | null) => (id ? map.get(id) || '成员' : '—')
  }, [members])

  const filters = useMemo(
    () => ({
      status,
      priority,
      assigneeId,
      labelId,
      q: q || undefined,
      overdue: overdue || undefined,
    }),
    [status, priority, assigneeId, labelId, q, overdue],
  )

  const { treeData, visibleIds, expandIds } = useMemo(() => {
    const { visible, expandIds: ids } = selectIssuesWithAncestors(issues, filters)
    const idSet = new Set(visible.map((i) => i.id))
    return {
      treeData: buildIssueTree(visible),
      visibleIds: idSet,
      expandIds: ids,
    }
  }, [issues, filters])

  // When filters change, auto-expand ancestor paths so matches stay visible (Linear-like).
  useEffect(() => {
    if (hasActiveFilters(filters)) {
      setExpandedRowKeys(expandIds)
    }
  }, [filters, expandIds])

  const handleCreate = async () => {
    const title = newTitle.trim()
    if (!title || creating) return
    if (!teamId) {
      message.warning('请先选择团队后再创建')
      return
    }
    setCreating(true)
    try {
      const { issue } = await createIssue(workspaceId, {
        project_id: projectId,
        title,
        status: 'todo',
        team_id: teamId,
      })
      message.success('已创建')
      setNewTitle('')
      setIssues((prev) => [issue, ...prev])
      setCurrent(issue)
      setOpen(true)
    } catch {
      // toasted by client
    } finally {
      setCreating(false)
    }
  }

  return (
    <div className="acb-page">
      <MockBanner />
      <div className="acb-page-bar">
        <Typography.Title level={4}>列表</Typography.Title>
        <Space size={8} wrap>
          <Input
            size="small"
            placeholder="新建 Issue，回车创建"
            value={newTitle}
            onChange={(e) => setNewTitle(e.target.value)}
            onPressEnter={() => void handleCreate()}
            onKeyDown={(e) => {
              if (e.key === 'Escape') {
                setNewTitle('')
                ;(e.target as HTMLInputElement).blur()
              }
            }}
            style={{ width: 220 }}
            disabled={creating}
            allowClear
          />
          <Button
            type="primary"
            size="small"
            icon={<PlusOutlined />}
            loading={creating}
            onClick={() => void handleCreate()}
            disabled={!newTitle.trim()}
          >
            新建 Issue
          </Button>
          <Button type="text" size="small" icon={<ReloadOutlined />} onClick={load}>
            刷新
          </Button>
        </Space>
      </div>
      <Space wrap className="acb-list-filters" size={6}>
        <Select
          allowClear
          placeholder="状态"
          size="small"
          style={{ width: 120 }}
          value={status}
          onChange={setStatus}
          options={ISSUE_STATUSES.map((s) => ({ value: s, label: STATUS_LABELS[s] }))}
        />
        <Select
          allowClear
          placeholder="优先级"
          size="small"
          style={{ width: 100 }}
          value={priority}
          onChange={setPriority}
          options={(Object.keys(PRIORITY_LABELS) as IssuePriority[]).map((p) => ({
            value: p,
            label: PRIORITY_LABELS[p],
          }))}
        />
        <Select
          allowClear
          placeholder="负责人"
          size="small"
          style={{ width: 120 }}
          value={assigneeId}
          onChange={setAssigneeId}
          options={members.map((m) => ({ value: m.id, label: memberLabel(m) }))}
        />
        <Select
          allowClear
          placeholder="标签"
          size="small"
          style={{ width: 120 }}
          value={labelId}
          onChange={setLabelId}
          options={labels.map((l) => ({ value: l.id, label: l.name }))}
        />
        <Input.Search
          allowClear
          placeholder="搜索标题 / 标识"
          size="small"
          style={{ width: 200 }}
          onSearch={setQ}
          onChange={(e) => {
            if (!e.target.value) setQ('')
          }}
        />
        <Checkbox checked={blocked} onChange={(e) => setFlag('blocked', e.target.checked)}>
          被阻塞
        </Checkbox>
        <Checkbox checked={overdue} onChange={(e) => setFlag('overdue', e.target.checked)}>
          已逾期
        </Checkbox>
      </Space>
      {loading ? (
        <div style={{ textAlign: 'center', padding: 48 }}>
          <Spin />
        </div>
      ) : (
        <Table<IssueTreeRow>
          className="acb-list-table acb-list-tree"
          size="small"
          rowKey="id"
          dataSource={treeData}
          indentSize={18}
          expandable={{
            expandedRowKeys,
            onExpandedRowsChange: (keys) => setExpandedRowKeys([...keys]),
            indentSize: 18,
          }}
          locale={{
            emptyText: (
              <div className="acb-empty-panel">
                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="还没有 Issue，新建一条">
                  <Button
                    type="primary"
                    size="small"
                    icon={<PlusOutlined />}
                    onClick={() => {
                      const el = document.querySelector<HTMLInputElement>('.acb-page-bar input')
                      el?.focus()
                    }}
                  >
                    新建 Issue
                  </Button>
                </Empty>
              </div>
            ),
          }}
          onRow={(record) => ({
            onClick: () => {
              setCurrent(record)
              setOpen(true)
            },
            style: { cursor: 'pointer' },
            className: record.parent_id ? 'acb-list-tree-child' : undefined,
          })}
          columns={[
            { title: '标识', dataIndex: 'identifier', width: 110 },
            {
              title: '标题',
              dataIndex: 'title',
              render: (title: string, record: IssueTreeRow) => (
                <span className="acb-list-title-cell">
                  {isOrphanSubitem(record, visibleIds) ? (
                    <Tag className="acb-list-subitem-hint">子项</Tag>
                  ) : null}
                  <span className="acb-list-title-text">{title}</span>
                </span>
              ),
            },
            {
              title: '状态',
              dataIndex: 'status',
              width: 110,
              render: (s: IssueStatus) => STATUS_LABELS[s],
            },
            {
              title: '优先级',
              dataIndex: 'priority',
              width: 90,
              render: (p: IssuePriority) => PRIORITY_LABELS[p],
            },
            {
              title: '负责人',
              dataIndex: 'assignee_id',
              width: 100,
              render: (id: string | null) => memberName(id),
            },
            {
              title: '标签',
              dataIndex: 'labels',
              render: (ls: Label[] | undefined) => (
                <Space size={4} wrap>
                  {(ls || []).map((l) => (
                    <Tag key={l.id} color={l.color}>
                      {l.name}
                    </Tag>
                  ))}
                </Space>
              ),
            },
            {
              title: '开始',
              dataIndex: 'start_at',
              width: 100,
              render: (v: string | null | undefined) => v || '—',
            },
            {
              title: '截止',
              dataIndex: 'due_at',
              width: 100,
              render: (v: string | null | undefined) => v || '—',
            },
            {
              title: '估时',
              dataIndex: 'estimate',
              width: 70,
              render: (v: number | null) => v ?? '—',
            },
          ]}
          pagination={{ pageSize: 20 }}
        />
      )}
      <IssueDrawer
        open={open}
        issue={current}
        workspaceId={workspaceId}
        projectId={projectId}
        teamId={teamId}
        onClose={() => setOpen(false)}
        onSaved={(saved) => {
          setIssues((prev) => prev.map((i) => (i.id === saved.id ? saved : i)))
          setCurrent(saved)
        }}
      />
    </div>
  )
}
