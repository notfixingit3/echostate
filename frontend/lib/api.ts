import type {
  ApiErrorResponse,
  ScanResponse,
  Report,
  ReportSummary,
  PaginatedResponse,
  CreateReportRequest,
  TargetDetail,
} from "./types"

// Empty string = same-origin /api (Docker nginx proxy). Unset = local dev against :8080.
export const API_BASE =
  process.env.NEXT_PUBLIC_API_URL !== undefined
    ? process.env.NEXT_PUBLIC_API_URL
    : "http://localhost:8080"

export class ApiError extends Error {
  status: number
  retryAfter?: number

  constructor(message: string, status: number, retryAfter?: number) {
    super(message)
    this.name = "ApiError"
    this.status = status
    this.retryAfter = retryAfter
  }
}

type ScanJobResponse = {
  id: string
  host: string
  status: "pending" | "running" | "completed" | "failed"
  snapshot_id?: string
  target_id?: string
  error?: string
  snapshot?: {
    id: string
    target_id: string
    scanned_at: string
    raw_data?: ScanResponse["raw_data"]
    changes?: string[]
  }
}

const SCAN_POLL_INTERVAL_MS = 2000
const SCAN_POLL_TIMEOUT_MS = 130_000

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

export async function fetchApi<T = unknown>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const url = `${API_BASE}${path}`
  const headers = new Headers(options.headers)

  if (!headers.has("Content-Type") && !(options.body instanceof FormData)) {
    headers.set("Content-Type", "application/json")
  }

  const response = await fetch(url, {
    ...options,
    headers,
    credentials: "include",
  })

  if (!response.ok) {
    const text = await response.text()
    let retryAfter: number | undefined

    try {
      const parsed = JSON.parse(text) as ApiErrorResponse
      if (typeof parsed.retry_after === "number") {
        retryAfter = parsed.retry_after
      }
      throw new ApiError(
        parsed.error || parsed.message || text || response.statusText,
        response.status,
        retryAfter
      )
    } catch (err) {
      if (err instanceof ApiError) {
        throw err
      }
      throw new ApiError(text || response.statusText, response.status)
    }
  }

  return response.json() as Promise<T>
}

async function waitForScanJob(
  jobId: string,
  onStatus?: (status: ScanJobResponse["status"]) => void
): Promise<ScanJobResponse> {
  const deadline = Date.now() + SCAN_POLL_TIMEOUT_MS

  while (Date.now() < deadline) {
    const job = await fetchApi<ScanJobResponse>(`/api/scans/${jobId}`)
    onStatus?.(job.status)
    if (job.status === "completed" && job.snapshot) {
      return job
    }
    if (job.status === "failed") {
      throw new ApiError(job.error || "Scan failed", 500)
    }
    await sleep(SCAN_POLL_INTERVAL_MS)
  }

  throw new ApiError("Scan timed out", 504)
}

export async function scanHost(
  host: string,
  onStatus?: (status: ScanJobResponse["status"]) => void
): Promise<ScanResponse> {
  const job = await fetchApi<ScanJobResponse>("/api/scan", {
    method: "POST",
    body: JSON.stringify({ host }),
  })
  onStatus?.(job.status)

  const completed =
    job.status === "completed" && job.snapshot
      ? job
      : await waitForScanJob(job.id, onStatus)

  const snapshot = completed.snapshot
  if (!snapshot) {
    throw new ApiError("Scan completed without snapshot data", 500)
  }

  return {
    snapshot_id: snapshot.id,
    target_id: snapshot.target_id,
    host: snapshot.raw_data?.host || host,
    scanned_at: snapshot.scanned_at,
    raw_data: snapshot.raw_data,
    changes: snapshot.changes,
  }
}

export async function createReport(
  snapshot_id: string,
  wait_for_enrichment = false
): Promise<Report> {
  return fetchApi<Report>("/api/reports", {
    method: "POST",
    body: JSON.stringify({
      snapshot_id,
      wait_for_enrichment,
    } satisfies CreateReportRequest),
  })
}

export async function getReport(report_id: string): Promise<Report> {
  return fetchApi<Report>(`/api/reports/${report_id}`)
}

export async function getLatestReportForSnapshot(
  snapshot_id: string
): Promise<Report | null> {
  const params = new URLSearchParams({
    snapshot_id,
    limit: "1",
    page: "1",
  })
  const list = await fetchApi<PaginatedResponse<ReportSummary>>(
    `/api/reports?${params.toString()}`
  )
  if (!list.data.length) {
    return null
  }
  return getReport(list.data[0].id)
}

export async function downloadReport(
  report: Pick<Report, "id" | "status" | "download_url">,
  filename?: string
): Promise<void> {
  if (report.status !== "completed") {
    throw new ApiError("Report not ready", 409)
  }

  const path = report.download_url || `/api/reports/${report.id}/download`
  const response = await fetch(`${API_BASE}${path}`, {
    credentials: "include",
  })

  if (!response.ok) {
    const text = await response.text()
    let message = text || "Download failed"
    try {
      const parsed = JSON.parse(text) as ApiErrorResponse
      message = parsed.error || parsed.message || message
    } catch {
      // keep plain-text message
    }
    throw new ApiError(message, response.status)
  }

  const blob = await response.blob()
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement("a")
  anchor.href = url
  anchor.download = filename || `echostate-report-${report.id}.pdf`
  document.body.appendChild(anchor)
  anchor.click()
  document.body.removeChild(anchor)
  URL.revokeObjectURL(url)
}

export async function getTarget(target_id: string): Promise<TargetDetail> {
  return fetchApi<TargetDetail>(`/api/targets/${target_id}`)
}

// Backward-compatible alias for existing callers.
export const fetchJson = fetchApi