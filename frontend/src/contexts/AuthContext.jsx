import { createContext, useContext, useState, useEffect, useCallback } from 'react'
import { authService } from '../services/authService'
import { TOKEN_KEY } from '../services/api'

const AuthContext = createContext(null)

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null)
  const [loading, setLoading] = useState(true)

  // On mount: restore session from stored token by fetching the current user.
  useEffect(() => {
    const token = localStorage.getItem(TOKEN_KEY)
    if (!token) {
      setLoading(false)
      return
    }

    authService
      .getMe()
      .then((res) => setUser(res.data.data))
      .catch(() => {
        localStorage.removeItem(TOKEN_KEY)
        setUser(null)
      })
      .finally(() => setLoading(false))
  }, [])

  const login = useCallback(async (credentials) => {
    const res = await authService.login(credentials)
    const { accessToken, user } = res.data.data
    localStorage.setItem(TOKEN_KEY, accessToken)
    setUser(user)
    return user
  }, [])

  const register = useCallback(async (data) => {
    const res = await authService.register(data)
    const { accessToken, user } = res.data.data
    localStorage.setItem(TOKEN_KEY, accessToken)
    setUser(user)
    return user
  }, [])

  const agentJoin = useCallback(async (data) => {
    const res = await authService.agentJoin(data)
    const { accessToken, user } = res.data.data
    localStorage.setItem(TOKEN_KEY, accessToken)
    setUser(user)
    return user
  }, [])

  const logout = useCallback(async () => {
    try {
      await authService.logout()
    } finally {
      localStorage.removeItem(TOKEN_KEY)
      setUser(null)
    }
  }, [])

  return (
    <AuthContext.Provider value={{ user, loading, login, register, agentJoin, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used inside <AuthProvider>')
  return ctx
}
