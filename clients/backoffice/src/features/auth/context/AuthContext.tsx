import { createContext, useState, useEffect, useRef } from 'react'
import type { ReactNode } from 'react'
import { authApi, type User } from '../api/authApi'

type AuthContextType = {
  user: User | null
  isLoading: boolean
  login: (username: string, password: string) => Promise<void>
  logout: () => Promise<void>
}

export const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const resolvedManually = useRef(false) // tracks if login() already settled auth state

  useEffect(() => {
    authApi
      .me()
      .then((u) => {
        if (!resolvedManually.current) setUser(u)
      })
      .catch(() => {
        if (!resolvedManually.current) setUser(null)
      })
      .finally(() => {
        if (!resolvedManually.current) setIsLoading(false)
      })
  }, [])

  const login = async (username: string, password: string) => {
    const res = await authApi.login(username, password)
    resolvedManually.current = true // stops the stale mount-check from overwriting this
    setUser(res.user)
    setIsLoading(false)
  }

  const logout = async () => {
    await authApi.logout()
    setUser(null)
  }

  return (
    <AuthContext.Provider value={{ user, isLoading, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}