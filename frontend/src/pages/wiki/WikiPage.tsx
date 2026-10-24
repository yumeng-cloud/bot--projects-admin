import { useCallback, useEffect, useMemo, useRef, useState, type ReactNode } from 'react'
import {
  Button,
  Dropdown,
  Empty,
  Form,
  Input,
  Layout,
  List,
  Modal,
  Select,
  Space,
  Spin,
  Tree,
  Typography,
  message,
  theme,
} from 'antd'
import type { DataNode, TreeProps } from 'antd/es/tree'
import type { MenuProps } from 'antd'
import {
  DeleteOutlined,
  EditOutlined,
  FileTextOutlined,
  MoreOutlined,
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
import { listProjects, type Project } from '../../api/projects'
import { listIssues } from '../../api/issues'
import type { Issue } from '../../types/issue'
import type { WikiPage as WikiPageT, WikiSpace, WikiTreeNode } from '../../types/wiki'
import { formatWikiTime } from '../../utils/time'
import { WikiBodyEditor, WikiBodyViewer } from '../../components/wiki'
import './wiki.css'

const { Sider, Content } = Layout

function toTreeData(
  nodes: WikiTreeNode[],
  renderTitle: (n: WikiTreeNode) => ReactNode,
): DataNode[] {
  return (nodes || []).map((n) => ({
    key: n.id,
    title: renderTitle(n),
    icon: <FileTextOutlined />,
    children: n.children?.length ? toTreeData(n.children, renderTitle) : undefined,
  }))
}


function issueSelectLabel(i: Pick<Issue, 'identifier' | 'title' | 'id'>): string {
  const idPart = (i.identifier || '').trim()
  const titlePart = (i.title || '').trim()
  if (idPart && titlePart) return `${idPart} ${titlePart}`
  if (idPart) return idPart
  if (titlePart) return titlePart
  return '未命名 Issue'
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
  const [creating, setCreating] = useState(false)
  const [saving, setSaving] = useState(false)
  const [q, setQ] = useState('')
  const [searchHits, setSearchHits] = useState<WikiPageT[] | null>(null)
  const [spaceModal, setSpaceModal] = useState(false)
  const [editing, setEditing] = useState(false)
  const [draftTitle, setDraftTitle] = useState('')
  const [draftBody, setDraftBody] = useState('')
  const [projects, setProjects] = useState<Project[]>([])
  const [issues, setIssues] = useState<Issue[]>([])
  const [spaceForm] = Form.useForm()
  const saveTimer = useRef<number | null>(null)
  const pageRef = useRef<WikiPageT | null>(null)
  pageRef.current = page

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

  const loadTreeAndPages = useCallback(
    async (sid: string) => {
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
    },
    [wsId],
  )

  useEffect(() => {
    void loadSpaces()
  }, [loadSpaces])

  useEffect(() => {
    if (!wsId) return
    let cancelled = false
    void listProjects(wsId)
      .then((res) => {
        if (!cancelled) setProjects(res.projects || [])
      })
      .catch(() => {
        if (!cancelled) setProjects([])
      })
    return () => {
      cancelled = true
    }
  }, [wsId])

  const selectedProjectId = page?.project_id || undefined

  useEffect(() => {
    if (!wsId) return
    let cancelled = false
    void listIssues(wsId, selectedProjectId ? { project_id: selectedProjectId } : {})
      .then((res) => {
        if (!cancelled) setIssues(res.issues || [])
      })
      .catch(() => {
        if (!cancelled) setIssues([])
      })
    return () => {
      cancelled = true
    }
  }, [wsId, selectedProjectId])

  useEffect(() => {
    if (spaceId) {
      setPage(null)
      setEditing(false)
      setSearchHits(null)
      void loadTreeAndPages(spaceId)
    }
  }, [spaceId, loadTreeAndPages])

  const openPage = async (pageId: string, opts?: { edit?: boolean }) => {
    if (!wsId) return
    setPageLoading(true)
    try {
      const res = await getWikiPage(wsId, pageId)
      setPage(res.page)
      setDraftTitle(res.page.title || '')
      setDraftBody(res.page.body_md || '')
      setEditing(!!opts?.edit)
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

  const createSpace = async () => {
    const v = await spaceForm.validateFields()
    await createWikiSpace(wsId, v)
    message.success('空间已创建')
    setSpaceModal(false)
    spaceForm.resetFields()
    await loadSpaces()
  }

  /** W1: 点新建即草稿，无三字段 Modal */
  const createDraftPage = async (parentId?: string | null) => {
    if (!spaceId || !wsId) {
      message.warning('请先选择或创建一个 Wiki 空间')
      return
    }
    setCreating(true)
    try {
      const res = await createWikiPage(wsId, spaceId, {
        title: '无标题',
        body_md: '',
        ...(parentId ? { parent_id: parentId } : {}),
      })
      await loadTreeAndPages(spaceId)
      setPage(res.page)
      setDraftTitle(res.page.title || '无标题')
      setDraftBody(res.page.body_md || '')
      setEditing(true)
      message.success('已创建草稿')
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : '创建失败')
    } finally {
      setCreating(false)
    }
  }

  const flushSave = useCallback(async () => {
    const cur = pageRef.current
    if (!cur || !wsId) return
    const title = draftTitle.trim() || '无标题'
    const body = draftBody
    if (title === (cur.title || '') && body === (cur.body_md || '')) return
    setSaving(true)
    try {
      const res = await updateWikiPage(wsId, cur.id, {
        title,
        body_md: body,
      })
      setPage(res.page)
      setDraftTitle(res.page.title || '')
      setDraftBody(res.page.body_md || '')
      if (spaceId) void loadTreeAndPages(spaceId)
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : '保存失败')
    } finally {
      setSaving(false)
    }
  }, [draftTitle, draftBody, wsId, spaceId, loadTreeAndPages])

  const scheduleSave = useCallback(() => {
    if (saveTimer.current) window.clearTimeout(saveTimer.current)
    saveTimer.current = window.setTimeout(() => {
      void flushSave()
    }, 600)
  }, [flushSave])

  useEffect(() => {
    return () => {
      if (saveTimer.current) window.clearTimeout(saveTimer.current)
    }
  }, [])

  useEffect(() => {
    if (!editing) return
    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 's') {
        e.preventDefault()
        if (saveTimer.current) window.clearTimeout(saveTimer.current)
        void flushSave()
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [editing, flushSave])

  const finishEdit = async () => {
    if (saveTimer.current) window.clearTimeout(saveTimer.current)
    await flushSave()
    setEditing(false)
  }

  const patchPageLinks = async (patch: { project_id?: string | null; issue_id?: string | null }) => {
    const cur = pageRef.current
    if (!cur || !wsId) return
    const prev = cur
    const optimistic: WikiPageT = {
      ...cur,
      ...(patch.project_id !== undefined ? { project_id: patch.project_id } : {}),
      ...(patch.issue_id !== undefined ? { issue_id: patch.issue_id } : {}),
    }
    setPage(optimistic)
    try {
      const res = await updateWikiPage(wsId, cur.id, patch)
      setPage(res.page)
    } catch (e: unknown) {
      setPage(prev)
      message.error(e instanceof Error ? e.message : '关联更新失败')
    }
  }

  const onProjectLinkChange = (value: string | null) => {
    const project_id = value || null
    const clearIssue =
      !project_id || (pageRef.current?.issue_id != null && pageRef.current.project_id !== project_id)
    void patchPageLinks(clearIssue ? { project_id, issue_id: null } : { project_id })
  }

  const onIssueLinkChange = (value: string | null) => {
    void patchPageLinks({ issue_id: value || null })
  }

  const projectOptions = useMemo(
    () => projects.map((p) => ({ label: p.name, value: p.id })),
    [projects],
  )

  const issueOptions = useMemo(() => {
    const opts = issues.map((i) => ({ label: issueSelectLabel(i), value: i.id }))
    const curId = page?.issue_id
    if (curId && !opts.some((o) => o.value === curId)) {
      opts.unshift({ label: '已关联 Issue', value: curId })
    }
    return opts
  }, [issues, page?.issue_id])

  const renderPageProps = () => (
    <div className="acb-wiki-props">
      <span className="acb-wiki-props-item">
        <span className="acb-wiki-props-label">关联项目</span>
        <Select
          allowClear
          showSearch
          optionFilterProp="label"
          placeholder="选择项目"
          style={{ minWidth: 160, maxWidth: 240 }}
          size="small"
          value={page?.project_id || undefined}
          options={projectOptions}
          onChange={(v) => onProjectLinkChange((v as string) ?? null)}
        />
      </span>
      <span className="acb-wiki-props-item">
        <span className="acb-wiki-props-label">关联 Issue</span>
        <Select
          allowClear
          showSearch
          optionFilterProp="label"
          placeholder="选择 Issue"
          style={{ minWidth: 200, maxWidth: 320 }}
          size="small"
          value={page?.issue_id || undefined}
          options={issueOptions}
          onChange={(v) => onIssueLinkChange((v as string) ?? null)}
        />
      </span>
    </div>
  )

  const removePage = (target?: WikiPageT | null) => {
    const p = target || page
    if (!p) return
    Modal.confirm({
      title: '删除页面？',
      content: p.title,
      okType: 'danger',
      onOk: async () => {
        await deleteWikiPage(wsId, p.id)
        message.success('已删除')
        if (page?.id === p.id) {
          setPage(null)
          setEditing(false)
        }
        if (spaceId) await loadTreeAndPages(spaceId)
      },
    })
  }

  const renderTreeTitle = (n: WikiTreeNode) => {
    const items: MenuProps['items'] = [
      {
        key: 'child',
        label: '新建子页',
        icon: <PlusOutlined />,
        onClick: ({ domEvent }) => {
          domEvent.stopPropagation()
          void createDraftPage(n.id)
        },
      },
      {
        key: 'del',
        label: '删除',
        icon: <DeleteOutlined />,
        danger: true,
        onClick: ({ domEvent }) => {
          domEvent.stopPropagation()
          const found = pages.find((x) => x.id === n.id)
          if (found) removePage(found)
          else {
            Modal.confirm({
              title: '删除页面？',
              content: n.title,
              okType: 'danger',
              onOk: async () => {
                await deleteWikiPage(wsId, n.id)
                message.success('已删除')
                if (page?.id === n.id) {
                  setPage(null)
                  setEditing(false)
                }
                if (spaceId) await loadTreeAndPages(spaceId)
              },
            })
          }
        },
      },
    ]
    return (
      <span className="acb-wiki-tree-node" style={{ display: 'flex', width: '100%', gap: 4 }}>
        <span className="acb-wiki-tree-node-title" style={{ flex: 1, minWidth: 0 }}>
          {n.title}
        </span>
        <Dropdown menu={{ items }} trigger={['click']}>
          <Button
            type="text"
            size="small"
            className="acb-wiki-tree-node-actions"
            icon={<MoreOutlined />}
            onClick={(e) => e.stopPropagation()}
          />
        </Dropdown>
      </span>
    )
  }

  const treeData = useMemo(() => toTreeData(tree, renderTreeTitle), [tree, pages, page, spaceId])
  const flatFallback = useMemo(
    () =>
      pages.map((p) => ({
        key: p.id,
        title: renderTreeTitle({
          id: p.id,
          title: p.title,
          slug: p.slug,
          parent_id: p.parent_id,
        }),
        icon: <FileTextOutlined />,
      })),
    [pages],
  )

  const onSelect: TreeProps['onSelect'] = (keys) => {
    const id = keys[0]
    if (typeof id === 'string') void openPage(id)
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
      className="acb-fill-page acb-wiki-shell"
      style={{ background: themeToken.colorBgContainer, height: '100%', overflow: 'hidden' }}
    >
      <Sider
        width={280}
        theme="light"
        className="acb-wiki-sider"
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
              <Button size="small" icon={<ReloadOutlined />} onClick={() => void loadSpaces()} />
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
            onSearch={() => void onSearch()}
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
                <List.Item style={{ cursor: 'pointer' }} onClick={() => void openPage(p.id)}>
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
                  disabled={!spaceId || creating}
                  loading={creating}
                  onClick={() => void createDraftPage(page?.id)}
                >
                  新建页
                </Button>
              </Space>
              {treeData.length || flatFallback.length ? (
                <Tree
                  className="acb-wiki-tree"
                  showIcon
                  treeData={treeData.length ? treeData : flatFallback}
                  selectedKeys={page ? [page.id] : []}
                  defaultExpandAll
                  onSelect={onSelect}
                />
              ) : (
                <div className="acb-wiki-sider-empty">
                  <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无页面" />
                  <Button
                    type="primary"
                    size="small"
                    className="acb-wiki-empty-cta"
                    icon={<PlusOutlined />}
                    disabled={!spaceId || creating}
                    loading={creating}
                    onClick={() => void createDraftPage()}
                  >
                    新建第一页
                  </Button>
                </div>
              )}
            </>
          )}
        </Space>
      </Sider>

      <Content className="acb-wiki-main">
        <div className="acb-wiki-main-inner">
          {pageLoading ? (
            <Spin />
          ) : !page ? (
            <div className="acb-wiki-empty">
              <p className="acb-wiki-empty-copy">还没有打开页面</p>
              <p className="acb-wiki-empty-hint">从左侧目录选择，或新建一页开始写团队手册</p>
              <Button
                type="primary"
                className="acb-wiki-empty-cta"
                icon={<PlusOutlined />}
                disabled={!spaceId || creating}
                loading={creating}
                onClick={() => void createDraftPage()}
              >
                新建第一页
              </Button>
            </div>
          ) : editing ? (
            <>
              <div className="acb-wiki-main-toolbar">
                <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                  {saving ? '保存中…' : '编辑中 · ⌘S 保存'}
                </Typography.Text>
                <Space>
                  <Button type="primary" onClick={() => void finishEdit()}>
                    完成
                  </Button>
                  <Button danger icon={<DeleteOutlined />} onClick={() => removePage()}>
                    删除
                  </Button>
                </Space>
              </div>
              <Input
                className="acb-wiki-title-input"
                value={draftTitle}
                placeholder="无标题"
                onChange={(e) => {
                  setDraftTitle(e.target.value)
                  scheduleSave()
                }}
                onBlur={() => void flushSave()}
              />
              <Typography.Paragraph className="acb-wiki-meta">
                更新于 {formatWikiTime(page.updated_at)}
              </Typography.Paragraph>
              {renderPageProps()}
              <WikiBodyEditor
                value={draftBody}
                onChange={(md) => {
                  setDraftBody(md)
                  scheduleSave()
                }}
                onBlur={() => void flushSave()}
              />
            </>
          ) : (
            <>
              <div className="acb-wiki-main-toolbar">
                <span />
                <Space>
                  <Button
                    icon={<EditOutlined />}
                    onClick={() => {
                      setDraftTitle(page.title || '')
                      setDraftBody(page.body_md || '')
                      setEditing(true)
                    }}
                  >
                    编辑
                  </Button>
                  <Button danger icon={<DeleteOutlined />} onClick={() => removePage()}>
                    删除
                  </Button>
                </Space>
              </div>
              <Typography.Title level={2} className="acb-wiki-title" style={{ margin: 0 }}>
                {page.title || '无标题'}
              </Typography.Title>
              <Typography.Paragraph className="acb-wiki-meta">
                更新于 {formatWikiTime(page.updated_at)}
              </Typography.Paragraph>
              {renderPageProps()}
              <WikiBodyViewer value={page.body_md || ''} />
            </>
          )}
        </div>
      </Content>

      <Modal
        title="新建 Wiki 空间"
        open={spaceModal}
        onCancel={() => setSpaceModal(false)}
        onOk={() => void createSpace()}
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
    </Layout>
  )
}
