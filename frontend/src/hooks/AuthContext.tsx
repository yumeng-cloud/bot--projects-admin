import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react'
import type { User } from '../types'
import * as authApi from '../api/auth'

type AuthCtx = {
  token: string | null
  user: User | null
  loading: boolean
  authed: boolean
  signOut: () => Promise<void>
  refresh: () => Promise<void>
}

const Ctx = createContext<AuthCtx | null>(null)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [token, setToken] = useState<string | null>(() => localStorage.getItem('token'))
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(!!localStorage.getItem('token'))

  const refresh = useCallback(async () => {
    const t = localStorage.getItem('token')
    setToken(t)
    if (!t) {
      setUser(null)
      setLoading(false)
      return
    }
    setLoading(true)
    try {
      const res = await authApi.me()
      setUser(res.user)
    } catch {
      localStorage.removeItem('token')
      setToken(null)
      setUser(null)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void refresh()
  }, [refresh])

  useEffect(() => {
    const sync = () => {
      const t = localStorage.getItem('token')
      setToken(t)
      if (t) void refresh()
      else {
        setUser(null)
        setLoading(false)
      }
    }
    window.addEventListener('storage', sync)
    window.addEventListener('auth-changed', sync)
    return () => {
      window.removeEventListener('storage', sync)
      window.removeEventListener('auth-changed', sync)
    }
  }, [refresh])

  const signOut = useCallback(async () => {
    await authApi.logout()
    localStorage.removeItem('token')
    localStorage.removeItem('workspaceId')
    setToken(null)
    setUser(null)
  }, [])

  const value = useMemo(
    () => ({ token, user, loading, authed: !!token, signOut, refresh }),
    [token, user, loading, signOut, refresh],
  )

  return <Ctx.Provider value={value}>{children}</Ctx.Provider>
}

export function useAuth() {
  const v = useContext(Ctx)
  if (!v) throw new Error('useAuth outside AuthProvider')
  return v
}
