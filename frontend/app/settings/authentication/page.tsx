"use client"

import { AdminGuard } from "@/components/admin-guard"
import { SettingsAuthPanel } from "@/components/settings-auth-panel"

export default function AuthenticationSettingsPage() {
  return (
    <AdminGuard>
      <SettingsAuthPanel />
    </AdminGuard>
  )
}