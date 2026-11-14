
import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from 'react'

import { api } from '../services/api'

interface User {
  id: string
  name: string
  email: string
}

interface AuthContextType {
  user: User | null
  loading: boolean
  login: (email: string, password: string) => Promise<void>
  register: (name: string, email: string, password: string) => Promise<void>
  logout: () => void
}

interface MeResponse {
  message: string
  userID: string
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

const TOKEN_KEY = 'mypoll_token'

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  // Restore the login session when the application loads
  useEffect(() => {
    const token = localStorage.getItem(TOKEN_KEY)

    if (!token) {
      setLoading(false)
      return
    }

    api
      .getMe()
      .then((response) => {
        const profile = response as unknown as MeResponse

        if (profile.userID) {
          setUser({
            id: profile.userID,
            name: 'User',
            email: '',
          })
        } else {
          throw new Error('Invalid authentication response')
        }
      })
      .catch(() => {
        localStorage.removeItem(TOKEN_KEY)
        setUser(null)
      })
      .finally(() => {
        setLoading(false)
      })
  }, [])

  // Login
  async function login(
    email: string,
    password: string
  ): Promise<void> {
    const response = await api.login(email, password)

    if (!response.token) {
      throw new Error('Login response did not contain a token')
    }

    if (!response.user) {
      throw new Error('Login response did not contain user details')
    }

    localStorage.setItem(TOKEN_KEY, response.token)

    setUser({
      id: response.user.id,
      name: response.user.name,
      email: response.user.email,
    })
  }

  // Register
 async function register(
  name: string,
  email: string,
  password: string,
): Promise<void> {
  await api.register(name, email, password)
}

  // Logout
  function logout(): void {
    localStorage.removeItem(TOKEN_KEY)
    setUser(null)
  }

  return (
    <AuthContext.Provider
      value={{
        user,
        loading,
        login,
        register,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext)

  if (!context) {
    throw new Error('useAuth must be used inside AuthProvider')
  }

  return context
}
