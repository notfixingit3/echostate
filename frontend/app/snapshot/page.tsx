"use client"

import * as React from "react"
import { Suspense } from "react"
import { useSearchParams } from "next/navigation"

import { SnapshotDetail } from "@/components/snapshot-detail"
import { fetchApi, ApiError } from "@/lib/api"
import type { Snapshot } from "@/lib/types"
import { Skeleton } from "@/components/ui/skeleton"
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert"
import { AlertCircleIcon } from "lucide-react"

function SnapshotPageContent() {
  const searchParams = useSearchParams()
  const id = searchParams.get("id")

  const [snapshot, setSnapshot] = React.useState<Snapshot | null>(null)
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState<string | null>(null)

  React.useEffect(() => {
    let cancelled = false

    async function load() {
      setLoading(true)
      setError(null)

      if (!id) {
        if (!cancelled) {
          setError("Missing snapshot ID.")
          setLoading(false)
        }
        return
      }

      try {
        const result = await fetchApi<Snapshot>(`/api/snapshots/${id}`)
        if (!cancelled) setSnapshot(result)
      } catch (err) {
        if (!cancelled) {
          if (err instanceof ApiError) {
            setError(err.message)
          } else if (err instanceof Error) {
            setError(err.message)
          } else {
            setError("Failed to load snapshot")
          }
        }
      } finally {
        if (!cancelled) setLoading(false)
      }
    }

    load()

    return () => {
      cancelled = true
    }
  }, [id])

  if (loading) {
    return (
      <section className="container flex flex-1 flex-col gap-6 py-8">
        <Skeleton className="h-8 w-64" />
        <Skeleton className="h-32 w-full" />
        <Skeleton className="h-64 w-full" />
      </section>
    )
  }

  if (error || !snapshot) {
    return (
      <section className="container flex flex-1 flex-col gap-6 py-8">
        <Alert variant="destructive">
          <AlertCircleIcon />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{error || "Snapshot not found."}</AlertDescription>
        </Alert>
      </section>
    )
  }

  return (
    <section className="container flex flex-1 flex-col gap-8 py-8 md:py-10">
      <SnapshotDetail snapshot={snapshot} />
    </section>
  )
}

export default function SnapshotPage() {
  return (
    <Suspense
      fallback={
        <section className="container flex flex-1 flex-col gap-6 py-8">
          <Skeleton className="h-8 w-64" />
          <Skeleton className="h-32 w-full" />
          <Skeleton className="h-64 w-full" />
        </section>
      }
    >
      <SnapshotPageContent />
    </Suspense>
  )
}
