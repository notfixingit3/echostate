import { PageHeader } from "@/components/page-header"
import { CollectionsManager } from "@/components/collections-manager"

export default function CollectionsPage() {
  return (
    <section className="container flex flex-1 flex-col gap-8 py-8 md:py-10">
      <PageHeader
        title="Collections"
        helpId="page.collections"
        description="Group targets into watchlists and queue bulk rescans for related infrastructure."
      />
      <CollectionsManager />
    </section>
  )
}