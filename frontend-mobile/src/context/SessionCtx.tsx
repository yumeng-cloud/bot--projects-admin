import { createContext, useContext, type ReactNode } from 'react'
import type { SessionContext, SessionContextPut, User } from '@acb/api-client'

type Value = {
  user: User
  session: SessionContext
  hydrateNote: string | null
  setSession: (ctx: SessionContext) => void
  saveSession: (patch: SessionContextPut) => Promise<SessionContext | null>
  onLogout: () => void
  onRehydrate: () => Promise<SessionContext>
}

const Ctx = createContext<Value | null>(null)

export function SessionProvider({
  value,
  children,
}: {
  value: Value
  children: ReactNode
}) {
  return <Ctx.Provider value={value}>{children}</Ctx.Provider>
}

export function useSession() {
  const v = useContext(Ctx)
  if (!v) throw new Error('useSession outside provider')
  return v
}
