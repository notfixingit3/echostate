import { PageHeader } from "@/components/page-header"
import { ReportsTable } from "@/components/reports-table"

export default function ReportsPage() {
  return (
    <section className="container flex flex-1 flex-col gap-8 py-8 md:py-10">
      <PageHeader
        title="Reports"
        helpId="page.reports"
        description="Browse and download PDF reports generated from snapshots."
      />
      <ReportsTable />
    </section>
  )
}
