"use client"

import * as React from "react"
import Link from "next/link"

import { Badge } from "@/components/ui/badge"
import { formatDistanceToNow } from "date-fns"
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { HelpTip } from "@/components/help-tip"
import { IntelSummary } from "@/components/intel-summary"
import { IntelPanels } from "@/components/intel-panels"
import { TargetSnapshotsTable } from "@/components/target-snapshots-table"
import { TargetSnapshotsChart } from "@/components/target-snapshots-chart"
import { TargetTags } from "@/components/target-tags"
import { CountryFlag } from "@/components/country-flag"
import { Button } from "@/components/ui/button"
import { InvestigationNotes } from "@/components/investigation-notes"
import { RescanTargetButton } from "@/components/rescan-target-button"
import { extractIntel } from "@/lib/intel"
import type { ScanResponse, TargetDetail } from "@/lib/types"
import {
  GlobeIcon,
  CalendarIcon,
  HistoryIcon,
  ExternalLinkIcon,
  RadarIcon,
  AlertCircleIcon,
} from "lucide-react"

export function TargetDetailView({
  target,
  refreshKey = 0,
  onRescanned,
}: {
  target: TargetDetail
  refreshKey?: number
  onRescanned?: (result: ScanResponse) => void | Promise<void>
}) {
  const latest = target.latest_snapshot
  const intel = latest
    ? extractIntel(latest.raw_data, latest)
    : extractIntel({ host: target.host })

  const latestCountry =
    intel.country || latest?.pwhois_country_code || null

  const warnings = []
  if (intel.expirationDate) {
    const daysToExpiry = (new Date(intel.expirationDate).getTime() - Date.now()) / (1000 * 60 * 60 * 24)
    if (daysToExpiry > 0 && daysToExpiry <= 30) {
      warnings.push(`Expires in ${Math.round(daysToExpiry)} days`)
    }
  }
  if (intel.domainStatus) {
    if (intel.domainStatus.some((s: string) => s.toLowerCase().includes("hold"))) {
      warnings.push("Domain on HOLD")
    }
  }

  return (
    <div className="flex flex-col gap-8">
      <Card className="overflow-hidden border-border/60 bg-card/90 backdrop-blur-sm">
        <CardHeader className="border-b border-border/50 bg-muted/20">
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div className="flex flex-col gap-2">
              <CardTitle className="flex items-center gap-2 text-xl md:text-2xl">
                <GlobeIcon className="size-5 text-primary" />
                {target.host}
              </CardTitle>
              <CardDescription className="font-mono text-xs">
                {target.id}
              </CardDescription>
              <TargetTags targetId={target.id} initialTags={target.tags} />
            </div>
            <div className="flex flex-wrap items-center gap-2">
              {warnings.map(w => (
                <Badge key={w} variant="destructive" className="font-mono gap-1.5">
                  <AlertCircleIcon className="size-3" />
                  {w}
                </Badge>
              ))}
              {intel.asn ? (
                <Badge variant="secondary" className="font-mono">
                  AS{intel.asn}
                </Badge>
              ) : null}
              {latestCountry ? (
                <Badge variant="outline" className="gap-1.5">
                  <CountryFlag code={latestCountry} showCode />
                </Badge>
              ) : null}
              <Badge variant="outline">
                {target.snapshot_count} snapshot
                {target.snapshot_count === 1 ? "" : "s"}
              </Badge>
            </div>
          </div>
        </CardHeader>
        <CardContent className="grid gap-4 pt-6 sm:grid-cols-2 lg:grid-cols-4">
          <div className="flex flex-col gap-1">
            <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
              Added
            </span>
            <span
              className="flex items-center gap-1.5 text-sm font-medium"
              title={new Date(target.created_at).toLocaleString()}
            >
              <CalendarIcon className="size-3.5 text-primary/70" />
              {formatDistanceToNow(new Date(target.created_at), { addSuffix: true })}
            </span>
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
              Latest scan
            </span>
            <span
              className="flex items-center gap-1.5 text-sm font-medium"
              title={target.latest_snapshot_at ? new Date(target.latest_snapshot_at).toLocaleString() : undefined}
            >
              <HistoryIcon className="size-3.5 text-primary/70" />
              {target.latest_snapshot_at
                ? formatDistanceToNow(new Date(target.latest_snapshot_at), { addSuffix: true })
                : "—"}
            </span>
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
              Latest ASN
            </span>
            <span className="font-mono text-sm text-muted-foreground">
              {target.latest_asn
                ? `AS${target.latest_asn}${target.latest_as_name ? ` · ${target.latest_as_name}` : ""}`
                : "—"}
            </span>
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
              Web title
            </span>
            <span className="truncate text-sm text-muted-foreground">
              {target.latest_web_title || "—"}
            </span>
          </div>
        </CardContent>
        <CardFooter className="flex flex-wrap gap-3 border-t border-border/50 bg-muted/10">
          <RescanTargetButton host={target.host} onSuccess={onRescanned} />
          {latest ? (
            <>
              <Button
                render={<Link href={`/snapshot?id=${latest.id}`} />}
                nativeButton={false}
                data-testid="target-view-latest-snapshot"
              >
                <ExternalLinkIcon data-icon="inline-start" />
                View latest snapshot
              </Button>
              <Button
                variant="outline"
                render={<Link href={`/target/snapshots?id=${target.id}`} />}
                nativeButton={false}
              >
                <HistoryIcon data-icon="inline-start" />
                All snapshots
              </Button>
            </>
          ) : null}
        </CardFooter>
      </Card>

      {latest ? (
        <>
          <div className="flex flex-col gap-3">
            <div className="flex items-center gap-2">
              <RadarIcon className="size-4 text-primary" />
              <h2 className="flex items-center gap-2 font-heading text-lg font-semibold tracking-tight">
                Latest intelligence
                <HelpTip id="snapshot.intel_summary" />
              </h2>
              <Badge variant="secondary" className="font-mono text-xs" title={new Date(latest.scanned_at).toLocaleString()}>
                {formatDistanceToNow(new Date(latest.scanned_at), { addSuffix: true })}
              </Badge>
            </div>
            <IntelSummary intel={intel} />
          </div>

          <IntelPanels
            key={`intel-${refreshKey}`}
            snapshotId={latest.id}
            targetId={target.id}
            raw={latest.raw_data}
            changes={latest.changes}
            changeDetails={latest.change_details}
            pwhois={latest.pwhois_data}
          />
        </>
      ) : (
        <Card className="border-dashed border-border/60 bg-card/60">
          <CardContent className="flex flex-col items-center gap-3 py-12 text-center">
            <RadarIcon className="size-8 text-muted-foreground" />
            <p className="text-muted-foreground">
              No snapshots yet for this target. Run a rescan to collect intelligence.
            </p>
            <RescanTargetButton host={target.host} onSuccess={onRescanned} />
          </CardContent>
        </Card>
      )}

      <InvestigationNotes
        targetId={target.id}
        snapshotId={latest?.id}
        compact
      />

      <div className="flex flex-col gap-4">
        <h2 className="font-heading text-lg font-semibold tracking-tight">
          Snapshot history
        </h2>
        <TargetSnapshotsChart
          key={`chart-${refreshKey}`}
          targetId={target.id}
        />
        <TargetSnapshotsTable
          key={`table-${refreshKey}`}
          targetId={target.id}
        />
      </div>
    </div>
  )
}