"use client"

import { AdminGuard } from "@/components/admin-guard"
import { SettingsSystemPanel } from "@/components/settings-system-panel"

export default function SystemSettingsPage() {
  return (
    <AdminGuard>
      <SettingsSystemPanel />
    </AdminGuard>
  )
}