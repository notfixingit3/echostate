"use client"

import * as React from "react"
import Link from "next/link"

import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Label } from "@/components/ui/label"
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
import { HelpTip, LabelWithHelp } from "@/components/help-tip"
import { IntelSummary } from "@/components/intel-summary"
import { IntelPanels } from "@/components/intel-panels"
import { CountryFlag } from "@/components/country-flag"
import {
  createReport,
  fetchApi,
  getReport,
  getLatestReportForSnapshot,
  downloadReport,
  ApiError,
} from "@/lib/api"
import { extractIntel } from "@/lib/intel"
import type { Snapshot, Report } from "@/lib/types"
import {
  GlobeIcon,
  CalendarIcon,
  ClockIcon,
  AlertCircleIcon,
  FileTextIcon,
  FileDownIcon,
  Loader2Icon,
  DownloadIcon,
} from "lucide-react"

export function SnapshotDetail({ snapshot: initialSnapshot }: { snapshot: Snapshot }) {
  const [snapshot, setSnapshot] = React.useState(initialSnapshot)
  const [isCreatingReport, setIsCreatingReport] = React.useState(false)
  const [reportError, setReportError] = React.useState<string | null>(null)
  const [report, setReport] = React.useState<Report | null>(null)
  const [waitForEnrichment, setWaitForEnrichment] = React.useState(false)

  React.useEffect(() => {
    setSnapshot(initialSnapshot)
  }, [initialSnapshot])

  const intel = extractIntel(snapshot.raw_data, snapshot)

  React.useEffect(() => {
    if (intel.enrichmentStatus !== "pending") {
      return
    }

    let cancelled = false
    const intervalId = setInterval(() => {
      void fetchApi<Snapshot>(`/api/snapshots/${snapshot.id}`)
        .then((updated) => {
          if (!cancelled) setSnapshot(updated)
        })
        .catch(() => {
          // Keep polling; transient API errors should not stop enrichment refresh.
        })
    }, 2000)

    return () => {
      cancelled = true
      clearInterval(intervalId)
    }
  }, [snapshot.id, intel.enrichmentStatus])

  React.useEffect(() => {
    let cancelled = false

    async function hydrateLatestReport() {
      try {
        const latest = await getLatestReportForSnapshot(snapshot.id)
        if (!cancelled && latest) {
          setReport(latest)
          setReportError(null)
        }
      } catch {
        // Leave create-report flow available if hydration fails.
      }
    }

    void hydrateLatestReport()

    return () => {
      cancelled = true
    }
  }, [snapshot.id])

  const handleCreateReport = async () => {
    setIsCreatingReport(true)
    setReportError(null)
    setReport(null)

    try {
      const created = await createReport(snapshot.id, waitForEnrichment)
      setReport(created)
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

  React.useEffect(() => {
    if (!report || report.status === "completed" || report.status === "failed") {
      return
    }

    let cancelled = false
    let intervalId: ReturnType<typeof setInterval> | null = null

    async function refresh() {
      try {
        const updated = await getReport(report!.id)
        if (cancelled) return
        setReport(updated)
        setReportError(null)
        if (
          updated.status !== "pending" &&
          updated.status !== "running" &&
          intervalId
        ) {
          clearInterval(intervalId)
          intervalId = null
        }
      } catch (err) {
        if (cancelled) return
        if (err instanceof ApiError) {
          setReportError(err.message)
        } else if (err instanceof Error) {
          setReportError(err.message)
        } else {
          setReportError("Failed to check report status")
        }
        if (intervalId) {
          clearInterval(intervalId)
          intervalId = null
        }
      }
    }

    void refresh()
    intervalId = setInterval(refresh, 2000)

    return () => {
      cancelled = true
      if (intervalId) clearInterval(intervalId)
    }
  }, [report])

  const handleDownloadReport = async () => {
    if (!report || report.status !== "completed") return

    try {
      await downloadReport(
        report,
        `echostate-${intel.host || "report"}-${report.id.slice(0, 8)}.pdf`
      )
      setReportError(null)
    } catch (err) {
      if (err instanceof ApiError) {
        setReportError(err.message)
      } else if (err instanceof Error) {
        setReportError(err.message)
      } else {
        setReportError("Failed to download report")
      }
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
              {intel.enrichmentStatus === "pending" ? (
                <Badge variant="secondary" className="gap-1">
                  <Loader2Icon className="size-3 animate-spin" />
                  Enrichment pending
                </Badge>
              ) : null}
              {intel.enrichmentStatus === "partial" ? (
                <Badge variant="outline">Enrichment partial</Badge>
              ) : null}
            </div>
          </div>
        </CardHeader>
        <CardContent className="grid gap-4 pt-6 sm:grid-cols-2 xl:grid-cols-4">
          <div className="flex flex-col gap-1">
            <span className="flex items-center gap-1.5 text-xs font-medium uppercase tracking-wider text-muted-foreground">
              Scanned at
              <HelpTip id="snapshot.scanned_at" />
            </span>
            <span className="flex items-center gap-1.5 text-sm font-medium">
              <CalendarIcon className="size-3.5 text-primary/70" />
              {new Date(snapshot.scanned_at).toLocaleString()}
            </span>
          </div>
          <div className="flex flex-col gap-1">
            <span className="flex items-center gap-1.5 text-xs font-medium uppercase tracking-wider text-muted-foreground">
              Last seen
              <HelpTip id="snapshot.last_seen" />
            </span>
            <span className="flex items-center gap-1.5 text-sm font-medium">
              <ClockIcon className="size-3.5 text-primary/70" />
              {new Date(snapshot.last_seen).toLocaleString()}
            </span>
          </div>
          <div className="flex flex-col gap-1">
            <span className="flex items-center gap-1.5 text-xs font-medium uppercase tracking-wider text-muted-foreground">
              Target
              <HelpTip id="snapshot.target" />
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
            <span className="flex items-center gap-1.5 text-xs font-medium uppercase tracking-wider text-muted-foreground">
              Data hash
              <HelpTip id="snapshot.data_hash" />
            </span>
            <span className="break-all font-mono text-xs text-muted-foreground">
              {snapshot.data_hash}
            </span>
          </div>
        </CardContent>
        <Separator />
        <CardFooter className="flex flex-col items-start gap-3">
          <div className="flex items-center gap-2 text-sm font-medium">
            PDF report
            <HelpTip id="snapshot.create_report" />
          </div>
          <div className="flex items-center gap-2">
            <Checkbox
              id="wait-for-enrichment"
              checked={waitForEnrichment}
              onCheckedChange={(checked) => setWaitForEnrichment(checked)}
            />
            <Label htmlFor="wait-for-enrichment" className="text-sm font-normal">
              <LabelWithHelp
                label="Wait for enrichment before generating PDF"
                helpId="snapshot.wait_for_enrichment"
              />
            </Label>
          </div>
          <div className="flex w-full flex-wrap justify-between gap-3">
            <div className="flex flex-wrap gap-2">
              {!report || report.status === "failed" ? (
                <Button
                  disabled={isCreatingReport}
                  onClick={() => void handleCreateReport()}
                  data-testid="snapshot-create-report-button"
                >
                  {isCreatingReport ? (
                    <>
                      <Loader2Icon
                        data-icon="inline-start"
                        className="animate-spin"
                      />
                      Creating report…
                    </>
                  ) : (
                    <>
                      <FileTextIcon data-icon="inline-start" />
                      Create report
                    </>
                  )}
                </Button>
              ) : report.status === "completed" ? (
                <>
                  <Button
                    variant="outline"
                    render={<Link href={`/report?id=${report.id}`} />}
                    nativeButton={false}
                    data-testid="snapshot-view-report-link"
                  >
                    <FileTextIcon data-icon="inline-start" />
                    View report
                  </Button>
                  <Button
                    onClick={() => void handleDownloadReport()}
                    data-testid="snapshot-download-report-button"
                  >
                    <DownloadIcon data-icon="inline-start" />
                    Download PDF
                  </Button>
                </>
              ) : (
                <Button disabled data-testid="snapshot-report-pending-button">
                  <Loader2Icon
                    data-icon="inline-start"
                    className="animate-spin"
                  />
                  Generating report…
                </Button>
              )}
            </div>
            <Button
              variant="outline"
              render={<Link href={`/reports?snapshot_id=${snapshot.id}`} />}
              nativeButton={false}
            >
              <FileDownIcon data-icon="inline-start" />
              View reports
            </Button>
          </div>
          {report &&
          (report.status === "pending" || report.status === "running") ? (
            <Alert data-testid="snapshot-report-status-alert">
              <Loader2Icon />
              <AlertTitle>Report in progress</AlertTitle>
              <AlertDescription>
                Report {report.id} is {report.status}. This updates automatically.
              </AlertDescription>
            </Alert>
          ) : null}
          {(reportError || (report && report.status === "failed")) && (
            <span className="flex items-center gap-1.5 text-sm text-destructive">
              <AlertCircleIcon className="size-4" />
              {report?.status === "failed"
                ? report.error || "Report generation failed."
                : reportError}
            </span>
          )}
        </CardFooter>
      </Card>

      <div className="flex flex-col gap-3">
        <h2 className="flex items-center gap-2 font-heading text-lg font-semibold tracking-tight">
          Intelligence summary
          <HelpTip id="snapshot.intel_summary" />
        </h2>
        <IntelSummary intel={intel} />
      </div>

      <IntelPanels
        snapshotId={snapshot.id}
        raw={snapshot.raw_data}
        changes={snapshot.changes}
        changeDetails={snapshot.change_details}
        pwhois={snapshot.pwhois_data}
      />
    </div>
  )
}