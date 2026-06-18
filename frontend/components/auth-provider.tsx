"use client"

import * as React from "react"
import { usePathname, useRouter } from "next/navigation"

import {
  fetchAuthConfig,
  fetchAuthSession,
  logout,
  storeUserId,
  type AuthConfig,
  type AuthUser,
} from "@/lib/auth"

type AuthContextValue = {
  loading: boolean
  config: AuthConfig | null
  user: AuthUser | null
  refresh: () => Promise<void>
  signOut: () => Promise<void>
  setUser: (user: AuthUser) => void
}

const AuthContext = React.createContext<AuthContextValue | null>(null)

const PUBLIC_PATHS = new Set(["/login"])

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const pathname = usePathname()
  const router = useRouter()
  const [loading, setLoading] = React.useState(true)
  const [config, setConfig] = React.useState<AuthConfig | null>(null)
  const [user, setUserState] = React.useState<AuthUser | null>(null)

  const refresh = React.useCallback(async () => {
    const [nextConfig, session] = await Promise.all([
      fetchAuthConfig(),
      fetchAuthSession(),
    ])
    setConfig(nextConfig)
    if (session.authenticated && session.user) {
      setUserState(session.user)
      storeUserId(session.user.id)
    } else {
      setUserState(null)
    }
  }, [])

  React.useEffect(() => {
    let active = true
    ;(async () => {
      try {
        await refresh()
      } catch {
        if (active) {
          setConfig(null)
          setUserState(null)
        }
      } finally {
        if (active) setLoading(false)
      }
    })()
    return () => {
      active = false
    }
  }, [refresh])

  React.useEffect(() => {
    if (loading || !config?.auth_required) return
    if (user) return
    if (PUBLIC_PATHS.has(pathname)) return
    router.replace("/login")
  }, [loading, config, user, pathname, router])

  const setUser = React.useCallback((next: AuthUser) => {
    setUserState(next)
    storeUserId(next.id)
  }, [])

  const signOut = React.useCallback(async () => {
    try {
      await logout()
    } finally {
      setUserState(null)
      router.replace("/login")
    }
  }, [router])

  return (
    <AuthContext.Provider value={{ loading, config, user, refresh, signOut, setUser }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = React.useContext(AuthContext)
  if (!ctx) {
    throw new Error("useAuth must be used within AuthProvider")
  }
  return ctx
}