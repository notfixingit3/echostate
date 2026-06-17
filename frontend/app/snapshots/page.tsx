import { PageHeader } from "@/components/page-header"
import { SnapshotsTable } from "@/components/snapshots-table"

export default function SnapshotsPage() {
  return (
    <section className="container flex flex-1 flex-col gap-8 py-8 md:py-10">
      <PageHeader
        title="Snapshots"
        description="Browse reconnaissance snapshots with ASN, web title, and submitter enrichment."
      />
      <SnapshotsTable />
    </section>
  )
}
