"use client"

import { AdminGuard } from "@/components/admin-guard"
import { SettingsIntegrationsPanel } from "@/components/settings-integrations-panel"

export default function IntegrationsSettingsPage() {
  return (
    <AdminGuard>
      <SettingsIntegrationsPanel />
    </AdminGuard>
  )
}