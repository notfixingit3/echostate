"use client"

import * as React from "react"
import Link from "next/link"

import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { IntelSummary } from "@/components/intel-summary"
import { IntelPanels } from "@/components/intel-panels"
import { TargetSnapshotsTable } from "@/components/target-snapshots-table"
import { CountryFlag } from "@/components/country-flag"
import { extractIntel } from "@/lib/intel"
import type { TargetDetail } from "@/lib/types"
import {
  GlobeIcon,
  CalendarIcon,
  HistoryIcon,
  ExternalLinkIcon,
  RadarIcon,
} from "lucide-react"

export function TargetDetailView({ target }: { target: TargetDetail }) {
  const latest = target.latest_snapshot
  const intel = latest
    ? extractIntel(latest.raw_data, latest)
    : extractIntel({ host: target.host })

  const latestCountry =
    intel.country || latest?.pwhois_country_code || null

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
            </div>
            <div className="flex flex-wrap items-center gap-2">
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
            <span className="flex items-center gap-1.5 text-sm font-medium">
              <CalendarIcon className="size-3.5 text-primary/70" />
              {new Date(target.created_at).toLocaleString()}
            </span>
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
              Latest scan
            </span>
            <span className="flex items-center gap-1.5 text-sm font-medium">
              <HistoryIcon className="size-3.5 text-primary/70" />
              {target.latest_snapshot_at
                ? new Date(target.latest_snapshot_at).toLocaleString()
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
        {latest ? (
          <CardFooter className="flex flex-wrap gap-3 border-t border-border/50 bg-muted/10">
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
              render={<Link href={`/targets/${target.id}/snapshots`} />}
              nativeButton={false}
            >
              <HistoryIcon data-icon="inline-start" />
              All snapshots
            </Button>
          </CardFooter>
        ) : null}
      </Card>

      {latest ? (
        <>
          <div className="flex flex-col gap-3">
            <div className="flex items-center gap-2">
              <RadarIcon className="size-4 text-primary" />
              <h2 className="font-heading text-lg font-semibold tracking-tight">
                Latest intelligence
              </h2>
              <Badge variant="secondary" className="font-mono text-xs">
                {new Date(latest.scanned_at).toLocaleString()}
              </Badge>
            </div>
            <IntelSummary intel={intel} />
          </div>

          <IntelPanels
            raw={latest.raw_data}
            changes={latest.changes}
            pwhois={latest.pwhois_data}
          />
        </>
      ) : (
        <Card className="border-dashed border-border/60 bg-card/60">
          <CardContent className="flex flex-col items-center gap-3 py-12 text-center">
            <RadarIcon className="size-8 text-muted-foreground" />
            <p className="text-muted-foreground">
              No snapshots yet for this target. Run a scan from the home page.
            </p>
            <Button render={<Link href="/" />} nativeButton={false}>
              Start a scan
            </Button>
          </CardContent>
        </Card>
      )}

      <div className="flex flex-col gap-4">
        <h2 className="font-heading text-lg font-semibold tracking-tight">
          Snapshot history
        </h2>
        <TargetSnapshotsTable targetId={target.id} />
      </div>
    </div>
  )
}