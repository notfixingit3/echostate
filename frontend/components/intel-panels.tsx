"use client"

import * as React from "react"
import Link from "next/link"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { CopyButton } from "@/components/copy-button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs"
import {
  ASN_CORE_FIELDS,
  ASN_ROUTING_FIELDS,
  PEERINGDB_FIELDS,
  WEB_FIELDS,
  WHOIS_FIELDS,
  PWHOIS_FIELDS,
  TLS_FIELDS,
  DNS_FIELDS,
  FAVICON_FIELDS,
  CRAWL_FIELDS,
  STORAGE_FIELDS,
  CT_FIELDS,
  TRACEROUTE_FIELDS,
  SCREENSHOT_FIELDS,
  ENRICHMENT_FIELDS,
  formatFieldLabel,
  formatValue,
} from "@/lib/intel"
import { ScreenshotTimeline } from "@/components/screenshot-timeline"
import type { RawIntel } from "@/lib/intel"
import type { ChangeDetail } from "@/lib/types"
import {
  GlobeIcon,
  NetworkIcon,
  MonitorIcon,
  AlertCircleIcon,
  GitCompareIcon,
  ExternalLinkIcon,
  ChevronRightIcon,
  ShieldIcon,
  DatabaseIcon,
  ScrollTextIcon,
  RouteIcon,
  CameraIcon,
} from "lucide-react"
import ReactDiffViewer from "react-diff-viewer-continued"
import { useTheme } from "next-themes"
import { fetchApi } from "@/lib/api"

function isStringRecord(value: unknown): value is Record<string, string> {
  if (!value || typeof value !== "object" || Array.isArray(value)) return false
  return Object.values(value).every((item) => typeof item === "string")
}

function WordPressItemsList({
  items,
  emptyLabel,
}: {
  items: unknown
  emptyLabel: string
}) {
  if (!Array.isArray(items) || items.length === 0) {
    return <p className="text-sm text-muted-foreground">{emptyLabel}</p>
  }

  return (
    <div className="flex flex-col gap-2">
      {items.map((item, index) => {
        if (!item || typeof item !== "object") return null
        const entry = item as Record<string, unknown>
        const slug = typeof entry.slug === "string" ? entry.slug : ""
        if (!slug) return null
        const version =
          typeof entry.version === "string" && entry.version.trim()
            ? entry.version.trim()
            : null

        return (
          <div
            key={`${slug}-${index}`}
            className="flex min-w-0 items-center justify-between gap-3 rounded-md border border-border/50 bg-background/60 px-3 py-2"
          >
            <span className="font-mono text-xs">{slug}</span>
            {version ? (
              <Badge variant="secondary" className="font-mono text-[11px]">
                v{version}
              </Badge>
            ) : (
              <span className="text-xs text-muted-foreground">version unknown</span>
            )}
          </div>
        )
      })}
    </div>
  )
}

function HeaderMap({ data }: { data: Record<string, string> }) {
  const entries = Object.entries(data)
  if (entries.length === 0) {
    return <p className="text-sm text-muted-foreground">No data available.</p>
  }

  return (
    <div className="flex flex-col gap-2">
      {entries.map(([header, val]) => (
        <div
          key={header}
          className="min-w-0 rounded-md border border-border/50 bg-background/60 px-3 py-2"
        >
          <div className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
            {header}
          </div>
          <div className="mt-1 break-all font-mono text-xs leading-relaxed">
            {val}
          </div>
        </div>
      ))}
    </div>
  )
}

function formatTracerouteVantageLabel(
  vantageId: unknown,
  label: unknown,
  index: number
): string {
  const id = typeof vantageId === "string" ? vantageId : ""
  const raw = typeof label === "string" ? label : ""
  const text = (raw || id).toLowerCase()
  if (id === "hackertarget" || id === "external" || text.includes("hackertarget")) {
    return "External vantage"
  }
  if (text.includes("hacker")) {
    return "External vantage"
  }
  return raw || `Vantage ${index + 1}`
}

function sanitizeTracerouteMessage(message: string): string {
  return message
    .replace(/hackertarget/gi, "external")
    .replace(/HackerTarget\s*\(external\)/gi, "External vantage")
}

function TracerouteVantageList({ vantages }: { vantages?: unknown }) {
  if (!Array.isArray(vantages) || vantages.length === 0) {
    return null
  }

  return (
    <div className="mt-6 flex flex-col gap-4 border-t border-border/50 pt-6">
      {vantages.map((vantage, index) => {
        if (!vantage || typeof vantage !== "object") return null
        const entry = vantage as Record<string, unknown>
        const label = formatTracerouteVantageLabel(entry.id, entry.label, index)
        const warning =
          typeof entry.warning === "string"
            ? sanitizeTracerouteMessage(entry.warning)
            : undefined

        return (
          <div key={label} className="rounded-lg border border-border/50 bg-muted/10 p-4">
            <div className="mb-3 flex items-center justify-between gap-2">
              <div className="text-sm font-medium">{label}</div>
              {typeof entry.hop_count === "number" ? (
                <Badge variant="secondary" className="font-mono text-[10px]">
                  {entry.hop_count} hops
                </Badge>
              ) : null}
            </div>
            {warning ? (
              <p className="mb-3 text-xs text-muted-foreground">{warning}</p>
            ) : null}
            <TracerouteHopList hops={entry.hops} />
          </div>
        )
      })}
    </div>
  )
}

function TracerouteHopList({ hops }: { hops?: unknown }) {
  if (!Array.isArray(hops) || hops.length === 0) {
    return null
  }

  return (
    <div className="flex flex-col gap-2">
      <div className="flex flex-col gap-2">
        {hops.map((hop, index) => {
          if (!hop || typeof hop !== "object") return null
          const entry = hop as Record<string, unknown>
          const hopNum = typeof entry.hop === "number" ? entry.hop : index + 1
          const ip = typeof entry.ip === "string" ? entry.ip : undefined
          const timeout = entry.timeout === true
          const rtt =
            typeof entry.rtt_ms === "number"
              ? `${entry.rtt_ms.toFixed(1)} ms`
              : undefined
          const country =
            typeof entry.country === "string" ? entry.country : undefined
          const city = typeof entry.city === "string" ? entry.city : undefined

          return (
            <div
              key={`${hopNum}-${ip ?? "timeout"}`}
              className="flex items-center gap-3 rounded-lg border border-border/50 bg-muted/15 px-3 py-2"
              data-testid={`traceroute-hop-${hopNum}`}
            >
              <Badge variant="outline" className="shrink-0 font-mono text-xs">
                {hopNum}
              </Badge>
              <span className="min-w-0 flex-1 font-mono text-sm">
                {timeout ? "*" : ip ?? "—"}
              </span>
              <div className="flex shrink-0 flex-col items-end text-xs text-muted-foreground">
                {rtt ? <span>{rtt}</span> : null}
                {country ? (
                  <span>
                    {city ? `${city}, ` : ""}
                    {country}
                  </span>
                ) : null}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

function DataGrid({
  data,
  fields,
  exclude = [],
}: {
  data?: Record<string, unknown> | null
  fields?: readonly string[]
  exclude?: string[]
}) {
  if (!data || Object.keys(data).length === 0) {
    return (
      <p className="text-sm text-muted-foreground">No data available.</p>
    )
  }

  const entries = fields
    ? fields
        .filter((key) => !exclude.includes(key) && data[key] !== undefined)
        .map((key) => [key, data[key]] as const)
    : Object.entries(data).filter(([key]) => !exclude.includes(key))

  if (entries.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">No data available.</p>
    )
  }

  return (
    <dl className="flex flex-col gap-3">
      {entries.map(([key, value]) => (
        <div
          key={key}
          className="min-w-0 rounded-lg border border-border/50 bg-muted/15 p-3 sm:p-4"
        >
          <dt className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
            {formatFieldLabel(key)}
          </dt>
          <dd className="mt-2 min-w-0 text-sm">
            {key === "name_servers" ||
            key === "status" ||
            key === "copyrights" ||
            key === "dns_names" ||
            key === "ip_sans" ||
            key === "A" ||
            key === "MX" ||
            key === "NS" ||
            key === "TXT" ||
            key === "DMARC" ||
            key === "tech_stack" ||
            key === "sitemap_urls" ||
            key === "visible_origins" ||
            key === "notes" ? (
              <div className="flex flex-wrap gap-1.5">
                {(Array.isArray(value) ? value : [value]).map((item) => (
                  <Badge
                    key={String(item)}
                    variant="outline"
                    className="h-auto max-w-full whitespace-normal break-all py-1 font-mono text-xs font-normal"
                  >
                    {String(item)}
                  </Badge>
                ))}
              </div>
            ) : key === "path_profile" && value && typeof value === "object" ? (
              <DataGrid data={value as Record<string, unknown>} />
            ) : key === "robots" && value && typeof value === "object" ? (
              <pre className="max-h-80 overflow-auto whitespace-pre-wrap break-words rounded-md border border-border/50 bg-background/60 p-3 font-mono text-xs leading-relaxed">
                {formatValue(value)}
              </pre>
            ) : key === "buckets" && Array.isArray(value) ? (
              <div className="flex flex-col gap-2">
                {value.map((bucket, index) => (
                  <div
                    key={index}
                    className="rounded-md border border-border/50 bg-background/60 px-3 py-2"
                  >
                    <div className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                      {typeof bucket === "object" && bucket && "provider" in bucket
                        ? String((bucket as Record<string, unknown>).provider)
                        : "bucket"}
                    </div>
                    <div className="mt-1 break-all font-mono text-xs">
                      {formatValue(bucket)}
                    </div>
                  </div>
                ))}
              </div>
            ) : key === "wordpress_plugins" && Array.isArray(value) ? (
              <WordPressItemsList items={value} emptyLabel="No plugins detected." />
            ) : key === "wordpress_themes" && Array.isArray(value) ? (
              <WordPressItemsList items={value} emptyLabel="No themes detected." />
            ) : key === "subdomains" && Array.isArray(value) ? (
              <div className="flex flex-wrap gap-1.5">
                {value.map((item) => (
                  <Badge
                    key={String(item)}
                    variant="outline"
                    className="h-auto max-w-full whitespace-normal break-all py-1 font-mono text-xs font-normal"
                    render={
                      <Link
                        href={`/?scan=${encodeURIComponent(String(item))}`}
                        data-testid={`ct-subdomain-${String(item)}`}
                      />
                    }
                  >
                    {String(item)}
                  </Badge>
                ))}
              </div>
            ) : (key === "headers" || key === "security_headers") &&
              isStringRecord(value) ? (
              <HeaderMap data={value} />
            ) : (key === "url" || key === "header_url") && typeof value === "string" ? (
              <a
                href={value}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex max-w-full items-start gap-1 break-all font-mono text-primary hover:underline"
              >
                <span className="min-w-0 break-all">{value}</span>
                <ExternalLinkIcon className="mt-0.5 size-3 shrink-0" />
              </a>
            ) : (
              <div className="flex min-w-0 items-start gap-2">
                <span
                  className={
                    key === "raw"
                      ? "block min-w-0 flex-1 whitespace-pre-wrap break-words font-mono text-xs leading-relaxed"
                      : "min-w-0 flex-1 break-words font-mono text-sm leading-relaxed"
                  }
                >
                  {formatValue(value)}
                </span>
                {(typeof value === "string" || typeof value === "number") && (
                  <CopyButton value={String(value)} />
                )}
              </div>
            )}
          </dd>
        </div>
      ))}
    </dl>
  )
}

function AsnIntelSections({ asn }: { asn?: Record<string, unknown> | null }) {
  const routing =
    asn?.routing && typeof asn.routing === "object"
      ? (asn.routing as Record<string, unknown>)
      : null
  const peeringdb =
    asn?.peeringdb && typeof asn.peeringdb === "object"
      ? (asn.peeringdb as Record<string, unknown>)
      : null

  return (
    <div className="flex flex-col gap-4">
      <DataGrid data={asn} fields={ASN_CORE_FIELDS} />
      {routing ? (
        <div className="border-t border-border/50 pt-4">
          <h3 className="mb-3 text-sm font-semibold">BGP routing</h3>
          <DataGrid data={routing} fields={ASN_ROUTING_FIELDS} />
        </div>
      ) : null}
      {peeringdb ? (
        <div className="border-t border-border/50 pt-4">
          <h3 className="mb-3 text-sm font-semibold">PeeringDB</h3>
          <DataGrid data={peeringdb} fields={PEERINGDB_FIELDS} />
        </div>
      ) : null}
    </div>
  )
}

function severityStyles(severity: string) {
  switch (severity) {
    case "critical":
      return "border-red-500/30 bg-red-500/10 text-red-600 dark:text-red-400"
    case "warning":
      return "border-amber-500/30 bg-amber-500/10 text-amber-600 dark:text-amber-400"
    default:
      return "border-blue-500/20 bg-blue-500/5 text-foreground"
  }
}

export function IntelPanels({
  snapshotId,
  targetId,
  raw,
  changes,
  changeDetails,
  pwhois,
}: {
  snapshotId?: string
  targetId?: string
  raw?: RawIntel | null
  changes?: string[]
  changeDetails?: ChangeDetail[]
  pwhois?: Record<string, unknown> | null
}) {
  const errorCount = raw?.errors?.length ?? 0
  const changeCount = changeDetails?.length ?? changes?.length ?? 0
  const ctCount =
    typeof raw?.ct?.count === "number"
      ? raw.ct.count
      : Array.isArray(raw?.ct?.subdomains)
        ? raw.ct.subdomains.length
        : 0
  const tracerouteVantages = Array.isArray(raw?.traceroute?.vantages)
    ? raw.traceroute.vantages
    : []
  const tracerouteCount = Array.isArray(raw?.traceroute?.hops)
    ? raw.traceroute.hops.length
    : tracerouteVantages.reduce((count, vantage) => {
        if (!vantage || typeof vantage !== "object") return count
        const hops = (vantage as Record<string, unknown>).hops
        return count + (Array.isArray(hops) ? hops.length : 0)
      }, 0)
  const hasScreenshot =
    typeof raw?.screenshot?.thumbnail === "string" ||
    typeof raw?.screenshot?.error === "string"

  const { resolvedTheme } = useTheme()
  const isDark = resolvedTheme === "dark"
  const [diffData, setDiffData] = React.useState<{ current: any, previous: any } | null>(null)
  const [diffLoading, setDiffLoading] = React.useState(false)

  async function loadDiff() {
    if (!snapshotId || diffData || diffLoading) return
    setDiffLoading(true)
    try {
      const res = await fetchApi<any>(`/api/snapshots/${snapshotId}/diff`)
      setDiffData(res)
    } catch (e) {
      console.error(e)
    } finally {
      setDiffLoading(false)
    }
  }

  return (
    <Tabs defaultValue="overview" className="w-full">
      <TabsList className="h-auto w-full flex-wrap justify-start gap-1 bg-muted/50 p-1 sm:w-auto">
        <TabsTrigger value="overview">Overview</TabsTrigger>
        <TabsTrigger value="whois">WHOIS</TabsTrigger>
        <TabsTrigger value="asn">ASN / BGP</TabsTrigger>
        <TabsTrigger value="dns">DNS</TabsTrigger>
        <TabsTrigger value="tls">TLS</TabsTrigger>
        <TabsTrigger value="web">Web</TabsTrigger>
        <TabsTrigger value="favicon">Favicon</TabsTrigger>
        <TabsTrigger value="crawl">Crawl</TabsTrigger>
        <TabsTrigger value="ct" className="gap-1.5">
          CT
          {ctCount > 0 ? (
            <Badge variant="secondary" className="size-5 justify-center p-0 text-[10px]">
              {ctCount}
            </Badge>
          ) : null}
        </TabsTrigger>
        <TabsTrigger value="traceroute" className="gap-1.5">
          Traceroute
          {tracerouteCount > 0 ? (
            <Badge variant="secondary" className="size-5 justify-center p-0 text-[10px]">
              {tracerouteCount}
            </Badge>
          ) : null}
        </TabsTrigger>
        <TabsTrigger value="screenshots" className="gap-1.5">
          <CameraIcon className="size-3" />
          Screenshots
          {hasScreenshot ? (
            <Badge variant="secondary" className="size-5 justify-center p-0 text-[10px]">
              1
            </Badge>
          ) : null}
        </TabsTrigger>
        <TabsTrigger value="storage">Storage</TabsTrigger>
        <TabsTrigger value="enrichment">Enrichment</TabsTrigger>
        <TabsTrigger value="pwhois">Submitter</TabsTrigger>
        <TabsTrigger value="errors" className="gap-1.5">
          Errors
          {errorCount > 0 ? (
            <Badge variant="destructive" className="size-5 justify-center p-0 text-[10px]">
              {errorCount}
            </Badge>
          ) : null}
        </TabsTrigger>
        <TabsTrigger value="changes" className="gap-1.5">
          Changes
          {changeCount > 0 ? (
            <Badge variant="secondary" className="size-5 justify-center p-0 text-[10px]">
              {changeCount}
            </Badge>
          ) : null}
        </TabsTrigger>
        {snapshotId && (
          <TabsTrigger value="raw-diff" onClick={loadDiff} className="gap-1.5">
            <GitCompareIcon className="size-3" />
            Raw Diff
          </TabsTrigger>
        )}
      </TabsList>

      <TabsContent value="overview" className="mt-4">
        <div className="grid items-start gap-4 md:grid-cols-2">
          <Card className="border-border/60 bg-card/80 backdrop-blur-sm">
            <CardHeader className="pb-3">
              <CardTitle className="flex items-center gap-2 text-base">
                <GlobeIcon className="size-4 text-primary" />
                WHOIS
              </CardTitle>
            </CardHeader>
            <CardContent>
              <DataGrid data={raw?.whois} fields={WHOIS_FIELDS} />
            </CardContent>
          </Card>
          <Card className="border-border/60 bg-card/80 backdrop-blur-sm">
            <CardHeader className="pb-3">
              <CardTitle className="flex items-center gap-2 text-base">
                <NetworkIcon className="size-4 text-primary" />
                ASN / BGP
              </CardTitle>
            </CardHeader>
            <CardContent>
              <AsnIntelSections asn={raw?.asn} />
            </CardContent>
          </Card>
          <Card className="border-border/60 bg-card/80 backdrop-blur-sm">
            <CardHeader className="pb-3">
              <CardTitle className="flex items-center gap-2 text-base">
                <MonitorIcon className="size-4 text-primary" />
                Web
              </CardTitle>
              {raw?.web?.url && typeof raw.web.url === "string" ? (
                <CardDescription className="truncate font-mono text-xs">
                  {raw.web.url}
                </CardDescription>
              ) : null}
            </CardHeader>
            <CardContent>
              <DataGrid data={raw?.web} fields={WEB_FIELDS} />
            </CardContent>
          </Card>
          <Card className="border-border/60 bg-card/80 backdrop-blur-sm md:col-span-2">
            <CardHeader className="pb-3">
              <CardTitle className="flex items-center gap-2 text-base">
                <ShieldIcon className="size-4 text-primary" />
                TLS Certificate
              </CardTitle>
            </CardHeader>
            <CardContent>
              <DataGrid data={raw?.tls} fields={TLS_FIELDS} />
            </CardContent>
          </Card>
        </div>
      </TabsContent>

      <TabsContent value="whois" className="mt-4">
        <Card className="border-border/60 bg-card/80 backdrop-blur-sm">
          <CardContent className="pt-6">
            <DataGrid data={raw?.whois} exclude={["raw"]} />
            {typeof raw?.whois?.raw === "string" ? (
              <details className="group mt-6 border-t pt-6">
                <summary className="mb-3 flex w-fit cursor-pointer list-none items-center gap-2 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground">
                  <ChevronRightIcon className="size-4 transition-transform group-open:rotate-90" />
                  View Raw WHOIS Data
                </summary>
                <div className="overflow-hidden rounded-lg border bg-muted/40">
                  <pre className="max-h-96 overflow-auto p-4 font-mono text-xs leading-relaxed whitespace-pre-wrap">
                    {raw.whois.raw}
                  </pre>
                </div>
                <div className="mt-2 flex justify-end">
                  <CopyButton value={raw.whois.raw as string} />
                </div>
              </details>
            ) : null}
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="asn" className="mt-4">
        <Card className="border-border/60 bg-card/80 backdrop-blur-sm">
          <CardContent className="pt-6">
            <AsnIntelSections asn={raw?.asn} />
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="dns" className="mt-4">
        <Card className="border-border/60 bg-card/80 backdrop-blur-sm">
          <CardContent className="pt-6">
            <DataGrid data={raw?.dns} fields={DNS_FIELDS} />
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="tls" className="mt-4">
        <Card className="border-border/60 bg-card/80 backdrop-blur-sm">
          <CardContent className="pt-6">
            <DataGrid data={raw?.tls} fields={TLS_FIELDS} />
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="favicon" className="mt-4">
        <Card className="border-border/60 bg-card/80 backdrop-blur-sm">
          <CardContent className="pt-6">
            <DataGrid data={raw?.favicon} fields={FAVICON_FIELDS} />
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="crawl" className="mt-4">
        <Card className="border-border/60 bg-card/80 backdrop-blur-sm">
          <CardContent className="pt-6">
            <DataGrid data={raw?.crawl} fields={CRAWL_FIELDS} />
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="ct" className="mt-4">
        <Card className="border-border/60 bg-card/80 backdrop-blur-sm">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              <ScrollTextIcon className="size-4 text-primary" />
              Certificate transparency
            </CardTitle>
            <CardDescription>
              Subdomains discovered via crt.sh. Click a hostname to scan it as a new target.
            </CardDescription>
          </CardHeader>
          <CardContent>
            {typeof raw?.ct?.skipped === "string" ? (
              <p className="text-sm text-muted-foreground">{raw.ct.skipped}</p>
            ) : (
              <DataGrid data={raw?.ct} fields={CT_FIELDS} />
            )}
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="traceroute" className="mt-4">
        <Card className="border-border/60 bg-card/80 backdrop-blur-sm">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              <RouteIcon className="size-4 text-primary" />
              Network path
            </CardTitle>
            <CardDescription>
              Hop-by-hop path from the scanner host to the resolved destination.
            </CardDescription>
          </CardHeader>
          <CardContent>
            {typeof raw?.traceroute?.skipped === "string" ? (
              <p className="text-sm text-muted-foreground">{raw.traceroute.skipped}</p>
            ) : (
              <>
                <DataGrid
                  data={raw?.traceroute}
                  fields={TRACEROUTE_FIELDS}
                  exclude={["hops", "vantages"]}
                />
                {tracerouteVantages.length > 0 ? (
                  <TracerouteVantageList vantages={tracerouteVantages} />
                ) : (
                  <div className="mt-6 border-t border-border/50 pt-6">
                    <TracerouteHopList hops={raw?.traceroute?.hops} />
                  </div>
                )}
              </>
            )}
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="screenshots" className="mt-4">
        <Card className="border-border/60 bg-card/80 backdrop-blur-sm">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              <CameraIcon className="size-4 text-primary" />
              Visual timeline
            </CardTitle>
            <CardDescription>
              JPEG thumbnails captured via browserless Chrome during each scan.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            {hasScreenshot ? (
              <DataGrid data={raw?.screenshot} fields={SCREENSHOT_FIELDS} />
            ) : null}
            {targetId ? (
              <ScreenshotTimeline targetId={targetId} />
            ) : hasScreenshot && typeof raw?.screenshot?.thumbnail === "string" ? (
              <div className="overflow-hidden rounded-md border border-border/50">
                <img
                  src={`data:image/jpeg;base64,${raw.screenshot.thumbnail}`}
                  alt="Latest screenshot thumbnail"
                  className="h-auto w-full max-w-xl object-cover"
                />
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">
                No screenshots captured for this snapshot.
              </p>
            )}
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="storage" className="mt-4">
        <Card className="border-border/60 bg-card/80 backdrop-blur-sm">
          <CardContent className="pt-6">
            <DataGrid data={raw?.storage} fields={STORAGE_FIELDS} />
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="web" className="mt-4">
        <Card className="border-border/60 bg-card/80 backdrop-blur-sm">
          <CardContent className="pt-6">
            <DataGrid data={raw?.web} fields={WEB_FIELDS} />
            {raw?.web?.url && typeof raw.web.url === "string" ? (
              <div className="mt-4">
                <Button
                  variant="outline"
                  size="sm"
                  render={
                    <Link
                      href={raw.web.url}
                      target="_blank"
                      rel="noopener noreferrer"
                    />
                  }
                  nativeButton={false}
                >
                  <ExternalLinkIcon data-icon="inline-start" />
                  Open page
                </Button>
              </div>
            ) : null}
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="enrichment" className="mt-4">
        <Card className="border-border/60 bg-card/80 backdrop-blur-sm">
          <CardHeader>
            <CardTitle className="text-base">Third-party enrichment</CardTitle>
            <CardDescription>
              Async Shodan, Censys, HIBP, RiskIQ, Wayback, and VirusTotal correlation.
            </CardDescription>
          </CardHeader>
          <CardContent>
            {raw?.enrichment &&
            Object.keys(raw.enrichment).length > 0 &&
            (raw.enrichment as Record<string, unknown>).status !== "pending" ? (
              <DataGrid data={raw.enrichment} fields={ENRICHMENT_FIELDS} />
            ) : (raw?.enrichment as Record<string, unknown> | undefined)?.status ===
              "pending" ? (
              <p className="text-sm text-muted-foreground">
                Enrichment is running asynchronously. This page refreshes automatically when complete.
              </p>
            ) : (
              <p className="text-sm text-muted-foreground">
                No enrichment data yet. Configure API keys in Settings and rescan.
              </p>
            )}
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="pwhois" className="mt-4">
        <Card className="border-border/60 bg-card/80 backdrop-blur-sm">
          <CardHeader>
            <CardTitle className="text-base">Submitter enrichment</CardTitle>
            <CardDescription>
              pWhois data for the IP address that submitted this scan.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <DataGrid data={pwhois} />
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="errors" className="mt-4">
        <Card className="border-border/60 bg-card/80 backdrop-blur-sm">
          <CardContent className="pt-6">
            {raw?.errors && raw.errors.length > 0 ? (
              <ul className="flex flex-col gap-2">
                {raw.errors.map((error) => (
                  <li
                    key={error}
                    className="flex items-start gap-2 rounded-lg border border-destructive/20 bg-destructive/5 px-3 py-2 text-sm text-destructive"
                  >
                    <AlertCircleIcon className="mt-0.5 size-4 shrink-0" />
                    {error}
                  </li>
                ))}
              </ul>
            ) : (
              <p className="text-sm text-muted-foreground">
                All gatherers completed without errors.
              </p>
            )}
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="changes" className="mt-4">
        <Card className="border-border/60 bg-card/80 backdrop-blur-sm">
          <CardContent className="pt-6">
            {changeDetails && changeDetails.length > 0 ? (
              <ul className="flex flex-col gap-2">
                {changeDetails.map((entry, index) => (
                  <li
                    key={`${entry.type}-${entry.summary}-${index}`}
                    className={`flex flex-col gap-1 rounded-lg border px-3 py-2 text-sm ${severityStyles(entry.severity)}`}
                  >
                    <div className="flex flex-wrap items-center gap-2">
                      <Badge variant="outline" className="font-mono text-[10px] uppercase">
                        {entry.type.replaceAll("_", " ")}
                      </Badge>
                      <Badge
                        variant={entry.severity === "critical" ? "destructive" : "secondary"}
                        className="text-[10px] uppercase"
                      >
                        {entry.severity}
                      </Badge>
                      <span className="font-medium">{entry.summary}</span>
                    </div>
                    {entry.detail ? (
                      <p className="font-mono text-xs text-muted-foreground">{entry.detail}</p>
                    ) : null}
                  </li>
                ))}
              </ul>
            ) : changes && changes.length > 0 ? (
              <ul className="flex flex-col gap-2">
                {changes.map((change) => {
                  const isAdded = change.startsWith("added ")
                  const isRemoved = change.startsWith("removed ")
                  const isChanged = change.startsWith("changed ")

                  return (
                    <li
                      key={change}
                      className={`flex items-center gap-2 rounded-lg border px-3 py-2 font-mono text-sm ${
                        isAdded ? "border-green-500/20 bg-green-500/10 text-green-600 dark:text-green-400" :
                        isRemoved ? "border-red-500/20 bg-red-500/10 text-red-600 dark:text-red-400" :
                        isChanged ? "border-amber-500/20 bg-amber-500/10 text-amber-600 dark:text-amber-400" :
                        "bg-muted/30"
                      }`}
                    >
                      {isAdded && <span className="font-bold">+</span>}
                      {isRemoved && <span className="font-bold">-</span>}
                      {isChanged && <span className="font-bold">~</span>}
                      {change}
                    </li>
                  )
                })}
              </ul>
            ) : (
              <p className="text-sm text-muted-foreground">
                No changes since the previous snapshot.
              </p>
            )}
          </CardContent>
        </Card>
      </TabsContent>

      {snapshotId && (
        <TabsContent value="raw-diff" className="mt-4">
          <Card className="border-border/60 bg-card/80 backdrop-blur-sm">
            <CardContent className="pt-6">
              {diffLoading ? (
                <p className="text-muted-foreground text-sm">Loading diff...</p>
              ) : diffData ? (
                diffData.previous ? (
                  <div className="overflow-hidden rounded-lg border bg-muted/40">
                    <ReactDiffViewer
                      oldValue={JSON.stringify(diffData.previous, null, 2)}
                      newValue={JSON.stringify(diffData.current, null, 2)}
                      splitView={true}
                      useDarkTheme={isDark}
                      hideLineNumbers={false}
                    />
                  </div>
                ) : (
                  <p className="text-muted-foreground text-sm">No previous snapshot to compare against.</p>
                )
              ) : (
                <p className="text-muted-foreground text-sm">Select to load diff.</p>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      )}
    </Tabs>
  )
}