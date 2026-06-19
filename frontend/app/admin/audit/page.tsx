"use client"

import { ScrollTextIcon } from "lucide-react"

import { AuditEventsTable } from "@/components/audit-events-table"
import { HelpTip } from "@/components/help-tip"

export default function AdminAuditPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2">
        <ScrollTextIcon className="size-5 text-primary" />
        <h2 className="text-xl font-semibold tracking-tight">Audit log</h2>
        <HelpTip id="settings.nav.audit" />
      </div>
      <p className="text-sm text-muted-foreground">
        Security-relevant actions: sign-ins, deletes, settings changes, imports, and user management.
      </p>
      <AuditEventsTable />
    </div>
  )
}