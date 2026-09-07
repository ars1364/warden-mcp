import { createContext, useCallback, useContext, useEffect, useState, type ReactNode } from 'react'
import { api, getToken, setToken, setUnauthorizedHandler } from '../lib/api'
import type { Me } from '../types/api'

interface AuthContextValue {
  me: Me | null
  loading: boolean
  isAuthenticated: boolean
  login: (token: string) => Promise<void>
  logout: () => void
  refreshMe: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [me, setMe] = useState<Me | null>(null)
  const [loading, setLoading] = useState(true)

  const logout = useCallback(() => {
    setToken(null)
    setMe(null)
  }, [])

  const refreshMe = useCallback(async () => {
    if (!getToken()) {
      setMe(null)
      return
    }
    try {
      const result = await api.get<Me>('/api/me')
      setMe(result)
    } catch {
      setMe(null)
      setToken(null)
    }
  }, [])

  const login = useCallback(async (token: string) => {
    setToken(token)
    await refreshMe()
  }, [refreshMe])

  useEffect(() => {
    setUnauthorizedHandler(() => {
      setMe(null)
    })
    refreshMe().finally(() => setLoading(false))
  }, [refreshMe])

  const value: AuthContextValue = {
    me,
    loading,
    isAuthenticated: !!getToken() && !!me,
    login,
    logout,
    refreshMe,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
