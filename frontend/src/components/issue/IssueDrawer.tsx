import { useCallback, useEffect, useMemo, useState } from 'react'
import { Drawer, Form, Space, Tag, message, Typography } from 'antd'
import { patchIssue, listProjectMembers, listProjectLabels, listIssues } from '../../api/issues'
import type { IssuePatch } from '../../api/issues'
import { listCycles } from '../../api/cycles'
import { listComments, createComment, deleteComment } from '../../api/comments'
import { uploadCommentAttachment } from '../../api/attachments'
import {
  listRelations,
  createRelation,
  deleteRelationOnIssue,
} from '../../api/relations'
import { listIssueActivities } from '../../api/activities'
import type {
  Activity,
  Cycle,
  Issue,
  IssuePriority,
  IssueRelation,
  IssueStatus,
  Label,
  MemberRef,
  Comment,
  RelationType,
} from '../../types/issue'
import { PRIORITY_LABELS, STATUS_LABELS, isIssueArchived, memberLabel } from '../../types/issue'
import { IssueDetailMain } from './IssueDetailMain'
import { IssuePropertyRail } from './IssuePropertyRail'
import { dayjs, TZ } from '../../utils/time'
import './issue-detail.css'

function calendarDayValue(v?: string | null) {
  if (!v) return undefined
  const s = String(v).trim()
  if (/^\d{4}-\d{2}-\d{2}$/.test(s)) return dayjs.tz(s, TZ)
  const d = dayjs.utc(s).tz(TZ)
  return d.isValid() ? d : undefined
}

function calendarDayString(v: unknown): string | null {
  if (v == null || v === '') return null
  if (typeof v === 'string') {
    if (/^\d{4}-\d{2}-\d{2}$/.test(v)) return v
    const d = dayjs(v)
    return d.isValid() ? d.tz(TZ).format('YYYY-MM-DD') : null
  }
  // Dayjs
  const d = v as { isValid?: () => boolean; tz: (z: string) => { format: (f: string) => string }; format: (f: string) => string }
  if (typeof d.isValid === 'function' && !d.isValid()) return null
  try {
    return d.tz(TZ).format('YYYY-MM-DD')
  } catch {
    return typeof d.format === 'function' ? d.format('YYYY-MM-DD') : null
  }
}

type Props = {
  open: boolean
  issue: Issue | null
  workspaceId: string
  projectId: string
  /** Current team pool (T6I); scopes parent/relation issue picker */
  teamId?: string | null
  onClose: () => void
  onSaved: (issue: Issue) => void
}

const ACTOR_KIND_LABELS: Record<string, string> = {
  human: '用户',
  bot: 'Bot',
  system: '系统',
}

function actorKindLabel(kind: string): string {
  return ACTOR_KIND_LABELS[kind] || kind
}

function parseActivityPayload(payload: string | null | undefined): Record<string, unknown> {
  if (!payload) return {}
  const s = payload.trim()
  if (!s || s === '{}' || s === '[]' || s === 'null') return {}
  try {
    const v = JSON.parse(s) as unknown
    if (v && typeof v === 'object' && !Array.isArray(v)) return v as Record<string, unknown>
  } catch {
    /* ignore */
  }
  return {}
}

function statusZh(v: unknown): string | null {
  if (typeof v !== 'string') return null
  return STATUS_LABELS[v as IssueStatus] || null
}

function priorityZh(v: unknown): string | null {
  if (typeof v !== 'string') return null
  return PRIORITY_LABELS[v as IssuePriority] || null
}

function activitySentence(action: string, payload: string | null | undefined): string {
  const p = parseActivityPayload(payload)
  switch (action) {
    case 'issue.created':
      return '创建了 Issue'
    case 'issue.deleted':
      return '删除了 Issue'
    case 'issue.status_changed': {
      const label = statusZh(p.status)
      return label ? `把状态改为${label}` : '改了状态'
    }
    case 'issue.priority_changed': {
      const label = priorityZh(p.priority)
      return label ? `把优先级改为${label}` : '改了优先级'
    }
    case 'issue.assignee_changed':
      return '变更了负责人'
    case 'issue.estimate_changed':
      return '变更了估时'
    case 'issue.cycle_changed':
      return '变更了 Cycle'
    case 'issue.title_changed':
      return '变更了标题'
    case 'issue.updated': {
      if (p.status !== undefined) {
        const label = statusZh(p.status)
        return label ? `把状态改为${label}` : '改了状态'
      }
      if (p.priority !== undefined) {
        const label = priorityZh(p.priority)
        return label ? `把优先级改为${label}` : '改了优先级'
      }
      if (p.title !== undefined) return '变更了标题'
      if (p.description !== undefined) return '更新了描述'
      if (p.assignee_id !== undefined) return '变更了负责人'
      if (p.estimate !== undefined) return '变更了估时'
      if (p.cycle_id !== undefined) return '变更了 Cycle'
      if (p.label_ids !== undefined) return '更新了标签'
      return '更新了 Issue'
    }
    case 'comment.created': {
      const text = [p.summary, p.comment, p.body]
        .find((v) => typeof v === 'string' && String(v).trim())
      if (typeof text === 'string' && text.trim()) {
        const one = text.trim().replace(/\s+/g, ' ')
        return one.length > 80 ? `评论：${one.slice(0, 80)}…` : `评论：${one}`
      }
      return '加了评论'
    }
    case 'seat.force_bind': {
      const text = [p.summary, p.comment, p.body]
        .find((v) => typeof v === 'string' && String(v).trim())
      if (typeof text === 'string' && text.trim()) return String(text).trim()
      return '强制补位了席位'
    }
    case 'comment.updated':
      return '更新了评论'
    case 'comment.deleted':
      return '删除了评论'
    case 'relation.created':
      return '添加了关联'
    case 'relation.deleted':
      return '删除了关联'
    case 'issue.label_attached':
    case 'label.attached':
      return '添加了标签'
    case 'issue.label_detached':
    case 'label.detached':
      return '移除了标签'
    case 'issue.cycle_completed_moved':
      return 'Cycle 完成后移动'
    default:
      return '更新了 Issue'
  }
}

function issueOptionLabel(i: Pick<Issue, 'identifier' | 'title'>): string {
  return `${i.identifier} · ${i.title}`
}


const ISSUE_DRAWER_WIDE = 940
/** 与 issue-detail.css @media (max-width: 639px) 轨上下布局断点一致 */
const ISSUE_DRAWER_NARROW_MQ = '(max-width: 639px)'
/** 规格 ISSUE_DETAIL_LAYOUT §5 / C6：窄屏显式贴近 min(100vw, 560) */
const ISSUE_DRAWER_NARROW_WIDTH = 'min(100vw, 560px)'

function useIssueDrawerNarrow(): boolean {
  const [narrow, setNarrow] = useState(() =>
    typeof window !== 'undefined' ? window.matchMedia(ISSUE_DRAWER_NARROW_MQ).matches : false,
  )
  useEffect(() => {
    const mq = window.matchMedia(ISSUE_DRAWER_NARROW_MQ)
    const onChange = () => setNarrow(mq.matches)
    onChange()
    mq.addEventListener('change', onChange)
    return () => mq.removeEventListener('change', onChange)
  }, [])
  return narrow
}

/**
 * Issue 详情壳：顶栏 chrome + 主栏 + Property 轨。
 * 布局见 frontend/ISSUE_DETAIL_LAYOUT.md
 */
export function IssueDrawer({ open, issue, workspaceId, projectId, teamId, onClose, onSaved }: Props) {
  const narrowDrawer = useIssueDrawerNarrow()
  const [form] = Form.useForm()
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [quickSaving, setQuickSaving] = useState(false)
  const [members, setMembers] = useState<MemberRef[]>([])
  const [labels, setLabels] = useState<Label[]>([])
  const [cycles, setCycles] = useState<Cycle[]>([])
  const [projectIssues, setProjectIssues] = useState<Issue[]>([])

  const [comments, setComments] = useState<Comment[]>([])
  const [relations, setRelations] = useState<IssueRelation[]>([])
  const [activities, setActivities] = useState<Activity[]>([])
  const [sideLoading, setSideLoading] = useState(false)
  const [commentBody, setCommentBody] = useState('')
  const [commentSending, setCommentSending] = useState(false)
  const [relType, setRelType] = useState<RelationType>('related')
  const [relToId, setRelToId] = useState<string | undefined>()
  const [relAdding, setRelAdding] = useState(false)

  const memberName = useMemo(() => {
    const map = new Map(members.map((m) => [m.id, memberLabel(m)]))
    return (id: string | null | undefined) => (id ? map.get(id) || null : null)
  }, [members])

  const issueById = useMemo(() => {
    const map = new Map(projectIssues.map((i) => [i.id, i]))
    return (id: string) => map.get(id)
  }, [projectIssues])

  const relationChoices = useMemo(
    () => projectIssues.filter((i) => i.id !== issue?.id),
    [projectIssues, issue?.id],
  )

  const loadSide = useCallback(
    async (issueId: string) => {
      setSideLoading(true)
      try {
        const [cRes, rRes, aRes] = await Promise.allSettled([
          listComments(workspaceId, issueId),
          listRelations(workspaceId, issueId),
          listIssueActivities(workspaceId, issueId),
        ])
        if (cRes.status === 'fulfilled') setComments(cRes.value.comments || [])
        else setComments([])
        if (rRes.status === 'fulfilled') setRelations(rRes.value.relations || [])
        else setRelations([])
        if (aRes.status === 'fulfilled') setActivities(aRes.value.activities || [])
        else setActivities([])
      } finally {
        setSideLoading(false)
      }
    },
    [workspaceId],
  )

  useEffect(() => {
    if (!open || !issue) return
    setTitle(issue.title)
    setDescription(issue.description || '')
    form.setFieldsValue({
      status: issue.status,
      priority: issue.priority,
      estimate: issue.estimate,
      assignee_id: issue.assignee_id,
      cycle_id: issue.cycle_id,
      label_ids: (issue.labels || []).map((l) => l.id),
      start_at: calendarDayValue(issue.start_at),
      due_at: calendarDayValue(issue.due_at),
    })
    setCommentBody('')
    setRelToId(undefined)
    setRelType('related')
    ;(async () => {
      const [m, l, c, iss] = await Promise.all([
        listProjectMembers(workspaceId),
        listProjectLabels(workspaceId),
        listCycles(workspaceId, projectId),
        listIssues(workspaceId, { project_id: projectId, team_id: teamId || undefined }),
      ])
      setMembers(m.members)
      setLabels(l.labels)
      setCycles(c.cycles)
      setProjectIssues(iss.issues || [])
    })()
    void loadSide(issue.id)
  }, [open, issue, form, workspaceId, projectId, teamId, loadSide])

  /** 标题 / 描述：blur 自动 PATCH（顶栏不再放实心保存） */
  const saveTitleDescription = async () => {
    if (!issue || quickSaving) return
    const t = title.trim()
    if (!t) {
      message.warning('请输入标题')
      setTitle(issue.title)
      return
    }
    const nextDesc = description ?? ''
    const prevDesc = issue.description || ''
    if (t === issue.title && nextDesc === prevDesc) return

    const prev = issue
    setQuickSaving(true)
    try {
      const { issue: saved } = await patchIssue(workspaceId, prev.id, {
        version: prev.version,
        title: t,
        description: nextDesc,
      })
      onSaved(saved)
      void loadSide(saved.id)
    } catch {
      setTitle(prev.title)
      setDescription(prev.description || '')
    } finally {
      setQuickSaving(false)
    }
  }

  /** Property 轨字段即时 PATCH */
  const quickPatch = async (
    body: Omit<IssuePatch, 'version'>,
    optimistic?: Partial<Issue>,
    revertForm?: () => void,
  ) => {
    if (!issue || quickSaving) return
    const prev = issue
    if (optimistic) onSaved({ ...issue, ...optimistic })
    setQuickSaving(true)
    try {
      const { issue: saved } = await patchIssue(workspaceId, prev.id, {
        version: prev.version,
        ...body,
      })
      onSaved(saved)
      void loadSide(saved.id)
    } catch {
      onSaved(prev)
      revertForm?.()
    } finally {
      setQuickSaving(false)
    }
  }

  const quickStatus = async (status: IssueStatus) => {
    if (!issue || issue.status === status) return
    const prevStatus = issue.status
    form.setFieldValue('status', status)
    await quickPatch({ status }, { status }, () => form.setFieldValue('status', prevStatus))
  }

  const quickPriority = async (priority: IssuePriority) => {
    if (!issue || issue.priority === priority) return
    const prevPriority = issue.priority
    form.setFieldValue('priority', priority)
    await quickPatch({ priority }, { priority }, () => form.setFieldValue('priority', prevPriority))
  }

  const quickAssignee = async (assignee_id: string | null) => {
    if (!issue || (issue.assignee_id ?? null) === (assignee_id ?? null)) return
    const prev = issue.assignee_id
    await quickPatch(
      { assignee_id },
      { assignee_id },
      () => form.setFieldValue('assignee_id', prev),
    )
  }

  const quickCycle = async (cycle_id: string | null) => {
    if (!issue || (issue.cycle_id ?? null) === (cycle_id ?? null)) return
    const prev = issue.cycle_id
    await quickPatch({ cycle_id }, { cycle_id }, () => form.setFieldValue('cycle_id', prev))
  }

  const quickEstimate = async (estimate: number | null) => {
    if (!issue) return
    const cur = issue.estimate ?? null
    const next = estimate ?? null
    if (cur === next) return
    await quickPatch({ estimate: next }, { estimate: next }, () =>
      form.setFieldValue('estimate', cur),
    )
  }

  const quickLabels = async (label_ids: string[]) => {
    if (!issue) return
    const prevIds = (issue.labels || []).map((l) => l.id)
    const same =
      prevIds.length === label_ids.length && prevIds.every((id) => label_ids.includes(id))
    if (same) return
    await quickPatch({ label_ids }, undefined, () => form.setFieldValue('label_ids', prevIds))
  }

  const quickStartAt = async (raw: unknown) => {
    if (!issue) return
    const next = calendarDayString(raw)
    const cur = issue.start_at ? String(issue.start_at).slice(0, 10) : null
    if (cur === next) return
    const revertTo = calendarDayValue(issue.start_at)
    await quickPatch({ start_at: next }, { start_at: next }, () =>
      form.setFieldValue('start_at', revertTo),
    )
  }

  const quickDueAt = async (raw: unknown) => {
    if (!issue) return
    const next = calendarDayString(raw)
    const cur = issue.due_at ? String(issue.due_at).slice(0, 10) : null
    if (cur === next) return
    const revertTo = calendarDayValue(issue.due_at)
    await quickPatch({ due_at: next }, { due_at: next }, () =>
      form.setFieldValue('due_at', revertTo),
    )
  }

  const otherIssueId = (rel: IssueRelation) =>
    rel.from_issue_id === issue?.id ? rel.to_issue_id : rel.from_issue_id

  const otherIssueLabel = (rel: IssueRelation) => {
    const id = otherIssueId(rel)
    const other = issueById(id)
    return other ? issueOptionLabel(other) : id
  }

  const addRelation = async () => {
    if (!issue || !relToId) {
      message.warning('请选择目标 Issue')
      return
    }
    setRelAdding(true)
    try {
      await createRelation(workspaceId, issue.id, {
        to_issue_id: relToId,
        relation_type: relType,
      })
      message.success('已添加关联')
      setRelToId(undefined)
      void loadSide(issue.id)
    } catch {
      /* toasted */
    } finally {
      setRelAdding(false)
    }
  }

  const removeRelation = async (rel: IssueRelation) => {
    if (!issue) return
    try {
      await deleteRelationOnIssue(workspaceId, issue.id, rel.id)
      message.success('已删除关联')
      void loadSide(issue.id)
    } catch {
      /* toasted */
    }
  }

  const sendComment = async (files: File[] = []) => {
    if (!issue || !commentBody.trim() || commentSending) return
    setCommentSending(true)
    try {
      const res = await createComment(workspaceId, issue.id, commentBody.trim())
      const cid = res.comment?.id
      if (cid && files.length) {
        for (const f of files) {
          await uploadCommentAttachment(workspaceId, issue.id, cid, f)
        }
      }
      message.success(files.length ? '已发送（含附件）' : '已发送')
      setCommentBody('')
      void loadSide(issue.id)
    } catch {
      /* toasted */
    } finally {
      setCommentSending(false)
    }
  }

  const removeComment = async (c: Comment) => {
    if (!issue) return
    try {
      await deleteComment(workspaceId, issue.id, c.id)
      message.success('已删除评论')
      void loadSide(issue.id)
    } catch {
      /* toasted */
    }
  }

  return (
    <Drawer
      title={
        <Space size={8} align="center" wrap>
          <span className="acb-issue-header-id">{issue ? issue.identifier : 'Issue'}</span>
          {issue && isIssueArchived(issue) ? (
            <Tag className="acb-issue-archived-tag" color="default">
              已归档
            </Tag>
          ) : null}
          {issue && (
            <Typography.Text type="secondary" style={{ fontSize: 11, fontWeight: 400 }}>
              v{issue.version}
            </Typography.Text>
          )}
        </Space>
      }
      width={narrowDrawer ? ISSUE_DRAWER_NARROW_WIDTH : ISSUE_DRAWER_WIDE}
      open={open}
      onClose={onClose}
      destroyOnHidden
      className="acb-drawer acb-issue-drawer"
      styles={{
        // 宽屏保留 92vw 防溢出；窄屏显式 min(100vw, 560)，勿再硬撑 820
        wrapper: narrowDrawer
          ? { width: ISSUE_DRAWER_NARROW_WIDTH, maxWidth: ISSUE_DRAWER_NARROW_WIDTH }
          : { maxWidth: '92vw' },
        body: { padding: 0, overflow: 'hidden', height: '100%' },
      }}
    >
      {issue && (
        <div className="acb-issue-shell">
          {isIssueArchived(issue) ? (
            <div className="acb-issue-archived-banner" role="status">
              此 Issue 已归档，默认列表与看板不可见。解档功能后置，本包仅只读提示。
            </div>
          ) : null}
          <div className="acb-issue-layout">
          <IssueDetailMain
            workspaceId={workspaceId}
            issueId={issue.id}
            title={title}
            description={description}
            onTitleChange={setTitle}
            onDescriptionChange={setDescription}
            onTitleBlur={() => void saveTitleDescription()}
            onDescriptionBlur={() => void saveTitleDescription()}
            sideLoading={sideLoading}
            comments={comments}
            memberName={memberName}
            commentBody={commentBody}
            commentSending={commentSending}
            onCommentBodyChange={setCommentBody}
            onSendComment={(files) => void sendComment(files)}
            onRemoveComment={(c) => void removeComment(c)}
            activities={activities}
            activitySentence={activitySentence}
            actorKindLabel={actorKindLabel}
          />
          <IssuePropertyRail
            form={form}
            quickSaving={quickSaving}
            members={members}
            labels={labels}
            cycles={cycles}
            onStatusChange={(s) => void quickStatus(s)}
            onPriorityChange={(p) => void quickPriority(p)}
            onAssigneeChange={(id) => void quickAssignee(id ?? null)}
            onCycleChange={(id) => void quickCycle(id ?? null)}
            onEstimateChange={(v) => void quickEstimate(v)}
            onLabelsChange={(ids) => void quickLabels(ids)}
            onStartAtChange={(v) => void quickStartAt(v)}
            onDueAtChange={(v) => void quickDueAt(v)}
            sideLoading={sideLoading}
            relations={relations}
            relationChoices={relationChoices}
            relType={relType}
            relToId={relToId}
            relAdding={relAdding}
            onRelTypeChange={setRelType}
            onRelToIdChange={setRelToId}
            onAddRelation={() => void addRelation()}
            onRemoveRelation={(rel) => void removeRelation(rel)}
            otherIssueLabel={otherIssueLabel}
            otherIssueId={otherIssueId}
            archivedAt={issue.archived_at}
          />
          </div>
        </div>
      )}
    </Drawer>
  )
}
