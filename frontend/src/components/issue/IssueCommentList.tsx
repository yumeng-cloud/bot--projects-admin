import {
  useCallback,
  useEffect,
  useState,
  type ClipboardEvent,
  type DragEvent,
  type KeyboardEvent,
} from 'react'
import { Button, Input, Popconfirm, Space, Spin, Typography, message } from 'antd'
import {
  CloseOutlined,
  DeleteOutlined,
  FileOutlined,
  PaperClipOutlined,
} from '@ant-design/icons'
import type { Comment, IssueAttachment } from '../../types/issue'
import { formatShanghai } from '../../utils/time'
import {
  deleteAttachment,
  listCommentAttachments,
  validateAttachmentFile,
} from '../../api/attachments'
import { IssueAttachmentList } from './IssueAttachmentList'
import { IssueAttachmentUpload } from './IssueAttachmentUpload'
import { WikiBodyViewer } from '../wiki/WikiBodyViewer'

type Props = {
  loading: boolean
  comments: Comment[]
  memberName: (id: string | null | undefined) => string | null
  commentBody: string
  commentSending: boolean
  onCommentBodyChange: (v: string) => void
  /** 发送：正文已在父级；files 为 composer 待挂附件 */
  onSend: (files: File[]) => void | Promise<void>
  onRemove: (c: Comment) => void
  workspaceId?: string
  issueId?: string
}

function isImageFile(f: File): boolean {
  return f.type.startsWith('image/') || /\.(png|jpe?g|gif|webp|bmp)$/i.test(f.name)
}

/** 待发芯片：图缩略 / 文件名，可删 */
function PendingChip({
  file,
  onRemove,
}: {
  file: File
  onRemove: () => void
}) {
  const [thumb, setThumb] = useState<string | null>(null)

  useEffect(() => {
    if (!isImageFile(file)) {
      setThumb(null)
      return
    }
    const url = URL.createObjectURL(file)
    setThumb(url)
    return () => URL.revokeObjectURL(url)
  }, [file])

  return (
    <span className="acb-issue-comment-pending-chip" title={file.name}>
      {thumb ? (
        <img src={thumb} alt="" className="acb-issue-comment-pending-thumb" />
      ) : (
        <span className="acb-issue-comment-pending-fileicon" aria-hidden>
          <FileOutlined />
        </span>
      )}
      <Typography.Text ellipsis className="acb-issue-comment-pending-name">
        {file.name}
      </Typography.Text>
      <Button
        type="text"
        size="small"
        icon={<CloseOutlined />}
        aria-label={`移除 ${file.name}`}
        className="acb-issue-comment-pending-remove"
        onClick={onRemove}
      />
    </span>
  )
}

function CommentAttachBlock({
  workspaceId,
  issueId,
  commentId,
}: {
  workspaceId: string
  issueId: string
  commentId: string
}) {
  const [attachments, setAttachments] = useState<IssueAttachment[]>([])
  const [loading, setLoading] = useState(false)
  const [deletingId, setDeletingId] = useState<string | undefined>()

  const reload = useCallback(() => {
    setLoading(true)
    void listCommentAttachments(workspaceId, issueId, commentId)
      .then((res) => setAttachments(res.attachments || []))
      .catch(() => setAttachments([]))
      .finally(() => setLoading(false))
  }, [workspaceId, issueId, commentId])

  useEffect(() => {
    reload()
  }, [reload])

  const handleDelete = (id: string) => {
    setDeletingId(id)
    void deleteAttachment(workspaceId, id)
      .then(() => {
        message.success('已删除')
        reload()
      })
      .catch(() => {})
      .finally(() => setDeletingId(undefined))
  }

  return (
    <div className="acb-issue-comment-attachments">
      <Spin spinning={loading} size="small">
        {attachments.length > 0 ? (
          <IssueAttachmentList
            attachments={attachments}
            onDelete={handleDelete}
            deletingId={deletingId}
            emptyText=""
          />
        ) : null}
      </Spin>
      {/* 补挂：hover 显轻量回形针，无常驻「挂附件」文案 */}
      <IssueAttachmentUpload
        compact
        workspaceId={workspaceId}
        issueId={issueId}
        commentId={commentId}
        className="acb-issue-comment-attach-hover"
        onUploaded={reload}
      />
    </div>
  )
}

export function IssueCommentList({
  loading,
  comments,
  memberName,
  commentBody,
  commentSending,
  onCommentBodyChange,
  onSend,
  onRemove,
  workspaceId,
  issueId,
}: Props) {
  const [pending, setPending] = useState<File[]>([])
  const [dragOver, setDragOver] = useState(false)

  const onKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
      e.preventDefault()
      void doSend()
    }
  }

  const doSend = async () => {
    if (!commentBody.trim() || commentSending) return
    const files = pending
    await onSend(files)
    setPending([])
  }

  const addPending = useCallback((file: File) => {
    const pre = validateAttachmentFile(file)
    if (pre) {
      message.warning(pre)
      return
    }
    setPending((prev) => {
      if (prev.some((f) => f.name === file.name && f.size === file.size)) {
        message.info('已在待发送列表')
        return prev
      }
      return [...prev, file]
    })
  }, [])

  const takeFiles = useCallback(
    (list: FileList | File[] | null | undefined) => {
      if (!list) return
      const arr = Array.from(list)
      for (const f of arr) addPending(f)
    },
    [addPending],
  )

  const onPaste = (e: ClipboardEvent<HTMLTextAreaElement>) => {
    const items = e.clipboardData?.items
    if (!items?.length) return
    const files: File[] = []
    for (const item of Array.from(items)) {
      if (item.kind === 'file') {
        const f = item.getAsFile()
        if (f) files.push(f)
      }
    }
    if (files.length === 0) return
    // 有图/文件时拦截默认粘贴路径里的 binary 噪音；纯文字仍走默认
    const hasImage = files.some((f) => f.type.startsWith('image/'))
    if (hasImage) e.preventDefault()
    takeFiles(files)
  }

  const onDragOver = (e: DragEvent<HTMLDivElement>) => {
    if (!e.dataTransfer?.types?.includes('Files')) return
    e.preventDefault()
    e.stopPropagation()
    setDragOver(true)
  }

  const onDragLeave = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault()
    // 仅离开 composer 根时关高亮
    if (e.currentTarget.contains(e.relatedTarget as Node)) return
    setDragOver(false)
  }

  const onDrop = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault()
    e.stopPropagation()
    setDragOver(false)
    takeFiles(e.dataTransfer?.files)
  }

  return (
    <div className="acb-issue-comments">
      <div className="acb-issue-section-label">
        评论{comments.length > 0 ? ` (${comments.length})` : ''}
      </div>
      <Spin spinning={loading} size="small">
        {comments.length === 0 && !loading ? (
          <div className="acb-issue-muted-empty">暂无评论</div>
        ) : (
          <div>
            {comments.map((c) => {
              const author =
                (c.author?.display_name && c.author.display_name.trim()) ||
                memberName(c.author_id)
              return (
                <div key={c.id} className="acb-issue-comment-item">
                  <div className="acb-issue-comment-head">
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <Space size={6} style={{ marginBottom: 2 }}>
                        <Typography.Text strong style={{ fontSize: 12 }}>
                          {author || '成员'}
                        </Typography.Text>
                        <Typography.Text type="secondary" style={{ fontSize: 11 }}>
                          {formatShanghai(c.created_at)}
                        </Typography.Text>
                      </Space>
                      <div className="acb-issue-comment-md">
                        <WikiBodyViewer value={c.body || ''} emptyText="" />
                      </div>
                      {workspaceId && issueId ? (
                        <CommentAttachBlock
                          workspaceId={workspaceId}
                          issueId={issueId}
                          commentId={c.id}
                        />
                      ) : null}
                    </div>
                    <Popconfirm
                      title="删除此评论？"
                      okText="删除"
                      cancelText="取消"
                      okButtonProps={{ danger: true, size: 'small' }}
                      onConfirm={() => onRemove(c)}
                    >
                      <Button
                        type="text"
                        size="small"
                        danger
                        className="acb-issue-comment-delete"
                        icon={<DeleteOutlined />}
                      />
                    </Popconfirm>
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </Spin>

      <div
        className={`acb-issue-comment-composer${dragOver ? ' is-dragover' : ''}`}
        onDragOver={onDragOver}
        onDragLeave={onDragLeave}
        onDrop={onDrop}
      >
        <Input.TextArea
          size="small"
          rows={2}
          placeholder="写评论…（Ctrl/⌘+Enter 发送；可粘贴图片或拖入文件）"
          value={commentBody}
          onChange={(e) => onCommentBodyChange(e.target.value)}
          onKeyDown={onKeyDown}
          onPaste={onPaste}
        />
        {pending.length > 0 ? (
          <div className="acb-issue-comment-pending">
            {pending.map((f) => (
              <PendingChip
                key={`${f.name}-${f.size}-${f.lastModified}`}
                file={f}
                onRemove={() => setPending((p) => p.filter((x) => x !== f))}
              />
            ))}
          </div>
        ) : null}
        <div className="acb-issue-comment-composer-actions">
          <IssueAttachmentUpload
            compact
            workspaceId={workspaceId}
            issueId={issueId}
            onPickPending={addPending}
          />
          <Button
            size="small"
            type="link"
            className="acb-issue-comment-send"
            loading={commentSending}
            onClick={() => void doSend()}
            disabled={!commentBody.trim()}
          >
            发送
          </Button>
        </div>
        {dragOver ? (
          <div className="acb-issue-comment-drop-hint" aria-hidden>
            <PaperClipOutlined /> 松开以添加附件
          </div>
        ) : null}
      </div>
    </div>
  )
}
