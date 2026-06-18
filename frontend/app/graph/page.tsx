import { PageHeader } from "@/components/page-header"
import { GraphView } from "@/components/graph-view"

export default function GraphPage() {
  return (
    <section className="container flex flex-1 flex-col gap-8 py-8 md:py-10">
      <PageHeader
        title="Graph"
        description="Interactive Neo4j maps for infrastructure clusters, BGP paths, traceroute geography, CT subdomains, DNS dependencies, and certificate SAN overlap."
      />
      <GraphView />
    </section>
  )
}