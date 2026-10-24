import { useRef, useState } from 'react'
import { Button, Upload, message } from 'antd'
import { PaperClipOutlined } from '@ant-design/icons'
import type { UploadProps } from 'antd'
import {
  ATTACHMENT_ACCEPT,
  uploadIssueAttachment,
  uploadCommentAttachment,
  validateAttachmentFile,
} from '../../api/attachments'
import { ApiError } from '../../api/client'

type Props = {
  workspaceId?: string
  /** Issue 级上传 / 评论上传均需要 */
  issueId?: string
  /** 有值则走评论附件 POST（须同时有 issueId） */
  commentId?: string | null
  /** 紧凑：评论区回形针 */
  compact?: boolean
  disabled?: boolean
  label?: string
  onUploaded?: () => void
  /** 选文件后不立刻上传（评论 composer：先挂再随发送落库） */
  onPickPending?: (file: File) => void
}

/**
 * 附件上传入口（antd Upload，非 Form 墙）。
 * 客户端预检类型/≤10MB；上传中防重复；失败中文 Toast。
 */
export function IssueAttachmentUpload({
  workspaceId,
  issueId,
  commentId,
  compact = false,
  disabled = false,
  label,
  onUploaded,
  onPickPending,
}: Props) {
  const [uploading, setUploading] = useState(false)
  const uploadingRef = useRef(false)

  const beforeUpload: UploadProps['beforeUpload'] = (file) => {
    const f = file as File
    const pre = validateAttachmentFile(f)
    if (pre) {
      message.warning(pre)
      return Upload.LIST_IGNORE
    }
    if (!workspaceId) {
      message.warning('缺少工作区')
      return Upload.LIST_IGNORE
    }
    if (!issueId) {
      message.warning('缺少 Issue')
      return Upload.LIST_IGNORE
    }
    if (onPickPending) {
      onPickPending(f)
      return Upload.LIST_IGNORE
    }

    if (uploadingRef.current) {
      message.info('正在上传，请稍候')
      return Upload.LIST_IGNORE
    }

    uploadingRef.current = true
    setUploading(true)

    const run = commentId
      ? uploadCommentAttachment(workspaceId, issueId, commentId, f)
      : uploadIssueAttachment(workspaceId, issueId, f)

    void run
      .then(() => {
        message.success('已上传')
        onUploaded?.()
      })
      .catch((err: unknown) => {
        // 上传路由若尚未齐：中文 Toast，不静默
        if (err instanceof ApiError) {
          message.error(err.message || '上传失败')
        } else {
          message.error(err instanceof Error ? err.message : '上传失败')
        }
      })
      .finally(() => {
        uploadingRef.current = false
        setUploading(false)
      })

    return false
  }

  const busy = disabled || uploading
  const btnLabel = label ?? (compact ? '附件' : '添加附件')

  return (
    <div className="acb-issue-attachment-upload">
      <Upload
        accept={ATTACHMENT_ACCEPT}
        showUploadList={false}
        beforeUpload={beforeUpload}
        disabled={busy}
        multiple={false}
      >
        {compact ? (
          <Button
            type="text"
            size="small"
            icon={<PaperClipOutlined />}
            loading={uploading}
            disabled={busy}
            title="添加附件（≤10MB）"
          >
            {btnLabel}
          </Button>
        ) : (
          <Button
            type="link"
            size="small"
            icon={<PaperClipOutlined />}
            loading={uploading}
            disabled={busy}
            className="acb-issue-attachment-add-btn"
          >
            {btnLabel}
          </Button>
        )}
      </Upload>
    </div>
  )
}
