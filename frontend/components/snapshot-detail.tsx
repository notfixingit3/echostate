"use client"

import * as React from "react"
import Link from "next/link"

import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert"
import { IntelSummary } from "@/components/intel-summary"
import { IntelPanels } from "@/components/intel-panels"
import { CountryFlag } from "@/components/country-flag"
import { fetchApi, ApiError } from "@/lib/api"
import { extractIntel } from "@/lib/intel"
import type { Snapshot, Report } from "@/lib/types"
import {
  GlobeIcon,
  CalendarIcon,
  ClockIcon,
  AlertCircleIcon,
  FileTextIcon,
  FileDownIcon,
  CheckCircleIcon,
} from "lucide-react"

type ReportResponse = Report

export function SnapshotDetail({ snapshot }: { snapshot: Snapshot }) {
  const [isCreatingReport, setIsCreatingReport] = React.useState(false)
  const [reportError, setReportError] = React.useState<string | null>(null)
  const [createdReport, setCreatedReport] = React.useState<ReportResponse | null>(
    null
  )

  const intel = extractIntel(snapshot.raw_data, snapshot)

  const handleCreateReport = async () => {
    setIsCreatingReport(true)
    setReportError(null)
    setCreatedReport(null)

    try {
      const report = await fetchApi<ReportResponse>("/api/reports", {
        method: "POST",
        body: JSON.stringify({ snapshot_id: snapshot.id }),
      })
      setCreatedReport(report)
    } catch (err) {
      if (err instanceof ApiError) {
        setReportError(err.message)
      } else if (err instanceof Error) {
        setReportError(err.message)
      } else {
        setReportError("Failed to create report")
      }
    } finally {
      setIsCreatingReport(false)
    }
  }

  return (
    <div className="flex flex-col gap-8">
      {reportError && (
        <Alert variant="destructive">
          <AlertCircleIcon />
          <AlertTitle>Report failed</AlertTitle>
          <AlertDescription>{reportError}</AlertDescription>
        </Alert>
      )}

      {createdReport && (
        <Alert>
          <CheckCircleIcon />
          <AlertTitle>Report queued</AlertTitle>
          <AlertDescription>
            Report {createdReport.id} is {createdReport.status}.
          </AlertDescription>
        </Alert>
      )}

      <Card className="overflow-hidden border-border/60 bg-card/90 backdrop-blur-sm">
        <CardHeader className="border-b border-border/50 bg-muted/20">
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div className="flex flex-col gap-2">
              <CardTitle className="flex items-center gap-2 text-xl md:text-2xl">
                <GlobeIcon className="size-5 text-primary" />
                {intel.host}
              </CardTitle>
              <CardDescription>
                Reconnaissance snapshot ·{" "}
                <span className="font-mono text-xs">{snapshot.id}</span>
              </CardDescription>
            </div>
            <div className="flex flex-wrap gap-2">
              {intel.asn ? (
                <Badge variant="secondary" className="font-mono">
                  AS{intel.asn}
                </Badge>
              ) : null}
              {intel.country ? (
                <Badge variant="outline" className="gap-1.5">
                  <CountryFlag code={intel.country} showCode />
                </Badge>
              ) : null}
              {intel.changeCount > 0 ? (
                <Badge variant="outline">
                  {intel.changeCount} change{intel.changeCount === 1 ? "" : "s"}
                </Badge>
              ) : null}
            </div>
          </div>
        </CardHeader>
        <CardContent className="grid gap-4 pt-6 sm:grid-cols-2 xl:grid-cols-4">
          <div className="flex flex-col gap-1">
            <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
              Scanned at
            </span>
            <span className="flex items-center gap-1.5 text-sm font-medium">
              <CalendarIcon className="size-3.5 text-primary/70" />
              {new Date(snapshot.scanned_at).toLocaleString()}
            </span>
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
              Last seen
            </span>
            <span className="flex items-center gap-1.5 text-sm font-medium">
              <ClockIcon className="size-3.5 text-primary/70" />
              {new Date(snapshot.last_seen).toLocaleString()}
            </span>
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
              Target
            </span>
            <Button
              variant="link"
              className="h-auto justify-start p-0 font-mono text-sm"
              render={
                <Link href={`/target/snapshots?id=${snapshot.target_id}`} />
              }
              nativeButton={false}
            >
              {snapshot.target_id}
            </Button>
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
              Data hash
            </span>
            <span className="break-all font-mono text-xs text-muted-foreground">
              {snapshot.data_hash}
            </span>
          </div>
        </CardContent>
        <Separator />
        <CardFooter className="flex flex-wrap justify-between gap-3">
          <Button
            disabled={isCreatingReport}
            onClick={handleCreateReport}
            data-testid="snapshot-create-report-button"
          >
            <FileTextIcon data-icon="inline-start" />
            {isCreatingReport ? "Creating report…" : "Create report"}
          </Button>
          <Button
            variant="outline"
            render={
              <Link href={`/reports?snapshot_id=${snapshot.id}`} />
            }
            nativeButton={false}
          >
            <FileDownIcon data-icon="inline-start" />
            View reports
          </Button>
        </CardFooter>
      </Card>

      <div className="flex flex-col gap-3">
        <h2 className="font-heading text-lg font-semibold tracking-tight">
          Intelligence summary
        </h2>
        <IntelSummary intel={intel} />
      </div>

      <IntelPanels
        snapshotId={snapshot.id}
        raw={snapshot.raw_data}
        changes={snapshot.changes}
        pwhois={snapshot.pwhois_data}
      />
    </div>
  )
}