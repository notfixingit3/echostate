"use client"

import * as React from "react"
import Link from "next/link"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
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
} from "lucide-react"

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
            {key === "name_servers" || key === "status" || key === "copyrights" ? (
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
              <span className={key === "raw" ? "block whitespace-pre-wrap font-mono text-xs leading-relaxed" : "font-mono"}>
                {formatValue(value)}
              </span>
            )}
          </dd>
        </React.Fragment>
      ))}
    </dl>
  )
}

export function IntelPanels({
  raw,
  changes,
  pwhois,
}: {
  raw?: RawIntel | null
  changes?: string[]
  pwhois?: Record<string, unknown> | null
}) {
  const errorCount = raw?.errors?.length ?? 0
  const changeCount = changes?.length ?? 0

  return (
    <Tabs defaultValue="overview" className="w-full">
      <TabsList className="h-auto w-full flex-wrap justify-start gap-1 bg-muted/50 p-1 sm:w-auto">
        <TabsTrigger value="overview">Overview</TabsTrigger>
        <TabsTrigger value="whois">WHOIS</TabsTrigger>
        <TabsTrigger value="asn">ASN / BGP</TabsTrigger>
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
        </div>
      </TabsContent>

      <TabsContent value="whois" className="mt-4">
        <Card className="border-border/60 bg-card/80 backdrop-blur-sm">
          <CardContent className="pt-6">
            <DataGrid data={raw?.whois} exclude={["raw"]} />
            {typeof raw?.whois?.raw === "string" ? (
              <div className="mt-6 border-t pt-6">
                <h4 className="mb-3 text-sm font-medium text-muted-foreground">
                  Raw WHOIS
                </h4>
                <pre className="max-h-96 overflow-auto rounded-lg border bg-muted/40 p-4 font-mono text-xs leading-relaxed whitespace-pre-wrap">
                  {raw.whois.raw}
                </pre>
              </div>
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
                {changes.map((change) => (
                  <li
                    key={change}
                    className="rounded-lg border bg-muted/30 px-3 py-2 font-mono text-sm"
                  >
                    {change}
                  </li>
                ))}
              </ul>
            ) : (
              <p className="text-sm text-muted-foreground">
                No changes since the previous snapshot.
              </p>
            )}
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>
  )
}