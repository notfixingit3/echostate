"use client"

import * as React from "react"
import Link from "next/link"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { fetchApi } from "@/lib/api"
import type { GraphRouteDiff, PaginatedResponse, SnapshotSummary } from "@/lib/types"

export function GraphHistoryBar({
  targetId,
  viewMode,
  routeDiff,
}: {
  targetId: string
  viewMode: "bgp" | "traceroute" | string
  routeDiff?: GraphRouteDiff
}) {
  const [snapshots, setSnapshots] = React.useState<SnapshotSummary[]>([])
  const [index, setIndex] = React.useState(0)
  const [loading, setLoading] = React.useState(false)

  React.useEffect(() => {
    let cancelled = false
    setLoading(true)
    void fetchApi<PaginatedResponse<SnapshotSummary>>(
      `/api/targets/${targetId}/snapshots?limit=50`
    )
      .then((result) => {
        if (!cancelled) {
          setSnapshots(result.data ?? [])
          setIndex(0)
        }
      })
      .catch(() => {
        if (!cancelled) setSnapshots([])
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [targetId])

  if (viewMode !== "bgp" && viewMode !== "traceroute") {
    return (
      <p className="text-xs text-muted-foreground">
        Graph topology reflects the latest scan. Open{" "}
        <Link href={`/target/snapshots?id=${targetId}`} className="text-primary hover:underline">
          target snapshots
        </Link>{" "}
        for full history.
      </p>
    )
  }

  if (loading) {
    return <p className="text-xs text-muted-foreground">Loading snapshot history…</p>
  }

  if (snapshots.length === 0) {
    return (
      <p className="text-xs text-muted-foreground">
        No snapshots yet for this target.
      </p>
    )
  }

  const selected = snapshots[index]
  const isLatest = index === 0
  const compareLabel = isLatest
    ? routeDiff?.has_previous
      ? "Comparing latest scan to previous snapshot"
      : "Latest scan (no previous snapshot to compare)"
    : `Selected snapshot from ${new Date(selected.scanned_at).toLocaleString()}`

  return (
    <div
      className="flex flex-col gap-3 rounded-xl border border-border/60 bg-card/60 p-4"
      data-testid="graph-history-bar"
    >
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="text-sm font-medium">Snapshot history</div>
        {selected.changes?.length ? (
          <Badge variant="secondary" className="font-mono text-xs">
            {selected.changes.length} change
            {selected.changes.length === 1 ? "" : "s"}
          </Badge>
        ) : null}
      </div>

      <input
        type="range"
        min={0}
        max={Math.max(0, snapshots.length - 1)}
        value={index}
        onChange={(event) => setIndex(Number(event.target.value))}
        className="w-full accent-primary"
        data-testid="graph-history-slider"
      />

      <div className="flex flex-wrap items-center justify-between gap-2 text-xs text-muted-foreground">
        <span>
          {isLatest ? "Latest" : new Date(selected.scanned_at).toLocaleString()}
        </span>
        <span>{compareLabel}</span>
      </div>

      <div className="flex flex-wrap gap-2">
        <Button
          variant="outline"
          size="sm"
          render={<Link href={`/snapshot?id=${selected.id}`} />}
        >
          Open snapshot
        </Button>
        {!isLatest ? (
          <Button
            variant="outline"
            size="sm"
            render={<Link href={`/snapshot?id=${selected.id}`} />}
          >
            View diff context
          </Button>
        ) : null}
        {isLatest && routeDiff?.previous_snapshot_id ? (
          <Button
            variant="outline"
            size="sm"
            render={
              <Link href={`/snapshot?id=${routeDiff.previous_snapshot_id}`} />
            }
          >
            Previous snapshot
          </Button>
        ) : null}
      </div>
    </div>
  )
}