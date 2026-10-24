import { Result } from 'antd'

export default function ModulePlaceholder({ title, hint }: { title: string; hint?: string }) {
  return (
    <Result
      status="info"
      title={title}
      subTitle={hint || '模块由对应负责人接入，此处为路由占位。'}
    />
  )
}
