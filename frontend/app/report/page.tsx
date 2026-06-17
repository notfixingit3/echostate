"use client"

import { Suspense } from "react"
import { useSearchParams } from "next/navigation"

import { ReportDetail } from "@/components/report-detail"
import { Skeleton } from "@/components/ui/skeleton"
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert"
import { AlertCircleIcon } from "lucide-react"

function ReportPageContent() {
  const searchParams = useSearchParams()
  const reportId = searchParams.get("id")

  if (!reportId) {
    return (
      <section className="container flex flex-1 flex-col gap-6 py-8">
        <Alert variant="destructive">
          <AlertCircleIcon />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>Missing report ID.</AlertDescription>
        </Alert>
      </section>
    )
  }

  return (
    <section className="container flex flex-1 flex-col gap-6 py-8">
      <ReportDetail reportId={reportId} />
    </section>
  )
}

export default function ReportPage() {
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
      <ReportPageContent />
    </Suspense>
  )
}
