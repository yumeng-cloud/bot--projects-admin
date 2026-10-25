import { useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { Button, Card, Form, Input, Typography, message, Tabs } from 'antd'
import { login as apiLogin, register as apiRegister } from '../api/auth'
import './login.css'

function safeReturnPath(from: unknown): string {
  if (typeof from !== 'string' || !from.startsWith('/') || from.startsWith('//')) return '/'
  // Stay on login / landing only if that was the intent; default home.
  if (from === '/login') return '/'
  return from
}

export default function LoginPage({ onAuth }: { onAuth: () => void }) {
  const [loading, setLoading] = useState(false)
  const location = useLocation()
  const nav = useNavigate()
  const stateFrom = (location.state as { from?: unknown } | null)?.from

  const finishAuth = () => {
    onAuth()
    // Only bounce when this page owns /login. Embedded in /join keeps JoinPage flow.
    if (location.pathname === '/login') {
      nav(safeReturnPath(stateFrom), { replace: true })
    }
  }

  const login = async (v: { email: string; password: string }) => {
    setLoading(true)
    try {
      await apiLogin(v)
      message.success('登录成功')
      finishAuth()
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
      finishAuth()
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : '注册失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="acb-login">
      <Card className="acb-login-card" bordered>
        <div className="acb-login-brand">
          <Typography.Title level={3} className="acb-login-title">
            育蒙协作看板
          </Typography.Title>
          <Typography.Paragraph className="acb-login-subtitle">
            人与 Bot 同场协作
          </Typography.Paragraph>
        </div>
        <Tabs
          items={[
            {
              key: 'login',
              label: '登录',
              children: (
                <Form layout="vertical" onFinish={login} initialValues={{ email: 'admin@yumeng.tech' }}>
                  <Form.Item name="email" label="邮箱" rules={[{ required: true, type: 'email' }]}>
                    <Input placeholder="请输入邮箱" autoComplete="username" />
                  </Form.Item>
                  <Form.Item name="password" label="密码" rules={[{ required: true }]}>
                    <Input.Password placeholder="请输入密码" autoComplete="current-password" />
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
                    <Input placeholder="可选" autoComplete="name" />
                  </Form.Item>
                  <Form.Item name="email" label="邮箱" rules={[{ required: true, type: 'email' }]}>
                    <Input placeholder="请输入邮箱" autoComplete="username" />
                  </Form.Item>
                  <Form.Item name="password" label="密码" rules={[{ required: true, min: 8 }]}>
                    <Input.Password placeholder="至少 8 位" autoComplete="new-password" />
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
