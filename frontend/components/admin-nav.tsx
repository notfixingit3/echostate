"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import { BellIcon, SettingsIcon, ShieldIcon, UsersIcon } from "lucide-react"

import { cn } from "@/lib/utils"

const adminItems = [
  { href: "/admin/integrations", label: "Integrations", icon: BellIcon },
  { href: "/admin/system", label: "System", icon: SettingsIcon },
  { href: "/admin/authentication", label: "Authentication", icon: ShieldIcon },
  { href: "/admin/users", label: "User management", icon: UsersIcon },
]

export function AdminNav() {
  const pathname = usePathname()

  return (
    <nav className="flex flex-col gap-1">
      {adminItems.map((item) => {
        const active = pathname === item.href || pathname.startsWith(`${item.href}/`)
        const Icon = item.icon
        return (
          <Link
            key={item.href}
            href={item.href}
            data-testid={`admin-nav-${item.label.toLowerCase().replace(/\s+/g, "-")}`}
            className={cn(
              "flex items-center gap-2.5 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
              active
                ? "bg-primary/10 text-primary"
                : "text-muted-foreground hover:bg-muted/60 hover:text-foreground"
            )}
          >
            <Icon className="size-4 shrink-0" />
            {item.label}
          </Link>
        )
      })}
    </nav>
  )
}