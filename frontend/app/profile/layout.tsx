"use client"

import { UserCircleIcon } from "lucide-react"

import { HelpTip } from "@/components/help-tip"

export default function ProfileLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="container max-w-3xl py-10">
      <div className="mb-8 flex items-center gap-2">
        <UserCircleIcon className="size-6 text-primary" />
        <h1 className="text-3xl font-bold tracking-tight">Profile</h1>
        <HelpTip id="page.profile" />
      </div>
      {children}
    </div>
  )
}