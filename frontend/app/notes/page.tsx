import { InvestigationNotes } from "@/components/investigation-notes"
import { PageHeader } from "@/components/page-header"

export default function NotesPage() {
  return (
    <section className="container flex flex-1 flex-col gap-8 py-8 md:py-10">
      <PageHeader
        title="Notes"
        helpId="page.notes"
        description="Investigation notes across targets, snapshots, collections, and graph nodes. Archive finished work or move mistakes to trash — purged automatically after 30 days."
      />
      <InvestigationNotes title="All notes" allowFreeform />
    </section>
  )
}