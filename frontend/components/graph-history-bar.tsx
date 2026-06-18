"use client"

import * as React from "react"
import Link from "next/link"

import { HelpTip } from "@/components/help-tip"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { fetchApi } from "@/lib/api"
import type { GraphRouteDiff, PaginatedResponse, SnapshotSummary } from "@/lib/types"

import type { GraphCompareMode } from "@/lib/types"

export type { GraphCompareMode }

function resolveCompareId(
  snapshots: SnapshotSummary[],
  index: number,
  compareMode: GraphCompareMode
): string | undefined {
  if (compareMode === "latest" && index > 0) {
    return snapshots[0]?.id
  }
  if (compareMode === "previous") {
    return snapshots[index + 1]?.id
  }
  return undefined
}

export function GraphHistoryBar({
  targetId,
  routeDiff,
  snapshotId,
  compareMode,
  onSnapshotChange,
  onCompareModeChange,
}: {
  targetId: string
  routeDiff?: GraphRouteDiff
  snapshotId?: string
  compareMode: GraphCompareMode
  onSnapshotChange: (
    snapshotId: string | undefined,
    compareSnapshotId?: string,
    mode?: GraphCompareMode
  ) => void
  onCompareModeChange: (mode: GraphCompareMode) => void
}) {
  const [snapshots, setSnapshots] = React.useState<SnapshotSummary[]>([])
  const [index, setIndex] = React.useState(0)
  const [loading, setLoading] = React.useState(false)

  const emitChange = React.useCallback(
    (nextIndex: number, mode: GraphCompareMode, rows: SnapshotSummary[]) => {
      if (rows.length === 0) return
      const current = nextIndex === 0 ? undefined : rows[nextIndex]?.id
      const compare = resolveCompareId(rows, nextIndex, mode)
      onSnapshotChange(current, compare, mode)
    },
    [onSnapshotChange]
  )

  React.useEffect(() => {
    let cancelled = false
    setLoading(true)
    void fetchApi<PaginatedResponse<SnapshotSummary>>(
      `/api/targets/${targetId}/snapshots?limit=50`
    )
      .then((result) => {
        if (!cancelled) {
          const rows = result.data ?? []
          setSnapshots(rows)
          const found = snapshotId
            ? rows.findIndex((row) => row.id === snapshotId)
            : 0
          const nextIndex = found >= 0 ? found : 0
          setIndex(nextIndex)
          emitChange(nextIndex, compareMode, rows)
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

  React.useEffect(() => {
    if (snapshots.length === 0) return
    emitChange(index, compareMode, snapshots)
  }, [compareMode])

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
  const compareLabel =
    compareMode === "latest" && !isLatest
      ? "Comparing to latest scan"
      : compareMode === "previous" && snapshots[index + 1]
        ? "Comparing to previous snapshot"
        : isLatest
          ? routeDiff?.has_previous
            ? "Latest scan (diff vs previous available)"
            : "Latest scan"
          : `Snapshot ${new Date(selected.scanned_at).toLocaleString()}`

  return (
    <div
      className="flex flex-col gap-3 rounded-xl border border-border/60 bg-card/60 p-4"
      data-testid="graph-history-bar"
    >
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="inline-flex items-center gap-1.5 text-sm font-medium">
          Snapshot history
          <HelpTip id="graph.history_slider" />
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <div className="inline-flex rounded-md border border-border/60 p-0.5">
            <Button
              type="button"
              variant={compareMode === "previous" ? "secondary" : "ghost"}
              size="sm"
              className="h-7 px-2.5 text-xs"
              onClick={() => onCompareModeChange("previous")}
            >
              vs previous
            </Button>
            <Button
              type="button"
              variant={compareMode === "latest" ? "secondary" : "ghost"}
              size="sm"
              className="h-7 px-2.5 text-xs"
              onClick={() => onCompareModeChange("latest")}
            >
              vs latest
            </Button>
          </div>
          {selected.changes?.length ? (
            <Badge variant="secondary" className="font-mono text-xs">
              {selected.changes.length} change
              {selected.changes.length === 1 ? "" : "s"}
            </Badge>
          ) : null}
        </div>
      </div>

      <input
        type="range"
        min={0}
        max={Math.max(0, snapshots.length - 1)}
        value={index}
        onChange={(event) => {
          const nextIndex = Number(event.target.value)
          setIndex(nextIndex)
          emitChange(nextIndex, compareMode, snapshots)
        }}
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
          <Button variant="outline" size="sm" onClick={() => {
            setIndex(0)
            emitChange(0, compareMode, snapshots)
          }}>
            Jump to latest
          </Button>
        ) : null}
      </div>
    </div>
  )
}