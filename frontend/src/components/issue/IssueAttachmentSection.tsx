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
 * 描述区旁「附件」区块（R3）：空态仅「添加附件…」；有文件再出列表+轻量添加。
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
      .catch(() => {})
      .finally(() => setDeletingId(undefined))
  }

  const empty = attachments.length === 0

  return (
    <div className="acb-issue-attachments">
      {empty ? (
        <div className="acb-issue-attachments-empty">
          <IssueAttachmentUpload
            workspaceId={workspaceId}
            issueId={issueId}
            label="添加附件…"
            onUploaded={reload}
          />
        </div>
      ) : (
        <>
          <div className="acb-issue-attachments-head">
            <div className="acb-issue-section-label" style={{ marginBottom: 0 }}>
              附件 ({attachments.length})
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
              emptyText=""
            />
          </Spin>
        </>
      )}
      {empty && loading ? <Spin size="small" /> : null}
    </div>
  )
}
