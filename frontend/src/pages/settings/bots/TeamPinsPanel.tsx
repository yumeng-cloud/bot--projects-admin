import { useCallback, useEffect, useState } from 'react'
import {
  Alert,
  Button,
  Card,
  Empty,
  Form,
  Input,
  Popconfirm,
  Space,
  Spin,
  Table,
  Typography,
  message,
} from 'antd'
import { DeleteOutlined, EditOutlined, PlusOutlined } from '@ant-design/icons'
import { ApiError } from '../../../api/client'
import {
  deletePin,
  isPinsServiceMissing,
  listPins,
  upsertPin,
  type TeamPin,
} from '../../../api/pins'
import { formatWikiTime } from '../../../utils/time'

type Props = {
  wsId: string
  teamId: string | null
}

/**
 * P5 最小记忆墙：本队共享 pins；切 team 必须换池、无跨队残留。
 * 与 Bot 私有「记忆」分清：标题「记忆墙」，说明「本队共享，非席位私有记忆」。
 */
export default function TeamPinsPanel({ wsId, teamId }: Props) {
  const [loading, setLoading] = useState(false)
  const [pins, setPins] = useState<TeamPin[]>([])
  const [apiMissing, setApiMissing] = useState(false)
  const [hint, setHint] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const [editingKey, setEditingKey] = useState<string | null>(null)
  const [form] = Form.useForm<{ key: string; value: string }>()

  const load = useCallback(async (tid: string) => {
    setLoading(true)
    setHint(null)
    try {
      const res = await listPins(wsId, tid)
      // 再次确认仍是当前队，避免竞态残留
      setPins(res.pins || [])
      setApiMissing(false)
    } catch (e: unknown) {
      setPins([])
      if (isPinsServiceMissing(e)) {
        setApiMissing(true)
        setHint('接口尚未就绪')
      } else {
        setApiMissing(false)
        setHint(e instanceof ApiError ? e.message : '加载记忆墙失败')
      }
    } finally {
      setLoading(false)
    }
  }, [wsId])

  useEffect(() => {
    // 换队：先清空，杜绝跨队残留（P5）
    setPins([])
    setApiMissing(false)
    setHint(null)
    setEditingKey(null)
    form.resetFields()
    if (!teamId) return
    void load(teamId)
  }, [wsId, teamId, load, form])

  const startAdd = () => {
    setEditingKey(null)
    form.setFieldsValue({ key: '', value: '' })
  }

  const startEdit = (row: TeamPin) => {
    setEditingKey(row.key)
    form.setFieldsValue({ key: row.key, value: row.value })
  }

  const onSave = async () => {
    if (!teamId) return
    try {
      const values = await form.validateFields()
      const key = values.key.trim()
      const value = values.value ?? ''
      if (!key) {
        message.warning('key 不能为空')
        return
      }
      setSaving(true)
      const { pin } = await upsertPin(wsId, teamId, { key, value }, crypto.randomUUID())
      setPins((prev) => {
        const rest = prev.filter((p) => p.key !== pin.key)
        return [...rest, pin].sort((a, b) => a.key.localeCompare(b.key))
      })
      setApiMissing(false)
      setHint(null)
      form.resetFields()
      setEditingKey(null)
      message.success(editingKey ? '已更新' : '已保存')
    } catch (e: unknown) {
      if (e && typeof e === 'object' && 'errorFields' in e) return
      if (isPinsServiceMissing(e)) {
        setApiMissing(true)
        setHint('接口尚未就绪')
        message.warning('接口尚未就绪')
      } else {
        message.error(e instanceof ApiError ? e.message : '保存失败')
      }
    } finally {
      setSaving(false)
    }
  }

  const onDelete = async (key: string) => {
    if (!teamId) return
    try {
      await deletePin(wsId, teamId, key, crypto.randomUUID())
      setPins((prev) => prev.filter((p) => p.key !== key))
      if (editingKey === key) {
        form.resetFields()
        setEditingKey(null)
      }
      message.success('已删除')
    } catch (e: unknown) {
      if (isPinsServiceMissing(e)) {
        setApiMissing(true)
        setHint('接口尚未就绪')
        message.warning('接口尚未就绪')
      } else {
        message.error(e instanceof ApiError ? e.message : '删除失败')
      }
    }
  }

  if (!teamId) {
    return (
      <Card size="small" className="acb-bots-pins" title="记忆墙" style={{ marginBottom: 12 }}>
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="请先选择团队" />
      </Card>
    )
  }

  return (
    <Card
      size="small"
      className="acb-bots-pins"
      style={{ marginBottom: 12 }}
      title={
        <Space size={8} wrap>
          <span>记忆墙</span>
          <Typography.Text type="secondary" style={{ fontSize: 12, fontWeight: 400 }}>
            本队共享，非席位私有记忆
          </Typography.Text>
        </Space>
      }
      extra={
        <Button
          size="small"
          type="link"
          icon={<PlusOutlined />}
          disabled={apiMissing}
          onClick={startAdd}
        >
          新增
        </Button>
      }
    >
      {apiMissing || hint ? (
        <Alert
          type={apiMissing ? 'info' : 'warning'}
          showIcon
          style={{ marginBottom: 8 }}
          message={hint || '接口尚未就绪'}
          description={
            apiMissing
              ? '团队 Pins REST 契约已齐、实现待接线。可稍后刷新；不会影响其它设置。'
              : undefined
          }
        />
      ) : null}

      <Form
        form={form}
        layout="inline"
        size="small"
        style={{ marginBottom: 8, rowGap: 8, flexWrap: 'wrap' }}
        onFinish={() => void onSave()}
      >
        <Form.Item
          name="key"
          rules={[{ required: true, message: '填写 key' }]}
          style={{ marginBottom: 0, minWidth: 120, flex: '1 1 120px' }}
        >
          <Input
            placeholder="key"
            disabled={apiMissing || Boolean(editingKey)}
            allowClear={!editingKey}
          />
        </Form.Item>
        <Form.Item name="value" style={{ marginBottom: 0, minWidth: 160, flex: '2 1 160px' }}>
          <Input placeholder="value" disabled={apiMissing} allowClear />
        </Form.Item>
        <Form.Item style={{ marginBottom: 0 }}>
          <Space size={4}>
            <Button type="primary" htmlType="submit" loading={saving} disabled={apiMissing}>
              {editingKey ? '更新' : '保存'}
            </Button>
            {editingKey ? (
              <Button
                onClick={() => {
                  setEditingKey(null)
                  form.resetFields()
                }}
              >
                取消
              </Button>
            ) : null}
          </Space>
        </Form.Item>
      </Form>

      {loading ? (
        <div style={{ textAlign: 'center', padding: 16 }}>
          <Spin size="small" />
        </div>
      ) : pins.length ? (
        <Table
          className="acb-list-table"
          size="small"
          rowKey={(r) => r.id || r.key}
          pagination={false}
          dataSource={pins}
          columns={[
            {
              title: 'key',
              dataIndex: 'key',
              ellipsis: true,
              width: '28%',
              render: (v: string) => (
                <Typography.Text style={{ fontFamily: 'var(--acb-font-mono, monospace)', fontSize: 12 }}>
                  {v}
                </Typography.Text>
              ),
            },
            {
              title: 'value',
              dataIndex: 'value',
              ellipsis: true,
              render: (v: string) => (
                <Typography.Text style={{ fontSize: 13 }} title={v}>
                  {v || '—'}
                </Typography.Text>
              ),
            },
            {
              title: '更新',
              dataIndex: 'updated_at',
              width: 110,
              render: (v: string) => (
                <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                  {v ? formatWikiTime(v) : '—'}
                </Typography.Text>
              ),
            },
            {
              title: '',
              width: 88,
              render: (_, row) => (
                <Space size={0} onClick={(e) => e.stopPropagation()}>
                  <Button
                    size="small"
                    type="link"
                    icon={<EditOutlined />}
                    style={{ padding: '0 4px' }}
                    disabled={apiMissing}
                    onClick={() => startEdit(row)}
                  />
                  <Popconfirm
                    title="删除此 pin？"
                    okText="删除"
                    cancelText="取消"
                    okButtonProps={{ danger: true }}
                    onConfirm={() => void onDelete(row.key)}
                  >
                    <Button
                      size="small"
                      type="link"
                      danger
                      icon={<DeleteOutlined />}
                      style={{ padding: '0 4px' }}
                      disabled={apiMissing}
                    />
                  </Popconfirm>
                </Space>
              ),
            },
          ]}
        />
      ) : !apiMissing ? (
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="本队暂无 pin" />
      ) : null}
    </Card>
  )
}
