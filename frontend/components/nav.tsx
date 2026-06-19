"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"

import { BrandLogo } from "@/components/brand-logo"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet"
import { Separator } from "@/components/ui/separator"
import { ThemeToggle } from "@/components/theme-toggle"
import { useAuth } from "@/components/auth-provider"
import { cn } from "@/lib/utils"
import { ChevronDownIcon, LogOutIcon, MenuIcon, UserIcon } from "lucide-react"

const appNavLinks = [
  { href: "/", label: "Home" },
  { href: "/targets", label: "Targets" },
  { href: "/collections", label: "Collections" },
  { href: "/notes", label: "Notes" },
  { href: "/snapshots", label: "Snapshots" },
  { href: "/graph", label: "Graph" },
  { href: "/reports", label: "Reports" },
] as const

function isNavLinkActive(href: string, pathname: string): boolean {
  if (href === "/") return pathname === "/"
  if (href === "/admin") return pathname.startsWith("/admin")
  return pathname === href || pathname.startsWith(`${href}/`)
}

export function Nav() {
  const pathname = usePathname()
  const { user, config, signOut } = useAuth()

  const navLinks = [
    ...appNavLinks,
    ...(user?.role === "admin"
      ? [{ href: "/admin", label: "Server settings" as const }]
      : []),
  ]

  return (
    <header className="sticky top-0 z-40 w-full border-b border-border/60 bg-background/80 pt-[env(safe-area-inset-top)] backdrop-blur-xl supports-backdrop-filter:bg-background/70">
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
              const isActive = isNavLinkActive(link.href, pathname)

              return (
                <Link
                  key={link.href}
                  href={link.href}
                  data-testid={`nav-link-${link.label.toLowerCase().replace(/\s+/g, "-")}`}
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
            <div className="hidden md:flex">
              <DropdownMenu>
                <DropdownMenuTrigger
                  render={
                    <Button
                      variant="ghost"
                      size="sm"
                      data-testid="nav-account-menu"
                      className="max-w-48"
                    />
                  }
                >
                  <span className="truncate">{user.display_name}</span>
                  <ChevronDownIcon data-icon="inline-end" />
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="min-w-40">
                  <DropdownMenuGroup>
                    <DropdownMenuLabel className="truncate">
                      {user.display_name}
                    </DropdownMenuLabel>
                    <DropdownMenuItem
                      render={
                        <Link href="/profile" data-testid="nav-account-profile" />
                      }
                    >
                      <UserIcon />
                      Profile
                    </DropdownMenuItem>
                  </DropdownMenuGroup>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem
                    variant="destructive"
                    onClick={() => void signOut()}
                    data-testid="nav-account-sign-out"
                  >
                    <LogOutIcon />
                    Sign out
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
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
                const isActive = isNavLinkActive(link.href, pathname)

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
                  <Separator className="my-2" />
                  <p className="truncate px-3 py-1 text-xs font-medium text-muted-foreground">
                    {user.display_name}
                  </p>
                  <Button
                    variant={pathname === "/profile" ? "secondary" : "ghost"}
                    className="justify-start"
                    render={
                      <Link href="/profile" data-testid="nav-account-profile" />
                    }
                  >
                    <UserIcon data-icon="inline-start" />
                    Profile
                  </Button>
                  <Button
                    variant="ghost"
                    className="justify-start"
                    onClick={() => void signOut()}
                    data-testid="nav-account-sign-out"
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