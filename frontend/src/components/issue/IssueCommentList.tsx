import { Button, Input, Popconfirm, Space, Spin, Typography } from 'antd'
import { DeleteOutlined } from '@ant-design/icons'
import type { KeyboardEvent } from 'react'
import type { Comment } from '../../types/issue'
import { formatShanghai } from '../../utils/time'
import { IssueAttachmentUpload } from './IssueAttachmentUpload'

type Props = {
  loading: boolean
  comments: Comment[]
  memberName: (id: string | null | undefined) => string | null
  commentBody: string
  commentSending: boolean
  onCommentBodyChange: (v: string) => void
  onSend: () => void
  onRemove: (c: Comment) => void
  workspaceId?: string
  issueId?: string
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
  const onKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
      e.preventDefault()
      onSend()
    }
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
              const author = memberName(c.author_id)
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
                    </div>
                    <Popconfirm
                      title="删除此评论？"
                      okText="删除"
                      cancelText="取消"
                      okButtonProps={{ danger: true, size: 'small' }}
                      onConfirm={() => onRemove(c)}
                    >
                      <Button type="text" size="small" danger icon={<DeleteOutlined />} />
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
        <div className="acb-issue-comment-composer-actions">
          <IssueAttachmentUpload
            compact
            workspaceId={workspaceId}
            issueId={issueId}
          />
          <Button
            size="small"
            type="primary"
            loading={commentSending}
            onClick={onSend}
            disabled={!commentBody.trim()}
          >
            发送评论
          </Button>
        </div>
      </div>
    </div>
  )
}
