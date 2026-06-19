"use client"

import * as React from "react"
import { useRouter } from "next/navigation"

import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  Pagination,
  PaginationContent,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination"
import { Skeleton } from "@/components/ui/skeleton"
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert"
import { useAuth } from "@/components/auth-provider"
import { useConfirm } from "@/components/confirm-dialog"
import { HelpTip } from "@/components/help-tip"
import { Label } from "@/components/ui/label"
import { fetchApi, createTarget, deleteTarget, ApiError } from "@/lib/api"
import { validateScanTarget } from "@/lib/validate-host"
import type { TargetSummary, PaginatedResponse } from "@/lib/types"
import { SearchIcon, AlertCircleIcon, Trash2Icon, PlusIcon } from "lucide-react"

type TargetsResponse = PaginatedResponse<TargetSummary>

export function TargetsTable() {
  const router = useRouter()
  const { user } = useAuth()
  const confirm = useConfirm()
  const canDelete = user?.role === "admin"
  const [q, setQ] = React.useState("")
  const [debouncedQ, setDebouncedQ] = React.useState("")
  const [page, setPage] = React.useState(1)
  const [data, setData] = React.useState<TargetsResponse | null>(null)
  const [loading, setLoading] = React.useState(false)
  const [deletingId, setDeletingId] = React.useState<string | null>(null)
  const [reloadKey, setReloadKey] = React.useState(0)
  const [error, setError] = React.useState<string | null>(null)
  const [newHost, setNewHost] = React.useState("")
  const [adding, setAdding] = React.useState(false)

  const limit = 20
  const columnCount = canDelete ? 7 : 6

  React.useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedQ(q), 300)
    return () => window.clearTimeout(timer)
  }, [q])

  React.useEffect(() => {
    let cancelled = false

    async function load() {
      setLoading(true)
      setError(null)

      const params = new URLSearchParams()
      params.set("page", String(page))
      params.set("limit", String(limit))
      if (debouncedQ.trim()) params.set("q", debouncedQ.trim())

      try {
        const result = await fetchApi<TargetsResponse>(
          `/api/targets?${params.toString()}`
        )
        if (!cancelled) setData(result)
      } catch (err) {
        if (!cancelled) {
          if (err instanceof ApiError) {
            setError(err.message)
          } else if (err instanceof Error) {
            setError(err.message)
          } else {
            setError("Failed to load targets")
          }
        }
      } finally {
        if (!cancelled) setLoading(false)
      }
    }

    load()

    return () => {
      cancelled = true
    }
  }, [page, debouncedQ, reloadKey])

  async function handleDelete(target: TargetSummary, event: React.MouseEvent) {
    event.stopPropagation()
    const ok = await confirm({
      title: "Delete target?",
      description: `Delete target ${target.host}? All snapshots and reports for this host will be removed.`,
      confirmLabel: "Delete",
      destructive: true,
    })
    if (!ok) return

    setDeletingId(target.id)
    setError(null)

    try {
      await deleteTarget(target.id)
      setReloadKey((key) => key + 1)
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else if (err instanceof Error) {
        setError(err.message)
      } else {
        setError("Failed to delete target")
      }
    } finally {
      setDeletingId(null)
    }
  }

  const totalPages = data ? Math.ceil(data.total / data.limit) : 0

  const handleSearchSubmit: React.FormEventHandler<HTMLFormElement> = (event) => {
    event.preventDefault()
    setPage(1)
    // The effect will re-fetch because q changed.
  }

  async function handleAddTarget(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError(null)

    let normalized: string
    try {
      normalized = validateScanTarget(newHost)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Invalid host")
      return
    }

    setAdding(true)
    try {
      const target = await createTarget(normalized)
      setNewHost("")
      setPage(1)
      setReloadKey((key) => key + 1)
      router.push(`/target?id=${target.id}`)
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else if (err instanceof Error) {
        setError(err.message)
      } else {
        setError("Failed to add target")
      }
    } finally {
      setAdding(false)
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <form
        onSubmit={handleAddTarget}
        className="flex flex-wrap items-end gap-2"
        data-testid="add-target-form"
      >
        <div className="space-y-1.5">
          <Label htmlFor="add-target-host" className="inline-flex items-center gap-1.5">
            Add target
            <HelpTip id="targets.add" />
          </Label>
          <Input
            id="add-target-host"
            placeholder="example.com or 203.0.113.10"
            value={newHost}
            onChange={(event) => setNewHost(event.target.value)}
            className="w-full min-w-[16rem] max-w-md"
            data-testid="add-target-host-input"
            disabled={adding}
          />
        </div>
        <Button type="submit" disabled={adding || !newHost.trim()} data-testid="add-target-submit">
          <PlusIcon data-icon="inline-start" />
          Add target
        </Button>
      </form>

      <form
        onSubmit={handleSearchSubmit}
        className="flex items-center gap-2"
      >
        <Input
          placeholder="Search by host"
          value={q}
          onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
            setQ(event.target.value)
            setPage(1)
          }}
          className="max-w-sm"
        />
        <Button type="submit" disabled={loading}>
          <SearchIcon data-icon="inline-start" />
          Search
        </Button>
      </form>

      {error && (
        <Alert variant="destructive">
          <AlertCircleIcon />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <div
        className="overflow-hidden rounded-xl border border-border/60 bg-card/60 backdrop-blur-sm [&_[data-slot=table-container]]:overflow-x-hidden"
        data-testid="targets-table-container"
      >
        <Table className="table-fixed">
          <TableHeader>
            <TableRow>
              <TableHead className="w-[24%]">Host</TableHead>
              <TableHead className="w-[18%]">ASN</TableHead>
              <TableHead className="w-[16%]">Web title</TableHead>
              <TableHead className="w-[14%]">Created</TableHead>
              <TableHead className="w-[8%]">Snapshots</TableHead>
              <TableHead className="w-[14%]">Latest snapshot</TableHead>
              {canDelete ? <TableHead className="w-12 px-1" /> : null}
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading && !data ? (
              ["a", "b", "c", "d", "e"].map((key) => (
                <TableRow key={key}>
                  <TableCell><Skeleton className="h-4 w-32" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-20" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-40" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-32" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-12" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-32" /></TableCell>
                  {canDelete ? <TableCell><Skeleton className="h-4 w-8" /></TableCell> : null}
                </TableRow>
              ))
            ) : !data ? (
              <TableRow>
                <TableCell colSpan={columnCount} className="text-center text-muted-foreground">
                  {error ? "Could not load targets." : "Loading targets…"}
                </TableCell>
              </TableRow>
            ) : data.data.length === 0 ? (
              <TableRow>
                <TableCell colSpan={columnCount} className="text-center text-muted-foreground">
                  No targets found.
                </TableCell>
              </TableRow>
            ) : (
              data.data.map((target) => {
                const asnLabel = target.latest_asn
                  ? `AS${target.latest_asn}${target.latest_as_name ? ` · ${target.latest_as_name}` : ""}`
                  : "—"
                const createdLabel = new Date(target.created_at).toLocaleString()
                const latestSnapshotLabel = target.latest_snapshot_at
                  ? new Date(target.latest_snapshot_at).toLocaleString()
                  : null

                return (
                <TableRow
                  key={target.id}
                  className="cursor-pointer"
                  onClick={() => router.push(`/target?id=${target.id}`)}
                  data-testid={`target-row-${target.id}`}
                >
                  <TableCell className="max-w-0 whitespace-normal">
                    <div className="truncate font-medium" title={target.host}>
                      {target.host}
                    </div>
                    {target.tags && target.tags.length > 0 && (
                      <div className="mt-1 flex flex-wrap gap-1">
                        {target.tags.map((t) => (
                          <span key={t} className="rounded bg-primary/10 px-1.5 py-0.5 text-[10px] text-primary">{t}</span>
                        ))}
                      </div>
                    )}
                  </TableCell>
                  <TableCell className="max-w-0 font-mono text-xs text-muted-foreground">
                    <div className="truncate" title={asnLabel}>
                      {asnLabel}
                    </div>
                  </TableCell>
                  <TableCell className="max-w-0 text-muted-foreground">
                    <div
                      className="truncate"
                      title={target.latest_web_title || undefined}
                    >
                      {target.latest_web_title || "—"}
                    </div>
                  </TableCell>
                  <TableCell className="max-w-0 text-xs">
                    <div className="truncate" title={createdLabel}>
                      {createdLabel}
                    </div>
                  </TableCell>
                  <TableCell className="text-center tabular-nums">
                    {target.snapshot_count}
                  </TableCell>
                  <TableCell className="max-w-0 text-xs">
                    {latestSnapshotLabel ? (
                      <div className="truncate" title={latestSnapshotLabel}>
                        {latestSnapshotLabel}
                      </div>
                    ) : (
                      "—"
                    )}
                  </TableCell>
                  {canDelete ? (
                    <TableCell className="w-12 px-1">
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon-sm"
                        disabled={deletingId === target.id}
                        onClick={(event) => void handleDelete(target, event)}
                        aria-label={`Delete ${target.host}`}
                        data-testid={`target-delete-${target.id}`}
                      >
                        <Trash2Icon />
                      </Button>
                    </TableCell>
                  ) : null}
                </TableRow>
                )
              })
            )}
          </TableBody>
        </Table>
      </div>

      {totalPages > 1 && (
        <Pagination>
          <PaginationContent>
            <PaginationItem>
              <PaginationPrevious
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                aria-disabled={page <= 1}
                className={page <= 1 ? "pointer-events-none opacity-50" : undefined}
              />
            </PaginationItem>
            {Array.from({ length: totalPages }, (_, index) => index + 1).map(
              (pageNumber) => (
                <PaginationItem key={pageNumber}>
                  <PaginationLink
                    isActive={pageNumber === page}
                    onClick={() => setPage(pageNumber)}
                  >
                    {pageNumber}
                  </PaginationLink>
                </PaginationItem>
              )
            )}
            <PaginationItem>
              <PaginationNext
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                aria-disabled={page >= totalPages}
                className={
                  page >= totalPages ? "pointer-events-none opacity-50" : undefined
                }
              />
            </PaginationItem>
          </PaginationContent>
        </Pagination>
      )}
    </div>
  )
}
