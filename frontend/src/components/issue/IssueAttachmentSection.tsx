import { useCallback, useEffect, useState } from 'react'
import { Spin, message } from 'antd'
import { deleteAttachment, listIssueAttachments } from '../../api/attachments'
import type { IssueAttachment } from '../../types/issue'
import { IssueAttachmentList } from './IssueAttachmentList'
import { IssueAttachmentUpload } from './IssueAttachmentUpload'

type Props = {
  workspaceId?: string
  issueId?: string
}

/**
 * 描述区旁「附件」区块：紧凑列表 +「添加附件」。
 * List/Delete 硬接 Issue 域；上传失败由 Upload 中文 Toast。
 */
export function IssueAttachmentSection({ workspaceId, issueId }: Props) {
  const [attachments, setAttachments] = useState<IssueAttachment[]>([])
  const [loading, setLoading] = useState(false)
  const [deletingId, setDeletingId] = useState<string | undefined>()

  const reload = useCallback(() => {
    if (!workspaceId || !issueId) {
      setAttachments([])
      return
    }
    setLoading(true)
    void listIssueAttachments(workspaceId, issueId)
      .then((res) => {
        setAttachments(res.attachments || [])
      })
      .catch(() => {
        // api() 已 Toast
        setAttachments([])
      })
      .finally(() => setLoading(false))
  }, [workspaceId, issueId])

  useEffect(() => {
    reload()
  }, [reload])

  const handleDelete = (attachmentId: string) => {
    if (!workspaceId) {
      message.warning('缺少工作区')
      return
    }
    setDeletingId(attachmentId)
    void deleteAttachment(workspaceId, attachmentId)
      .then(() => {
        message.success('已删除')
        reload()
      })
      .catch(() => {
        // api() 已 Toast
      })
      .finally(() => setDeletingId(undefined))
  }

  return (
    <div className="acb-issue-attachments">
      <div className="acb-issue-attachments-head">
        <div className="acb-issue-section-label" style={{ marginBottom: 0 }}>
          附件{attachments.length > 0 ? ` (${attachments.length})` : ''}
        </div>
        <IssueAttachmentUpload
          workspaceId={workspaceId}
          issueId={issueId}
          label="添加附件"
          onUploaded={reload}
        />
      </div>
      <Spin spinning={loading} size="small">
        <IssueAttachmentList
          attachments={attachments}
          onDelete={handleDelete}
          deletingId={deletingId}
        />
      </Spin>
    </div>
  )
}
