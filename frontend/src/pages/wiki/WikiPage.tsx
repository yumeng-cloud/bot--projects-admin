import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  Button,
  Empty,
  Form,
  Input,
  Layout,
  List,
  Modal,
  Space,
  Spin,
  Tree,
  Typography,
  message,
  theme,
} from 'antd'
import type { DataNode } from 'antd/es/tree'
import {
  DeleteOutlined,
  EditOutlined,
  FileTextOutlined,
  PlusOutlined,
  ReloadOutlined,
  SearchOutlined,
} from '@ant-design/icons'
import { useParams } from 'react-router-dom'
import {
  createWikiPage,
  createWikiSpace,
  deleteWikiPage,
  getWikiPage,
  listWikiPages,
  listWikiSpaces,
  searchWiki,
  updateWikiPage,
  wikiTree,
} from '../../api/wiki'
import type { WikiPage as WikiPageT, WikiSpace, WikiTreeNode } from '../../types/wiki'

const { Sider, Content } = Layout
const { TextArea } = Input

function toTreeData(nodes: WikiTreeNode[]): DataNode[] {
  return (nodes || []).map((n) => ({
    key: n.id,
    title: n.title,
    icon: <FileTextOutlined />,
    children: n.children?.length ? toTreeData(n.children) : undefined,
  }))
}

export default function WikiPage() {
  const { wsId = '' } = useParams()
  const { token: themeToken } = theme.useToken()
  const [loading, setLoading] = useState(true)
  const [spaces, setSpaces] = useState<WikiSpace[]>([])
  const [spaceId, setSpaceId] = useState<string>()
  const [tree, setTree] = useState<WikiTreeNode[]>([])
  const [pages, setPages] = useState<WikiPageT[]>([])
  const [page, setPage] = useState<WikiPageT | null>(null)
  const [pageLoading, setPageLoading] = useState(false)
  const [q, setQ] = useState('')
  const [searchHits, setSearchHits] = useState<WikiPageT[] | null>(null)
  const [spaceModal, setSpaceModal] = useState(false)
  const [pageModal, setPageModal] = useState(false)
  const [editing, setEditing] = useState(false)
  const [spaceForm] = Form.useForm()
  const [pageForm] = Form.useForm()
  const [editForm] = Form.useForm()

  const loadSpaces = useCallback(async () => {
    if (!wsId) return
    setLoading(true)
    try {
      const res = await listWikiSpaces(wsId)
      const list = res.spaces || []
      setSpaces(list)
      setSpaceId((prev) => prev || list[0]?.id)
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : '加载空间失败')
    } finally {
      setLoading(false)
    }
  }, [wsId])

  const loadTreeAndPages = useCallback(async (sid: string) => {
    if (!wsId || !sid) return
    try {
      const [t, p] = await Promise.all([
        wikiTree(wsId, sid).catch(() => ({ tree: [] as WikiTreeNode[] })),
        listWikiPages(wsId, sid).catch(() => ({ pages: [] as WikiPageT[] })),
      ])
      setTree(t.tree || [])
      setPages(p.pages || [])
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : '加载目录失败')
    }
  }, [wsId])

  useEffect(() => {
    loadSpaces()
  }, [loadSpaces])

  useEffect(() => {
    if (spaceId) {
      setPage(null)
      setSearchHits(null)
      loadTreeAndPages(spaceId)
    }
  }, [spaceId, loadTreeAndPages])

  const openPage = async (pageId: string) => {
    if (!wsId) return
    setPageLoading(true)
    setEditing(false)
    try {
      const res = await getWikiPage(wsId, pageId)
      setPage(res.page)
      editForm.setFieldsValue({ title: res.page.title, body_md: res.page.body_md })
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : '加载页面失败')
    } finally {
      setPageLoading(false)
    }
  }

  const onSearch = async () => {
    if (!wsId || !q.trim()) {
      setSearchHits(null)
      return
    }
    try {
      const res = await searchWiki(wsId, q.trim())
      setSearchHits(res.pages || [])
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : '搜索失败')
    }
  }

  const treeData = useMemo(() => toTreeData(tree), [tree])
  const flatFallback = useMemo(
    () =>
      pages.map((p) => ({
        key: p.id,
        title: p.title,
        icon: <FileTextOutlined />,
      })),
    [pages],
  )

  const createSpace = async () => {
    const v = await spaceForm.validateFields()
    await createWikiSpace(wsId, v)
    message.success('空间已创建')
    setSpaceModal(false)
    spaceForm.resetFields()
    await loadSpaces()
  }

  const createPage = async () => {
    if (!spaceId) return
    const v = await pageForm.validateFields()
    const res = await createWikiPage(wsId, spaceId, {
      title: v.title,
      body_md: v.body_md || '',
      parent_id: v.parent_id || undefined,
    })
    message.success('页面已创建')
    setPageModal(false)
    pageForm.resetFields()
    await loadTreeAndPages(spaceId)
    await openPage(res.page.id)
  }

  const saveEdit = async () => {
    if (!page) return
    const v = await editForm.validateFields()
    const res = await updateWikiPage(wsId, page.id, {
      title: v.title,
      body_md: v.body_md,
    })
    setPage(res.page)
    setEditing(false)
    message.success('已保存')
    if (spaceId) await loadTreeAndPages(spaceId)
  }

  const removePage = () => {
    if (!page) return
    Modal.confirm({
      title: '删除页面？',
      content: page.title,
      okType: 'danger',
      onOk: async () => {
        await deleteWikiPage(wsId, page.id)
        message.success('已删除')
        setPage(null)
        if (spaceId) await loadTreeAndPages(spaceId)
      },
    })
  }

  if (loading) {
    return (
      <div style={{ padding: 48, textAlign: 'center' }}>
        <Spin />
      </div>
    )
  }

  return (
    <Layout
      className="acb-fill-page"
      style={{ background: themeToken.colorBgContainer, height: '100%', overflow: 'hidden' }}
    >
      <Sider
        width={280}
        theme="light"
        style={{
          borderRight: `1px solid ${themeToken.colorBorderSecondary}`,
          padding: 12,
          overflow: 'auto',
          height: '100%',
        }}
      >
        <Space direction="vertical" style={{ width: '100%' }} size="middle">
          <Space style={{ width: '100%', justifyContent: 'space-between' }}>
            <Typography.Text strong>Wiki</Typography.Text>
            <Space size={4}>
              <Button size="small" icon={<ReloadOutlined />} onClick={loadSpaces} />
              <Button size="small" icon={<PlusOutlined />} onClick={() => setSpaceModal(true)} />
            </Space>
          </Space>

          <List
            size="small"
            dataSource={spaces}
            locale={{ emptyText: '暂无空间' }}
            renderItem={(s) => (
              <List.Item
                style={{
                  cursor: 'pointer',
                  background: s.id === spaceId ? themeToken.colorPrimaryBg : undefined,
                  padding: '6px 8px',
                  borderRadius: 6,
                }}
                onClick={() => setSpaceId(s.id)}
              >
                <Typography.Text ellipsis>{s.name}</Typography.Text>
              </List.Item>
            )}
          />

          <Input.Search
            placeholder="搜索页面"
            value={q}
            onChange={(e) => setQ(e.target.value)}
            onSearch={onSearch}
            enterButton={<SearchOutlined />}
            allowClear
            onClear={() => setSearchHits(null)}
          />

          {searchHits ? (
            <List
              size="small"
              header={<Typography.Text type="secondary">搜索结果</Typography.Text>}
              dataSource={searchHits}
              locale={{ emptyText: '无匹配' }}
              renderItem={(p) => (
                <List.Item style={{ cursor: 'pointer' }} onClick={() => openPage(p.id)}>
                  {p.title}
                </List.Item>
              )}
            />
          ) : (
            <>
              <Space style={{ width: '100%', justifyContent: 'space-between' }}>
                <Typography.Text type="secondary">目录</Typography.Text>
                <Button
                  size="small"
                  type="link"
                  icon={<PlusOutlined />}
                  disabled={!spaceId}
                  onClick={() => {
                    pageForm.resetFields()
                    setPageModal(true)
                  }}
                >
                  新建页
                </Button>
              </Space>
              {treeData.length || flatFallback.length ? (
                <Tree
                  showIcon
                  treeData={treeData.length ? treeData : flatFallback}
                  selectedKeys={page ? [page.id] : []}
                  defaultExpandAll
                  onSelect={(keys) => {
                    const id = keys[0]
                    if (typeof id === 'string') openPage(id)
                  }}
                />
              ) : (
                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无页面" />
              )}
            </>
          )}
        </Space>
      </Sider>

      <Content style={{ padding: 20, overflow: 'auto', minWidth: 0, minHeight: 0, flex: 1 }}>
        {pageLoading ? (
          <Spin />
        ) : !page ? (
          <Empty description="选择或新建一个页面" />
        ) : editing ? (
          <Space direction="vertical" style={{ width: '100%' }} size="middle">
            <Space>
              <Button type="primary" onClick={saveEdit}>
                保存
              </Button>
              <Button onClick={() => setEditing(false)}>取消</Button>
            </Space>
            <Form form={editForm} layout="vertical">
              <Form.Item name="title" label="标题" rules={[{ required: true, message: '必填' }]}>
                <Input />
              </Form.Item>
              <Form.Item name="body_md" label="正文 (Markdown)">
                <TextArea rows={20} style={{ fontFamily: 'ui-monospace, monospace' }} />
              </Form.Item>
            </Form>
          </Space>
        ) : (
          <Space direction="vertical" style={{ width: '100%' }} size="middle">
            <Space style={{ width: '100%', justifyContent: 'space-between' }}>
              <Typography.Title level={3} style={{ margin: 0 }}>
                {page.title}
              </Typography.Title>
              <Space>
                <Button icon={<EditOutlined />} onClick={() => setEditing(true)}>
                  编辑
                </Button>
                <Button danger icon={<DeleteOutlined />} onClick={removePage}>
                  删除
                </Button>
              </Space>
            </Space>
            <Typography.Paragraph type="secondary" style={{ margin: 0 }}>
              更新于 {page.updated_at || '—'}
            </Typography.Paragraph>
            <pre
              style={{
                whiteSpace: 'pre-wrap',
                wordBreak: 'break-word',
                fontFamily: 'ui-monospace, SFMono-Regular, Menlo, monospace',
                fontSize: 13,
                lineHeight: 1.6,
                margin: 0,
                padding: 16,
                background: themeToken.colorFillQuaternary,
                borderRadius: 6,
                border: `1px solid ${themeToken.colorBorderSecondary}`,
              }}
            >
              {page.body_md || '（空）'}
            </pre>
          </Space>
        )}
      </Content>

      <Modal
        title="新建 Wiki 空间"
        open={spaceModal}
        onCancel={() => setSpaceModal(false)}
        onOk={createSpace}
        destroyOnHidden
      >
        <Form form={spaceForm} layout="vertical">
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="slug" label="Slug">
            <Input placeholder="可选" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={2} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="新建页面"
        open={pageModal}
        onCancel={() => setPageModal(false)}
        onOk={createPage}
        width={720}
        destroyOnHidden
      >
        <Form form={pageForm} layout="vertical" initialValues={{ body_md: '' }}>
          <Form.Item name="title" label="标题" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="parent_id" label="父页面 ID">
            <Input placeholder="可选" />
          </Form.Item>
          <Form.Item name="body_md" label="正文 (Markdown)">
            <TextArea rows={12} style={{ fontFamily: 'ui-monospace, monospace' }} />
          </Form.Item>
        </Form>
      </Modal>
    </Layout>
  )
}
