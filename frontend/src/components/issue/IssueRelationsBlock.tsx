import { Button, Popconfirm, Select, Space, Spin, Typography } from 'antd'
import { DeleteOutlined } from '@ant-design/icons'
import type { Issue, IssueRelation, RelationType } from '../../types/issue'
import { RELATION_TYPE_LABELS } from '../../types/issue'

type Props = {
  loading: boolean
  relations: IssueRelation[]
  relationChoices: Issue[]
  relType: RelationType
  relToId: string | undefined
  relAdding: boolean
  onRelTypeChange: (t: RelationType) => void
  onRelToIdChange: (id: string | undefined) => void
  onAdd: () => void
  onRemove: (rel: IssueRelation) => void
  otherIssueLabel: (rel: IssueRelation) => string
  otherIssueId: (rel: IssueRelation) => string
}

const RELATION_OPTIONS: { value: RelationType; label: string }[] = (
  Object.keys(RELATION_TYPE_LABELS) as RelationType[]
).map((k) => ({ value: k, label: RELATION_TYPE_LABELS[k] }))

function issueOptionLabel(i: Pick<Issue, 'identifier' | 'title'>): string {
  return `${i.identifier} · ${i.title}`
}

export function IssueRelationsBlock({
  loading,
  relations,
  relationChoices,
  relType,
  relToId,
  relAdding,
  onRelTypeChange,
  onRelToIdChange,
  onAdd,
  onRemove,
  otherIssueLabel,
  otherIssueId,
}: Props) {
  return (
    <div className="acb-props-rail-section">
      <div className="acb-props-rail-section-title">
        关联{relations.length > 0 ? ` (${relations.length})` : ''}
      </div>
      <Spin spinning={loading} size="small">
        {relations.length === 0 && !loading ? (
          <div className="acb-issue-muted-empty" style={{ padding: '2px 2px 8px' }}>
            暂无
          </div>
        ) : (
          <div className="acb-rel-list">
            {relations.map((rel) => (
              <div key={rel.id} className="acb-rel-item">
                <Typography.Text type="secondary" style={{ flexShrink: 0, fontSize: 11 }}>
                  {RELATION_TYPE_LABELS[rel.relation_type as RelationType] || rel.relation_type}
                </Typography.Text>
                <Typography.Text ellipsis title={otherIssueId(rel)} style={{ fontSize: 12 }}>
                  {otherIssueLabel(rel)}
                </Typography.Text>
                <Popconfirm
                  title="删除此关联？"
                  okText="删除"
                  cancelText="取消"
                  okButtonProps={{ danger: true, size: 'small' }}
                  onConfirm={() => onRemove(rel)}
                >
                  <Button
                    type="text"
                    size="small"
                    danger
                    icon={<DeleteOutlined />}
                    style={{ flexShrink: 0, width: 22, height: 22, padding: 0 }}
                  />
                </Popconfirm>
              </div>
            ))}
          </div>
        )}
        <Space.Compact style={{ width: '100%' }} size="small">
          <Select
            size="small"
            value={relType}
            onChange={onRelTypeChange}
            options={RELATION_OPTIONS}
            style={{ width: 72 }}
            popupMatchSelectWidth={120}
          />
          <Select
            size="small"
            showSearch
            allowClear
            placeholder="Issue"
            value={relToId}
            onChange={(v) => onRelToIdChange(v)}
            style={{ flex: 1, minWidth: 0 }}
            optionFilterProp="label"
            options={relationChoices.map((i) => ({
              value: i.id,
              label: issueOptionLabel(i),
            }))}
          />
          <Button size="small" type="primary" loading={relAdding} onClick={onAdd}>
            +
          </Button>
        </Space.Compact>
      </Spin>
    </div>
  )
}
