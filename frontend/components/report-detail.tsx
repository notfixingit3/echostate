"use client"

import * as React from "react"

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
import { Skeleton } from "@/components/ui/skeleton"
import { Separator } from "@/components/ui/separator"
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert"
import { fetchApi, ApiError, API_BASE } from "@/lib/api"
import type { Report, ReportStatus } from "@/lib/types"
import {
  FileTextIcon,
  DownloadIcon,
  AlertCircleIcon,
  CalendarIcon,
  ClockIcon,
  GlobeIcon,
} from "lucide-react"

function statusBadgeVariant(status: ReportStatus) {
  switch (status) {
    case "completed":
      return "default"
    case "failed":
      return "destructive"
    case "running":
      return "secondary"
    case "pending":
    default:
      return "outline"
  }
}

export function ReportDetail({ reportId }: { reportId: string }) {

  const [report, setReport] = React.useState<Report | null>(null)
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState<string | null>(null)

  React.useEffect(() => {
    let cancelled = false
    let intervalId: ReturnType<typeof setInterval> | null = null

    async function load() {
      try {
        const result = await fetchApi<Report>(`/api/reports/${reportId}`)
        if (!cancelled) {
          setReport(result)
          setError(null)
          if (
            result.status !== "pending" &&
            result.status !== "running" &&
            intervalId
          ) {
            clearInterval(intervalId)
            intervalId = null
          }
        }
      } catch (err) {
        if (!cancelled) {
          if (err instanceof ApiError) {
            setError(err.message)
          } else if (err instanceof Error) {
            setError(err.message)
          } else {
            setError("Failed to load report")
          }
          if (intervalId) {
            clearInterval(intervalId)
            intervalId = null
          }
        }
      } finally {
        if (!cancelled) setLoading(false)
      }
    }

    load()
    intervalId = setInterval(load, 5000)

    return () => {
      cancelled = true
      if (intervalId) clearInterval(intervalId)
    }
  }, [reportId])

  const handleDownload = () => {
    if (!report || report.status !== "completed") return
    const url = report.download_url
      ? `${API_BASE}${report.download_url}`
      : `${API_BASE}/api/reports/${report.id}/download`
    const a = document.createElement("a")
    a.href = url
    a.download = `echostate-report-${report.id}.pdf`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
  }

  if (loading && !report) {
    return (
      <div className="flex flex-col gap-6">
        <Card>
          <CardHeader>
            <Skeleton className="h-6 w-48" />
            <Skeleton className="h-4 w-64" />
          </CardHeader>
          <CardContent className="grid gap-4 sm:grid-cols-2 lg:grid-cols-5">
            {["a", "b", "c", "d", "e"].map((key) => (
              <div key={key} className="flex flex-col gap-1">
                <Skeleton className="h-3 w-20" />
                <Skeleton className="h-4 w-32" />
              </div>
            ))}
          </CardContent>
        </Card>
      </div>
    )
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertCircleIcon />
        <AlertTitle>Error</AlertTitle>
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    )
  }

  if (!report) {
    return (
      <Alert variant="destructive">
        <AlertCircleIcon />
        <AlertTitle>Not found</AlertTitle>
        <AlertDescription>Report not found.</AlertDescription>
      </Alert>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileTextIcon />
            Report {report.id}
          </CardTitle>
          <CardDescription>PDF report status and download</CardDescription>
        </CardHeader>
          <CardContent className="grid gap-4 sm:grid-cols-2 lg:grid-cols-5">
            <div className="flex flex-col gap-1">
              <span className="text-xs text-muted-foreground">Status</span>
              <span>
                <Badge variant={statusBadgeVariant(report.status)} data-testid="report-detail-status">
                  {report.status}
                </Badge>
              </span>
            </div>
            <div className="flex flex-col gap-1">
              <span className="text-xs text-muted-foreground">Snapshot</span>
              <span className="text-sm font-mono">{report.snapshot_id}</span>
            </div>
            <div className="flex flex-col gap-1">
              <span className="text-xs text-muted-foreground">Host</span>
              <span className="flex items-center gap-1.5 text-sm font-medium">
                <GlobeIcon className="size-3.5 text-muted-foreground" />
                {report.host || "—"}
              </span>
            </div>
            <div className="flex flex-col gap-1">
              <span className="text-xs text-muted-foreground">Created</span>
              <span className="flex items-center gap-1.5 text-sm font-medium">
                <CalendarIcon className="size-3.5 text-muted-foreground" />
                {new Date(report.created_at).toLocaleString()}
              </span>
            </div>
            <div className="flex flex-col gap-1">
              <span className="text-xs text-muted-foreground">Completed</span>
              <span className="flex items-center gap-1.5 text-sm font-medium">
                <ClockIcon className="size-3.5 text-muted-foreground" />
                {report.completed_at
                  ? new Date(report.completed_at).toLocaleString()
                  : "—"}
              </span>
            </div>
          </CardContent>
        <Separator />
        <CardFooter className="flex justify-between">
          <Button
            disabled={report.status !== "completed"}
            onClick={handleDownload}
            data-testid="report-detail-download-button"
          >
            <DownloadIcon data-icon="inline-start" />
            Download PDF
          </Button>
        </CardFooter>
      </Card>

      {report.status === "failed" && (
        <Alert variant="destructive">
          <AlertCircleIcon />
          <AlertTitle>Report failed</AlertTitle>
          <AlertDescription>
            {report.error || "Report generation failed."}
          </AlertDescription>
        </Alert>
      )}

      {(report.status === "pending" || report.status === "running") && (
        <Alert>
          <ClockIcon />
          <AlertTitle>Report in progress</AlertTitle>
          <AlertDescription>
            Status refreshes automatically every 5 seconds.
          </AlertDescription>
        </Alert>
      )}
    </div>
  )
}
