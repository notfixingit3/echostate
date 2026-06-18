import { Badge } from "@/components/ui/badge"
import { CountryFlag } from "@/components/country-flag"
import type { IntelHighlights } from "@/lib/intel"
import { cn } from "@/lib/utils"
import {
  GlobeIcon,
  NetworkIcon,
  ServerIcon,
  MapPinIcon,
  FileTextIcon,
  AlertTriangleIcon,
  GitCompareIcon,
  ShieldIcon,
  DatabaseIcon,
  ImageIcon,
  FingerprintIcon,
  CloudIcon,
  FolderSearchIcon,
  ScrollTextIcon,
} from "lucide-react"

function StatCard({
  label,
  value,
  icon: Icon,
  mono = false,
  badge,
  className,
}: {
  label: string
  value?: string
  icon: React.ComponentType<{ className?: string }>
  mono?: boolean
  badge?: string
  className?: string
}) {
  return (
    <div
      className={cn(
        "intel-stat flex min-w-0 flex-col gap-2 rounded-xl border bg-card/80 p-4 backdrop-blur-sm",
        className
      )}
    >
      <div className="flex items-center gap-2 text-xs font-medium uppercase tracking-wider text-muted-foreground">
        <Icon className="size-3.5 shrink-0 text-primary/70" />
        {label}
      </div>
      <div className="flex min-w-0 flex-wrap items-center gap-2">
        {badge ? (
          <Badge variant="secondary" className="max-w-full font-mono text-xs">
            <span className="truncate">{badge}</span>
          </Badge>
        ) : null}
        <span
          className={cn(
            "min-w-0 text-sm leading-snug",
            mono
              ? "break-all font-mono"
              : "line-clamp-3 break-words font-medium"
          )}
          title={value}
        >
          {value || "—"}
        </span>
      </div>
    </div>
  )
}

export function IntelSummary({ intel }: { intel: IntelHighlights }) {
  const asnLabel =
    intel.asn && intel.asName
      ? `AS${intel.asn} · ${intel.asName}`
      : intel.asn
        ? `AS${intel.asn}`
        : intel.asName

  return (
    <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
      <StatCard label="Host" value={intel.host} icon={GlobeIcon} mono />
      <StatCard
        label="Resolved IP"
        value={intel.resolvedIp}
        icon={ServerIcon}
        mono
      />
      <StatCard label="ASN" value={asnLabel} icon={NetworkIcon} />
      <StatCard
        label="BGP Prefix"
        value={intel.prefix}
        icon={NetworkIcon}
        mono
      />
      <div className="intel-stat flex min-w-0 flex-col gap-2 rounded-xl border bg-card/80 p-4 backdrop-blur-sm">
        <div className="flex items-center gap-2 text-xs font-medium uppercase tracking-wider text-muted-foreground">
          <MapPinIcon className="size-3.5 shrink-0 text-primary/70" />
          Country
        </div>
        <CountryFlag code={intel.country} />
      </div>
      <StatCard label="Web title" value={intel.webTitle} icon={FileTextIcon} />
      <StatCard label="Registrar" value={intel.registrar} icon={GlobeIcon} />
      <StatCard
        label="Cert expires"
        value={intel.certExpires}
        icon={ShieldIcon}
      />
      <StatCard
        label="Cert issuer"
        value={intel.certIssuer}
        icon={ShieldIcon}
      />
      <StatCard
        label="DNS A records"
        value={
          intel.dnsARecords?.length
            ? intel.dnsARecords.join(", ")
            : undefined
        }
        icon={DatabaseIcon}
        mono
      />
      <StatCard
        label="Favicon MMH3"
        value={intel.faviconMMH3}
        icon={ImageIcon}
        mono
      />
      <StatCard label="JARM" value={intel.jarm} icon={FingerprintIcon} mono />
      <StatCard
        label="BGP hijack risk"
        value={intel.hijackRisk}
        icon={ShieldIcon}
      />
      {intel.bucketCount ? (
        <StatCard
          label="Cloud buckets"
          value={String(intel.bucketCount)}
          icon={CloudIcon}
        />
      ) : null}
      {intel.crawlPathCount ? (
        <StatCard
          label="Crawl paths"
          value={String(intel.crawlPathCount)}
          icon={FolderSearchIcon}
        />
      ) : null}
      {intel.ctSubdomainCount ? (
        <StatCard
          label="CT subdomains"
          value={String(intel.ctSubdomainCount)}
          icon={ScrollTextIcon}
        />
      ) : null}
      <StatCard
        label="Submitter IP"
        value={intel.clientIp}
        icon={ServerIcon}
        mono
      />
      {intel.pwhoisOrg ? (
        <StatCard
          label="Submitter org"
          value={intel.pwhoisOrg}
          icon={NetworkIcon}
        />
      ) : null}
      {intel.pwhoisCity ? (
        <StatCard
          label="Submitter city"
          value={intel.pwhoisCity}
          icon={MapPinIcon}
        />
      ) : null}
      {intel.errorCount > 0 ? (
        <StatCard
          label="Gather errors"
          value={`${intel.errorCount} issue${intel.errorCount === 1 ? "" : "s"}`}
          icon={AlertTriangleIcon}
        />
      ) : null}
      {intel.changeCount > 0 ? (
        <StatCard
          label="Changes"
          value={`${intel.changeCount} since last scan`}
          icon={GitCompareIcon}
        />
      ) : null}
    </div>
  )
}