import { Input } from 'antd'
import type { Activity, Comment } from '../../types/issue'
import { IssueDescriptionEditor } from './IssueDescriptionEditor'
import { IssueAttachmentSection } from './IssueAttachmentSection'
import { IssueCommentList } from './IssueCommentList'
import { IssueActivityFold } from './IssueActivityFold'

type Props = {
  workspaceId?: string
  issueId?: string
  title: string
  description: string
  onTitleChange: (v: string) => void
  onDescriptionChange: (v: string) => void
  /** R1: blur 自动保存标题/描述 */
  onTitleBlur?: () => void
  onDescriptionBlur?: () => void
  // comments
  sideLoading: boolean
  comments: Comment[]
  memberName: (id: string | null | undefined) => string | null
  commentBody: string
  commentSending: boolean
  onCommentBodyChange: (v: string) => void
  onSendComment: () => void
  onRemoveComment: (c: Comment) => void
  // activity
  activities: Activity[]
  activitySentence: (action: string, payload: string | null | undefined) => string
  actorKindLabel: (kind: string) => string
}

/** 主栏：大标题 + MD 描述 + 附件 + 评论流 + 活动折叠（ISSUE_DETAIL_LAYOUT §3） */
export function IssueDetailMain({
  workspaceId,
  issueId,
  title,
  description,
  onTitleChange,
  onDescriptionChange,
  onTitleBlur,
  onDescriptionBlur,
  sideLoading,
  comments,
  memberName,
  commentBody,
  commentSending,
  onCommentBodyChange,
  onSendComment,
  onRemoveComment,
  activities,
  activitySentence,
  actorKindLabel,
}: Props) {
  return (
    <div className="acb-issue-main">
      <div className="acb-issue-main-inner">
        <div className="acb-issue-title-wrap">
          <Input
            className="acb-issue-title-input"
            variant="borderless"
            placeholder="Issue 标题"
            value={title}
            onChange={(e) => onTitleChange(e.target.value)}
            onBlur={() => onTitleBlur?.()}
          />
        </div>

        <div className="acb-issue-desc-wrap">
          <IssueDescriptionEditor
            value={description}
            onChange={onDescriptionChange}
            onBlur={onDescriptionBlur}
            placeholder="添加描述…"
          />
        </div>

        <IssueAttachmentSection workspaceId={workspaceId} issueId={issueId} />

        <div className="acb-issue-main-sections">
          <IssueCommentList
            loading={sideLoading}
            comments={comments}
            memberName={memberName}
            commentBody={commentBody}
            commentSending={commentSending}
            onCommentBodyChange={onCommentBodyChange}
            onSend={onSendComment}
            onRemove={onRemoveComment}
            workspaceId={workspaceId}
            issueId={issueId}
          />
          <IssueActivityFold
            loading={sideLoading}
            activities={activities}
            memberName={memberName}
            activitySentence={activitySentence}
            actorKindLabel={actorKindLabel}
            issueId={issueId}
          />
        </div>
      </div>
    </div>
  )
}
