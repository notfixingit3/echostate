"use client"

import * as React from "react"
import { useRouter } from "next/navigation"

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
import { fetchApi, ApiError } from "@/lib/api"
import type { SnapshotSummary, PaginatedResponse } from "@/lib/types"
import { AlertCircleIcon } from "lucide-react"

type SnapshotsResponse = PaginatedResponse<SnapshotSummary>

export function TargetSnapshotsTable({ targetId }: { targetId: string }) {
  const router = useRouter()

  const [page, setPage] = React.useState(1)
  const [data, setData] = React.useState<SnapshotsResponse | null>(null)
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

      try {
        const result = await fetchApi<SnapshotsResponse>(
          `/api/targets/${targetId}/snapshots?${params.toString()}`
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
  }, [page, targetId])

  const totalPages = data ? Math.ceil(data.total / data.limit) : 0

  return (
    <div className="flex flex-col gap-4">
      {error && (
        <Alert variant="destructive">
          <AlertCircleIcon />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <div className="overflow-hidden rounded-xl border border-border/60 bg-card/60 backdrop-blur-sm">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Scanned</TableHead>
              <TableHead>ASN</TableHead>
              <TableHead>Web title</TableHead>
              <TableHead>IP</TableHead>
              <TableHead>Client IP</TableHead>
              <TableHead>Country</TableHead>
              <TableHead>Organization</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading && !data ? (
              ["a", "b", "c", "d", "e"].map((key) => (
                <TableRow key={key}>
                  <TableCell><Skeleton className="h-4 w-32" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-20" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-40" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-24" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-24" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-12" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-32" /></TableCell>
                </TableRow>
              ))
            ) : data?.data.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="text-center text-muted-foreground">
                  No snapshots found for this target.
                </TableCell>
              </TableRow>
            ) : (
              data?.data.map((snapshot) => (
                <TableRow
                  key={snapshot.id}
                  className="cursor-pointer"
                  onClick={() => router.push(`/snapshot?id=${snapshot.id}`)}
                >
                  <TableCell>
                    {new Date(snapshot.scanned_at).toLocaleString()}
                  </TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground">
                    {snapshot.asn
                      ? `AS${snapshot.asn}${snapshot.as_name ? ` · ${snapshot.as_name}` : ""}`
                      : "—"}
                  </TableCell>
                  <TableCell className="max-w-[12rem] truncate text-muted-foreground">
                    {snapshot.web_title || "—"}
                  </TableCell>
                  <TableCell className="font-mono text-xs">
                    {snapshot.resolved_ip || "—"}
                  </TableCell>
                  <TableCell>
                    {snapshot.client_ip && snapshot.client_ip !== "unknown"
                      ? snapshot.client_ip
                      : "unknown"}
                  </TableCell>
                  <TableCell>
                    <CountryFlag code={snapshot.pwhois_country_code} />
                  </TableCell>
                  <TableCell>{snapshot.pwhois_org_name || "—"}</TableCell>
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
