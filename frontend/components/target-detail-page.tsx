"use client"

import * as React from "react"
import { useSearchParams } from "next/navigation"

import { PageHeader } from "@/components/page-header"
import { TargetDetailView } from "@/components/target-detail"
import { fetchApi, ApiError } from "@/lib/api"
import type { TargetDetail } from "@/lib/types"
import { Skeleton } from "@/components/ui/skeleton"
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert"
import { AlertCircleIcon } from "lucide-react"

export function TargetDetailPageContent() {
  const searchParams = useSearchParams()
  const id = searchParams.get("id")

  const [target, setTarget] = React.useState<TargetDetail | null>(null)
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState<string | null>(null)
  const [refreshKey, setRefreshKey] = React.useState(0)

  const loadTarget = React.useCallback(async () => {
    if (!id) {
      setError("Missing target ID.")
      setLoading(false)
      return null
    }

    const result = await fetchApi<TargetDetail>(`/api/targets/${id}`)
    setTarget(result)
    return result
  }, [id])

  React.useEffect(() => {
    let cancelled = false

    async function load() {
      setLoading(true)
      setError(null)

      try {
        await loadTarget()
      } catch (err) {
        if (!cancelled) {
          if (err instanceof ApiError) {
            setError(err.message)
          } else if (err instanceof Error) {
            setError(err.message)
          } else {
            setError("Failed to load target")
          }
        }
      } finally {
        if (!cancelled) setLoading(false)
      }
    }

    void load()

    return () => {
      cancelled = true
    }
  }, [loadTarget])

  const handleRescanned = React.useCallback(async () => {
    await loadTarget()
    setRefreshKey((key) => key + 1)
  }, [loadTarget])

  if (loading) {
    return (
      <section className="container flex flex-1 flex-col gap-8 py-8 md:py-10">
        <Skeleton className="h-10 w-64" />
        <Skeleton className="h-40 w-full" />
        <Skeleton className="h-64 w-full" />
      </section>
    )
  }

  if (error || !target) {
    return (
      <section className="container flex flex-1 flex-col gap-8 py-8 md:py-10">
        <Alert variant="destructive">
          <AlertCircleIcon />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{error || "Target not found."}</AlertDescription>
        </Alert>
      </section>
    )
  }

  return (
    <section className="container flex flex-1 flex-col gap-8 py-8 md:py-10">
      <PageHeader
        title="Target detail"
        description={`Surveillance overview for ${target.host}.`}
      />
      <TargetDetailView
        target={target}
        refreshKey={refreshKey}
        onRescanned={handleRescanned}
      />
    </section>
  )
}