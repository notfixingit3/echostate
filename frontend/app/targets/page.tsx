import { PageHeader } from "@/components/page-header"
import { TargetsTable } from "@/components/targets-table"

export default function TargetsPage() {
  return (
    <section className="container flex flex-1 flex-col gap-8 py-8 md:py-10">
      <PageHeader
        title="Targets"
        description="Browse hosts and IPs under surveillance, with ASN and web intel from the latest snapshot."
      />
      <TargetsTable />
    </section>
  )
}
