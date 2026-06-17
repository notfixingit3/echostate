"use client"

import * as React from "react"
import { useTheme } from "next-themes"
import { MoonIcon, SunIcon, MonitorIcon } from "lucide-react"

import { Button } from "@/components/ui/button"

const themes = ["light", "dark", "system"] as const

export function ThemeToggle() {
  const { theme, setTheme, resolvedTheme } = useTheme()
  const [mounted, setMounted] = React.useState(false)

  React.useEffect(() => {
    setMounted(true)
  }, [])

  const cycleTheme = () => {
    const current = theme ?? "system"
    const index = themes.indexOf(current as (typeof themes)[number])
    const next = themes[(index + 1) % themes.length]
    setTheme(next)
  }

  const Icon =
    !mounted || theme === "system"
      ? MonitorIcon
      : resolvedTheme === "dark"
        ? MoonIcon
        : SunIcon

  const label =
    !mounted || theme === "system"
      ? "System theme"
      : resolvedTheme === "dark"
        ? "Dark mode"
        : "Light mode"

  return (
    <Button
      variant="outline"
      size="icon"
      onClick={cycleTheme}
      aria-label={`Theme: ${label}. Click to change.`}
      data-testid="theme-toggle"
      className="shrink-0"
    >
      <Icon className="size-4" />
    </Button>
  )
}