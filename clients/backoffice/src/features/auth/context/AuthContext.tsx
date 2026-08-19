import { createContext, useState, useEffect, useRef, useCallback } from 'react'
import type { ReactNode } from 'react'
import {
  authApi,
  resolveUserPermissions,
  resolveBranchContexts,
  type User,
  type BranchContext,
} from '../api/authApi'

const ACTIVE_CONTEXT_KEY = 'backoffice.activeContext'

type AuthContextType = {
  user: User | null
  isLoading: boolean
  contexts: BranchContext[]
  activeContext: BranchContext | null
  login: (username: string, password: string) => Promise<void>
  logout: () => Promise<void>
  selectContext: (context: BranchContext) => Promise<void>
}

export const AuthContext = createContext<AuthContextType | undefined>(undefined)

function readStoredContext(contexts: BranchContext[]): BranchContext | null {
  const raw = localStorage.getItem(ACTIVE_CONTEXT_KEY)
  if (!raw) return null
  try {
    const parsed = JSON.parse(raw) as BranchContext
    return (
      contexts.find(
        (c) => c.merchant.id === parsed.merchant.id && c.branch.id === parsed.branch.id,
      ) ?? null
    )
  } catch {
    return null
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [contexts, setContexts] = useState<BranchContext[]>([])
  const [activeContext, setActiveContext] = useState<BranchContext | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const resolvedManually = useRef(false) // tracks if login() already settled auth state

  const settleUser = useCallback(async (u: User) => {
    u.permissions = await resolveUserPermissions(u)
    const resolved = await resolveBranchContexts(u)
    setUser(u)
    setContexts(resolved)
    const stored = readStoredContext(resolved)
    if (stored) {
      setActiveContext(stored)
    } else if (resolved.length === 1) {
      setActiveContext(resolved[0])
    } else {
      setActiveContext(null)
    }
  }, [])

  /** Apply the selected context to the backend so the access token carries the
   *  matching mid/bid claims. Used when a stored context must be restored after
   *  a fresh login (where the session starts on the default role). */
  const applyContextToBackend = useCallback(async (context: BranchContext | null) => {
    if (!context) return
    try {
      await authApi.switchContext(context.userRoleId)
      localStorage.setItem(ACTIVE_CONTEXT_KEY, JSON.stringify(context))
    } catch {
      // ignore — context is retained locally; next explicit switch will sync it
    }
  }, [])

  useEffect(() => {
    authApi
      .me()
      .then(async (u) => {
        if (!resolvedManually.current) {
          await settleUser(u)
        }
      })
      .catch(() => {
        if (!resolvedManually.current) {
          setUser(null)
          setContexts([])
          setActiveContext(null)
        }
      })
      .finally(() => {
        if (!resolvedManually.current) setIsLoading(false)
      })
  }, [settleUser])

  const login = async (username: string, password: string) => {
    const res = await authApi.login(username, password)
    resolvedManually.current = true
    const stored = readStoredContext(await resolveBranchContexts(res.user))
    await settleUser(res.user)
    if (stored) {
      await applyContextToBackend(stored)
    }
    setIsLoading(false)
  }

  const logout = async () => {
    await authApi.logout()
    localStorage.removeItem(ACTIVE_CONTEXT_KEY)
    setUser(null)
    setContexts([])
    setActiveContext(null)
  }

  const selectContext = async (context: BranchContext) => {
    try {
      await authApi.switchContext(context.userRoleId)
    } catch {
      // staff roles may not map to a scoped IAM role — keep the context locally
    }
    localStorage.setItem(ACTIVE_CONTEXT_KEY, JSON.stringify(context))
    setActiveContext(context)
  }

  return (
    <AuthContext.Provider
      value={{ user, isLoading, contexts, activeContext, login, logout, selectContext }}
    >
      {children}
    </AuthContext.Provider>
  )
}
