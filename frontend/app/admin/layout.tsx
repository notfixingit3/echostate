"use client"

import { SlidersHorizontalIcon } from "lucide-react"

import { AdminGuard } from "@/components/admin-guard"
import { AdminNav } from "@/components/admin-nav"
import { HelpTip } from "@/components/help-tip"

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  return (
    <AdminGuard>
      <div className="container py-10">
        <div className="mb-8 flex items-center gap-2">
          <SlidersHorizontalIcon className="size-6 text-primary" />
          <h1 className="text-3xl font-bold tracking-tight">Administration</h1>
          <HelpTip id="page.administration" />
        </div>

        <div className="grid gap-8 lg:grid-cols-[220px_minmax(0,1fr)]">
          <aside className="lg:sticky lg:top-24 lg:self-start">
            <AdminNav />
          </aside>
          <div className="min-w-0 max-w-4xl">{children}</div>
        </div>
      </div>
    </AdminGuard>
  )
}