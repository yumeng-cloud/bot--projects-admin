import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  Button,
  DatePicker,
  Form,
  Input,
  InputNumber,
  Modal,
  Select,
  Space,
  Switch,
  Table,
  Tabs,
  Typography,
  message,
} from 'antd'
import { DownloadOutlined, PlusOutlined, ReloadOutlined } from '@ant-design/icons'
import { useParams } from 'react-router-dom'
import {
  billingSummary,
  createTimeEntry,
  deleteTimeEntry,
  exportBillingCsv,
  listBillingRates,
  listTimeEntries,
  upsertBillingRate,
} from '../../api/time'
import { listIssues, listProjectMembers } from '../../api/issues'
import type { BillingRate, BillingSummary, TimeEntry } from '../../types/time'
import type { Issue, MemberRef } from '../../types/issue'
import { memberLabel } from '../../types/issue'
import { issueReadable } from '../../utils/labels'
import type { Dayjs } from 'dayjs'
import { dayjs, formatShanghai, TZ } from '../../utils/time'

const { RangePicker } = DatePicker

export default function TimePage() {
  const { wsId = '' } = useParams()
  const [entries, setEntries] = useState<TimeEntry[]>([])
  const [rates, setRates] = useState<BillingRate[]>([])
  const [summary, setSummary] = useState<BillingSummary | null>(null)
  const [issues, setIssues] = useState<Issue[]>([])
  const [members, setMembers] = useState<MemberRef[]>([])
  const [loading, setLoading] = useState(false)
  const [entryModal, setEntryModal] = useState(false)
  const [rateModal, setRateModal] = useState(false)
  const [groupBy, setGroupBy] = useState<'project' | 'member' | 'date'>('project')
  const [range, setRange] = useState<[Dayjs, Dayjs] | null>(null)
  const [entryForm] = Form.useForm()
  const [rateForm] = Form.useForm()

  const issueById = useMemo(() => {
    const m = new Map<string, Issue>()
    for (const i of issues) m.set(i.id, i)
    return m
  }, [issues])

  const memberById = useMemo(() => {
    const m = new Map<string, MemberRef>()
    for (const x of members) m.set(x.id, x)
    return m
  }, [members])

  const resolveIssue = (r: TimeEntry) => {
    const iss = issueById.get(r.issue_id)
    return issueReadable(r.issue_identifier || iss?.identifier, iss?.title, r.issue_id)
  }

  const resolveMember = (memberId: string | null | undefined, name?: string | null) => {
    if (name && name.trim()) return name.trim()
    if (!memberId) return '工作区默认'
    const m = memberById.get(memberId)
    if (m) return memberLabel(m)
    return '未知成员'
  }

  const loadMeta = useCallback(async () => {
    if (!wsId) return
    try {
      const [i, m] = await Promise.all([
        listIssues(wsId, {}).catch(() => ({ issues: [] as Issue[] })),
        listProjectMembers(wsId).catch(() => ({ members: [] as MemberRef[] })),
      ])
      setIssues(i.issues || [])
      setMembers(m.members || [])
    } catch {
      /* ignore */
    }
  }, [wsId])

  const loadEntries = useCallback(async () => {
    if (!wsId) return
    setLoading(true)
    try {
      const q: Record<string, string> = {}
      if (range) {
        q.from = range[0].startOf('day').utc().toISOString()
        q.to = range[1].endOf('day').utc().toISOString()
      }
      const res = await listTimeEntries(wsId, q)
      setEntries(res.time_entries || [])
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : '加载工时失败')
    } finally {
      setLoading(false)
    }
  }, [wsId, range])

  const loadRates = useCallback(async () => {
    if (!wsId) return
    try {
      const res = await listBillingRates(wsId)
      setRates(res.billing_rates || [])
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : '加载费率失败')
    }
  }, [wsId])

  const loadSummary = useCallback(async () => {
    if (!wsId) return
    try {
      const q: { from?: string; to?: string; group_by: 'project' | 'member' | 'date' } = {
        group_by: groupBy,
      }
      if (range) {
        q.from = range[0].format('YYYY-MM-DD')
        q.to = range[1].format('YYYY-MM-DD')
      }
      const res = await billingSummary(wsId, q)
      setSummary(res.summary)
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : '加载汇总失败')
    }
  }, [wsId, groupBy, range])

  useEffect(() => {
    void loadMeta()
  }, [loadMeta])

  useEffect(() => {
    void loadEntries()
  }, [loadEntries])

  useEffect(() => {
    void loadRates()
  }, [loadRates])

  useEffect(() => {
    void loadSummary()
  }, [loadSummary])

  const submitEntry = async () => {
    const v = await entryForm.validateFields()
    await createTimeEntry(wsId, {
      issue_id: v.issue_id,
      hours: v.hours,
      note: v.note,
      billable: v.billable,
      member_id: v.member_id || undefined,
      worked_at: v.worked_at ? dayjs(v.worked_at).utc().toISOString() : undefined,
    })
    message.success('已记工时')
    setEntryModal(false)
    entryForm.resetFields()
    await loadEntries()
    await loadSummary()
  }

  const submitRate = async () => {
    const v = await rateForm.validateFields()
    await upsertBillingRate(wsId, {
      member_id: v.member_id || null,
      hourly_rate: v.hourly_rate,
      currency: v.currency || 'CNY',
    })
    message.success('费率已保存')
    setRateModal(false)
    rateForm.resetFields()
    await loadRates()
  }

  const onExport = async () => {
    try {
      await exportBillingCsv(wsId, {
        from: range?.[0]?.format('YYYY-MM-DD'),
        to: range?.[1]?.format('YYYY-MM-DD'),
      })
      message.success('已导出 CSV')
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : '导出失败')
    }
  }

  const issueOptions = issues.map((i) => ({
    value: i.id,
    label: issueReadable(i.identifier, i.title, i.id),
  }))

  const memberOptions = members.map((m) => ({
    value: m.id,
    label: memberLabel(m),
  }))

  return (
    <div className="acb-page">
      <div className="acb-page-bar">
        <Typography.Title level={4} style={{ margin: 0 }}>
          工时 / 计费
        </Typography.Title>
        <Space wrap size={8}>
          <Typography.Text type="secondary" style={{ fontSize: 12 }}>
            时区 {TZ}
          </Typography.Text>
          <RangePicker
            value={range}
            onChange={(v) => setRange(v as [Dayjs, Dayjs] | null)}
            allowClear
            size="small"
          />
          <Button
            size="small"
            icon={<ReloadOutlined />}
            onClick={() => {
              void loadMeta()
              void loadEntries()
              void loadSummary()
              void loadRates()
            }}
          >
            刷新
          </Button>
          <Button size="small" icon={<DownloadOutlined />} onClick={() => void onExport()}>
            导出 CSV
          </Button>
          <Button type="primary" size="small" icon={<PlusOutlined />} onClick={() => setEntryModal(true)}>
            记一笔
          </Button>
        </Space>
      </div>

      <div className="acb-surface-panel">
      <Tabs
        items={[
          {
            key: 'entries',
            label: '工时列表',
            children: (
              <Table
                className="acb-list-table"
                size="small"
                rowKey="id"
                loading={loading}
                dataSource={entries}
                pagination={{ pageSize: 20 }}
                columns={[
                  {
                    title: 'Issue',
                    key: 'issue',
                    ellipsis: true,
                    render: (_, r) => resolveIssue(r),
                  },
                  { title: '小时', dataIndex: 'hours', width: 72 },
                  {
                    title: '可计费',
                    dataIndex: 'billable',
                    width: 80,
                    render: (v: boolean) => (v ? '是' : '否'),
                  },
                  { title: '备注', dataIndex: 'note', ellipsis: true },
                  {
                    title: '工作日',
                    dataIndex: 'worked_at',
                    width: 120,
                    render: (v: string) => formatShanghai(v, 'YYYY-MM-DD'),
                  },
                  {
                    title: '成员',
                    key: 'member',
                    width: 120,
                    ellipsis: true,
                    render: (_, r) => resolveMember(r.member_id, r.member_name),
                  },
                  {
                    title: '操作',
                    width: 72,
                    render: (_, r) => (
                      <Button
                        type="link"
                        danger
                        size="small"
                        onClick={() => {
                          Modal.confirm({
                            title: '删除这条工时？',
                            onOk: async () => {
                              await deleteTimeEntry(wsId, r.id)
                              message.success('已删除')
                              await loadEntries()
                              await loadSummary()
                            },
                          })
                        }}
                      >
                        删除
                      </Button>
                    ),
                  },
                ]}
              />
            ),
          },
          {
            key: 'summary',
            label: '汇总',
            children: (
              <Space direction="vertical" style={{ width: '100%' }} size="middle">
                <Space>
                  <Typography.Text>分组</Typography.Text>
                  <Select
                    style={{ width: 140 }}
                    value={groupBy}
                    onChange={setGroupBy}
                    options={[
                      { value: 'project', label: '项目' },
                      { value: 'member', label: '成员' },
                      { value: 'date', label: '日期' },
                    ]}
                  />
                </Space>
                {summary && (
                  <Typography.Paragraph type="secondary" style={{ margin: 0 }}>
                    合计 {summary.total_hours} h · {summary.total_amount} {summary.currency || ''}
                    {summary.from || summary.to
                      ? ` · ${summary.from || '…'} ~ ${summary.to || '…'}`
                      : ''}
                  </Typography.Paragraph>
                )}
                <Table
                  className="acb-list-table"
                  size="small"
                  rowKey="group_key"
                  dataSource={summary?.rows || []}
                  pagination={false}
                  columns={[
                    {
                      title: '分组',
                      dataIndex: 'group_label',
                      render: (v: string, r) => {
                        if (groupBy === 'member') return resolveMember(r.group_key, v)
                        if (/^[0-9a-f-]{16,}$/i.test((v || r.group_key || ''))) {
                          const iss = issueById.get(r.group_key)
                          return issueReadable(iss?.identifier, iss?.title || v, r.group_key)
                        }
                        return v || '—'
                      },
                    },
                    { title: '小时', dataIndex: 'hours' },
                    { title: '可计费小时', dataIndex: 'billable_hours' },
                    { title: '金额', dataIndex: 'amount' },
                    { title: '币种', dataIndex: 'currency' },
                    { title: '条目数', dataIndex: 'entries' },
                  ]}
                />
              </Space>
            ),
          },
          {
            key: 'rates',
            label: '费率',
            children: (
              <Space direction="vertical" style={{ width: '100%' }} size="middle">
                <Button type="primary" size="small" icon={<PlusOutlined />} onClick={() => setRateModal(true)}>
                  新增 / 更新费率
                </Button>
                <Table
                  className="acb-list-table"
                  size="small"
                  rowKey="id"
                  dataSource={rates}
                  pagination={false}
                  columns={[
                    {
                      title: '成员',
                      dataIndex: 'member_id',
                      render: (v?: string | null) => resolveMember(v),
                    },
                    { title: '时薪', dataIndex: 'hourly_rate' },
                    { title: '币种', dataIndex: 'currency' },
                    {
                      title: '更新',
                      dataIndex: 'updated_at',
                      render: (v?: string) => formatShanghai(v),
                    },
                  ]}
                />
              </Space>
            ),
          },
        ]}
      />
      </div>

      <Modal
        title="记一笔工时"
        open={entryModal}
        onCancel={() => setEntryModal(false)}
        onOk={() => void submitEntry()}
        destroyOnHidden
      >
        <Form
          form={entryForm}
          layout="vertical"
          initialValues={{ billable: true, hours: 1, worked_at: dayjs().tz(TZ) }}
        >
          <Form.Item name="issue_id" label="Issue" rules={[{ required: true, message: '请选择 Issue' }]}>
            <Select
              showSearch
              optionFilterProp="label"
              placeholder="选择 Issue"
              options={issueOptions}
            />
          </Form.Item>
          <Form.Item name="member_id" label="成员">
            <Select
              allowClear
              showSearch
              optionFilterProp="label"
              placeholder="默认当前成员"
              options={memberOptions}
            />
          </Form.Item>
          <Form.Item name="hours" label="小时" rules={[{ required: true }]}>
            <InputNumber min={0.1} step={0.25} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="worked_at" label="工作日">
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="note" label="备注">
            <Input.TextArea rows={2} />
          </Form.Item>
          <Form.Item name="billable" label="可计费" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="设置费率"
        open={rateModal}
        onCancel={() => setRateModal(false)}
        onOk={() => void submitRate()}
        destroyOnHidden
      >
        <Form form={rateForm} layout="vertical" initialValues={{ currency: 'CNY', hourly_rate: 500 }}>
          <Form.Item name="member_id" label="成员（空=工作区默认）">
            <Select
              allowClear
              showSearch
              optionFilterProp="label"
              placeholder="工作区默认"
              options={memberOptions}
            />
          </Form.Item>
          <Form.Item name="hourly_rate" label="时薪" rules={[{ required: true }]}>
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="currency" label="币种">
            <Input />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
