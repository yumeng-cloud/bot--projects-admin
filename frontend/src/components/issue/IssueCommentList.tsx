import { useCallback, useEffect, useState, type KeyboardEvent } from 'react'
import { Button, Input, Popconfirm, Space, Spin, Typography, message } from 'antd'
import { CloseOutlined, DeleteOutlined, PaperClipOutlined } from '@ant-design/icons'
import type { Comment, IssueAttachment } from '../../types/issue'
import { formatShanghai } from '../../utils/time'
import {
  deleteAttachment,
  listCommentAttachments,
} from '../../api/attachments'
import { IssueAttachmentList } from './IssueAttachmentList'
import { IssueAttachmentUpload } from './IssueAttachmentUpload'

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
    <div className="acb-issue-comment-attachments" style={{ marginTop: 6 }}>
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
      <IssueAttachmentUpload
        compact
        workspaceId={workspaceId}
        issueId={issueId}
        commentId={commentId}
        label="挂附件"
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

  const addPending = (file: File) => {
    setPending((prev) => {
      if (prev.some((f) => f.name === file.name && f.size === file.size)) {
        message.info('已在待发送列表')
        return prev
      }
      message.success(`已挂「${file.name}」，发送后写入评论`)
      return [...prev, file]
    })
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
                      <Typography.Paragraph
                        style={{
                          marginBottom: 0,
                          whiteSpace: 'pre-wrap',
                          fontSize: 12.5,
                          lineHeight: 1.5,
                        }}
                      >
                        {c.body}
                      </Typography.Paragraph>
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

      <div className="acb-issue-comment-composer">
        <Input.TextArea
          size="small"
          rows={2}
          placeholder="写评论…（Ctrl/⌘+Enter 发送）"
          value={commentBody}
          onChange={(e) => onCommentBodyChange(e.target.value)}
          onKeyDown={onKeyDown}
        />
        {pending.length > 0 ? (
          <div className="acb-issue-comment-pending" style={{ marginTop: 6 }}>
            {pending.map((f) => (
              <Space key={`${f.name}-${f.size}`} size={4} style={{ marginRight: 8 }}>
                <PaperClipOutlined />
                <Typography.Text style={{ fontSize: 12 }}>{f.name}</Typography.Text>
                <Button
                  type="text"
                  size="small"
                  icon={<CloseOutlined />}
                  aria-label={`移除 ${f.name}`}
                  onClick={() => setPending((p) => p.filter((x) => x !== f))}
                />
              </Space>
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
      </div>
    </div>
  )
}
