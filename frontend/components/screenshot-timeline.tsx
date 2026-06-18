"use client"

import * as React from "react"
import Link from "next/link"

import { Badge } from "@/components/ui/badge"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { fetchApi } from "@/lib/api"
import type { PaginatedResponse, ScreenshotEntry } from "@/lib/types"
import { CameraIcon, ExternalLinkIcon } from "lucide-react"

export function ScreenshotTimeline({
  targetId,
  initialEntries,
}: {
  targetId?: string
  initialEntries?: ScreenshotEntry[]
}) {
  const [entries, setEntries] = React.useState<ScreenshotEntry[]>(
    initialEntries ?? []
  )
  const [loading, setLoading] = React.useState(!initialEntries && !!targetId)
  const [error, setError] = React.useState<string | null>(null)

  React.useEffect(() => {
    if (!targetId || initialEntries) return

    setLoading(true)
    setError(null)
    void fetchApi<PaginatedResponse<ScreenshotEntry>>(
      `/api/targets/${targetId}/screenshots?limit=20`
    )
      .then((data) => setEntries(data.data ?? []))
      .catch((err) => {
        setEntries([])
        setError(err instanceof Error ? err.message : "Failed to load screenshots")
      })
      .finally(() => setLoading(false))
  }, [targetId, initialEntries])

  if (loading) {
    return (
      <p className="text-sm text-muted-foreground">Loading screenshot timeline…</p>
    )
  }

  if (error) {
    return <p className="text-sm text-destructive">{error}</p>
  }

  if (entries.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        No screenshots yet. Re-scan with browserless Chrome enabled to capture page thumbnails.
      </p>
    )
  }

  return (
    <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
      {entries.map((entry) => (
        <ScreenshotCard key={entry.snapshot_id} entry={entry} />
      ))}
    </div>
  )
}

function ScreenshotCard({ entry }: { entry: ScreenshotEntry }) {
  const scannedAt = new Date(entry.scanned_at)
  const dateLabel = Number.isNaN(scannedAt.getTime())
    ? entry.scanned_at
    : scannedAt.toLocaleString()

  return (
    <Card className="overflow-hidden border-border/60" data-testid={`screenshot-${entry.snapshot_id}`}>
      <CardHeader className="space-y-1 pb-3">
        <CardTitle className="flex items-center gap-2 text-sm">
          <CameraIcon className="size-4 text-primary" />
          {dateLabel}
        </CardTitle>
        {entry.url ? (
          <CardDescription className="truncate font-mono text-xs">
            {entry.url}
          </CardDescription>
        ) : null}
      </CardHeader>
      <CardContent className="space-y-3">
        {entry.thumbnail ? (
          <div className="overflow-hidden rounded-md border border-border/50 bg-muted/20">
            <img
              src={`data:image/jpeg;base64,${entry.thumbnail}`}
              alt={`Screenshot captured ${dateLabel}`}
              className="h-auto w-full object-cover"
            />
          </div>
        ) : (
          <div className="rounded-md border border-dashed border-border/60 px-3 py-6 text-center text-sm text-muted-foreground">
            {entry.error || "Capture failed"}
          </div>
        )}
        <div className="flex flex-wrap items-center gap-2">
          {entry.width && entry.height ? (
            <Badge variant="outline" className="font-mono text-[10px]">
              {entry.width}×{entry.height}
            </Badge>
          ) : null}
          {entry.url ? (
            <Link
              href={entry.url}
              target="_blank"
              rel="noreferrer"
              className="inline-flex items-center gap-1 text-xs text-primary hover:underline"
            >
              Open page
              <ExternalLinkIcon className="size-3" />
            </Link>
          ) : null}
        </div>
      </CardContent>
    </Card>
  )
}