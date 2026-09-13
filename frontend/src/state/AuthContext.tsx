import { createContext, ReactNode, useContext, useEffect, useMemo, useState } from 'react'
import { CurrentUserResponse, fetchCurrentUser, logout as requestLogout } from '../api/auth'
import { currentUser } from '../config/currentUser'

type AuthState = {
  user: CurrentUserResponse | null
  isLoading: boolean
  error: string
  login: () => void
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthState | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<CurrentUserResponse | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    let cancelled = false
    const load = async () => {
      try {
        const currentUser = await fetchCurrentUser()
        if (!cancelled) setUser(currentUser ?? localDevelopmentUser())
      } catch {
        if (!cancelled) {
          const fallbackUser = localDevelopmentUser()
          if (fallbackUser) {
            setUser(fallbackUser)
          } else {
            setError('ログイン状態の確認に失敗しました。')
          }
        }
      } finally {
        if (!cancelled) setIsLoading(false)
      }
    }
    void load()
    return () => {
      cancelled = true
    }
  }, [])

  const value = useMemo<AuthState>(
    () => ({
      user,
      isLoading,
      error,
      login: () => {
        window.location.href = '/auth/login'
      },
      logout: async () => {
        await requestLogout()
        setUser(null)
      },
    }),
    [user, isLoading, error],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

function localDevelopmentUser(): CurrentUserResponse | null {
  if (!import.meta.env.DEV) return null
  return {
    id: currentUser.id,
    handle: currentUser.handle.replace(/^@/, ''),
    displayName: currentUser.displayName,
    bio: currentUser.bio,
    createdAt: new Date().toISOString(),
  }
}

export function useAuth() {
  const value = useContext(AuthContext)
  if (!value) throw new Error('AuthProvider is required')
  return value
}

export function useOptionalAuth() {
  return useContext(AuthContext)
}
