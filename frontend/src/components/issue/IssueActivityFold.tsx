import { Collapse, Spin, Timeline, Typography } from 'antd'
import type { Activity } from '../../types/issue'
import { formatShanghai } from '../../utils/time'

type Props = {
  loading: boolean
  activities: Activity[]
  memberName: (id: string | null | undefined) => string | null
  activitySentence: (action: string, payload: string | null | undefined) => string
  actorKindLabel: (kind: string) => string
}

/** 活动时间线 — 默认折叠，不占首屏（ISSUE_DETAIL_LAYOUT A5） */
export function IssueActivityFold({
  loading,
  activities,
  memberName,
  activitySentence,
  actorKindLabel,
}: Props) {
  const sorted = [...activities].sort((a, b) => {
    const t = b.created_at.localeCompare(a.created_at)
    if (t !== 0) return t
    return b.id.localeCompare(a.id)
  })

  return (
    <div className="acb-issue-activity">
      <Collapse
        ghost
        size="small"
        defaultActiveKey={[]}
        items={[
          {
            key: 'activity',
            label: `活动${sorted.length > 0 ? ` (${sorted.length})` : ''}`,
            children: (
              <Spin spinning={loading} size="small">
                {sorted.length === 0 && !loading ? (
                  <div className="acb-issue-muted-empty">暂无活动</div>
                ) : (
                  <Timeline
                    className="acb-activity-timeline"
                    style={{ marginTop: 4, paddingLeft: 2 }}
                    items={sorted.map((a) => {
                      const actor = a.actor_id ? memberName(a.actor_id) : null
                      const who =
                        actor ||
                        (a.actor_id
                          ? a.actor_kind === 'bot' || a.actor_kind === 'system'
                            ? actorKindLabel(a.actor_kind)
                            : '有人'
                          : '有人')
                      const verb = activitySentence(a.action, a.payload)
                      return {
                        key: a.id,
                        color: 'gray' as const,
                        children: (
                          <div style={{ paddingBottom: 4 }}>
                            <div
                              style={{
                                fontSize: 13,
                                lineHeight: 1.45,
                                color: 'var(--acb-text)',
                              }}
                            >
                              {`${who} ${verb}`}
                            </div>
                            <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                              {formatShanghai(a.created_at)}
                            </Typography.Text>
                          </div>
                        ),
                      }
                    })}
                  />
                )}
              </Spin>
            ),
          },
        ]}
      />
    </div>
  )
}
