import type {
  ApiErrorResponse,
  ScanResponse,
  Report,
  CreateReportRequest,
  TargetDetail,
} from "./types"

export const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080"

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

export async function scanHost(host: string): Promise<ScanResponse> {
  const data = (await fetchApi("/api/scan", {
    method: "POST",
    body: JSON.stringify({ host }),
  })) as {
    id: string
    target_id: string
    scanned_at: string
    raw_data?: ScanResponse["raw_data"]
    changes?: string[]
  }

  return {
    snapshot_id: data.id,
    target_id: data.target_id,
    host: data.raw_data?.host || host,
    scanned_at: data.scanned_at,
    raw_data: data.raw_data,
    changes: data.changes,
  }
}

export async function createReport(snapshot_id: string): Promise<Report> {
  return fetchApi<Report>("/api/reports", {
    method: "POST",
    body: JSON.stringify({ snapshot_id } satisfies CreateReportRequest),
  })
}

export async function getReport(report_id: string): Promise<Report> {
  return fetchApi<Report>(`/api/reports/${report_id}`)
}

export async function getTarget(target_id: string): Promise<TargetDetail> {
  return fetchApi<TargetDetail>(`/api/targets/${target_id}`)
}

// Backward-compatible alias for existing callers.
export const fetchJson = fetchApi
