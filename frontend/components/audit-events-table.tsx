"use client"

import * as React from "react"

import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
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
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination"
import { Skeleton } from "@/components/ui/skeleton"
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert"
import { fetchApi, ApiError } from "@/lib/api"
import type { AuditEvent, PaginatedResponse } from "@/lib/types"
import { cn } from "@/lib/utils"
import { SearchIcon, AlertCircleIcon, ChevronRightIcon } from "lucide-react"

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

const ACTION_OPTIONS = Object.entries(ACTION_LABELS)
  .map(([value, label]) => ({ value, label }))
  .sort((a, b) => a.label.localeCompare(b.label))

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

function formatDetailJSON(detail: Record<string, unknown>): string {
  if (Object.keys(detail).length === 0) return "{}"
  return JSON.stringify(detail, null, 2)
}

function paginationItems(current: number, total: number): Array<number | "ellipsis"> {
  if (total <= 7) {
    return Array.from({ length: total }, (_, index) => index + 1)
  }

  const items: Array<number | "ellipsis"> = [1]
  const start = Math.max(2, current - 1)
  const end = Math.min(total - 1, current + 1)

  if (start > 2) items.push("ellipsis")
  for (let page = start; page <= end; page += 1) {
    items.push(page)
  }
  if (end < total - 1) items.push("ellipsis")
  items.push(total)
  return items
}

type AuditResponse = PaginatedResponse<AuditEvent>

export function AuditEventsTable() {
  const [q, setQ] = React.useState("")
  const [debouncedQ, setDebouncedQ] = React.useState("")
  const [actionFilter, setActionFilter] = React.useState("all")
  const [page, setPage] = React.useState(1)
  const [expandedId, setExpandedId] = React.useState<string | null>(null)
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
      if (actionFilter !== "all") params.set("action", actionFilter)

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
  }, [page, debouncedQ, actionFilter])

  const totalPages = data ? Math.ceil(data.total / data.limit) : 0

  return (
    <div className="flex flex-col gap-4">
      <form
        onSubmit={(event) => {
          event.preventDefault()
          setPage(1)
        }}
        className="flex flex-wrap items-center gap-2"
      >
        <Input
          placeholder="Search actor, IP, resource, or detail"
          value={q}
          onChange={(event) => {
            setQ(event.target.value)
            setPage(1)
          }}
          className="max-w-md min-w-[12rem] flex-1"
          data-testid="audit-search-input"
        />
        <Select
          value={actionFilter}
          onValueChange={(value) => {
            if (value) {
              setActionFilter(value)
              setPage(1)
            }
          }}
        >
          <SelectTrigger
            className="w-full min-w-[11rem] sm:w-52"
            data-testid="audit-action-filter"
          >
            <SelectValue placeholder="All actions" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All actions</SelectItem>
            {ACTION_OPTIONS.map((option) => (
              <SelectItem key={option.value} value={option.value}>
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
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
              <TableHead className="w-10" />
              <TableHead className="w-[15%]">When</TableHead>
              <TableHead className="w-[13%]">Actor</TableHead>
              <TableHead className="w-[16%]">Action</TableHead>
              <TableHead className="w-[13%]">Resource</TableHead>
              <TableHead className="w-[20%]">Detail</TableHead>
              <TableHead className="w-[15%]">Client IP</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading && !data ? (
              ["a", "b", "c", "d", "e"].map((key) => (
                <TableRow key={key}>
                  <TableCell><Skeleton className="size-6" /></TableCell>
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
                <TableCell colSpan={7} className="text-center text-muted-foreground">
                  {error ? "Could not load audit events." : "Loading audit events…"}
                </TableCell>
              </TableRow>
            ) : data.data.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="text-center text-muted-foreground">
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
                const expanded = expandedId === event.id

                return (
                  <React.Fragment key={event.id}>
                    <TableRow data-testid={`audit-row-${event.id}`}>
                      <TableCell className="px-2">
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon-xs"
                          aria-expanded={expanded}
                          aria-label={expanded ? "Collapse row" : "Expand row"}
                          data-testid={`audit-expand-${event.id}`}
                          onClick={() =>
                            setExpandedId((current) =>
                              current === event.id ? null : event.id
                            )
                          }
                        >
                          <ChevronRightIcon
                            className={cn(
                              "transition-transform",
                              expanded && "rotate-90"
                            )}
                          />
                        </Button>
                      </TableCell>
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
                          <div
                            className="truncate text-[11px] leading-snug"
                            title={ipCell.secondary}
                          >
                            {ipCell.secondary}
                          </div>
                        ) : null}
                      </TableCell>
                    </TableRow>
                    {expanded ? (
                      <TableRow data-testid={`audit-row-detail-${event.id}`}>
                        <TableCell colSpan={7} className="bg-muted/20 p-4">
                          <div className="flex flex-col gap-3 text-sm">
                            <div className="flex flex-wrap gap-x-6 gap-y-1 text-xs text-muted-foreground">
                              <span>
                                <span className="font-medium text-foreground">Event ID:</span>{" "}
                                <span className="font-mono">{event.id}</span>
                              </span>
                              {event.user_id ? (
                                <span>
                                  <span className="font-medium text-foreground">User ID:</span>{" "}
                                  <span className="font-mono">{event.user_id}</span>
                                </span>
                              ) : null}
                              <span>
                                <span className="font-medium text-foreground">Action code:</span>{" "}
                                <span className="font-mono">{event.action}</span>
                              </span>
                            </div>
                            {event.user_agent ? (
                              <div>
                                <p className="mb-1 text-xs font-medium uppercase tracking-wider text-muted-foreground">
                                  User agent
                                </p>
                                <p className="break-all font-mono text-xs text-muted-foreground">
                                  {event.user_agent}
                                </p>
                              </div>
                            ) : null}
                            <div>
                              <p className="mb-1 text-xs font-medium uppercase tracking-wider text-muted-foreground">
                                Detail
                              </p>
                              <pre className="max-h-48 overflow-auto rounded-md border border-border/50 bg-background/70 p-3 font-mono text-xs leading-relaxed whitespace-pre-wrap">
                                {formatDetailJSON(event.detail ?? {})}
                              </pre>
                            </div>
                          </div>
                        </TableCell>
                      </TableRow>
                    ) : null}
                  </React.Fragment>
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
            {paginationItems(page, totalPages).map((item, index) =>
              item === "ellipsis" ? (
                <PaginationItem key={`ellipsis-${index}`}>
                  <PaginationEllipsis />
                </PaginationItem>
              ) : (
                <PaginationItem key={item}>
                  <PaginationLink
                    isActive={item === page}
                    onClick={() => setPage(item)}
                  >
                    {item}
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