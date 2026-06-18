"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"

import { BrandLogo } from "@/components/brand-logo"
import { Button } from "@/components/ui/button"
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet"
import { ThemeToggle } from "@/components/theme-toggle"
import { useAuth } from "@/components/auth-provider"
import { cn } from "@/lib/utils"
import { LogOutIcon, MenuIcon } from "lucide-react"

const appNavLinks = [
  { href: "/", label: "Home" },
  { href: "/targets", label: "Targets" },
  { href: "/collections", label: "Collections" },
  { href: "/notes", label: "Notes" },
  { href: "/snapshots", label: "Snapshots" },
  { href: "/graph", label: "Graph" },
  { href: "/reports", label: "Reports" },
] as const

export function Nav() {
  const pathname = usePathname()
  const { user, config, signOut } = useAuth()

  const navLinks = [
    ...appNavLinks,
    ...(user?.role === "admin" ? [{ href: "/admin", label: "Admin" as const }] : []),
  ]

  return (
    <header className="sticky top-0 z-40 w-full border-b border-border/60 bg-background/80 backdrop-blur-xl supports-backdrop-filter:bg-background/70">
      <div className="container flex h-16 items-center justify-between">
        <Link
          href="/"
          className="group flex items-center gap-3"
          data-testid="nav-logo"
        >
          <BrandLogo variant="banner" className="h-8 w-auto" />
        </Link>

        <div className="flex items-center gap-2">
          <nav className="hidden items-center gap-1 md:flex">
            {navLinks.map((link) => {
              const isActive =
                link.href === "/"
                  ? pathname === "/"
                  : pathname.startsWith(link.href)

              return (
                <Link
                  key={link.href}
                  href={link.href}
                  data-testid={`nav-link-${link.label.toLowerCase()}`}
                  className={cn(
                    "rounded-lg px-3.5 py-2 text-sm font-medium transition-colors",
                    isActive
                      ? "bg-primary/10 text-primary"
                      : "text-muted-foreground hover:bg-muted/60 hover:text-foreground"
                  )}
                >
                  {link.label}
                </Link>
              )
            })}
          </nav>
          {user ? (
            <div className="hidden items-center gap-2 md:flex">
              <Button
                variant="ghost"
                size="sm"
                className={cn(
                  "max-w-40 truncate",
                  pathname.startsWith("/profile") && "bg-primary/10 text-primary"
                )}
                render={<Link href="/profile" />}
              >
                {user.display_name}
              </Button>
              <Button variant="ghost" size="sm" onClick={() => void signOut()}>
                <LogOutIcon data-icon="inline-start" />
                Sign out
              </Button>
            </div>
          ) : config?.auth_required ? (
            <Button
              variant="outline"
              size="sm"
              className="hidden md:inline-flex"
              render={<Link href="/login" />}
            >
              Sign in
            </Button>
          ) : null}
          <ThemeToggle />
          <Sheet>
          <SheetTrigger
            render={
              <Button variant="outline" size="icon" aria-label="Open menu" />
            }
            className="md:hidden"
          >
            <MenuIcon />
          </SheetTrigger>
          <SheetContent side="right" className="w-72">
            <SheetHeader>
              <SheetTitle>
                <Link href="/" className="flex items-center gap-2">
                  <BrandLogo variant="banner" className="h-7 w-auto" />
                </Link>
              </SheetTitle>
            </SheetHeader>
            <nav className="flex flex-col gap-1 p-4">
              {navLinks.map((link) => {
                const isActive =
                  link.href === "/"
                    ? pathname === "/"
                    : pathname.startsWith(link.href)

                return (
                  <Button
                    key={link.href}
                    variant={isActive ? "secondary" : "ghost"}
                    className="justify-start"
                    render={<Link href={link.href} />}
                  >
                    {link.label}
                  </Button>
                )
              })}
              {user ? (
                <>
                  <div className="my-2 border-t border-border/60" />
                  <Button
                    variant={pathname.startsWith("/profile") ? "secondary" : "ghost"}
                    className="justify-start"
                    render={<Link href="/profile" />}
                  >
                    {user.display_name}
                  </Button>
                  {user.role === "admin" ? (
                    <Button
                      variant={pathname.startsWith("/admin") ? "secondary" : "ghost"}
                      className="justify-start"
                      render={<Link href="/admin" />}
                    >
                      Administration
                    </Button>
                  ) : null}
                  <Button
                    variant="ghost"
                    className="justify-start"
                    onClick={() => void signOut()}
                  >
                    <LogOutIcon data-icon="inline-start" />
                    Sign out
                  </Button>
                </>
              ) : config?.auth_required ? (
                <Button variant="outline" className="mt-2 justify-start" render={<Link href="/login" />}>
                  Sign in
                </Button>
              ) : null}
            </nav>
          </SheetContent>
          </Sheet>
        </div>
      </div>
    </header>
  )
}