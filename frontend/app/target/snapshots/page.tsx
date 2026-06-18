"use client"

import * as React from "react"
import { Suspense } from "react"
import Link from "next/link"
import { useSearchParams } from "next/navigation"

import { PageHeader } from "@/components/page-header"
import { TargetSnapshotsTable } from "@/components/target-snapshots-table"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { fetchApi } from "@/lib/api"
import type { TargetDetail } from "@/lib/types"
import { AlertCircleIcon, ArrowLeftIcon } from "lucide-react"

function TargetSnapshotsPageContent() {
  const searchParams = useSearchParams()
  const targetId = searchParams.get("id")
  const [host, setHost] = React.useState<string | null>(null)

  React.useEffect(() => {
    if (!targetId) return

    let cancelled = false
    void fetchApi<TargetDetail>(`/api/targets/${targetId}`)
      .then((target) => {
        if (!cancelled) setHost(target.host)
      })
      .catch(() => {
        if (!cancelled) setHost(null)
      })

    return () => {
      cancelled = true
    }
  }, [targetId])

  if (!targetId) {
    return (
      <section className="container flex flex-1 flex-col gap-8 py-8 md:py-10">
        <Alert variant="destructive">
          <AlertCircleIcon />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>Missing target ID.</AlertDescription>
        </Alert>
      </section>
    )
  }

  return (
    <section className="container flex flex-1 flex-col gap-8 py-8 md:py-10">
      <PageHeader
        title="Target snapshots"
        description={
          host
            ? `Reconnaissance history for ${host}.`
            : `Reconnaissance history for target ${targetId}.`
        }
      >
        <Button
          variant="outline"
          render={<Link href={`/target?id=${targetId}`} />}
          nativeButton={false}
        >
          <ArrowLeftIcon data-icon="inline-start" />
          Target overview
        </Button>
      </PageHeader>
      <TargetSnapshotsTable targetId={targetId} />
    </section>
  )
}

export default function TargetSnapshotsPage() {
  return (
    <Suspense
      fallback={
        <section className="container flex flex-1 flex-col gap-8 py-8 md:py-10">
          <Skeleton className="h-10 w-64" />
          <Skeleton className="h-40 w-full" />
        </section>
      }
    >
      <TargetSnapshotsPageContent />
    </Suspense>
  )
}