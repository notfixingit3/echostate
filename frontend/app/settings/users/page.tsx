"use client"

import { AdminGuard } from "@/components/admin-guard"
import { AuthUsersPanel } from "@/components/auth-users"

export default function UsersSettingsPage() {
  return (
    <AdminGuard>
      <AuthUsersPanel />
    </AdminGuard>
  )
}