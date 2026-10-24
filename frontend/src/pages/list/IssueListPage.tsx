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
import { PlusOutlined, ReloadOutlined, CloseOutlined } from '@ant-design/icons'
import {
  archiveIssues,
  createIssue,
  listIssues,
  listProjectLabels,
  listProjectMembers,
  patchIssue,
  unarchiveIssues,
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

/** 列表归档视图：与 C1 显隐对齐，E3 补仅归档 */
type ArchiveView = 'hide' | 'include' | 'only'

export default function IssueListPage({ workspaceId, projectId }: IssueListPageProps) {
  const [searchParams, setSearchParams] = useSearchParams()
  const blocked = searchParams.get('blocked') === '1'
  const overdue = searchParams.get('overdue') === '1'
  /** B3/B4：qaEmpty=1 强制无 Issue 空态供箱截 */
  const forceEmpty = searchParams.get('qaEmpty') === '1'
  /** C2：qaBulk=1 自动勾选前几条供箱截（操作条可见） */
  const forceBulk = searchParams.get('qaBulk') === '1'
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
  /** C1+E3：hide 剔归档；include=显示已归档；only=仅已归档（互斥） */
  const [archiveView, setArchiveView] = useState<ArchiveView>('hide')
  const [open, setOpen] = useState(false)
  const [current, setCurrent] = useState<Issue | null>(null)
  const [newTitle, setNewTitle] = useState('')
  const [creating, setCreating] = useState(false)
  const [expandedRowKeys, setExpandedRowKeys] = useState<Key[]>([])
  /** C2：列表多选 */
  const [selectedRowKeys, setSelectedRowKeys] = useState<Key[]>([])
  const [bulkStatus, setBulkStatus] = useState<IssueStatus | undefined>()
  const [bulkBusy, setBulkBusy] = useState(false)

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
      // C1/E3：hide 不传；include → include_archived=1；only → archived_only（API archived=only）
      const { issues: list } = await listIssues(workspaceId, {
        project_id: projectId,
        blocked: blocked || undefined,
        team_id: teamId || undefined,
        ...(archiveView === 'only'
          ? { archived_only: 1 as const }
          : archiveView === 'include'
            ? { include_archived: 1 as const }
            : {}),
      })
      // 防御性客户端滤：only 只留归档；hide 剔归档；include 原样
      const next =
        archiveView === 'only'
          ? list.filter((i) => isIssueArchived(i))
          : archiveView === 'include'
            ? list
            : list.filter((i) => !isIssueArchived(i))
      setIssues(next)
    } catch {
      setIssues([])
      // 4xx toast from client; empty pool, no white screen
    } finally {
      setLoading(false)
    }
  }, [workspaceId, projectId, blocked, teamId, archiveView])

  useEffect(() => {
    loadMeta()
  }, [loadMeta])

  useEffect(() => {
    if (teamLoading) return
    void load()
  }, [load, teamLoading])

  // C2：列表变短时清掉失效勾选
  useEffect(() => {
    setSelectedRowKeys((prev) => {
      if (prev.length === 0) return prev
      const idSet = new Set(issues.map((i) => i.id))
      const next = prev.filter((k) => idSet.has(String(k)))
      return next.length === prev.length ? prev : next
    })
  }, [issues])

  const selectedIssues = useMemo(() => {
    const idSet = new Set(selectedRowKeys.map(String))
    return issues.filter((i) => idSet.has(i.id))
  }, [issues, selectedRowKeys])

  const selectedDoneCount = useMemo(
    () => selectedIssues.filter((i) => i.status === 'done' && !isIssueArchived(i)).length,
    [selectedIssues],
  )
  const selectedArchivedCount = useMemo(
    () => selectedIssues.filter((i) => isIssueArchived(i)).length,
    [selectedIssues],
  )

  const clearSelection = useCallback(() => {
    setSelectedRowKeys([])
    setBulkStatus(undefined)
  }, [])

  const handleBulkStatus = async () => {
    if (!bulkStatus || selectedIssues.length === 0 || bulkBusy) return
    setBulkBusy(true)
    let ok = 0
    let fail = 0
    const updates: Issue[] = []
    try {
      // 无 batch status API → 串行 PATCH（C2-2 缺口另报 Issue 域）
      for (const iss of selectedIssues) {
        if (iss.status === bulkStatus) {
          ok += 1
          continue
        }
        try {
          const { issue: saved } = await patchIssue(workspaceId, iss.id, {
            version: iss.version,
            status: bulkStatus,
          })
          updates.push(saved)
          ok += 1
        } catch {
          fail += 1
        }
      }
      if (updates.length > 0) {
        const map = new Map(updates.map((u) => [u.id, u]))
        setIssues((prev) => prev.map((i) => map.get(i.id) ?? i))
      }
      if (fail === 0) {
        message.success(`已将 ${ok} 条改为「${STATUS_LABELS[bulkStatus]}」`)
        clearSelection()
      } else {
        message.warning(`改状态完成：成功 ${ok}，失败 ${fail}`)
      }
    } finally {
      setBulkBusy(false)
    }
  }

  const handleBulkArchive = async () => {
    const ids = selectedIssues
      .filter((i) => i.status === 'done' && !isIssueArchived(i))
      .map((i) => i.id)
    if (ids.length === 0) {
      message.warning('请选择已完成且未归档的 Issue')
      return
    }
    if (bulkBusy) return
    setBulkBusy(true)
    try {
      await archiveIssues(workspaceId, ids)
      const idSet = new Set(ids)
      const now = new Date().toISOString()
      if (archiveView === 'hide') {
        setIssues((prev) => prev.filter((i) => !idSet.has(i.id)))
      } else {
        setIssues((prev) =>
          prev.map((i) =>
            idSet.has(i.id)
              ? { ...i, archived_at: i.archived_at || now, version: i.version + 1 }
              : i,
          ),
        )
      }
      message.success(`已归档 ${ids.length} 条`)
      clearSelection()
    } catch (err) {
      message.error((err as Error).message || '归档失败')
    } finally {
      setBulkBusy(false)
    }
  }

  const handleBulkUnarchive = async () => {
    const ids = selectedIssues.filter((i) => isIssueArchived(i)).map((i) => i.id)
    if (ids.length === 0) {
      message.warning('请选择已归档的 Issue')
      return
    }
    if (bulkBusy) return
    setBulkBusy(true)
    try {
      await unarchiveIssues(workspaceId, ids)
      const idSet = new Set(ids)
      if (archiveView === 'only') {
        setIssues((prev) => prev.filter((i) => !idSet.has(i.id)))
      } else {
        setIssues((prev) =>
          prev.map((i) =>
            idSet.has(i.id) ? { ...i, archived_at: null, version: i.version + 1 } : i,
          ),
        )
      }
      message.success(`已解档 ${ids.length} 条`)
      clearSelection()
    } catch (err) {
      message.error((err as Error).message || '解档失败')
    } finally {
      setBulkBusy(false)
    }
  }

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

  // C2 箱截：qaBulk=1 时自动勾选可见前 3 条（含至少一条 done 若有）
  useEffect(() => {
    if (!forceBulk || loading || issues.length === 0) return
    if (selectedRowKeys.length > 0) return
    const done = issues.filter((i) => i.status === 'done' && !isIssueArchived(i))
    const rest = issues.filter((i) => !(i.status === 'done' && !isIssueArchived(i)))
    const pick = [...done.slice(0, 1), ...rest.slice(0, 2)].slice(0, 3)
    const keys = (pick.length > 0 ? pick : issues.slice(0, 3)).map((i) => i.id)
    setSelectedRowKeys(keys)
    setBulkStatus('in_progress')
  }, [forceBulk, loading, issues, selectedRowKeys.length])

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
        <Select
          size="small"
          className="acb-list-archived-view"
          style={{ width: 130 }}
          value={archiveView}
          onChange={(v: ArchiveView) => setArchiveView(v)}
          aria-label="归档筛选"
          options={[
            { value: 'hide', label: '隐藏已归档' },
            { value: 'include', label: '显示已归档' },
            { value: 'only', label: '仅已归档' },
          ]}
        />
      </Space>
      {selectedRowKeys.length > 0 ? (
        <div className="acb-list-bulk-bar" role="toolbar" aria-label="批量操作">
          <Space size={8} wrap>
            <Typography.Text className="acb-list-bulk-count">
              已选 {selectedRowKeys.length} 条
            </Typography.Text>
            <Select
              allowClear
              size="small"
              placeholder="改状态为…"
              style={{ width: 130 }}
              value={bulkStatus}
              onChange={(v: IssueStatus | undefined) => setBulkStatus(v)}
              options={ISSUE_STATUSES.map((s) => ({ value: s, label: STATUS_LABELS[s] }))}
              disabled={bulkBusy}
            />
            <Button
              size="small"
              type="primary"
              disabled={!bulkStatus || bulkBusy}
              loading={bulkBusy}
              onClick={() => void handleBulkStatus()}
            >
              批量改状态
            </Button>
            <Button
              size="small"
              disabled={selectedDoneCount === 0 || bulkBusy}
              loading={bulkBusy}
              onClick={() => void handleBulkArchive()}
            >
              批量归档{selectedDoneCount > 0 ? `（${selectedDoneCount}）` : ''}
            </Button>
            {archiveView !== 'hide' ? (
              <Button
                size="small"
                disabled={selectedArchivedCount === 0 || bulkBusy}
                loading={bulkBusy}
                onClick={() => void handleBulkUnarchive()}
              >
                批量解档{selectedArchivedCount > 0 ? `（${selectedArchivedCount}）` : ''}
              </Button>
            ) : null}
            <Button
              size="small"
              type="text"
              icon={<CloseOutlined />}
              disabled={bulkBusy}
              onClick={clearSelection}
            >
              取消选择
            </Button>
          </Space>
        </div>
      ) : null}
      {loading ? (
        <div style={{ textAlign: 'center', padding: 48 }}>
          <Spin />
        </div>
      ) : forceEmpty ||
        (issues.length === 0 && !hasActiveFilters(filters) && archiveView === 'hide') ? (
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
          rowSelection={{
            selectedRowKeys,
            onChange: (keys) => setSelectedRowKeys(keys),
            preserveSelectedRowKeys: true,
            columnWidth: 40,
          }}
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
                  description={
                    archiveView === 'only' ? '暂无已归档的 Issue' : '没有符合条件的 Issue'
                  }
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
              onClick: (e) => {
                const t = e.target as HTMLElement
                if (t.closest('.ant-checkbox-wrapper, .ant-table-selection-column')) return
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
          // hide：归档行踢出；only：解档行踢出；include：始终保留/更新
          if (archiveView === 'hide' && archived) {
            setIssues((prev) => prev.filter((i) => i.id !== saved.id))
            setCurrent(saved)
            return
          }
          if (archiveView === 'only' && !archived) {
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
