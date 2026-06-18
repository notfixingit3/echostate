"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import {
  BellIcon,
  KeyRoundIcon,
  SettingsIcon,
  ShieldIcon,
  UserCogIcon,
  UsersIcon,
} from "lucide-react"

import { useAuth } from "@/components/auth-provider"
import { cn } from "@/lib/utils"

const adminItems = [
  { href: "/settings/integrations", label: "Integrations", icon: BellIcon, helpId: "settings.nav.integrations" },
  { href: "/settings/system", label: "System", icon: SettingsIcon, helpId: "settings.nav.system" },
  { href: "/settings/authentication", label: "Authentication", icon: ShieldIcon, helpId: "settings.nav.authentication" },
  { href: "/settings/users", label: "User management", icon: UsersIcon, helpId: "settings.nav.users" },
]

export function SettingsNav() {
  const pathname = usePathname()
  const { user } = useAuth()
  const isAdmin = user?.role === "admin"

  const items = [
    { href: "/settings/account", label: "Account & passkeys", icon: UserCogIcon, helpId: "settings.nav.account" },
    ...(isAdmin ? adminItems : []),
  ]

  return (
    <nav className="flex flex-col gap-1">
      {items.map((item) => {
        const active = pathname === item.href || pathname.startsWith(`${item.href}/`)
        const Icon = item.icon
        return (
          <Link
            key={item.href}
            href={item.href}
            data-testid={`settings-nav-${item.label.toLowerCase().replace(/\s+/g, "-")}`}
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