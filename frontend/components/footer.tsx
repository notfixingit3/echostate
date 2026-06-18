import Link from "next/link"
import { ExternalLinkIcon } from "lucide-react"

import { BrandLogo } from "@/components/brand-logo"
import { Badge } from "@/components/ui/badge"
import { VersionBadge } from "@/components/version-badge"

const capabilities = ["WHOIS", "ASN / BGP", "DNS & TLS", "Snapshots", "PDF reports"]

const footerLinks = [
  {
    href: "https://github.com/notfixingit3/echostate",
    label: "GitHub",
  },
  {
    href: "https://github.com/notfixingit3/echostate/blob/dev/CHANGELOG.md",
    label: "Changelog",
  },
] as const

export function Footer() {
  const year = new Date().getFullYear()

  return (
    <footer
      className="mt-auto border-t border-border/60 bg-linear-to-b from-muted/20 to-muted/40"
      data-testid="footer"
    >
      <div className="container py-10 md:py-12">
        <div className="grid gap-10 lg:grid-cols-[minmax(0,1.2fr)_minmax(0,1fr)_minmax(0,0.8fr)] lg:items-start lg:gap-8">
          <div className="flex flex-col gap-4">
            <Link href="/" className="group flex w-fit items-center gap-3">
              <BrandLogo variant="icon" className="size-9" />
              <div className="flex flex-col gap-0.5">
                <span className="font-heading text-base font-semibold tracking-tight">
                  EchoState
                </span>
                <span className="text-xs text-muted-foreground">
                  Passive reconnaissance platform
                </span>
              </div>
            </Link>
            <p className="max-w-sm text-sm leading-relaxed text-muted-foreground">
              Gather WHOIS, routing, DNS, TLS, and web intelligence — then
              track changes with snapshots and exportable PDF reports.
            </p>
          </div>

          <div className="flex flex-col gap-3">
            <span className="font-heading text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              Capabilities
            </span>
            <div className="flex flex-wrap gap-2">
              {capabilities.map((cap) => (
                <Badge
                  key={cap}
                  variant="secondary"
                  className="font-mono text-[11px]"
                >
                  {cap}
                </Badge>
              ))}
            </div>
          </div>

          <div className="flex flex-col gap-4 lg:items-end">
            <div className="flex flex-wrap gap-4 text-sm">
              {footerLinks.map((link) => (
                <a
                  key={link.href}
                  href={link.href}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-1.5 text-muted-foreground transition-colors hover:text-primary"
                >
                  {link.label}
                  <ExternalLinkIcon className="size-3" />
                </a>
              ))}
            </div>
            <VersionBadge />
          </div>
        </div>

        <div className="mt-8 flex flex-col items-center justify-between gap-2 border-t border-border/40 pt-6 text-xs text-muted-foreground sm:flex-row">
          <span>© {year} EchoState</span>
          <span>Look before you leap.</span>
        </div>
      </div>
    </footer>
  )
}