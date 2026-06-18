"use client"

import * as React from "react"

import { Badge } from "@/components/ui/badge"
import { fetchApi } from "@/lib/api"
import { APP_VERSION } from "@/lib/version"
import { cn } from "@/lib/utils"

type VersionInfo = {
  current: string
  branch: string
  latest?: string
  update_available: boolean
  checked_at?: string
  check_error?: string
}

type VersionStatus = "current" | "update" | "unknown"

function statusFromInfo(info: VersionInfo | null): VersionStatus {
  if (!info || info.check_error) {
    return "unknown"
  }
  if (info.update_available) {
    return "update"
  }
  return "current"
}

function statusTitle(info: VersionInfo | null, status: VersionStatus): string {
  if (!info) {
    return `v${APP_VERSION} — checking for updates…`
  }
  if (info.check_error) {
    return `v${APP_VERSION} (${info.branch}) — update check unavailable`
  }
  if (status === "update" && info.latest) {
    return `v${APP_VERSION} (${info.branch}) — v${info.latest} is available`
  }
  if (info.latest) {
    return `v${APP_VERSION} (${info.branch}) — up to date`
  }
  return `v${APP_VERSION} (${info.branch})`
}

const statusStyles: Record<VersionStatus, string> = {
  current: "bg-emerald-500 shadow-[0_0_6px_rgba(16,185,129,0.55)]",
  update: "bg-amber-500 shadow-[0_0_6px_rgba(245,158,11,0.55)]",
  unknown: "bg-muted-foreground/40",
}

export function VersionBadge() {
  const [info, setInfo] = React.useState<VersionInfo | null>(null)

  React.useEffect(() => {
    let cancelled = false

    void fetchApi<VersionInfo>("/api/version")
      .then((result) => {
        if (!cancelled) {
          setInfo(result)
        }
      })
      .catch(() => {
        if (!cancelled) {
          setInfo(null)
        }
      })

    return () => {
      cancelled = true
    }
  }, [])

  const status = statusFromInfo(info)

  return (
    <div className="inline-flex items-center gap-2">
      <span
        className={cn("size-2 shrink-0 rounded-full", statusStyles[status])}
        data-testid="footer-version-status"
        data-status={status}
        title={statusTitle(info, status)}
        aria-label={statusTitle(info, status)}
        role="status"
      />
      <Badge
        variant="outline"
        className="font-mono text-xs"
        data-testid="footer-version"
      >
        v{APP_VERSION}
      </Badge>
    </div>
  )
}