import { useState } from 'react'
import { Button, Card, Form, Input, Typography, message, Tabs } from 'antd'
import { login as apiLogin, register as apiRegister } from '../api/auth'

export default function LoginPage({ onAuth }: { onAuth: () => void }) {
  const [loading, setLoading] = useState(false)

  const login = async (v: { email: string; password: string }) => {
    setLoading(true)
    try {
      await apiLogin(v)
      message.success('登录成功')
      onAuth()
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : '登录失败')
    } finally {
      setLoading(false)
    }
  }

  const register = async (v: { email: string; password: string; name?: string }) => {
    setLoading(true)
    try {
      await apiRegister(v)
      message.success('注册成功')
      onAuth()
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : '注册失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ minHeight: '100%', display: 'grid', placeItems: 'center', padding: 24 }}>
      <Card style={{ width: 400, boxShadow: '0 1px 2px rgba(0, 0, 0, 0.04)' }}>
        <Typography.Title level={3} style={{ textAlign: 'center', marginBottom: 4 }}>
          AI 协作看板
        </Typography.Title>
        <Typography.Paragraph type="secondary" style={{ textAlign: 'center' }}>
          育蒙科技 · Linear-like + Grok Bot
        </Typography.Paragraph>
        <Tabs
          items={[
            {
              key: 'login',
              label: '登录',
              children: (
                <Form layout="vertical" onFinish={login} initialValues={{ email: 'admin@yumeng.tech' }}>
                  <Form.Item name="email" label="邮箱" rules={[{ required: true, type: 'email' }]}>
                    <Input placeholder="admin@yumeng.tech" />
                  </Form.Item>
                  <Form.Item name="password" label="密码" rules={[{ required: true }]}>
                    <Input.Password placeholder="密码" />
                  </Form.Item>
                  <Button type="primary" htmlType="submit" block loading={loading}>
                    登录
                  </Button>
                </Form>
              ),
            },
            {
              key: 'register',
              label: '注册',
              children: (
                <Form layout="vertical" onFinish={register}>
                  <Form.Item name="name" label="姓名">
                    <Input />
                  </Form.Item>
                  <Form.Item name="email" label="邮箱" rules={[{ required: true, type: 'email' }]}>
                    <Input />
                  </Form.Item>
                  <Form.Item name="password" label="密码" rules={[{ required: true, min: 8 }]}>
                    <Input.Password />
                  </Form.Item>
                  <Button type="primary" htmlType="submit" block loading={loading}>
                    注册
                  </Button>
                </Form>
              ),
            },
          ]}
        />
      </Card>
    </div>
  )
}
