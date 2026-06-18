"use client"

import * as React from "react"
import { useTheme } from "next-themes"

import { useAuth } from "@/components/auth-provider"
import type { UserTheme } from "@/lib/auth"

const themes: UserTheme[] = ["light", "dark", "system"]

function isUserTheme(value: string | undefined): value is UserTheme {
  return !!value && themes.includes(value as UserTheme)
}

export function UserThemeSync() {
  const { user } = useAuth()
  const { setTheme } = useTheme()

  React.useEffect(() => {
    if (user?.theme && isUserTheme(user.theme)) {
      setTheme(user.theme)
    }
  }, [user?.theme, setTheme])

  return null
}