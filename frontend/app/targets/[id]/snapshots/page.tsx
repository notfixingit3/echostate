import Link from "next/link"

import { PageHeader } from "@/components/page-header"
import { TargetSnapshotsTable } from "@/components/target-snapshots-table"
import { Button } from "@/components/ui/button"
import { ArrowLeftIcon } from "lucide-react"

export async function generateStaticParams() {
  return [{ id: "placeholder" }]
}

export default async function TargetSnapshotsPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params

  return (
    <section className="container flex flex-1 flex-col gap-8 py-8 md:py-10">
      <PageHeader
        title="Target snapshots"
        description={`Reconnaissance history for target ${id}.`}
      >
        <Button
          variant="outline"
          render={<Link href={`/targets/${id}`} />}
          nativeButton={false}
        >
          <ArrowLeftIcon data-icon="inline-start" />
          Target overview
        </Button>
      </PageHeader>
      <TargetSnapshotsTable targetId={id} />
    </section>
  )
}
