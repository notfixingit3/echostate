"use client"

import * as React from "react"
import { useRouter, useSearchParams } from "next/navigation"

import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
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
import { fetchApi, downloadReport, deleteReport, ApiError } from "@/lib/api"
import { useTableSelection } from "@/lib/use-table-selection"
import type { ReportSummary, ReportStatus, PaginatedResponse } from "@/lib/types"
import { SearchIcon, AlertCircleIcon, DownloadIcon, Trash2Icon } from "lucide-react"

type ReportsResponse = PaginatedResponse<ReportSummary>

const statusOptions: { value: ReportStatus; label: string }[] = [
  { value: "pending", label: "Pending" },
  { value: "running", label: "Running" },
  { value: "completed", label: "Completed" },
  { value: "failed", label: "Failed" },
]

function statusBadgeVariant(status: ReportStatus) {
  switch (status) {
    case "completed":
      return "default"
    case "failed":
      return "destructive"
    case "running":
      return "secondary"
    case "pending":
    default:
      return "outline"
  }
}

export function ReportsTable() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const { user } = useAuth()
  const confirm = useConfirm()
  const canDelete = Boolean(user)
  const [status, setStatus] = React.useState<ReportStatus | "">("")
  const [snapshotId, setSnapshotId] = React.useState(
    () => searchParams.get("snapshot_id")?.trim() ?? ""
  )
  const [page, setPage] = React.useState(1)
  const [reloadKey, setReloadKey] = React.useState(0)
  const [data, setData] = React.useState<ReportsResponse | null>(null)
  const [loading, setLoading] = React.useState(false)
  const [deletingId, setDeletingId] = React.useState<string | null>(null)
  const [bulkDeleting, setBulkDeleting] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)
  const [downloadError, setDownloadError] = React.useState<string | null>(null)

  const limit = 20
  const columnCount = canDelete ? 8 : 7

  React.useEffect(() => {
    const fromUrl = searchParams.get("snapshot_id")?.trim() ?? ""
    setSnapshotId((current) => (current === fromUrl ? current : fromUrl))
    setPage(1)
  }, [searchParams])

  React.useEffect(() => {
    let cancelled = false

    async function load() {
      setLoading(true)
      setError(null)

      const params = new URLSearchParams()
      params.set("page", String(page))
      params.set("limit", String(limit))
      if (status) params.set("status", status)
      if (snapshotId.trim()) params.set("snapshot_id", snapshotId.trim())

      try {
        const result = await fetchApi<ReportsResponse>(
          `/api/reports?${params.toString()}`
        )
        if (!cancelled) setData(result)
      } catch (err) {
        if (!cancelled) {
          if (err instanceof ApiError) {
            setError(err.message)
          } else if (err instanceof Error) {
            setError(err.message)
          } else {
            setError("Failed to load reports")
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
  }, [page, status, snapshotId, reloadKey])

  const visibleIds = React.useMemo(
    () => (data?.data ?? []).map((report) => report.id),
    [data?.data]
  )
  const selection = useTableSelection(visibleIds)

  async function deleteReports(ids: string[]) {
    const results = await Promise.allSettled(ids.map((id) => deleteReport(id)))
    const failed = results.filter((result) => result.status === "rejected").length
    if (failed > 0) {
      throw new Error(
        failed === ids.length
          ? "Failed to delete selected reports"
          : `Deleted ${ids.length - failed} of ${ids.length} reports; ${failed} failed`
      )
    }
  }

  async function handleBulkDelete() {
    const ids = [...selection.selected]
    if (ids.length === 0) return

    const noun = ids.length === 1 ? "report" : `${ids.length} reports`
    const ok = await confirm({
      title: "Delete reports?",
      description: `Delete ${noun}?`,
      confirmLabel: "Delete",
      destructive: true,
    })
    if (!ok) return

    setBulkDeleting(true)
    setError(null)

    try {
      await deleteReports(ids)
      selection.clear()
      setReloadKey((key) => key + 1)
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else if (err instanceof Error) {
        setError(err.message)
      } else {
        setError("Failed to delete selected reports")
      }
      setReloadKey((key) => key + 1)
    } finally {
      setBulkDeleting(false)
    }
  }

  async function handleDelete(report: ReportSummary) {
    const label = report.host || report.id
    const ok = await confirm({
      title: "Delete report?",
      description: `Delete report for ${label}?`,
      confirmLabel: "Delete",
      destructive: true,
    })
    if (!ok) return

    setDeletingId(report.id)
    setError(null)

    try {
      await deleteReport(report.id)
      setReloadKey((key) => key + 1)
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message)
      } else if (err instanceof Error) {
        setError(err.message)
      } else {
        setError("Failed to delete report")
      }
    } finally {
      setDeletingId(null)
    }
  }

  const totalPages = data ? Math.ceil(data.total / data.limit) : 0

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
        <Select
          value={status}
          onValueChange={(value) => {
            setStatus(value as ReportStatus)
            setPage(1)
          }}
        >
          <SelectTrigger className="w-40">
            <SelectValue placeholder="All statuses" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All statuses</SelectItem>
            {statusOptions.map((option) => (
              <SelectItem key={option.value} value={option.value}>
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Input
          placeholder="Filter by snapshot ID"
          value={snapshotId}
          onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
            setSnapshotId(event.target.value)
            setPage(1)
          }}
          className="max-w-xs"
        />
        <Button type="submit" disabled={loading}>
          <SearchIcon data-icon="inline-start" />
          Filter
        </Button>
      </form>

      {(error || downloadError) && (
        <Alert variant="destructive">
          <AlertCircleIcon />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{error || downloadError}</AlertDescription>
        </Alert>
      )}

      {canDelete ? (
        <BulkActionBar
          count={selection.count}
          noun="report"
          deleting={bulkDeleting}
          onClear={selection.clear}
          onDelete={() => void handleBulkDelete()}
        />
      ) : null}

      <div className="overflow-x-auto rounded-xl border border-border/60 bg-card/60 backdrop-blur-sm" data-testid="reports-table-container">
        <Table>
          <TableHeader>
            <TableRow>
              {canDelete ? (
                <TableHead className="w-10">
                  <Checkbox
                    checked={selection.allSelected}
                    onCheckedChange={() => selection.toggleAll()}
                    aria-label="Select all reports on this page"
                    data-testid="reports-select-all"
                    onClick={(event) => event.stopPropagation()}
                  />
                </TableHead>
              ) : null}
              <TableHead>ID</TableHead>
              <TableHead>Snapshot</TableHead>
              <TableHead>Host</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Created</TableHead>
              <TableHead>Completed</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading && !data ? (
              ["a", "b", "c", "d", "e"].map((key) => (
                <TableRow key={key}>
                  <TableCell><Skeleton className="h-4 w-32" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-32" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-24" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-16" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-32" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-32" /></TableCell>
                  {canDelete ? (
                    <TableCell>
                      <Skeleton className="h-4 w-4" />
                    </TableCell>
                  ) : null}
                  <TableCell className="text-right">
                    <Skeleton className="h-8 w-24" />
                  </TableCell>
                </TableRow>
              ))
            ) : !data ? (
              <TableRow>
                <TableCell colSpan={columnCount} className="text-center text-muted-foreground">
                  {error ? "Could not load reports." : "Loading reports…"}
                </TableCell>
              </TableRow>
            ) : data.data.length === 0 ? (
              <TableRow>
                <TableCell colSpan={columnCount} className="text-center text-muted-foreground">
                  No reports found.
                </TableCell>
              </TableRow>
            ) : (
              data.data.map((report) => (
                <TableRow
                  key={report.id}
                  className="cursor-pointer"
                  onClick={() => router.push(`/report?id=${report.id}`)}
                  data-testid={`report-row-${report.id}`}
                >
                  {canDelete ? (
                    <TableCell onClick={(event) => event.stopPropagation()}>
                      <Checkbox
                        checked={selection.selected.has(report.id)}
                        onCheckedChange={() => selection.toggle(report.id)}
                        aria-label={`Select report for ${report.host || report.id}`}
                        data-testid={`report-select-${report.id}`}
                      />
                    </TableCell>
                  ) : null}
                  <TableCell className="font-mono text-xs">{report.id}</TableCell>
                  <TableCell className="font-mono text-xs">
                    {report.snapshot_id}
                  </TableCell>
                  <TableCell>{report.host || "—"}</TableCell>
                  <TableCell>
                    <Badge variant={statusBadgeVariant(report.status)} data-testid={`report-status-${report.id}`}>
                      {report.status}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    {new Date(report.created_at).toLocaleString()}
                  </TableCell>
                  <TableCell>
                    {report.completed_at
                      ? new Date(report.completed_at).toLocaleString()
                      : "—"}
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex items-center justify-end gap-1">
                      <Button
                        variant="outline"
                        size="sm"
                        disabled={report.status !== "completed"}
                        onClick={(event) => {
                          event.stopPropagation()
                          if (report.status !== "completed") return
                          void downloadReport(report).catch((err) => {
                            if (err instanceof ApiError) {
                              setDownloadError(err.message)
                            } else if (err instanceof Error) {
                              setDownloadError(err.message)
                            } else {
                              setDownloadError("Failed to download report")
                            }
                          })
                        }}
                      >
                        <DownloadIcon data-icon="inline-start" />
                        Download
                      </Button>
                      {canDelete ? (
                        <Button
                          variant="ghost"
                          size="icon"
                          className="text-destructive hover:bg-destructive/10 hover:text-destructive"
                          disabled={deletingId === report.id || bulkDeleting}
                          aria-label={`Delete report for ${report.host || report.id}`}
                          onClick={(event) => {
                            event.stopPropagation()
                            void handleDelete(report)
                          }}
                        >
                          <Trash2Icon />
                        </Button>
                      ) : null}
                    </div>
                  </TableCell>
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
