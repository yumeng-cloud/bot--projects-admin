import { useMemo, useState } from 'react'
import {
  Button,
  Drawer,
  Dropdown,
  Empty,
  Form,
  Input,
  List,
  Modal,
  Space,
  Typography,
  message,
} from 'antd'
import type { MenuProps } from 'antd'
import {
  DeleteOutlined,
  EditOutlined,
  MoreOutlined,
  PlusOutlined,
  SettingOutlined,
} from '@ant-design/icons'
import {
  createWikiSpace,
  deleteWikiSpace,
  updateWikiSpace,
} from '../../api/wiki'
import type { WikiSpace } from '../../types/wiki'
import { formatWikiTime } from '../../utils/time'

export type WikiSpaceAdminProps = {
  open: boolean
  onClose: () => void
  workspaceId: string
  spaces: WikiSpace[]
  /** page count per space id (optional display) */
  pageCountBySpace?: Record<string, number>
  currentSpaceId?: string
  onSelectSpace: (spaceId: string) => void
  onSpacesChanged: () => Promise<void> | void
}

type Panel = 'list' | 'create' | 'edit'

/**
 * 知识库管理面（P2）：右侧抽屉，与写作面分离。
 * 最小集：库列表 / 新建 / 改名描述；成员权限灰置「即将支持」。
 */
export function WikiSpaceAdmin({
  open,
  onClose,
  workspaceId,
  spaces,
  pageCountBySpace,
  currentSpaceId,
  onSelectSpace,
  onSpacesChanged,
}: WikiSpaceAdminProps) {
  const [panel, setPanel] = useState<Panel>('list')
  const [editing, setEditing] = useState<WikiSpace | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [form] = Form.useForm()

  const title = useMemo(() => {
    if (panel === 'create') return '新建知识库'
    if (panel === 'edit') return '知识库设置'
    return '知识库管理'
  }, [panel])

  const resetToList = () => {
    setPanel('list')
    setEditing(null)
    form.resetFields()
  }

  const handleClose = () => {
    resetToList()
    onClose()
  }

  const openCreate = () => {
    form.resetFields()
    setEditing(null)
    setPanel('create')
  }

  const openEdit = (s: WikiSpace) => {
    setEditing(s)
    form.setFieldsValue({
      name: s.name,
      description: s.description || '',
    })
    setPanel('edit')
  }

  const submitCreate = async () => {
    const v = await form.validateFields()
    setSubmitting(true)
    try {
      const res = await createWikiSpace(workspaceId, {
        name: v.name.trim(),
        description: (v.description || '').trim() || undefined,
      })
      message.success('知识库已创建')
      await onSpacesChanged()
      if (res.space?.id) onSelectSpace(res.space.id)
      resetToList()
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : '创建失败')
    } finally {
      setSubmitting(false)
    }
  }

  const submitEdit = async () => {
    if (!editing) return
    const v = await form.validateFields()
    setSubmitting(true)
    try {
      await updateWikiSpace(workspaceId, editing.id, {
        name: v.name.trim(),
        description: (v.description || '').trim(),
      })
      message.success('已保存')
      await onSpacesChanged()
      resetToList()
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : '保存失败')
    } finally {
      setSubmitting(false)
    }
  }

  const confirmDelete = (s: WikiSpace) => {
    Modal.confirm({
      title: '删除知识库？',
      content: `将删除「${s.name}」及其页面，不可恢复。`,
      okType: 'danger',
      okText: '删除',
      cancelText: '取消',
      onOk: async () => {
        await deleteWikiSpace(workspaceId, s.id)
        message.success('已删除')
        await onSpacesChanged()
      },
    })
  }

  const rowMenu = (s: WikiSpace): MenuProps['items'] => [
    {
      key: 'edit',
      label: '设置',
      icon: <EditOutlined />,
      onClick: () => openEdit(s),
    },
    {
      key: 'members',
      label: '成员与权限（即将支持）',
      icon: <SettingOutlined />,
      disabled: true,
    },
    { type: 'divider' },
    {
      key: 'del',
      label: '删除',
      icon: <DeleteOutlined />,
      danger: true,
      onClick: () => confirmDelete(s),
    },
  ]

  return (
    <Drawer
      rootClassName="acb-wiki-admin"
      className="acb-wiki-admin"
      title={title}
      open={open}
      onClose={handleClose}
      width={420}
      destroyOnHidden
      extra={
        panel === 'list' ? (
          <Button type="primary" className="acb-wiki-admin-create" icon={<PlusOutlined />} onClick={openCreate}>
            新建知识库
          </Button>
        ) : (
          <Button type="link" onClick={resetToList}>
            返回列表
          </Button>
        )
      }
    >
      <div className="acb-wiki-admin-body">
        {panel === 'list' ? (
          spaces.length === 0 ? (
            <div className="acb-wiki-admin-empty">
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description="还没有知识库"
              />
              <Button type="primary" className="acb-wiki-admin-create" icon={<PlusOutlined />} onClick={openCreate}>
                新建知识库
              </Button>
            </div>
          ) : (
            <List
              className="acb-wiki-admin-list"
              dataSource={spaces}
              renderItem={(s) => {
                const count = pageCountBySpace?.[s.id]
                const metaParts: string[] = []
                if (typeof count === 'number') metaParts.push(`${count} 页`)
                if (s.updated_at) metaParts.push(`更新于 ${formatWikiTime(s.updated_at)}`)
                return (
                  <List.Item
                    className={
                      s.id === currentSpaceId
                        ? 'acb-wiki-admin-row acb-wiki-admin-row--active'
                        : 'acb-wiki-admin-row'
                    }
                    actions={[
                      <Dropdown key="more" menu={{ items: rowMenu(s) }} trigger={['click']}>
                        <Button
                          type="text"
                          size="small"
                          icon={<MoreOutlined />}
                          aria-label="更多"
                        />
                      </Dropdown>,
                    ]}
                  >
                    <List.Item.Meta
                      title={
                        <Typography.Link
                          onClick={() => {
                            onSelectSpace(s.id)
                            handleClose()
                          }}
                        >
                          {s.name}
                        </Typography.Link>
                      }
                      description={
                        <Space direction="vertical" size={0}>
                          {s.description ? (
                            <Typography.Text type="secondary" ellipsis>
                              {s.description}
                            </Typography.Text>
                          ) : null}
                          {metaParts.length ? (
                            <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                              {metaParts.join(' · ')}
                            </Typography.Text>
                          ) : null}
                        </Space>
                      }
                    />
                  </List.Item>
                )
              }}
            />
          )
        ) : (
          <Form form={form} layout="vertical" requiredMark={false}>
            <Form.Item
              name="name"
              label="名称"
              rules={[{ required: true, message: '请填写知识库名称' }]}
            >
              <Input placeholder="例如：产品手册" maxLength={80} />
            </Form.Item>
            <Form.Item name="description" label="描述">
              <Input.TextArea rows={3} placeholder="可选" maxLength={400} />
            </Form.Item>
            {panel === 'edit' ? (
              <Form.Item label="成员与权限">
                <Button disabled block>
                  即将支持
                </Button>
              </Form.Item>
            ) : null}
            <Space>
              <Button
                type="primary"
                loading={submitting}
                onClick={() => void (panel === 'create' ? submitCreate() : submitEdit())}
              >
                {panel === 'create' ? '创建' : '保存'}
              </Button>
              <Button onClick={resetToList}>取消</Button>
            </Space>
          </Form>
        )}
      </div>
    </Drawer>
  )
}

export default WikiSpaceAdmin
