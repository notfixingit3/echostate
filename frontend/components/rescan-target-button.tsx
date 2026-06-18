"use client"

import * as React from "react"

import { Button } from "@/components/ui/button"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { scanHost, ApiError } from "@/lib/api"
import type { ScanResponse } from "@/lib/types"
import { AlertCircleIcon, Loader2Icon, ScanIcon } from "lucide-react"

export function RescanTargetButton({
  host,
  onSuccess,
  variant = "default",
  size = "default",
  className,
}: {
  host: string
  onSuccess?: (result: ScanResponse) => void | Promise<void>
  variant?: "default" | "outline" | "secondary"
  size?: "default" | "sm"
  className?: string
}) {
  const [loading, setLoading] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)

  async function handleRescan() {
    const trimmedHost = host.trim()
    if (!trimmedHost) {
      setError("Missing target host.")
      return
    }

    setLoading(true)
    setError(null)

    try {
      const result = await scanHost(trimmedHost)
      await onSuccess?.(result)
    } catch (err) {
      if (err instanceof ApiError && err.status === 429 && typeof err.retryAfter === "number") {
        setError(`Too many scans. Try again in ${err.retryAfter} seconds.`)
        return
      }
      if (err instanceof ApiError) {
        setError(err.message)
        return
      }
      if (err instanceof Error) {
        setError(err.message || "Rescan failed.")
        return
      }
      setError("Rescan failed.")
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className={className}>
      <Button
        type="button"
        variant={variant}
        size={size}
        onClick={() => void handleRescan()}
        disabled={loading}
        data-testid="target-rescan-button"
      >
        {loading ? (
          <Loader2Icon className="size-4 animate-spin" data-icon="inline-start" />
        ) : (
          <ScanIcon className="size-4" data-icon="inline-start" />
        )}
        {loading ? "Scanning…" : "Rescan target"}
      </Button>
      {error ? (
        <Alert variant="destructive" className="mt-3">
          <AlertCircleIcon />
          <AlertTitle>Rescan failed</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
    </div>
  )
}