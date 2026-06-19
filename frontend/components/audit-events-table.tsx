"use client"

import * as React from "react"

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
import type { AuditEvent, PaginatedResponse } from "@/lib/types"
import { SearchIcon, AlertCircleIcon } from "lucide-react"

const ACTION_LABELS: Record<string, string> = {
  "auth.login": "Sign in",
  "auth.logout": "Sign out",
  "auth.passkey_register": "Passkey registered",
  "auth.device_code_issued": "Device code issued",
  "auth.enrollment_code_issued": "Enrollment code issued",
  "user.create": "User created",
  "user.credential_delete": "Passkey removed",
  "target.create": "Target added",
  "target.delete": "Target deleted",
  "target.tags_update": "Target tags updated",
  "snapshot.delete": "Snapshot deleted",
  "report.delete": "Report deleted",
  "scan.cancel": "Scan cancelled",
  "settings.update": "Settings updated",
  "data.export": "Data exported",
  "data.import": "Data imported",
  "webhook.create": "Webhook created",
  "webhook.update": "Webhook updated",
  "webhook.delete": "Webhook deleted",
}

function formatAction(action: string): string {
  return ACTION_LABELS[action] ?? action
}

function isIPInfo(value: unknown): value is Record<string, unknown> {
  return !!value && typeof value === "object" && !Array.isArray(value)
}

function formatDetail(detail: Record<string, unknown>): string {
  const keys = Object.keys(detail).filter((key) => key !== "ip_info")
  if (keys.length === 0) return "—"
  return keys
    .slice(0, 4)
    .map((key) => {
      const value = detail[key]
      if (value == null) return key
      if (typeof value === "object") return `${key}: …`
      return `${key}: ${String(value)}`
    })
    .join(" · ")
}

function formatIPCell(event: AuditEvent): { primary: string; secondary?: string } {
  const ip = event.ip?.trim()
  if (!ip) return { primary: "—" }

  const info = event.detail?.ip_info
  if (!isIPInfo(info)) return { primary: ip }

  const parts: string[] = []
  if (typeof info.origin_as === "string" && info.origin_as.trim()) {
    parts.push(info.origin_as.trim())
  }
  if (typeof info.org_name === "string" && info.org_name.trim()) {
    parts.push(info.org_name.trim())
  }
  const location = [info.city, info.country_code]
    .filter((value): value is string => typeof value === "string" && value.trim() !== "")
    .map((value) => value.trim())
    .join(", ")
  if (location) parts.push(location)
  if (typeof info.prefix === "string" && info.prefix.trim()) {
    parts.push(info.prefix.trim())
  }

  return {
    primary: ip,
    secondary: parts.length > 0 ? parts.join(" · ") : undefined,
  }
}

type AuditResponse = PaginatedResponse<AuditEvent>

export function AuditEventsTable() {
  const [q, setQ] = React.useState("")
  const [debouncedQ, setDebouncedQ] = React.useState("")
  const [page, setPage] = React.useState(1)
  const [data, setData] = React.useState<AuditResponse | null>(null)
  const [loading, setLoading] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)

  const limit = 50

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
        const result = await fetchApi<AuditResponse>(`/api/audit?${params.toString()}`)
        if (!cancelled) setData(result)
      } catch (err) {
        if (!cancelled) {
          if (err instanceof ApiError) {
            setError(err.message)
          } else if (err instanceof Error) {
            setError(err.message)
          } else {
            setError("Failed to load audit events")
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

  return (
    <div className="flex flex-col gap-4">
      <form
        onSubmit={(event) => {
          event.preventDefault()
          setPage(1)
        }}
        className="flex items-center gap-2"
      >
        <Input
          placeholder="Search actor, IP, resource, or detail"
          value={q}
          onChange={(event) => {
            setQ(event.target.value)
            setPage(1)
          }}
          className="max-w-md"
          data-testid="audit-search-input"
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
        data-testid="audit-table-container"
      >
        <Table className="table-fixed">
          <TableHeader>
            <TableRow>
              <TableHead className="w-[16%]">When</TableHead>
              <TableHead className="w-[14%]">Actor</TableHead>
              <TableHead className="w-[18%]">Action</TableHead>
              <TableHead className="w-[14%]">Resource</TableHead>
              <TableHead className="w-[22%]">Detail</TableHead>
              <TableHead className="w-[16%]">Client IP</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading && !data ? (
              ["a", "b", "c", "d", "e"].map((key) => (
                <TableRow key={key}>
                  <TableCell><Skeleton className="h-4 w-28" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-20" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-24" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-24" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-40" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-20" /></TableCell>
                </TableRow>
              ))
            ) : !data ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center text-muted-foreground">
                  {error ? "Could not load audit events." : "Loading audit events…"}
                </TableCell>
              </TableRow>
            ) : data.data.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center text-muted-foreground">
                  No audit events found.
                </TableCell>
              </TableRow>
            ) : (
              data.data.map((event) => {
                const when = new Date(event.created_at).toLocaleString()
                const actor = event.actor_name
                  ? `${event.actor_name}${event.actor_role ? ` (${event.actor_role})` : ""}`
                  : "—"
                const resource = event.resource_type
                  ? `${event.resource_type}${event.resource_id ? ` · ${event.resource_id}` : ""}`
                  : "—"
                const detail = formatDetail(event.detail ?? {})
                const ipCell = formatIPCell(event)
                const ipTitle = ipCell.secondary
                  ? `${ipCell.primary}\n${ipCell.secondary}`
                  : ipCell.primary

                return (
                  <TableRow key={event.id} data-testid={`audit-row-${event.id}`}>
                    <TableCell className="max-w-0 text-xs">
                      <div className="truncate" title={when}>{when}</div>
                    </TableCell>
                    <TableCell className="max-w-0 text-sm">
                      <div className="truncate" title={actor}>{actor}</div>
                    </TableCell>
                    <TableCell className="max-w-0 text-sm">
                      <div className="truncate" title={event.action}>
                        {formatAction(event.action)}
                      </div>
                    </TableCell>
                    <TableCell className="max-w-0 font-mono text-xs text-muted-foreground">
                      <div className="truncate" title={resource}>{resource}</div>
                    </TableCell>
                    <TableCell className="max-w-0 text-xs text-muted-foreground">
                      <div className="truncate" title={detail}>{detail}</div>
                    </TableCell>
                    <TableCell className="max-w-0 text-xs text-muted-foreground">
                      <div className="truncate font-mono" title={ipTitle}>
                        {ipCell.primary}
                      </div>
                      {ipCell.secondary ? (
                        <div className="truncate text-[11px] leading-snug" title={ipCell.secondary}>
                          {ipCell.secondary}
                        </div>
                      ) : null}
                    </TableCell>
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