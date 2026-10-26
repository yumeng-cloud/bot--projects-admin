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
  /** 紧凑：仅回形针图标（composer / 已发补挂 / Issue 区） */
  compact?: boolean
  disabled?: boolean
  /**
   * 图标旁弱提示文案（非按钮字）。空态可用，如「点击添加（≤10MB）」。
   * 禁止把「添加附件 / 挂附件」塞进按钮 children。
   */
  hint?: string
  /** 额外 class（如 hover 显隐） */
  className?: string
  onUploaded?: () => void
  /** 选文件后不立刻上传（评论 composer：先挂再随发送落库） */
  onPickPending?: (file: File) => void
}

const TOOLTIP = '添加附件（≤10MB）'

/**
 * 附件上传入口（antd Upload，非 Form 墙）。
 * 仅回形针图标 + tooltip；可选旁侧弱提示。禁老 OA 文案钮。
 */
export function IssueAttachmentUpload({
  workspaceId,
  issueId,
  commentId,
  compact = false,
  disabled = false,
  hint,
  className,
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
  const rootClass = ['acb-issue-attachment-upload', className].filter(Boolean).join(' ')

  return (
    <div className={rootClass}>
      <Upload
        accept={ATTACHMENT_ACCEPT}
        showUploadList={false}
        beforeUpload={beforeUpload}
        disabled={busy}
        multiple={false}
      >
        <Button
          type="text"
          size="small"
          icon={<PaperClipOutlined />}
          loading={uploading}
          disabled={busy}
          title={TOOLTIP}
          aria-label={TOOLTIP}
          className={
            compact
              ? 'acb-issue-attachment-clip-btn'
              : 'acb-issue-attachment-clip-btn acb-issue-attachment-add-btn'
          }
        />
      </Upload>
      {hint ? <span className="acb-issue-attachment-upload-hint">{hint}</span> : null}
    </div>
  )
}
