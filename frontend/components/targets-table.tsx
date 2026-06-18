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
import { fetchApi, ApiError } from "@/lib/api"
import type { TargetSummary, PaginatedResponse } from "@/lib/types"
import { SearchIcon, AlertCircleIcon } from "lucide-react"

type TargetsResponse = PaginatedResponse<TargetSummary>

export function TargetsTable() {
  const router = useRouter()
  const [q, setQ] = React.useState("")
  const [debouncedQ, setDebouncedQ] = React.useState("")
  const [page, setPage] = React.useState(1)
  const [data, setData] = React.useState<TargetsResponse | null>(null)
  const [loading, setLoading] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)

  const limit = 20

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
  }, [page, debouncedQ])

  const totalPages = data ? Math.ceil(data.total / data.limit) : 0

  const handleSearchSubmit: React.FormEventHandler<HTMLFormElement> = (event) => {
    event.preventDefault()
    setPage(1)
    // The effect will re-fetch because q changed.
  }

  return (
    <div className="flex flex-col gap-4">
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

      <div className="overflow-hidden rounded-xl border border-border/60 bg-card/60 backdrop-blur-sm" data-testid="targets-table-container">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Host</TableHead>
              <TableHead>ASN</TableHead>
              <TableHead>Web title</TableHead>
              <TableHead>Created</TableHead>
              <TableHead>Snapshots</TableHead>
              <TableHead>Latest snapshot</TableHead>
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
                </TableRow>
              ))
            ) : !data ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center text-muted-foreground">
                  {error ? "Could not load targets." : "Loading targets…"}
                </TableCell>
              </TableRow>
            ) : data.data.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center text-muted-foreground">
                  No targets found.
                </TableCell>
              </TableRow>
            ) : (
              data.data.map((target) => (
                <TableRow
                  key={target.id}
                  className="cursor-pointer"
                  onClick={() => router.push(`/target?id=${target.id}`)}
                  data-testid={`target-row-${target.id}`}
                >
                  <TableCell>
                    <div className="font-medium">{target.host}</div>
                    {target.tags && target.tags.length > 0 && (
                      <div className="flex flex-wrap gap-1 mt-1">
                        {target.tags.map((t) => (
                          <span key={t} className="rounded bg-primary/10 px-1.5 py-0.5 text-[10px] text-primary">{t}</span>
                        ))}
                      </div>
                    )}
                  </TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground">
                    {target.latest_asn
                      ? `AS${target.latest_asn}${target.latest_as_name ? ` · ${target.latest_as_name}` : ""}`
                      : "—"}
                  </TableCell>
                  <TableCell className="max-w-[12rem] truncate text-muted-foreground">
                    {target.latest_web_title || "—"}
                  </TableCell>
                  <TableCell>
                    {new Date(target.created_at).toLocaleString()}
                  </TableCell>
                  <TableCell>{target.snapshot_count}</TableCell>
                  <TableCell>
                    {target.latest_snapshot_at
                      ? new Date(target.latest_snapshot_at).toLocaleString()
                      : "—"}
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
