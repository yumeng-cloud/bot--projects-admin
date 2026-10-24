import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Spin, Typography, Button, Input, Space, Empty, message } from 'antd'
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons'
import { archiveIssues, createIssue } from '../../api/issues'
import { ApiError } from '../../api/client'
import {
  BoardPackStrip,
  ColumnTemplatePicker,
  KanbanBoard,
} from '../../components/board'
import {
  getBoardColumnPrefs,
  putBoardColumnPrefs,
  resetBoardColumnPrefs,
} from '../../api/boardColumnPrefs'
import {
  defaultPrefs,
  isDefaultLayout,
  layoutFromApiColumns,
  layoutToApiColumns,
  type BoardColumnLayout,
} from '../../components/board/boardTemplates'
import { IssueDrawer } from '../../components/issue'
import { MockBanner } from '../../components/common/MockBanner'
import { useCurrentTeam } from '../../hooks/useCurrentTeam'
import { useBoardPackStats } from '../../hooks/useBoardPackStats'
import type { Issue } from '../../types/issue'
import '../../components/board/board.css'

export type BoardPageProps = {
  workspaceId: string
  projectId: string
}

export default function BoardPage({ workspaceId, projectId }: BoardPageProps) {
  const navigate = useNavigate()
  const { teamId, loading: teamLoading } = useCurrentTeam(workspaceId)
  const pack = useBoardPackStats({
    workspaceId,
    projectId,
    teamId,
    enabled: !teamLoading,
  })
  const { issues, setIssues, loading } = pack
  const [open, setOpen] = useState(false)
  const [current, setCurrent] = useState<Issue | null>(null)
  const [newTitle, setNewTitle] = useState('')
  const [creating, setCreating] = useState(false)
  const [colPrefs, setColPrefs] = useState<BoardColumnLayout>(() => defaultPrefs())
  const [colsLoading, setColsLoading] = useState(true)
  const [archiving, setArchiving] = useState(false)

  useEffect(() => {
    let cancelled = false
    setColsLoading(true)
    setColPrefs(defaultPrefs())
    ;(async () => {
      try {
        const prefs = await getBoardColumnPrefs(workspaceId, projectId)
        if (cancelled) return
        setColPrefs(layoutFromApiColumns(prefs?.columns))
      } catch {
        if (!cancelled) setColPrefs(defaultPrefs())
      } finally {
        if (!cancelled) setColsLoading(false)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [workspaceId, projectId])

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

  const applyColPrefs = async (next: BoardColumnLayout) => {
    if (isDefaultLayout(next)) {
      await resetBoardColumnPrefs(workspaceId, projectId)
    } else {
      await putBoardColumnPrefs(workspaceId, projectId, layoutToApiColumns(next))
    }
    setColPrefs(next)
  }


  /** C1/C2：POST archive；乐观去卡，失败回滚 + 中文 Toast */
  const handleArchiveIssues = useCallback(
    async (ids: string[]) => {
      if (!ids.length || archiving) return
      const prev = issues
      const idSet = new Set(ids)
      setArchiving(true)
      setIssues((list) => list.filter((i) => !idSet.has(i.id)))
      if (current && idSet.has(current.id)) {
        setOpen(false)
        setCurrent(null)
      }
      try {
        await archiveIssues(workspaceId, ids)
        message.success(ids.length === 1 ? '已归档' : `已归档 ${ids.length} 条`)
      } catch (err) {
        setIssues(prev)
        const msg =
          err instanceof ApiError
            ? err.message
            : err instanceof Error
              ? err.message
              : '归档失败'
        message.error(msg || '归档接口尚未就绪')
        throw err
      } finally {
        setArchiving(false)
      }
    },
    [archiving, issues, current, setIssues, workspaceId],
  )

  const goCycles = () => {
    navigate(`/w/${workspaceId}/projects/${projectId}/cycles`)
  }

  const goBlockers = () => {
    navigate(`/w/${workspaceId}/projects/${projectId}/issues?blocked=1&overdue=1`)
  }

  return (
    <div className="acb-board-page">
      <MockBanner />
      <div className="acb-page-bar">
        <Typography.Title level={4}>看板</Typography.Title>
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
          <Button
            type="text"
            size="small"
            icon={<ReloadOutlined />}
            onClick={() => void pack.reload()}
          >
            刷新
          </Button>
          <ColumnTemplatePicker prefs={colPrefs} onApply={applyColPrefs} />
        </Space>
      </div>

      {!loading && !colsLoading ? (
        <BoardPackStrip
          done={pack.done}
          total={pack.total}
          activeCycle={pack.activeCycle}
          collaborators={pack.collaborators}
          blockerCount={pack.blockerCount}
          onCycleClick={goCycles}
          onBlockersClick={goBlockers}
        />
      ) : null}

      {loading || colsLoading ? (
        <div style={{ textAlign: 'center', padding: 32 }}>
          <Spin size="small" />
        </div>
      ) : issues.length === 0 ? (
        <Empty style={{ marginTop: 48 }} description="还没有 Issue，新建一条或套用列模板">
          <Space wrap>
            <Input
              size="small"
              placeholder="标题"
              value={newTitle}
              onChange={(e) => setNewTitle(e.target.value)}
              onPressEnter={() => void handleCreate()}
              onKeyDown={(e) => {
                if (e.key === 'Escape') {
                  setNewTitle('')
                  ;(e.target as HTMLInputElement).blur()
                }
              }}
              style={{ width: 240 }}
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
            <ColumnTemplatePicker
              prefs={colPrefs}
              onApply={applyColPrefs}
              variant="cta"
            />
          </Space>
        </Empty>
      ) : (
        <KanbanBoard
          workspaceId={workspaceId}
          issues={issues}
          selectedId={current?.id}
          columns={colPrefs.columns}
          onChange={setIssues}
          onOpen={(issue) => {
            setCurrent(issue)
            setOpen(true)
          }}
          onArchiveIssues={handleArchiveIssues}
          archiving={archiving}
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
          const archived =
            saved.archived_at != null && String(saved.archived_at).trim() !== ''
          if (archived) {
            setIssues((prev) => prev.filter((i) => i.id !== saved.id))
            setCurrent(saved)
            return
          }
          // B2：解档后回默认看板（若先前被剔出则 upsert）
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
