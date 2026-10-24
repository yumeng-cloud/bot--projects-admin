import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Alert,
  Button,
  Col,
  Dropdown,
  Empty,
  Progress,
  Row,
  Select,
  Space,
  Tag,
  Typography,
  message,
  Spin,
  List,
} from 'antd'
import type { MenuProps } from 'antd'
import { MoreOutlined, ReloadOutlined } from '@ant-design/icons'
import { listCycles, patchCycle, cycleEstimateSum } from '../../api/cycles'
import { listIssues, patchIssue } from '../../api/issues'
import { BoardPackStrip } from '../../components/board'
import { IssueDrawer } from '../../components/issue'
import { MockBanner } from '../../components/common/MockBanner'
import { useBoardPackStats } from '../../hooks/useBoardPackStats'
import type { Cycle, Issue } from '../../types/issue'
import { STATUS_LABELS } from '../../types/issue'
import { cycleStatusLabel } from '../../utils/labels'
import '../../components/board/board.css'

export type CyclesPageProps = {
  workspaceId: string
  projectId: string
}

export default function CyclesPage({ workspaceId, projectId }: CyclesPageProps) {
  const navigate = useNavigate()
  const pack = useBoardPackStats({ workspaceId, projectId })
  const [loading, setLoading] = useState(true)
  const [cycles, setCycles] = useState<Cycle[]>([])
  const [issues, setIssues] = useState<Issue[]>([])
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [open, setOpen] = useState(false)
  const [current, setCurrent] = useState<Issue | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const [c, i] = await Promise.all([listCycles(workspaceId, projectId), listIssues(workspaceId, { project_id: projectId })])
      setCycles(c.cycles)
      setIssues(i.issues)
      if (!selectedId && c.cycles[0]) setSelectedId(c.cycles[0].id)
    } finally {
      setLoading(false)
    }
  }, [workspaceId, projectId, selectedId])

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [projectId])

  const selected = cycles.find((c) => c.id === selectedId) || null
  const cycleIssues = useMemo(
    () => issues.filter((i) => i.cycle_id === selectedId),
    [issues, selectedId],
  )
  const used = selected ? cycleEstimateSum(selected.id, issues) : 0
  const overload = selected ? used > selected.capacity : false
  const pct = selected && selected.capacity > 0 ? Math.round((used / selected.capacity) * 100) : 0

  const completeCycle = async () => {
    if (!selected) return
    try {
      const { cycle } = await patchCycle(workspaceId, selected.id, { status: 'completed' })
      message.success(`已结束 ${cycle.name}，未完成项已回待办池（演示）`)
      await load()
    } catch (e) {
      message.error((e as Error).message || '结束失败')
    }
  }

  const assignIssue = async (issueId: string, cycleId: string | null) => {
    const issue = issues.find((i) => i.id === issueId)
    if (!issue) return
    try {
      const { issue: saved } = await patchIssue(workspaceId, issue.id, {
        version: issue.version,
        cycle_id: cycleId,
      })
      setIssues((prev) => prev.map((i) => (i.id === saved.id ? saved : i)))
      message.success(cycleId ? '已划入 Cycle' : '已移出 Cycle')
    } catch {
      // client 已 Toast（含 version_conflict / conflict）
    }
  }

  if (loading) {
    return (
      <div style={{ padding: 48, textAlign: 'center' }}>
        <Spin />
      </div>
    )
  }

  return (
    <div className="acb-page">
      <MockBanner />
      <div className="acb-page-bar">
        <Typography.Title level={4}>Cycles</Typography.Title>
        <Button
          type="text"
          size="small"
          icon={<ReloadOutlined />}
          onClick={() => {
            void load()
            void pack.reload()
          }}
        >
          刷新
        </Button>
      </div>
      {!pack.loading ? (
        <BoardPackStrip
          done={pack.done}
          total={pack.total}
          activeCycle={pack.activeCycle}
          collaborators={pack.collaborators}
          blockerCount={pack.blockerCount}
          onCycleClick={() => {
            if (pack.activeCycle) {
              setSelectedId(pack.activeCycle.id)
              const el = document.getElementById(`acb-cycle-${pack.activeCycle.id}`)
              el?.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
            }
          }}
          onBlockersClick={() => {
            navigate(`/w/${workspaceId}/projects/${projectId}/issues?blocked=1&overdue=1`)
          }}
        />
      ) : null}
      <Row gutter={16} className="acb-cycles-row">
        <Col span={8}>
          <div className="acb-cycles-panel">
            <Typography.Text strong style={{ display: 'block', marginBottom: 8 }}>
              周期列表
            </Typography.Text>
            <List
              size="small"
              dataSource={cycles}
              locale={{
                emptyText: (
                  <div className="acb-empty-panel" style={{ minHeight: 120 }}>
                    <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无 Cycle" />
                  </div>
                ),
              }}
              renderItem={(c) => {
                const u = cycleEstimateSum(c.id, issues)
                const over = u > c.capacity
                return (
                  <List.Item
                    id={`acb-cycle-${c.id}`}
                    onClick={() => setSelectedId(c.id)}
                    style={{
                      cursor: 'pointer',
                      background: c.id === selectedId ? 'color-mix(in srgb, var(--acb-primary) 12%, var(--acb-bg-container))' : undefined,
                      padding: 8,
                      borderRadius: 6,
                    }}
                  >
                    <Space direction="vertical" size={0} style={{ width: '100%' }}>
                      <Space>
                        <Typography.Text strong>{c.name}</Typography.Text>
                        <Tag>{cycleStatusLabel(c.status)}</Tag>
                        {over && <Tag color="error">过载</Tag>}
                      </Space>
                      <Typography.Text type="secondary">
                        {u} / {c.capacity} pt
                      </Typography.Text>
                    </Space>
                  </List.Item>
                )
              }}
            />
          </div>
        </Col>
        <Col span={16}>
          {selected ? (
            <div className="acb-cycles-panel">
              <div className="acb-page-bar" style={{ marginBottom: 8, padding: 0 }}>
                <Typography.Text strong>{selected.name}</Typography.Text>
                {selected.status !== 'completed' ? (
                  <Button danger size="small" onClick={() => void completeCycle()}>
                    结束 Cycle
                  </Button>
                ) : null}
              </div>
              {overload && (
                <Alert
                  type="error"
                  showIcon
                  message={`容量过载：估时合计 ${used} pt，容量 ${selected.capacity} pt`}
                  style={{ marginBottom: 12 }}
                />
              )}
              <Progress
                percent={Math.min(pct, 100)}
                status={overload ? 'exception' : 'active'}
                format={() => `${used}/${selected.capacity} pt`}
              />
              <Typography.Paragraph type="secondary" style={{ marginTop: 12 }}>
                {selected.start_date || '—'} ~ {selected.end_date || '—'}
              </Typography.Paragraph>
              <Typography.Title level={5}>划入 Issue</Typography.Title>
              <Select
                style={{ width: '100%', marginBottom: 12 }}
                placeholder="选择未划入的 Issue"
                options={issues
                  .filter((i) => i.cycle_id !== selected.id)
                  .map((i) => ({
                    value: i.id,
                    label: `${i.identifier} ${i.title}`,
                  }))}
                onSelect={(id: string) => assignIssue(id, selected.id)}
              />
              <List
                dataSource={cycleIssues}
                locale={{
                  emptyText: (
                    <div className="acb-empty-panel" style={{ minHeight: 120 }}>
                      <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="本 Cycle 暂无 Issue" />
                    </div>
                  ),
                }}
                renderItem={(i) => {
                  const menu: MenuProps = {
                    items: [
                      {
                        key: 'open',
                        label: '打开',
                        onClick: () => {
                          setCurrent(i)
                          setOpen(true)
                        },
                      },
                      {
                        key: 'rm',
                        label: '移出',
                        danger: true,
                        onClick: () => void assignIssue(i.id, null),
                      },
                    ],
                  }
                  return (
                    <List.Item
                      actions={[
                        <Dropdown key="more" menu={menu} trigger={['click']}>
                          <Button type="text" size="small" icon={<MoreOutlined />} aria-label="更多" />
                        </Dropdown>,
                      ]}
                    >
                      <List.Item.Meta
                        title={`${i.identifier} ${i.title}`}
                        description={`${STATUS_LABELS[i.status]} · 估时 ${i.estimate ?? 0}pt`}
                      />
                    </List.Item>
                  )
                }}
              />
            </div>
          ) : (
            <div className="acb-cycles-panel">
              <div className="acb-empty-panel" style={{ minHeight: 140 }}>
                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="选择左侧 Cycle 查看容量与 Issue" />
              </div>
            </div>
          )}
        </Col>
      </Row>
      <IssueDrawer
        open={open}
        issue={current}
        workspaceId={workspaceId}
        projectId={projectId}
        onClose={() => setOpen(false)}
        onSaved={(saved) => {
          setIssues((prev) => prev.map((i) => (i.id === saved.id ? saved : i)))
          setCurrent(saved)
        }}
      />
    </div>
  )
}
