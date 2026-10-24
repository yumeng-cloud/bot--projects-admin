import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import {
  Button,
  Card,
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

export default function BotDetailPage() {
  const { wsId = '', memberId = '' } = useParams()
  const [loading, setLoading] = useState(true)
  const [memories, setMemories] = useState<BotMemory[]>([])
  const [todos, setTodos] = useState<BotTodo[]>([])
  const [handoff, setHandoff] = useState<BotHandoff | null>(null)
  const [q, setQ] = useState('')
  const [importJson, setImportJson] = useState('')
  const [importing, setImporting] = useState(false)

  const load = async () => {
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
      message.error(e instanceof Error ? e.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load()
  }, [wsId, memberId])

  if (loading && !handoff) return <Spin />

  return (
    <div className="acb-page">
      <Space style={{ marginBottom: 16 }}>
        <Link to={`/w/${wsId}/settings/bots`}>← 名册</Link>
        <Typography.Title level={4} style={{ margin: 0 }}>
          Bot 详情
        </Typography.Title>
        <Typography.Text type="secondary" code>
          {memberId}
        </Typography.Text>
        {handoff?.claimed_role && <Tag color="blue">{handoff.claimed_role.name}</Tag>}
      </Space>

      <Tabs
        items={[
          {
            key: 'memories',
            label: `记忆 (${memories.length})`,
            children: (
              <>
                <Space style={{ marginBottom: 12 }} wrap>
                  <Input.Search
                    placeholder="搜索 q"
                    allowClear
                    onSearch={(v) => {
                      setQ(v)
                      void listMemories(wsId, memberId, v || undefined).then((r) => setMemories(r.memories || []))
                    }}
                    style={{ width: 240 }}
                  />
                </Space>
                <Card size="small" title="新增 / Upsert" style={{ marginBottom: 16 }}>
                  <Form
                    layout="inline"
                    onFinish={async (v) => {
                      await upsertMemory(wsId, memberId, {
                        key_name: v.key_name,
                        value_text: v.value_text,
                        tags: v.tags || '[]',
                      }, crypto.randomUUID())
                      message.success('已保存')
                      void load()
                    }}
                  >
                    <Form.Item name="key_name" rules={[{ required: true }]}>
                      <Input placeholder="key_name" />
                    </Form.Item>
                    <Form.Item name="value_text" rules={[{ required: true }]}>
                      <Input placeholder="value_text" style={{ width: 280 }} />
                    </Form.Item>
                    <Form.Item name="tags">
                      <Input placeholder='tags JSON 默认 []' style={{ width: 160 }} />
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
                      { title: 'key', dataIndex: 'key_name', width: 140 },
                      { title: 'value', dataIndex: 'value_text', ellipsis: true },
                      { title: 'tags', dataIndex: 'tags', width: 120 },
                      {
                        title: '',
                        width: 80,
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
            label: `Todos (${todos.length})`,
            children: (
              <>
                <Card size="small" title="新建待办" style={{ marginBottom: 16 }}>
                  <Form
                    layout="inline"
                    onFinish={async (v) => {
                      await createTodo(wsId, memberId, { title: v.title, status: v.status || 'todo' }, crypto.randomUUID())
                      message.success('已创建')
                      void load()
                    }}
                  >
                    <Form.Item name="title" rules={[{ required: true }]}>
                      <Input placeholder="title" style={{ width: 280 }} />
                    </Form.Item>
                    <Form.Item name="status" initialValue="todo">
                      <Select
                        style={{ width: 120 }}
                        options={[
                          { value: 'todo', label: 'todo' },
                          { value: 'doing', label: 'doing' },
                          { value: 'done', label: 'done' },
                          { value: 'canceled', label: 'canceled' },
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
                              { value: 'todo', label: 'todo' },
                              { value: 'doing', label: 'doing' },
                              { value: 'done', label: 'done' },
                              { value: 'canceled', label: 'canceled' },
                            ]}
                            onChange={async (status) => {
                              await patchTodo(wsId, memberId, row.id, { status }, crypto.randomUUID())
                              void load()
                            }}
                          />
                        ),
                      },
                      { title: 'issue', dataIndex: 'issue_id', render: (v) => v || '—' },
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
            label: 'Handoff',
            children: (
              <Space direction="vertical" style={{ width: '100%' }} size="middle">
                {handoff ? (
                  <>
                    <Typography.Text type="secondary">
                      生成于 {handoff.generated_at}（UTC 存库）
                    </Typography.Text>
                    <Card size="small" title={`开放 Todos (${handoff.todos?.length || 0})`}>
                      {(handoff.todos || []).map((t) => (
                        <div key={t.id}>
                          <Tag>{t.status}</Tag> {t.title}
                        </div>
                      ))}
                      {!handoff.todos?.length && <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />}
                    </Card>
                    <Card size="small" title={`指派 Issues (${handoff.issues?.length || 0})`}>
                      {(handoff.issues || []).map((i) => (
                        <div key={i.id}>
                          <Tag color="purple">{i.identifier}</Tag> {i.title} <Tag>{i.status}</Tag>
                        </div>
                      ))}
                      {!handoff.issues?.length && <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />}
                    </Card>
                    <Card size="small" title={`Memories (${handoff.memories?.length || 0})`}>
                      <pre style={{ maxHeight: 240, overflow: 'auto', margin: 0 }}>
                        {JSON.stringify(handoff.memories, null, 2)}
                      </pre>
                    </Card>
                    <Space>
                      <Button onClick={() => void load()}>刷新 handoff</Button>
                      <Button
                        onClick={() => {
                          const blob = new Blob(
                            [JSON.stringify({ handoff }, null, 2)],
                            { type: 'application/json' },
                          )
                          const url = URL.createObjectURL(blob)
                          const a = document.createElement('a')
                          a.href = url
                          a.download = `handoff-${memberId}.json`
                          a.click()
                          URL.revokeObjectURL(url)
                        }}
                      >
                        导出 JSON
                      </Button>
                    </Space>
                  </>
                ) : (
                  <Empty description="暂无 handoff" />
                )}
                <Card size="small" title="导入 Handoff">
                  <Input.TextArea
                    rows={8}
                    placeholder='粘贴 JSON：{ "handoff": {...} } 或直接手写 handoff 对象'
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
                    导入
                  </Button>
                </Card>
              </Space>
            ),
          },
        ]}
      />
    </div>
  )
}
