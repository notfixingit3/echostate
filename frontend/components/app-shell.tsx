"use client"

import Link from "next/link"

import { useAuth } from "@/components/auth-provider"
import { BrandLogo } from "@/components/brand-logo"
import { Footer } from "@/components/footer"
import { Nav } from "@/components/nav"
import { ThemeToggle } from "@/components/theme-toggle"

export function AppShell({ children }: { children: React.ReactNode }) {
  const { loading, config, user } = useAuth()
  const gated = config?.auth_required && !user

  if (loading && !user) {
    return (
      <div className="relative flex min-h-svh flex-col">
        <div className="absolute right-4 top-4 z-10">
          <ThemeToggle />
        </div>
        <main className="flex flex-1 items-center justify-center px-4">
          <p className="text-sm text-muted-foreground">Loading…</p>
        </main>
      </div>
    )
  }

  if (gated) {
    return (
      <div className="relative flex min-h-svh flex-col">
        <div className="absolute right-4 top-4 z-10">
          <ThemeToggle />
        </div>
        <main className="flex flex-1 flex-col items-center justify-center px-4 py-8">
          <Link href="/login" className="mb-8" aria-label="EchoState">
            <BrandLogo variant="banner" className="h-9 w-auto" />
          </Link>
          <div className="w-full max-w-md">{children}</div>
        </main>
      </div>
    )
  }

  return (
    <div className="flex min-h-svh flex-col">
      <Nav />
      <main className="min-w-0 flex-1">{children}</main>
      <Footer />
    </div>
  )
}