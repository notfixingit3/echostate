"use client"

import Link from "next/link"

import { useAuth } from "@/components/auth-provider"
import { Button } from "@/components/ui/button"

export function AdminGuard({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuth()

  if (loading) {
    return <p className="text-sm text-muted-foreground">Loading…</p>
  }

  if (!user || user.role !== "admin") {
    return (
      <div className="rounded-lg border border-dashed p-8 text-center">
        <p className="text-muted-foreground">Admin access required.</p>
        <Button variant="link" className="mt-2" render={<Link href="/settings/account" />}>
          Go to Account & passkeys
        </Button>
      </div>
    )
  }

  return <>{children}</>
}