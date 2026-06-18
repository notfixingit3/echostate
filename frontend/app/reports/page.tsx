import { Suspense } from "react"

import { PageHeader } from "@/components/page-header"
import { ReportsTable } from "@/components/reports-table"
import { Skeleton } from "@/components/ui/skeleton"

export default function ReportsPage() {
  return (
    <section className="container flex flex-1 flex-col gap-8 py-8 md:py-10">
      <PageHeader
        title="Reports"
        helpId="page.reports"
        description="Browse and download PDF reports generated from snapshots."
      />
      <Suspense
        fallback={
          <div className="flex flex-col gap-4">
            <Skeleton className="h-10 w-full max-w-xl" />
            <Skeleton className="h-64 w-full" />
          </div>
        }
      >
        <ReportsTable />
      </Suspense>
    </section>
  )
}
