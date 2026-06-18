import { Badge } from "@/components/ui/badge"
import { CountryFlag } from "@/components/country-flag"
import type { IntelHighlights } from "@/lib/intel"
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
} from "lucide-react"

function StatCard({
  label,
  value,
  icon: Icon,
  mono = false,
  badge,
}: {
  label: string
  value?: string
  icon: React.ComponentType<{ className?: string }>
  mono?: boolean
  badge?: string
}) {
  return (
    <div className="intel-stat flex flex-col gap-2 rounded-xl border bg-card/80 p-4 backdrop-blur-sm">
      <div className="flex items-center gap-2 text-xs font-medium uppercase tracking-wider text-muted-foreground">
        <Icon className="size-3.5 text-primary/70" />
        {label}
      </div>
      <div className="flex flex-wrap items-center gap-2">
        {badge ? (
          <Badge variant="secondary" className="font-mono text-xs">
            {badge}
          </Badge>
        ) : null}
        <span
          className={
            mono
              ? "font-mono text-sm leading-snug break-all"
              : "text-sm font-medium leading-snug"
          }
        >
          {value || "—"}
        </span>
      </div>
    </div>
  )
}

export function IntelSummary({ intel }: { intel: IntelHighlights }) {
  return (
    <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      <StatCard label="Host" value={intel.host} icon={GlobeIcon} mono />
      <StatCard
        label="Resolved IP"
        value={intel.resolvedIp}
        icon={ServerIcon}
        mono
      />
      <StatCard
        label="ASN"
        value={
          intel.asn && intel.asName
            ? `AS${intel.asn} · ${intel.asName}`
            : intel.asn
              ? `AS${intel.asn}`
              : intel.asName
        }
        icon={NetworkIcon}
      />
      <StatCard
        label="BGP Prefix"
        value={intel.prefix}
        icon={NetworkIcon}
        mono
      />
      <div className="intel-stat flex flex-col gap-2 rounded-xl border bg-card/80 p-4 backdrop-blur-sm">
        <div className="flex items-center gap-2 text-xs font-medium uppercase tracking-wider text-muted-foreground">
          <MapPinIcon className="size-3.5 text-primary/70" />
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