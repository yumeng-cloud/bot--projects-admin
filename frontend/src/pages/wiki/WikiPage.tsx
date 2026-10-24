import { useCallback, useEffect, useMemo, useRef, useState, type ReactNode } from 'react'
import {
  Button,
  Dropdown,
  Empty,
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
import type { DataNode, TreeProps } from 'antd/es/tree'
import type { MenuProps } from 'antd'
import {
  BookOutlined,
  DeleteOutlined,
  DownOutlined,
  EditOutlined,
  FileTextOutlined,
  MoreOutlined,
  PlusOutlined,
  ReloadOutlined,
  SearchOutlined,
  SettingOutlined,
} from '@ant-design/icons'
import { useParams } from 'react-router-dom'
import {
  createWikiPage,
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
import { WikiBodyEditor, WikiBodyViewer, WikiPropChips, WikiSpaceAdmin } from '../../components/wiki'
import type { ChipDef } from '../../components/wiki'
import './wiki.css'

const { Sider, Content } = Layout

/**
 * Wiki P2 知识库工作台壳（对标飞书文档 / 语雀 / Notion 中文）：
 * 顶条切换知识库 + 管理入口；目录像书；舞台大标题→属性芯片→正文；管理与写作分面。
 * 验收：WIKI_CN_REBUILD_P0 C1–C8 · WIKI_IA_REDESIGN_P2 R1–R5 · WIKI_VISUAL_SPEC_P2。
 */

function isRenderableTreeNode(n: WikiTreeNode | null | undefined): n is WikiTreeNode {
  return !!n?.id
}

function toTreeData(
  nodes: WikiTreeNode[],
  renderTitle: (n: WikiTreeNode) => ReactNode,
): DataNode[] {
  return (nodes || [])
    .filter(isRenderableTreeNode)
    .map((n) => {
      const kids = toTreeData(n.children || [], renderTitle)
      return {
        key: n.id,
        title: renderTitle(n),
        children: kids.length ? kids : undefined,
      }
    })
}

function firstTreePageId(nodes: WikiTreeNode[]): string | undefined {
  for (const n of nodes || []) {
    if (isRenderableTreeNode(n)) return n.id
    const child = firstTreePageId((n as WikiTreeNode | undefined)?.children || [])
    if (child) return child
  }
  return undefined
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
  const [adminOpen, setAdminOpen] = useState(false)
  const [editing, setEditing] = useState(false)
  const [draftTitle, setDraftTitle] = useState('')
  const [draftBody, setDraftBody] = useState('')
  const [projects, setProjects] = useState<Project[]>([])
  const [issues, setIssues] = useState<Issue[]>([])
  const saveTimer = useRef<number | null>(null)
  const pageRef = useRef<WikiPageT | null>(null)
  pageRef.current = page
  const autoOpenSpaceRef = useRef<string | null>(null)

  const loadSpaces = useCallback(async () => {
    if (!wsId) return
    setLoading(true)
    try {
      const res = await listWikiSpaces(wsId)
      const list = res.spaces || []
      setSpaces(list)
      setSpaceId((prev) => {
        if (prev && list.some((s) => s.id === prev)) return prev
        return list[0]?.id
      })
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : '加载知识库失败')
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
      autoOpenSpaceRef.current = null
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

  useEffect(() => {
    if (!spaceId || page || pageLoading || loading) return
    if (autoOpenSpaceRef.current === spaceId) return
    const firstId = firstTreePageId(tree) || pages[0]?.id
    if (!firstId) return
    autoOpenSpaceRef.current = spaceId
    void openPage(firstId)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [spaceId, tree, pages, page, pageLoading, loading])

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

  /** C2: 点新建即草稿，无三字段 Modal */
  const createDraftPage = async (parentId?: string | null) => {
    if (!spaceId || !wsId) {
      message.warning('请先选择或新建一个知识库')
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

  const patchPageLinks = async (patch: {
    project_id?: string | null
    issue_id?: string | null
    parent_id?: string | null
  }) => {
    const cur = pageRef.current
    if (!cur || !wsId) return
    const prev = cur
    const optimistic: WikiPageT = {
      ...cur,
      ...(patch.project_id !== undefined ? { project_id: patch.project_id } : {}),
      ...(patch.issue_id !== undefined ? { issue_id: patch.issue_id } : {}),
      ...(patch.parent_id !== undefined ? { parent_id: patch.parent_id } : {}),
    }
    setPage(optimistic)
    try {
      const res = await updateWikiPage(wsId, cur.id, patch)
      setPage(res.page)
      if (spaceId && patch.parent_id !== undefined) void loadTreeAndPages(spaceId)
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

  const onParentLinkChange = (value: string | null) => {
    void patchPageLinks({ parent_id: value || null })
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

  const parentOptions = useMemo(() => {
    if (!page) return []
    return pages
      .filter((p) => p.id !== page.id)
      .map((p) => ({ label: p.title || '无标题', value: p.id }))
  }, [pages, page])

  const projectDisplay = useMemo(() => {
    if (!page?.project_id) return ''
    return projects.find((p) => p.id === page.project_id)?.name || '已关联项目'
  }, [page?.project_id, projects])

  const issueDisplay = useMemo(() => {
    if (!page?.issue_id) return ''
    const hit = issues.find((i) => i.id === page.issue_id)
    return hit ? issueSelectLabel(hit) : '已关联 Issue'
  }, [page?.issue_id, issues])

  const parentDisplay = useMemo(() => {
    if (!page?.parent_id) return ''
    return pages.find((p) => p.id === page.parent_id)?.title || '已设父级'
  }, [page?.parent_id, pages])

  const propChips: ChipDef[] = useMemo(() => {
    if (!page) return []
    return [
      {
        kind: 'project',
        label: '项目',
        value: page.project_id,
        display: projectDisplay,
        emptyText: '未关联',
        options: projectOptions,
        onChange: onProjectLinkChange,
      },
      {
        kind: 'issue',
        label: 'Issue',
        value: page.issue_id,
        display: issueDisplay,
        emptyText: '未关联',
        options: issueOptions,
        onChange: onIssueLinkChange,
      },
      {
        kind: 'parent',
        label: '父级',
        value: page.parent_id,
        display: parentDisplay,
        emptyText: '无',
        options: parentOptions,
        onChange: onParentLinkChange,
      },
    ]
  }, [
    page,
    projectDisplay,
    issueDisplay,
    parentDisplay,
    projectOptions,
    issueOptions,
    parentOptions,
  ])

  const spaceName = spaces.find((s) => s.id === spaceId)?.name || '知识库'
  const currentSpace = spaces.find((s) => s.id === spaceId)

  const pageMoreItems: MenuProps['items'] = [
    ...(editing
      ? [
          {
            key: 'exit',
            label: '退出编辑',
            onClick: () => {
              void finishEdit()
            },
          } as const,
        ]
      : [
          {
            key: 'edit',
            label: '编辑',
            icon: <EditOutlined />,
            onClick: () => {
              if (!page) return
              setDraftTitle(page.title || '')
              setDraftBody(page.body_md || '')
              setEditing(true)
            },
          } as const,
        ]),
    {
      key: 'del',
      label: '删除',
      icon: <DeleteOutlined />,
      danger: true,
      onClick: () => removePage(),
    },
  ]

  const dirty =
    !!page &&
    editing &&
    ((draftTitle.trim() || '无标题') !== (page.title || '') || draftBody !== (page.body_md || ''))

  const saveHint = saving ? '保存中…' : dirty ? '编辑中 · ⌘S 保存' : editing ? '已保存 · ⌘S' : ''

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
      <span className="acb-wiki-tree-node">
        <FileTextOutlined className="acb-wiki-tree-node-icon" aria-hidden />
        <span className="acb-wiki-tree-node-title">{n.title?.trim() || '无标题'}</span>
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
      pages
        .filter((p) => !!p?.id)
        .map((p) => ({
          key: p.id,
          title: renderTreeTitle({
            id: p.id,
            title: p.title,
            slug: p.slug,
            parent_id: p.parent_id,
          }),
        })),
    [pages],
  )

  const onSelect: TreeProps['onSelect'] = (keys) => {
    const id = keys[0]
    if (typeof id === 'string') void openPage(id)
  }

  const spaceSwitchItems: MenuProps['items'] = spaces.map((s) => ({
    key: s.id,
    label: s.name,
    icon: <BookOutlined />,
    onClick: () => setSpaceId(s.id),
  }))

  if (loading) {
    return (
      <div style={{ padding: 48, textAlign: 'center' }}>
        <Spin />
      </div>
    )
  }

  /* C6: 无任何知识库 */
  if (!spaces.length) {
    return (
      <div className="acb-fill-page acb-wiki-shell acb-wiki-shell--empty">
        <div className="acb-wiki-topbar">
          <div className="acb-wiki-topbar-left">
            <span className="acb-wiki-topbar-space">知识库</span>
          </div>
          <div className="acb-wiki-topbar-right">
            <Button
              type="text"
              size="small"
              className="acb-wiki-topbar-admin acb-wiki-topbar-action"
              icon={<SettingOutlined />}
              onClick={() => setAdminOpen(true)}
            >
              管理
            </Button>
          </div>
        </div>
        <div className="acb-wiki-empty-workspace">
          <p className="acb-wiki-empty-copy">还没有知识库</p>
          <p className="acb-wiki-empty-hint">新建一个知识库，开始写团队手册</p>
          <Button
            type="primary"
            className="acb-wiki-empty-cta"
            icon={<PlusOutlined />}
            onClick={() => setAdminOpen(true)}
          >
            新建知识库
          </Button>
        </div>
        <WikiSpaceAdmin
          open={adminOpen}
          onClose={() => setAdminOpen(false)}
          workspaceId={wsId}
          spaces={spaces}
          currentSpaceId={spaceId}
          onSelectSpace={setSpaceId}
          onSpacesChanged={loadSpaces}
        />
      </div>
    )
  }

  return (
    <div className="acb-fill-page acb-wiki-shell">
      {/* C1: 知识库顶条 — 当前库名▾ + 管理 */}
      <header className="acb-wiki-topbar">
        <div className="acb-wiki-topbar-left">
          <Dropdown
            menu={{
              items: [
                ...spaceSwitchItems,
                { type: 'divider' as const },
                {
                  key: '__manage',
                  label: '管理知识库…',
                  icon: <SettingOutlined />,
                  onClick: () => setAdminOpen(true),
                },
              ],
              selectedKeys: spaceId ? [spaceId] : [],
            }}
            trigger={['click']}
          >
            <Button type="text" className="acb-wiki-topbar-space">
              <BookOutlined className="acb-wiki-topbar-space-icon" aria-hidden />
              <span className="acb-wiki-topbar-space-name">{spaceName}</span>
              <DownOutlined className="acb-wiki-topbar-space-caret" aria-hidden />
            </Button>
          </Dropdown>
        </div>
        <div className="acb-wiki-topbar-right acb-wiki-topbar-actions">
          <Button
            type="text"
            size="small"
            className="acb-wiki-topbar-admin acb-wiki-topbar-action"
            icon={<SettingOutlined />}
            onClick={() => setAdminOpen(true)}
          >
            管理
          </Button>
        </div>
      </header>

      <Layout
        className="acb-wiki-body-row"
        style={{
          background: themeToken.colorBgLayout,
          flex: '1 1 0',
          minHeight: 0,
          height: '100%',
          alignSelf: 'stretch',
          overflow: 'hidden',
        }}
      >
        {/* C3: 目录栏 ≥300 */}
        <Sider
          width={300}
          theme="light"
          className="acb-wiki-sider acb-wiki-dir"
          style={{
            borderRight: `1px solid ${themeToken.colorBorderSecondary}`,
            background: themeToken.colorFillSecondary,
            padding: 14,
            overflow: 'auto',
            alignSelf: 'stretch',
            height: '100%',
            minHeight: 0,
          }}
        >
          <Space direction="vertical" style={{ width: '100%' }} size="middle">
            <Space style={{ width: '100%', justifyContent: 'space-between' }}>
              <Typography.Text strong>目录</Typography.Text>
              <Button
                size="small"
                type="text"
                icon={<ReloadOutlined />}
                aria-label="刷新"
                onClick={() => spaceId && void loadTreeAndPages(spaceId)}
              />
            </Space>

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
                locale={{ emptyText: '无匹配页面' }}
                renderItem={(p) => (
                  <List.Item style={{ cursor: 'pointer' }} onClick={() => void openPage(p.id)}>
                    {p.title || '无标题'}
                  </List.Item>
                )}
              />
            ) : (
              <>
                <Space style={{ width: '100%', justifyContent: 'space-between' }}>
                  <Typography.Text type="secondary">页面</Typography.Text>
                  <Button
                    size="small"
                    type="link"
                    icon={<PlusOutlined />}
                    disabled={!spaceId || creating}
                    loading={creating}
                    onClick={() => void createDraftPage()}
                  >
                    新建页
                  </Button>
                </Space>
                {treeData.length || flatFallback.length ? (
                  <Tree
                    className="acb-wiki-tree"
                    showIcon={false}
                    blockNode
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

        {/* C4: 页面舞台 */}
        <Content className="acb-wiki-main acb-wiki-stage">
          <div className="acb-wiki-main-inner acb-wiki-stage-rail">
            {pageLoading || (!page && (tree.length > 0 || pages.length > 0)) ? (
              <Spin />
            ) : !page ? (
              <div className="acb-wiki-empty">
                <p className="acb-wiki-empty-copy">还没有页面</p>
                <p className="acb-wiki-empty-hint">新建一页开始写团队手册</p>
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
                <div className="acb-wiki-breadcrumb">
                  <Typography.Text type="secondary">{spaceName}</Typography.Text>
                </div>
                <div className="acb-wiki-main-toolbar">
                  <Typography.Text className="acb-wiki-save-hint" type="secondary">
                    {saveHint}
                  </Typography.Text>
                  <Dropdown menu={{ items: pageMoreItems }} trigger={['click']}>
                    <Button
                      type="text"
                      size="small"
                      className="acb-wiki-more-btn"
                      icon={<MoreOutlined />}
                      aria-label="更多"
                    />
                  </Dropdown>
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
                <WikiPropChips chips={propChips} />
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
                <div className="acb-wiki-breadcrumb">
                  <Typography.Text type="secondary">{spaceName}</Typography.Text>
                  {currentSpace?.description ? (
                    <>
                      <Typography.Text type="secondary"> · </Typography.Text>
                      <Typography.Text type="secondary" ellipsis>
                        {page.title || '无标题'}
                      </Typography.Text>
                    </>
                  ) : null}
                </div>
                <div className="acb-wiki-main-toolbar">
                  <span />
                  <Space size={4}>
                    <Button
                      type="text"
                      size="small"
                      icon={<EditOutlined />}
                      onClick={() => {
                        setDraftTitle(page.title || '')
                        setDraftBody(page.body_md || '')
                        setEditing(true)
                      }}
                    >
                      编辑
                    </Button>
                    <Dropdown menu={{ items: pageMoreItems }} trigger={['click']}>
                      <Button
                        type="text"
                        size="small"
                        className="acb-wiki-more-btn"
                        icon={<MoreOutlined />}
                        aria-label="更多"
                      />
                    </Dropdown>
                  </Space>
                </div>
                <Typography.Title level={2} className="acb-wiki-title" style={{ margin: 0 }}>
                  {page.title || '无标题'}
                </Typography.Title>
                <Typography.Paragraph className="acb-wiki-meta">
                  更新于 {formatWikiTime(page.updated_at)}
                </Typography.Paragraph>
                <WikiPropChips chips={propChips} />
                <WikiBodyViewer value={page.body_md || ''} />
              </>
            )}
          </div>
        </Content>
      </Layout>

      {/* C8: 管理面 — 建库/改名只在此 */}
      <WikiSpaceAdmin
        open={adminOpen}
        onClose={() => setAdminOpen(false)}
        workspaceId={wsId}
        spaces={spaces}
        pageCountBySpace={spaceId ? { [spaceId]: pages.length } : undefined}
        currentSpaceId={spaceId}
        onSelectSpace={setSpaceId}
        onSpacesChanged={loadSpaces}
      />
    </div>
  )
}
