import { useEffect, useMemo, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import {
  Alert,
  Button,
  Card,
  Collapse,
  Form,
  Input,
  Select,
  Space,
  Table,
  Tabs,
  Tag,
  Typography,
  message,
  Spin,
  Empty,
} from 'antd'
import {
  createTodo,
  getHandoff,
  importHandoff,
  listMemories,
  listTodos,
  patchTodo,
  upsertMemory,
  deleteMemory,
  type BotHandoff,
  type BotMemory,
  type BotTodo,
} from '../../../api/bots'
import { ApiError } from '../../../api/client'
import { readStoredProjectId } from '../../../utils/currentProject'
import { formatShanghai, formatWikiTime } from '../../../utils/time'
import { openTodos, pickLastFocus, statusCn } from './seatHandoffText'
import './bots.css'

export default function BotDetailPage() {
  const { wsId = '', memberId = '' } = useParams()
  const nav = useNavigate()
  const [loading, setLoading] = useState(true)
  const [memories, setMemories] = useState<BotMemory[]>([])
  const [todos, setTodos] = useState<BotTodo[]>([])
  const [handoff, setHandoff] = useState<BotHandoff | null>(null)
  const [q, setQ] = useState('')
  const [importJson, setImportJson] = useState('')
  const [importing, setImporting] = useState(false)
  /** X-R3：当前仓项目上下文（acb.project.<wsId>） */
  const projectId = readStoredProjectId(wsId)

  const clearBotData = () => {
    setMemories([])
    setTodos([])
    setHandoff(null)
  }

  // X-R2：切仓/换 member 先 clear，再 fetch（等价 queryKey 含 wsId）；cancelled 防竞态
  useEffect(() => {
    let cancelled = false
    clearBotData()
    setQ('')
    setImportJson('')
    if (!wsId || !memberId) {
      setLoading(false)
      return () => {
        cancelled = true
      }
    }
    setLoading(true)
    ;(async () => {
      try {
        const [m, t, h] = await Promise.all([
          listMemories(wsId, memberId),
          listTodos(wsId, memberId),
          getHandoff(wsId, memberId),
        ])
        if (cancelled) return
        setMemories(m.memories || [])
        setTodos(t.todos || [])
        setHandoff(h.handoff)
      } catch (e: unknown) {
        if (cancelled) return
        // X-R2：失败保持空态，禁止继续渲染上一仓 memories/handoff
        clearBotData()
        const status = e instanceof ApiError ? e.status : 0
        if (status === 404 || status === 403) {
          message.error('该 Bot 不属于当前工作区')
          nav(`/w/${wsId}/settings/bots`, { replace: true })
          return
        }
        message.error(e instanceof Error ? e.message : '加载失败')
      } finally {
        if (!cancelled) setLoading(false)
      }
    })()
    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [wsId, memberId])

  const load = async () => {
    if (!wsId || !memberId) {
      clearBotData()
      setLoading(false)
      return
    }
    setLoading(true)
    try {
      const [m, t, h] = await Promise.all([
        listMemories(wsId, memberId, q || undefined),
        listTodos(wsId, memberId),
        getHandoff(wsId, memberId),
      ])
      setMemories(m.memories || [])
      setTodos(t.todos || [])
      setHandoff(h.handoff)
    } catch (e: unknown) {
      clearBotData()
      const status = e instanceof ApiError ? e.status : 0
      if (status === 404 || status === 403) {
        message.error('该 Bot 不属于当前工作区')
        nav(`/w/${wsId}/settings/bots`, { replace: true })
        return
      }
      message.error(e instanceof Error ? e.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }

  // X-R3 / B1：有项目上下文时 handoff Issues 仅当前项目
  const handoffIssues = useMemo(() => {
    const all = handoff?.issues || []
    if (!projectId) return all
    return all.filter((i) => i.project_id === projectId)
  }, [handoff?.issues, projectId])

  if (loading && !handoff) {
    return (
      <div className="acb-page acb-bots" style={{ textAlign: 'center', paddingTop: 48 }}>
        <Spin />
      </div>
    )
  }

  const roleName = handoff?.claimed_role?.name
  const focus = pickLastFocus(handoff?.memories, handoff?.claimed_role?.slug)
  const openTd = openTodos(handoff?.todos)

  return (
    <div className="acb-page acb-bots">
      <Space style={{ marginBottom: 12 }} wrap>
        <Link to={`/w/${wsId}/settings/bots`}>← 席位</Link>
        <Typography.Title level={4} style={{ margin: 0 }}>
          Bot 记忆与交接
        </Typography.Title>
        {roleName ? <Tag color="blue">席位 · {roleName}</Tag> : <Tag>未占席</Tag>}
      </Space>
      <Typography.Paragraph type="secondary" style={{ marginBottom: 12, fontSize: 13 }}>
        演员级记忆 / Todo；换机请回席位页按「绑席位 → handoff get」续，不要日常依赖下方 JSON 导入导出。
      </Typography.Paragraph>

      {handoff ? (
        <Card size="small" style={{ marginBottom: 16 }} title="交接摘要（人话）">
          <Space direction="vertical" size={8} style={{ width: '100%' }}>
            <div>
              <Typography.Text type="secondary">当前焦点 · </Typography.Text>
              {focus || '（暂无）'}
            </div>
            <div>
              <Typography.Text type="secondary">进行中 Todos · </Typography.Text>
              {openTd.length
                ? openTd.map((t) => `[${statusCn(t.status)}] ${t.title}`).join('；')
                : '无'}
            </div>
            <div>
              <Typography.Text type="secondary">开放 Issues · </Typography.Text>
              {handoffIssues.length
                ? handoffIssues.map((i) => `${i.identifier} ${i.title}`).join('；')
                : '无'}
            </div>
            <Typography.Text type="secondary" style={{ fontSize: 12 }}>
              生成于 {formatWikiTime(handoff.generated_at)}（{formatShanghai(handoff.generated_at)}）
            </Typography.Text>
          </Space>
        </Card>
      ) : null}

      <Tabs
        items={[
          {
            key: 'memories',
            label: `记忆 (${memories.length})`,
            children: (
              <>
                <Space style={{ marginBottom: 12 }} wrap>
                  <Input.Search
                    placeholder="搜索记忆"
                    allowClear
                    onSearch={(v) => {
                      setQ(v)
                      void listMemories(wsId, memberId, v || undefined).then((r) =>
                        setMemories(r.memories || []),
                      )
                    }}
                    style={{ width: 240 }}
                  />
                </Space>
                <Card size="small" title="新增 / 更新" style={{ marginBottom: 16 }}>
                  <Form
                    layout="inline"
                    onFinish={async (v) => {
                      await upsertMemory(
                        wsId,
                        memberId,
                        {
                          key_name: v.key_name,
                          value_text: v.value_text,
                          tags: v.tags || '[]',
                        },
                        crypto.randomUUID(),
                      )
                      message.success('已保存')
                      void load()
                    }}
                  >
                    <Form.Item name="key_name" rules={[{ required: true }]}>
                      <Input placeholder="键名，如 last_focus" />
                    </Form.Item>
                    <Form.Item name="value_text" rules={[{ required: true }]}>
                      <Input placeholder="内容" style={{ width: 280 }} />
                    </Form.Item>
                    <Form.Item name="tags" hidden>
                      <Input />
                    </Form.Item>
                    <Button type="primary" htmlType="submit">
                      保存
                    </Button>
                  </Form>
                </Card>
                {memories.length ? (
                  <Table
                    rowKey="id"
                    size="small"
                    dataSource={memories}
                    pagination={false}
                    columns={[
                      { title: '键', dataIndex: 'key_name', width: 160 },
                      { title: '值', dataIndex: 'value_text', ellipsis: true },
                      {
                        title: '更新',
                        dataIndex: 'updated_at',
                        width: 140,
                        render: (v: string) => formatWikiTime(v),
                      },
                      {
                        title: '',
                        width: 72,
                        render: (_, row) => (
                          <Button
                            danger
                            size="small"
                            onClick={async () => {
                              await deleteMemory(wsId, memberId, row.id)
                              message.success('已删')
                              void load()
                            }}
                          >
                            删
                          </Button>
                        ),
                      },
                    ]}
                  />
                ) : (
                  <Empty description="暂无记忆" />
                )}
              </>
            ),
          },
          {
            key: 'todos',
            label: `待办 (${todos.length})`,
            children: (
              <>
                <Card size="small" title="新建待办" style={{ marginBottom: 16 }}>
                  <Form
                    layout="inline"
                    onFinish={async (v) => {
                      await createTodo(
                        wsId,
                        memberId,
                        { title: v.title, status: v.status || 'todo' },
                        crypto.randomUUID(),
                      )
                      message.success('已创建')
                      void load()
                    }}
                  >
                    <Form.Item name="title" rules={[{ required: true }]}>
                      <Input placeholder="标题" style={{ width: 280 }} />
                    </Form.Item>
                    <Form.Item name="status" initialValue="todo">
                      <Select
                        style={{ width: 120 }}
                        options={[
                          { value: 'todo', label: '待办' },
                          { value: 'doing', label: '进行中' },
                          { value: 'done', label: '已完成' },
                          { value: 'canceled', label: '已取消' },
                        ]}
                      />
                    </Form.Item>
                    <Button type="primary" htmlType="submit">
                      添加
                    </Button>
                  </Form>
                </Card>
                {todos.length ? (
                  <Table
                    rowKey="id"
                    size="small"
                    dataSource={todos}
                    pagination={false}
                    columns={[
                      { title: '标题', dataIndex: 'title' },
                      {
                        title: '状态',
                        dataIndex: 'status',
                        width: 160,
                        render: (s: string, row) => (
                          <Select
                            size="small"
                            value={s}
                            style={{ width: 120 }}
                            options={[
                              { value: 'todo', label: '待办' },
                              { value: 'doing', label: '进行中' },
                              { value: 'done', label: '已完成' },
                              { value: 'canceled', label: '已取消' },
                            ]}
                            onChange={async (status) => {
                              await patchTodo(wsId, memberId, row.id, { status }, crypto.randomUUID())
                              void load()
                            }}
                          />
                        ),
                      },
                    ]}
                  />
                ) : (
                  <Empty description="暂无待办" />
                )}
              </>
            ),
          },
          {
            key: 'handoff',
            label: '交接包',
            children: (
              <Space direction="vertical" style={{ width: '100%' }} size="middle">
                {handoff ? (
                  <>
                    <Card size="small" title={`进行中 Todos（${openTd.length}）`}>
                      {openTd.map((t) => (
                        <div key={t.id}>
                          <Tag>{statusCn(t.status)}</Tag> {t.title}
                        </div>
                      ))}
                      {!openTd.length && <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />}
                    </Card>
                    <Card size="small" title={`开放 Issues（${handoffIssues.length}）`}>
                      {handoffIssues.map((i) => (
                        <div key={i.id}>
                          <Tag color="purple">{i.identifier}</Tag> {i.title}{' '}
                          <Tag>{statusCn(i.status)}</Tag>
                        </div>
                      ))}
                      {!handoffIssues.length && <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />}
                    </Card>
                    <Button onClick={() => void load()}>刷新交接</Button>
                  </>
                ) : (
                  <Empty description="暂无交接数据" />
                )}

                <Collapse
                  size="small"
                  items={[
                    {
                      key: 'backup',
                      label: '可选备份 / 救场（非默认换机路径）',
                      children: (
                        <Space direction="vertical" style={{ width: '100%' }} size="middle">
                          <Alert
                            type="warning"
                            showIcon
                            message="日常换机请走：绑席位 → handoff get"
                            description="下方 JSON 导出/导入仅作冷备或服务异常时的救场，不要写成默认剧本。"
                          />
                          <Space>
                            <Button
                              disabled={!handoff}
                              onClick={() => {
                                if (!handoff) return
                                const blob = new Blob([JSON.stringify({ handoff }, null, 2)], {
                                  type: 'application/json',
                                })
                                const url = URL.createObjectURL(blob)
                                const a = document.createElement('a')
                                a.href = url
                                a.download = `handoff-backup.json`
                                a.click()
                                URL.revokeObjectURL(url)
                              }}
                            >
                              导出备份 JSON
                            </Button>
                          </Space>
                          <Card size="small" title="导入备份（救场）">
                            <Input.TextArea
                              rows={6}
                              placeholder='粘贴备份 JSON：{ "handoff": {...} }'
                              value={importJson}
                              onChange={(e) => setImportJson(e.target.value)}
                              style={{ fontFamily: 'monospace', fontSize: 12 }}
                            />
                            <Button
                              type="primary"
                              loading={importing}
                              style={{ marginTop: 12 }}
                              onClick={async () => {
                                let parsed: unknown
                                try {
                                  parsed = JSON.parse(importJson)
                                } catch {
                                  message.error('JSON 解析失败')
                                  return
                                }
                                const body =
                                  parsed &&
                                  typeof parsed === 'object' &&
                                  'handoff' in (parsed as object)
                                    ? (parsed as { handoff: BotHandoff })
                                    : { handoff: parsed as BotHandoff }
                                setImporting(true)
                                try {
                                  const res = await importHandoff(
                                    wsId,
                                    memberId,
                                    body,
                                    crypto.randomUUID(),
                                  )
                                  const warnings = res.warnings || []
                                  if (warnings.length) {
                                    warnings.forEach((w) => message.warning(w))
                                  }
                                  message.success(
                                    `导入完成：记忆 ${res.imported_memories} / Todo ${res.imported_todos}`,
                                  )
                                  setImportJson('')
                                  void load()
                                } catch {
                                  // api/client already toasts
                                } finally {
                                  setImporting(false)
                                }
                              }}
                            >
                              导入备份
                            </Button>
                          </Card>
                        </Space>
                      ),
                    },
                  ]}
                />
              </Space>
            ),
          },
        ]}
      />
    </div>
  )
}
