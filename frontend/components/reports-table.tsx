"use client"

import * as React from "react"
import { useRouter } from "next/navigation"

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
import { fetchApi, ApiError, API_BASE } from "@/lib/api"
import type { ReportSummary, ReportStatus, PaginatedResponse } from "@/lib/types"
import { SearchIcon, AlertCircleIcon, DownloadIcon } from "lucide-react"

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
  const [status, setStatus] = React.useState<ReportStatus | "">("")
  const [snapshotId, setSnapshotId] = React.useState("")
  const [page, setPage] = React.useState(1)
  const [data, setData] = React.useState<ReportsResponse | null>(null)
  const [loading, setLoading] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)

  const limit = 20

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
  }, [page, status, snapshotId])

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

      {error && (
        <Alert variant="destructive">
          <AlertCircleIcon />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <div className="overflow-hidden rounded-xl border border-border/60 bg-card/60 backdrop-blur-sm" data-testid="reports-table-container">
        <Table>
          <TableHeader>
            <TableRow>
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
                  <TableCell className="text-right">
                    <Skeleton className="h-8 w-24" />
                  </TableCell>
                </TableRow>
              ))
            ) : data?.data.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="text-center text-muted-foreground">
                  No reports found.
                </TableCell>
              </TableRow>
            ) : (
              data?.data.map((report) => (
                <TableRow
                  key={report.id}
                  className="cursor-pointer"
                  onClick={() => router.push(`/report?id=${report.id}`)}
                  data-testid={`report-row-${report.id}`}
                >
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
                    <Button
                      variant="outline"
                      size="sm"
                      disabled={report.status !== "completed"}
                      onClick={(event) => {
                        event.stopPropagation()
                        if (report.status !== "completed") return
                        const a = document.createElement("a")
                        a.href = `${API_BASE}/api/reports/${report.id}/download`
                        a.download = `echostate-report-${report.id}.pdf`
                        document.body.appendChild(a)
                        a.click()
                        document.body.removeChild(a)
                      }}
                    >
                      <DownloadIcon data-icon="inline-start" />
                      Download
                    </Button>
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
