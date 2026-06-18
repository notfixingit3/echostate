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
  ASN_FIELDS,
  WEB_FIELDS,
  WHOIS_FIELDS,
  PWHOIS_FIELDS,
  TLS_FIELDS,
  DNS_FIELDS,
  formatFieldLabel,
  formatValue,
} from "@/lib/intel"
import type { RawIntel } from "@/lib/intel"
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
} from "lucide-react"
import ReactDiffViewer from "react-diff-viewer-continued"
import { useTheme } from "next-themes"
import { fetchApi } from "@/lib/api"

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
    <dl className="grid grid-cols-1 gap-3 sm:grid-cols-[minmax(8rem,auto)_1fr] sm:gap-x-6 sm:gap-y-3">
      {entries.map(([key, value]) => (
        <React.Fragment key={key}>
          <dt className="text-sm font-medium text-muted-foreground">
            {formatFieldLabel(key)}
          </dt>
          <dd className="text-sm break-words">
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
            key === "tech_stack" ? (
              <div className="flex flex-wrap gap-1.5">
                {(Array.isArray(value) ? value : [value]).map((item) => (
                  <Badge key={String(item)} variant="outline" className="font-mono text-xs font-normal">
                    {String(item)}
                  </Badge>
                ))}
              </div>
            ) : key === "url" && typeof value === "string" ? (
              <a
                href={value}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1 font-mono text-primary hover:underline"
              >
                {value}
                <ExternalLinkIcon className="size-3" />
              </a>
            ) : (
              <div className="flex items-center gap-1">
                <span className={key === "raw" ? "block whitespace-pre-wrap font-mono text-xs leading-relaxed" : "font-mono"}>
                  {formatValue(value)}
                </span>
                {(typeof value === "string" || typeof value === "number") && (
                  <CopyButton value={String(value)} />
                )}
              </div>
            )}
          </dd>
        </React.Fragment>
      ))}
    </dl>
  )
}

export function IntelPanels({
  snapshotId,
  raw,
  changes,
  pwhois,
}: {
  snapshotId?: string
  raw?: RawIntel | null
  changes?: string[]
  pwhois?: Record<string, unknown> | null
}) {
  const errorCount = raw?.errors?.length ?? 0
  const changeCount = changes?.length ?? 0

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
        <div className="grid gap-4 lg:grid-cols-3">
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
              <DataGrid data={raw?.asn} fields={ASN_FIELDS} />
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
          <Card className="border-border/60 bg-card/80 backdrop-blur-sm lg:col-span-3">
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
            <DataGrid data={raw?.asn} fields={ASN_FIELDS} />
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
            {changes && changes.length > 0 ? (
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