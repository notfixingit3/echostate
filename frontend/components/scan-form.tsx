"use client"

import * as React from "react"
import Link from "next/link"
import { useSearchParams } from "next/navigation"

import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from "@/components/ui/card"
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert"
import { IntelSummary } from "@/components/intel-summary"
import { scanHost, createReport, getReport, ApiError } from "@/lib/api"
import { extractIntel } from "@/lib/intel"
import type { ScanResponse, Report } from "@/lib/types"
import { Loader2Icon, ScanIcon, AlertCircleIcon, CheckCircleIcon, FileTextIcon } from "lucide-react"

export function ScanForm() {
  const searchParams = useSearchParams()
  const [host, setHost] = React.useState("")

  React.useEffect(() => {
    const preset = searchParams.get("scan")?.trim()
    if (preset) {
      setHost(preset)
    }
  }, [searchParams])
  const [result, setResult] = React.useState<ScanResponse | null>(null)
  const [error, setError] = React.useState<string | null>(null)
  const [isLoading, setIsLoading] = React.useState(false)
  const [scanStatus, setScanStatus] = React.useState<string | null>(null)
  const [report, setReport] = React.useState<Report | null>(null)
  const [reportError, setReportError] = React.useState<string | null>(null)
  const [isCreatingReport, setIsCreatingReport] = React.useState(false)
  const [withReport, setWithReport] = React.useState(false)

  const busy = isLoading || isCreatingReport

  async function runScan(trimmedHost: string, queueReport: boolean) {
    setIsLoading(true)
    setWithReport(queueReport)
    setError(null)
    setResult(null)
    setReport(null)
    setReportError(null)
    setIsCreatingReport(false)
    setScanStatus(null)

    try {
      const scanResult = await scanHost(trimmedHost, (status) => {
        setScanStatus(status)
      })
      setResult(scanResult)

      if (queueReport) {
        setIsCreatingReport(true)
        try {
          const created = await createReport(scanResult.snapshot_id)
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
    } catch (err) {
      if (err instanceof ApiError && err.status === 429 && typeof err.retryAfter === "number") {
        setError(`Too many scans. Try again in ${err.retryAfter} seconds.`)
        return
      }

      if (err instanceof Error) {
        setError(err.message || "Something went wrong. Please try again.")
        return
      }

      setError("Something went wrong. Please try again.")
    } finally {
      setIsLoading(false)
      setWithReport(false)
      setScanStatus(null)
    }
  }

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()

    const trimmedHost = host.trim()
    if (!trimmedHost) {
      setError("Please enter a host, IP address, or URL.")
      return
    }

    await runScan(trimmedHost, false)
  }

  const handleScanAndReport = async () => {
    const trimmedHost = host.trim()
    if (!trimmedHost) {
      setError("Please enter a host, IP address, or URL.")
      return
    }

    await runScan(trimmedHost, true)
  }

  const handleGenerateReport = async () => {
    if (!result) return

    setIsCreatingReport(true)
    setReportError(null)
    setReport(null)

    try {
      const created = await createReport(result.snapshot_id)
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
    if (!report || report.status === "completed" || report.status === "failed") return

    let cancelled = false
    const intervalId = setInterval(async () => {
      try {
        const updated = await getReport(report.id)
        if (cancelled) return
        setReport(updated)
        setReportError(null)
      } catch (err) {
        if (cancelled) return
        if (err instanceof ApiError) {
          setReportError(err.message)
        } else if (err instanceof Error) {
          setReportError(err.message)
        } else {
          setReportError("Failed to check report status")
        }
      }
    }, 5000)

    return () => {
      cancelled = true
      clearInterval(intervalId)
    }
  }, [report])

  const intel = result ? extractIntel(result.raw_data) : null

  return (
    <div className="flex flex-col gap-6">
      <form onSubmit={handleSubmit} className="scan-form flex flex-col gap-4">
        <div className="flex flex-col gap-2">
          <label htmlFor="host" className="text-sm font-medium">
            Host / IP / URL
          </label>
          <Input
            id="host"
            name="host"
            type="text"
            placeholder="example.com, 8.8.8.8, or https://example.com"
            value={host}
            onChange={(event) => setHost(event.target.value)}
            disabled={busy}
            aria-invalid={!!error}
            data-testid="scan-host-input"
          />
        </div>

        <div className="flex flex-wrap gap-2">
          <Button
            type="submit"
            disabled={busy || !host.trim()}
            data-testid="scan-submit-button"
          >
            {isLoading && !withReport ? (
              <>
                <Loader2Icon data-icon="inline-start" className="animate-spin" />
                Scanning…
              </>
            ) : (
              <>
                <ScanIcon data-icon="inline-start" />
                Start Scan
              </>
            )}
          </Button>
          <Button
            type="button"
            variant="secondary"
            disabled={busy || !host.trim()}
            onClick={() => void handleScanAndReport()}
            data-testid="scan-and-report-button"
          >
            {isLoading && withReport ? (
              <>
                <Loader2Icon data-icon="inline-start" className="animate-spin" />
                Scanning…
              </>
            ) : isCreatingReport ? (
              <>
                <Loader2Icon data-icon="inline-start" className="animate-spin" />
                Building PDF…
              </>
            ) : (
              <>
                <FileTextIcon data-icon="inline-start" />
                Scan & Report
              </>
            )}
          </Button>
        </div>
        {isLoading && scanStatus ? (
          <p className="text-sm text-muted-foreground" data-testid="scan-status">
            Scan job: <span className="font-medium text-foreground">{scanStatus}</span>
          </p>
        ) : null}
      </form>

      {error && (
        <Alert variant="destructive" data-testid="scan-error-alert">
          <AlertCircleIcon />
          <AlertTitle>Scan failed</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {result && (
        <>
        <Card
          className="border-primary/20 bg-card/90 backdrop-blur-sm"
          data-testid="scan-result-card"
        >
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <CheckCircleIcon className="text-primary" />
              Scan complete
            </CardTitle>
            <CardDescription>
              Snapshot taken at{" "}
              {new Date(result.scanned_at).toLocaleString()}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="flex min-w-0 flex-col gap-1">
                <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                  Host
                </span>
                <span className="break-all font-mono text-sm" data-testid="scan-result-host">
                  {result.host}
                </span>
              </div>
              <div className="flex min-w-0 flex-col gap-1">
                <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                  Snapshot ID
                </span>
                <span className="break-all font-mono text-sm" data-testid="scan-result-snapshot-id">
                  {result.snapshot_id}
                </span>
              </div>
            </div>
          </CardContent>
          <CardFooter className="flex flex-col items-start gap-2">
            <div className="flex w-full flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              {!report || report.status === "failed" ? (
                <Button
                  disabled={isCreatingReport}
                  onClick={handleGenerateReport}
                  data-testid="scan-result-generate-report-button"
                >
                  {isCreatingReport ? (
                    <>
                      <Loader2Icon data-icon="inline-start" className="animate-spin" />
                      Generating report…
                    </>
                  ) : (
                    <>
                      <FileTextIcon data-icon="inline-start" />
                      Generate report
                    </>
                  )}
                </Button>
              ) : report.status === "completed" ? (
                <Button
                  variant="outline"
                  render={<Link href={`/report?id=${report.id}`} />}
                  nativeButton={false}
                  data-testid="scan-result-view-report-link"
                >
                  View report →
                </Button>
              ) : (
                <Button disabled data-testid="scan-result-report-pending-button">
                  <Loader2Icon data-icon="inline-start" className="animate-spin" />
                  Generating report…
                </Button>
              )}
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  render={<Link href={`/target?id=${result.target_id}`} />}
                  nativeButton={false}
                  data-testid="scan-result-target-link"
                >
                  View target →
                </Button>
                <Button
                  variant="outline"
                  render={<Link href={`/snapshot?id=${result.snapshot_id}`} />}
                  nativeButton={false}
                  data-testid="scan-result-view-link"
                >
                  View snapshot →
                </Button>
              </div>
            </div>
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

        {intel ? (
          <div className="flex w-full flex-col gap-3">
            <h3 className="font-heading text-base font-semibold tracking-tight">
              Intelligence summary
            </h3>
            <IntelSummary intel={intel} />
          </div>
        ) : null}
        </>
      )}
    </div>
  )
}
