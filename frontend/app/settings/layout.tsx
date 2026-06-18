"use client"

import { SlidersHorizontalIcon } from "lucide-react"

import { SettingsNav } from "@/components/settings-nav"
import { HelpTip } from "@/components/help-tip"

export default function SettingsLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="container py-10">
      <div className="mb-8 flex items-center gap-2">
        <SlidersHorizontalIcon className="size-6 text-primary" />
        <h1 className="text-3xl font-bold tracking-tight">Admin</h1>
        <HelpTip id="page.admin" />
      </div>

      <div className="grid gap-8 lg:grid-cols-[220px_minmax(0,1fr)]">
        <aside className="lg:sticky lg:top-24 lg:self-start">
          <SettingsNav />
        </aside>
        <div className="min-w-0 max-w-4xl">{children}</div>
      </div>
    </div>
  )
}