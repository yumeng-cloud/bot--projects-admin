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
  Switch,
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
  /** B3/B4：qaEmpty=1 强制无 Issue 空态供箱截 */
  const forceEmpty = searchParams.get('qaEmpty') === '1'
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
  /** C1：默认关；开 → include_archived=1 */
  const [showArchived, setShowArchived] = useState(false)
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
      // C1：默认不传 include_archived；开开关才传 1。
      const { issues: list } = await listIssues(workspaceId, {
        project_id: projectId,
        blocked: blocked || undefined,
        team_id: teamId || undefined,
        include_archived: showArchived ? 1 : undefined,
      })
      // 关时防御性剔归档，避免主列表虚显；开时保留归档行
      setIssues(showArchived ? list : list.filter((i) => !isIssueArchived(i)))
    } catch {
      setIssues([])
      // 4xx toast from client; empty pool, no white screen
    } finally {
      setLoading(false)
    }
  }, [workspaceId, projectId, blocked, teamId, showArchived])

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
        <span className="acb-list-archived-toggle">
          显示已归档
          <Switch
            size="small"
            checked={showArchived}
            onChange={setShowArchived}
            aria-label="显示已归档"
          />
        </span>
      </Space>
      {loading ? (
        <div style={{ textAlign: 'center', padding: 48 }}>
          <Spin />
        </div>
      ) : forceEmpty || (issues.length === 0 && !hasActiveFilters(filters)) ? (
        <div
          className="acb-empty-panel"
          style={{ minHeight: 120, padding: '20px 16px', textAlign: 'center', marginTop: 8 }}
        >
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description={
              <span>
                <Typography.Text strong style={{ display: 'block', fontSize: 15, marginBottom: 4 }}>
                  还没有 Issue
                </Typography.Text>
                <Typography.Text type="secondary" style={{ fontSize: 13 }}>
                  写一条标题，立刻开工
                </Typography.Text>
              </span>
            }
          >
            <Space wrap>
              <Input
                size="small"
                placeholder="标题"
                value={newTitle}
                onChange={(e) => setNewTitle(e.target.value)}
                onPressEnter={() => void handleCreate()}
                style={{ width: 220 }}
                disabled={creating || forceEmpty}
                allowClear
              />
              <Button
                type="primary"
                size="small"
                icon={<PlusOutlined />}
                loading={creating}
                onClick={() => {
                  if (forceEmpty) {
                    const el = document.querySelector<HTMLInputElement>('.acb-page-bar input')
                    el?.focus()
                    return
                  }
                  void handleCreate()
                }}
                disabled={forceEmpty ? false : !newTitle.trim()}
              >
                新建
              </Button>
            </Space>
          </Empty>
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
              <div className="acb-empty-panel" style={{ minHeight: 100, padding: '16px' }}>
                <Empty
                  image={Empty.PRESENTED_IMAGE_SIMPLE}
                  description="没有符合条件的 Issue"
                />
              </div>
            ),
          }}
          onRow={(record) => {
            const archived = isIssueArchived(record)
            const classes = [
              record.parent_id ? 'acb-list-tree-child' : '',
              archived ? 'acb-list-row-archived' : '',
            ]
              .filter(Boolean)
              .join(' ')
            return {
              onClick: () => {
                setCurrent(record)
                setOpen(true)
              },
              style: { cursor: 'pointer' },
              className: classes || undefined,
            }
          }}
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
                  {isIssueArchived(record) ? (
                    <Tag className="acb-list-archived-tag">已归档</Tag>
                  ) : null}
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
          const archived = isIssueArchived(saved)
          if (archived && !showArchived) {
            setIssues((prev) => prev.filter((i) => i.id !== saved.id))
            setCurrent(saved)
            return
          }
          setIssues((prev) => {
            const idx = prev.findIndex((i) => i.id === saved.id)
            if (idx < 0) return [saved, ...prev]
            return prev.map((i) => (i.id === saved.id ? saved : i))
          })
          setCurrent(saved)
        }}
      />
    </div>
  )
}
