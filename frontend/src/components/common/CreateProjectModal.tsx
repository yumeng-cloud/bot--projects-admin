import { useEffect, useState } from 'react'
import { Form, Input, Modal, message } from 'antd'
import { createProject, type Project } from '../../api/projects'

type Props = {
  open: boolean
  workspaceId: string
  onClose: () => void
  onCreated: (project: Project) => void
}

/** 从名称抽 1–8 位大写字母数字；中文名等抽不出时回落 PRJ */
export function suggestKeyPrefix(name: string): string {
  const raw = name
    .trim()
    .toUpperCase()
    .replace(/[^A-Z0-9]/g, '')
    .slice(0, 8)
  return raw || 'PRJ'
}

export default function CreateProjectModal({ open, workspaceId, onClose, onCreated }: Props) {
  const [form] = Form.useForm<{ name: string; key_prefix: string; description?: string }>()
  const [busy, setBusy] = useState(false)
  const [prefixTouched, setPrefixTouched] = useState(false)

  useEffect(() => {
    if (!open) return
    setPrefixTouched(false)
    form.setFieldsValue({ name: '', key_prefix: 'PRJ', description: '' })
  }, [open, form])

  const submit = async () => {
    let values: { name: string; key_prefix: string; description?: string }
    try {
      values = await form.validateFields()
    } catch {
      return
    }
    const name = values.name.trim()
    const key_prefix = suggestKeyPrefix(values.key_prefix || name)
    if (!name) {
      message.warning('请填写项目名称')
      return
    }
    setBusy(true)
    try {
      const { project } = await createProject(workspaceId, {
        name,
        key_prefix,
        description: values.description?.trim() || undefined,
      })
      message.success(`已创建项目「${project.name}」`)
      onCreated(project)
      onClose()
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : '创建项目失败')
    } finally {
      setBusy(false)
    }
  }

  return (
    <Modal
      title="创建项目"
      open={open}
      onCancel={onClose}
      onOk={() => void submit()}
      okText="创建"
      cancelText="取消"
      confirmLoading={busy}
      destroyOnClose
      width={440}
    >
      <Form
        form={form}
        layout="vertical"
        requiredMark={false}
        style={{ marginTop: 8 }}
        onValuesChange={(changed) => {
          if ('name' in changed && !prefixTouched) {
            form.setFieldsValue({ key_prefix: suggestKeyPrefix(String(changed.name || '')) })
          }
        }}
      >
        <Form.Item
          name="name"
          label="项目名称"
          rules={[{ required: true, message: '请填写项目名称' }]}
        >
          <Input placeholder="例如：产品协作" maxLength={64} autoFocus onPressEnter={() => void submit()} />
        </Form.Item>
        <Form.Item
          name="key_prefix"
          label="编号前缀"
          extra="Issue 编号用，如 PRJ-1；1–8 位字母或数字"
          rules={[
            { required: true, message: '请填写前缀' },
            {
              pattern: /^[A-Za-z0-9]{1,8}$/,
              message: '须为 1–8 位字母或数字',
            },
          ]}
        >
          <Input
            placeholder="PRJ"
            maxLength={8}
            style={{ textTransform: 'uppercase' }}
            onChange={() => setPrefixTouched(true)}
          />
        </Form.Item>
        <Form.Item name="description" label="简介（可选）">
          <Input.TextArea rows={2} maxLength={200} placeholder="一句话说明这个项目" />
        </Form.Item>
      </Form>
    </Modal>
  )
}
