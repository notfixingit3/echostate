import { Suspense } from "react"

import { TargetDetailPageContent } from "@/components/target-detail-page"
import { Skeleton } from "@/components/ui/skeleton"

export function generateStaticParams() {
  return [{ id: "placeholder" }]
}

export default function TargetDetailPage() {
  return (
    <Suspense
      fallback={
        <section className="container flex flex-1 flex-col gap-8 py-8 md:py-10">
          <Skeleton className="h-10 w-64" />
          <Skeleton className="h-40 w-full" />
          <Skeleton className="h-64 w-full" />
        </section>
      }
    >
      <TargetDetailPageContent />
    </Suspense>
  )
}