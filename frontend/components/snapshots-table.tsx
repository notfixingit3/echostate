"use client"

import * as React from "react"
import { useRouter } from "next/navigation"

import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { CountryFlag } from "@/components/country-flag"
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
import { Checkbox } from "@/components/ui/checkbox"
import { BulkActionBar } from "@/components/bulk-action-bar"
import { useAuth } from "@/components/auth-provider"
import { useConfirm } from "@/components/confirm-dialog"
import { fetchApi, deleteSnapshot, ApiError } from "@/lib/api"
import { useTableSelection } from "@/lib/use-table-selection"
import type { SnapshotSummary, PaginatedResponse } from "@/lib/types"
import { SearchIcon, AlertCircleIcon, Trash2Icon } from "lucide-react"

type SnapshotsResponse = PaginatedResponse<SnapshotSummary>

export function SnapshotsTable() {
  const router = useRouter()
  const { user } = useAuth()
  const confirm = useConfirm()
  const canDelete = user?.role === "admin"
  const [targetId, setTargetId] = React.useState("")
  const [debouncedTargetId, setDebouncedTargetId] = React.useState("")
  const [country, setCountry] = React.useState("")
  const [page, setPage] = React.useState(1)
  const [reloadKey, setReloadKey] = React.useState(0)
  const [data, setData] = React.useState<SnapshotsResponse | null>(null)
  const [loading, setLoading] = React.useState(false)
  const [deletingId, setDeletingId] = React.useState<string | null>(null)
  const [bulkDeleting, setBulkDeleting] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)

  const limit = 20
  const columnCount = canDelete ? 9 : 7

  React.useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedTargetId(targetId), 300)
    return () => window.clearTimeout(timer)
  }, [targetId])

  React.useEffect(() => {
    let cancelled = false

    async function load() {
      setLoading(true)
      setError(null)

      const params = new URLSearchParams()
      params.set("page", String(page))
      params.set("limit", String(limit))
      if (debouncedTargetId.trim()) {
        params.set("target_id", debouncedTargetId.trim())
      }

      try {
        const result = await fetchApi<SnapshotsResponse>(
          `/api/snapshots?${params.toString()}`
        )
        if (!cancelled) setData(result)
      } catch (err) {
        if (!cancelled) {
          if (err instanceof ApiError) {
            setError(err.message)
          } else if (err instanceof Error) {
            setError(err.message)
          } else {
            setError("Failed to load snapshots")
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
  }, [page, debouncedTargetId, reloadKey])

  async function handleDelete(snapshot: SnapshotSummary) {
    const label = snapshot.host || snapshot.target_id
    const ok = await confirm({
      title: "Delete snapshot?",
      description: `Delete snapshot for ${label}? Reports for this snapshot will also be removed.`,
      confirmLabel: "Delete",
      destructive: true,
    })
    if (!ok) return

    setDeletingId(snapshot.id)
    setError(null)

    try {
      await deleteSnapshot(snapshot.id)
      setReloadKey((key) => key + 1)
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else if (err instanceof Error) {
        setError(err.message)
      } else {
        setError("Failed to delete snapshot")
      }
    } finally {
      setDeletingId(null)
    }
  }

  const filteredData = React.useMemo(() => {
    if (!data) return null
    if (!country.trim()) return data
    const normalizedCountry = country.trim().toLowerCase()
    return {
      ...data,
      data: data.data.filter(
        (snapshot) =>
          snapshot.pwhois_country_code?.toLowerCase() === normalizedCountry
      ),
    }
  }, [data, country])

  const visibleIds = React.useMemo(
    () => (filteredData?.data ?? []).map((snapshot) => snapshot.id),
    [filteredData?.data]
  )
  const selection = useTableSelection(visibleIds)

  async function deleteSnapshots(ids: string[]) {
    const results = await Promise.allSettled(ids.map((id) => deleteSnapshot(id)))
    const failed = results.filter((result) => result.status === "rejected").length
    if (failed > 0) {
      throw new Error(
        failed === ids.length
          ? "Failed to delete selected snapshots"
          : `Deleted ${ids.length - failed} of ${ids.length} snapshots; ${failed} failed`
      )
    }
  }

  async function handleBulkDelete() {
    const ids = [...selection.selected]
    if (ids.length === 0) return

    const noun = ids.length === 1 ? "snapshot" : `${ids.length} snapshots`
    const ok = await confirm({
      title: "Delete snapshots?",
      description: `Delete ${noun}? Reports for these snapshots will also be removed.`,
      confirmLabel: "Delete",
      destructive: true,
    })
    if (!ok) return

    setBulkDeleting(true)
    setError(null)

    try {
      await deleteSnapshots(ids)
      selection.clear()
      setReloadKey((key) => key + 1)
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else if (err instanceof Error) {
        setError(err.message)
      } else {
        setError("Failed to delete selected snapshots")
      }
      setReloadKey((key) => key + 1)
    } finally {
      setBulkDeleting(false)
    }
  }

  const totalPages = filteredData
    ? Math.ceil(filteredData.total / filteredData.limit)
    : 0

  const handleFilterSubmit: React.FormEventHandler<HTMLFormElement> = (event) => {
    event.preventDefault()
    setPage(1)
  }

  return (
    <div className="flex flex-col gap-4">
      <form
        onSubmit={handleFilterSubmit}
        className="flex flex-wrap items-center gap-2"
      >
        <Input
          placeholder="Filter by target ID"
          value={targetId}
          onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
            setTargetId(event.target.value)
            setPage(1)
          }}
          className="max-w-xs"
        />
        <Input
          placeholder="Filter by country code"
          value={country}
          onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
            setCountry(event.target.value)
            setPage(1)
          }}
          className="max-w-xs"
        />
        <Button type="submit" disabled={loading}>
          <SearchIcon data-icon="inline-start" />
          Filter
        </Button>
      </form>

      {error && (
        <Alert variant="destructive">
          <AlertCircleIcon />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {canDelete ? (
        <BulkActionBar
          count={selection.count}
          noun="snapshot"
          deleting={bulkDeleting}
          onClear={selection.clear}
          onDelete={() => void handleBulkDelete()}
        />
      ) : null}

      <div className="overflow-x-auto rounded-xl border border-border/60 bg-card/60 backdrop-blur-sm" data-testid="snapshots-table-container">
        <Table>
          <TableHeader>
            <TableRow>
              {canDelete ? (
                <TableHead className="w-10">
                  <Checkbox
                    checked={selection.allSelected}
                    onCheckedChange={() => selection.toggleAll()}
                    aria-label="Select all snapshots on this page"
                    data-testid="snapshots-select-all"
                    onClick={(event) => event.stopPropagation()}
                  />
                </TableHead>
              ) : null}
              <TableHead>Host</TableHead>
              <TableHead>ASN</TableHead>
              <TableHead>Web title</TableHead>
              <TableHead>Scanned</TableHead>
              <TableHead>Client IP</TableHead>
              <TableHead>Country</TableHead>
              <TableHead>Organization</TableHead>
              {canDelete ? (
                <TableHead className="text-right">Actions</TableHead>
              ) : null}
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading && !filteredData ? (
              ["a", "b", "c", "d", "e"].map((key) => (
                <TableRow key={key}>
                  <TableCell><Skeleton className="h-4 w-32" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-20" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-40" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-32" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-24" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-12" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-32" /></TableCell>
                  {canDelete ? (
                    <>
                      <TableCell>
                        <Skeleton className="h-4 w-4" />
                      </TableCell>
                      <TableCell className="text-right">
                        <Skeleton className="ml-auto h-8 w-8" />
                      </TableCell>
                    </>
                  ) : null}
                </TableRow>
              ))
            ) : !filteredData ? (
              <TableRow>
                <TableCell colSpan={columnCount} className="text-center text-muted-foreground">
                  {error ? "Could not load snapshots." : "Loading snapshots…"}
                </TableCell>
              </TableRow>
            ) : filteredData.data.length === 0 ? (
              <TableRow>
                <TableCell colSpan={columnCount} className="text-center text-muted-foreground">
                  No snapshots found.
                </TableCell>
              </TableRow>
            ) : (
              filteredData.data.map((snapshot) => (
                <TableRow
                  key={snapshot.id}
                  className="cursor-pointer"
                  onClick={() => router.push(`/snapshot?id=${snapshot.id}`)}
                  data-testid={`snapshot-row-${snapshot.id}`}
                >
                  {canDelete ? (
                    <TableCell onClick={(event) => event.stopPropagation()}>
                      <Checkbox
                        checked={selection.selected.has(snapshot.id)}
                        onCheckedChange={() => selection.toggle(snapshot.id)}
                        aria-label={`Select snapshot for ${snapshot.host || snapshot.target_id}`}
                        data-testid={`snapshot-select-${snapshot.id}`}
                      />
                    </TableCell>
                  ) : null}
                  <TableCell className="font-medium">
                    {snapshot.host || snapshot.target_id}
                  </TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground">
                    {snapshot.asn
                      ? `AS${snapshot.asn}${snapshot.as_name ? ` · ${snapshot.as_name}` : ""}`
                      : "—"}
                  </TableCell>
                  <TableCell className="max-w-[12rem] truncate text-muted-foreground">
                    {snapshot.web_title || "—"}
                  </TableCell>
                  <TableCell>
                    {new Date(snapshot.scanned_at).toLocaleString()}
                  </TableCell>
                  <TableCell>
                    {snapshot.client_ip && snapshot.client_ip !== "unknown"
                      ? snapshot.client_ip
                      : "unknown"}
                  </TableCell>
                  <TableCell>
                    <CountryFlag code={snapshot.pwhois_country_code} />
                  </TableCell>
                  <TableCell>
                    {snapshot.pwhois_org_name || "—"}
                  </TableCell>
                  {canDelete ? (
                    <TableCell className="text-right">
                      <Button
                        variant="ghost"
                        size="icon"
                        className="text-destructive hover:bg-destructive/10 hover:text-destructive"
                        disabled={deletingId === snapshot.id || bulkDeleting}
                        aria-label={`Delete snapshot for ${snapshot.host || snapshot.target_id}`}
                        onClick={(event) => {
                          event.stopPropagation()
                          void handleDelete(snapshot)
                        }}
                      >
                        <Trash2Icon />
                      </Button>
                    </TableCell>
                  ) : null}
                </TableRow>
              ))
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
