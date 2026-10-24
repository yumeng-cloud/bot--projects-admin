import { Button, DatePicker, Form, InputNumber, Select } from 'antd'
import type { FormInstance } from 'antd/es/form'
import type {
  Cycle,
  Issue,
  IssuePriority,
  IssueRelation,
  IssueStatus,
  Label,
  MemberRef,
  RelationType,
} from '../../types/issue'
import {
  ISSUE_STATUSES,
  PRIORITY_LABELS,
  STATUS_LABELS,
  memberLabel,
} from '../../types/issue'
import { IssuePropertyRow } from './IssuePropertyRow'
import { IssueRelationsBlock } from './IssueRelationsBlock'

type Props = {
  form: FormInstance
  quickSaving: boolean
  members: MemberRef[]
  labels: Label[]
  cycles: Cycle[]
  onStatusChange: (s: IssueStatus) => void
  onPriorityChange: (p: IssuePriority) => void
  /** R1: 其余属性即时 PATCH（可选，与 status/priority 同路径） */
  onAssigneeChange?: (id: string | null) => void
  onCycleChange?: (id: string | null) => void
  onEstimateChange?: (v: number | null) => void
  onLabelsChange?: (ids: string[]) => void
  onStartAtChange?: (v: unknown) => void
  onDueAtChange?: (v: unknown) => void
  // relations
  sideLoading: boolean
  relations: IssueRelation[]
  relationChoices: Issue[]
  relType: RelationType
  relToId: string | undefined
  relAdding: boolean
  onRelTypeChange: (t: RelationType) => void
  onRelToIdChange: (id: string | undefined) => void
  onAddRelation: () => void
  onRemoveRelation: (rel: IssueRelation) => void
  otherIssueLabel: (rel: IssueRelation) => string
  otherIssueId: (rel: IssueRelation) => string
  /** C4：已归档只读态 */
  archivedAt?: string | null
  /** WHEEL B2：取消归档 */
  unarchiving?: boolean
  onUnarchive?: () => void
}

const statusOptions = ISSUE_STATUSES.map((s) => ({ value: s, label: STATUS_LABELS[s] }))
const priorityOptions = (Object.keys(PRIORITY_LABELS) as IssuePriority[]).map((p) => ({
  value: p,
  label: PRIORITY_LABELS[p],
}))

/** Property 轨：状态…关联，一行一项 label|value（ISSUE_DETAIL_LAYOUT §2–§4） */
export function IssuePropertyRail({
  form,
  quickSaving,
  members,
  labels,
  cycles,
  onStatusChange,
  onPriorityChange,
  onAssigneeChange,
  onCycleChange,
  onEstimateChange,
  onLabelsChange,
  onStartAtChange,
  onDueAtChange,
  sideLoading,
  relations,
  relationChoices,
  relType,
  relToId,
  relAdding,
  onRelTypeChange,
  onRelToIdChange,
  onAddRelation,
  onRemoveRelation,
  otherIssueLabel,
  otherIssueId,
  archivedAt,
  unarchiving,
  onUnarchive,
}: Props) {
  const archived =
    archivedAt != null && String(archivedAt).trim() !== ''
  return (
    <aside className="acb-props-rail acb-property-rail">
      <Form form={form} size="small" requiredMark={false} colon={false} component="div">
        {archived ? (
          <IssuePropertyRow label="归档">
            <div className="acb-property-archive">
              <span className="acb-property-archive__state">已归档</span>
              {onUnarchive ? (
                <Button
                  size="small"
                  type="link"
                  className="acb-property-archive__action"
                  loading={unarchiving}
                  onClick={() => onUnarchive()}
                >
                  取消归档
                </Button>
              ) : null}
            </div>
          </IssuePropertyRow>
        ) : null}
        <IssuePropertyRow label="状态">
          <Form.Item name="status" noStyle>
            <Select
              variant="borderless"
              size="small"
              disabled={quickSaving}
              options={statusOptions}
              onChange={(v) => onStatusChange(v as IssueStatus)}
              popupMatchSelectWidth={false}
            />
          </Form.Item>
        </IssuePropertyRow>

        <IssuePropertyRow label="优先级">
          <Form.Item name="priority" noStyle>
            <Select
              variant="borderless"
              size="small"
              disabled={quickSaving}
              options={priorityOptions}
              onChange={(v) => onPriorityChange(v as IssuePriority)}
              popupMatchSelectWidth={false}
            />
          </Form.Item>
        </IssuePropertyRow>

        <IssuePropertyRow label="负责人">
          <Form.Item name="assignee_id" noStyle>
            <Select
              variant="borderless"
              size="small"
              allowClear
              disabled={quickSaving}
              options={members.map((m) => ({ value: m.id, label: memberLabel(m) }))}
              placeholder="未指派"
              popupMatchSelectWidth={false}
              onChange={(v) => onAssigneeChange?.(v ?? null)}
            />
          </Form.Item>
        </IssuePropertyRow>

        <IssuePropertyRow label="Cycle">
          <Form.Item name="cycle_id" noStyle>
            <Select
              variant="borderless"
              size="small"
              allowClear
              disabled={quickSaving}
              options={cycles.map((c) => ({ value: c.id, label: c.name }))}
              placeholder="未划入"
              popupMatchSelectWidth={false}
              onChange={(v) => onCycleChange?.(v ?? null)}
            />
          </Form.Item>
        </IssuePropertyRow>

        <IssuePropertyRow label="开始">
          <Form.Item name="start_at" noStyle>
            <DatePicker
              variant="borderless"
              size="small"
              allowClear
              format="YYYY-MM-DD"
              placeholder="未设"
              style={{ width: '100%' }}
              disabled={quickSaving}
              onChange={(v) => onStartAtChange?.(v)}
            />
          </Form.Item>
        </IssuePropertyRow>

        <IssuePropertyRow label="截止">
          <Form.Item name="due_at" noStyle>
            <DatePicker
              variant="borderless"
              size="small"
              allowClear
              format="YYYY-MM-DD"
              placeholder="未设"
              style={{ width: '100%' }}
              disabled={quickSaving}
              onChange={(v) => onDueAtChange?.(v)}
            />
          </Form.Item>
        </IssuePropertyRow>

        <IssuePropertyRow label="估时">
          <Form.Item name="estimate" noStyle>
            <InputNumber
              variant="borderless"
              size="small"
              min={0}
              placeholder="—"
              style={{ width: '100%' }}
              disabled={quickSaving}
              addonAfter={<span className="acb-property-suffix">pt</span>}
              onChange={(v) => onEstimateChange?.(typeof v === 'number' ? v : null)}
            />
          </Form.Item>
        </IssuePropertyRow>

        <IssuePropertyRow label="标签">
          <Form.Item name="label_ids" noStyle>
            <Select
              variant="borderless"
              size="small"
              mode="multiple"
              maxTagCount="responsive"
              disabled={quickSaving}
              options={labels.map((l) => ({ value: l.id, label: l.name }))}
              placeholder="添加标签"
              popupMatchSelectWidth={false}
              onChange={(v) => onLabelsChange?.(v as string[])}
            />
          </Form.Item>
        </IssuePropertyRow>
      </Form>

      <IssueRelationsBlock
        loading={sideLoading}
        relations={relations}
        relationChoices={relationChoices}
        relType={relType}
        relToId={relToId}
        relAdding={relAdding}
        onRelTypeChange={onRelTypeChange}
        onRelToIdChange={onRelToIdChange}
        onAdd={onAddRelation}
        onRemove={onRemoveRelation}
        otherIssueLabel={otherIssueLabel}
        otherIssueId={otherIssueId}
      />
    </aside>
  )
}
