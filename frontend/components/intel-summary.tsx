import { HelpTip } from "@/components/help-tip"
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
  PuzzleIcon,
  LayoutTemplateIcon,
} from "lucide-react"

function StatCard({
  label,
  value,
  icon: Icon,
  mono = false,
  badge,
  className,
  helpId,
}: {
  label: string
  value?: string
  icon: React.ComponentType<{ className?: string }>
  mono?: boolean
  badge?: string
  className?: string
  helpId?: string
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
        {helpId ? <HelpTip id={helpId} /> : null}
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
    <div className="grid items-start gap-4 sm:grid-cols-2 xl:grid-cols-3">
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
        helpId="intel.cert_expires"
        label="Cert expires"
        value={
          intel.certDaysRemaining !== undefined
            ? `${intel.certExpires || "—"} (${intel.certDaysRemaining}d)`
            : intel.certExpires
        }
        icon={ShieldIcon}
        badge={
          intel.certStatus === "critical"
            ? "Expired / critical"
            : intel.certStatus === "warning"
              ? "Expiring soon"
              : undefined
        }
        className={
          intel.certStatus === "critical"
            ? "border-destructive/40"
            : intel.certStatus === "warning"
              ? "border-amber-500/40"
              : undefined
        }
      />
      <StatCard
        helpId="intel.cert_issuer"
        label="Cert issuer"
        value={intel.certIssuer}
        icon={ShieldIcon}
      />
      <StatCard
        helpId="intel.dns_a"
        label="DNS A records"
        value={
          intel.dnsARecords?.length
            ? intel.dnsARecords.join(", ")
            : undefined
        }
        icon={DatabaseIcon}
        mono
      />
      {intel.dnsAAAARecords?.length ? (
        <StatCard
          helpId="intel.dns_aaaa"
          label="DNS AAAA records"
          value={intel.dnsAAAARecords.join(", ")}
          icon={DatabaseIcon}
          mono
        />
      ) : null}
      {intel.ptrRecords?.length ? (
        <StatCard
          helpId="intel.dns_ptr"
          label="PTR records"
          value={intel.ptrRecords.join(", ")}
          icon={DatabaseIcon}
          mono
        />
      ) : null}
      {intel.cdnProvider ? (
        <StatCard
          helpId="intel.cdn_provider"
          label="CDN provider"
          value={intel.cdnProvider}
          icon={CloudIcon}
        />
      ) : null}
      {intel.mailProvider ? (
        <StatCard
          helpId="intel.mail_provider"
          label="Mail provider"
          value={intel.mailProvider}
          icon={ShieldIcon}
        />
      ) : null}
      {intel.bimiRecord ? (
        <StatCard
          helpId="intel.bimi"
          label="BIMI"
          value="Published"
          icon={ShieldIcon}
          badge="Brand indicator"
        />
      ) : null}
      {intel.metaDescription ? (
        <StatCard
          helpId="intel.meta_description"
          label="Meta description"
          value={intel.metaDescription}
          icon={FileTextIcon}
        />
      ) : null}
      {intel.providerHintCount ? (
        <StatCard
          helpId="intel.provider_hints"
          label="Provider hints"
          value={`${intel.providerHintCount} detected`}
          icon={CloudIcon}
        />
      ) : null}
      {intel.humansTxtLines ? (
        <StatCard
          label="humans.txt"
          value={`${intel.humansTxtLines} lines`}
          icon={ScrollTextIcon}
        />
      ) : null}
      {intel.adsTxtLines ? (
        <StatCard
          label="ads.txt"
          value={`${intel.adsTxtLines} lines`}
          icon={ScrollTextIcon}
        />
      ) : null}
      {intel.rdapSource ? (
        <StatCard
          label="WHOIS source"
          value={intel.rdapSource}
          icon={GlobeIcon}
          mono
        />
      ) : null}
      {intel.enrichmentStatus === "pending" ? (
        <StatCard
          helpId="intel.enrichment_status"
          label="Enrichment"
          value="Pending"
          icon={CloudIcon}
          badge="Async"
        />
      ) : null}
      {intel.shodanHostOrg ? (
        <StatCard
          helpId="intel.shodan_host"
          label="Shodan org"
          value={intel.shodanHostOrg}
          icon={CloudIcon}
          badge={
            intel.shodanOpenPorts
              ? `${intel.shodanOpenPorts} ports`
              : undefined
          }
        />
      ) : null}
      {intel.hibpBreachedEmails ? (
        <StatCard
          label="HIBP breaches"
          value={`${intel.hibpBreachedEmails} email${intel.hibpBreachedEmails === 1 ? "" : "s"}`}
          icon={ShieldIcon}
          badge="Breached"
          className="border-destructive/40"
        />
      ) : null}
      {intel.waybackUrlCount ? (
        <StatCard
          helpId="intel.wayback_urls"
          label="Wayback URLs"
          value={String(intel.waybackUrlCount)}
          icon={ScrollTextIcon}
        />
      ) : null}
      {intel.virustotalRecordCount ? (
        <StatCard
          label="VT passive DNS"
          value={`${intel.virustotalRecordCount} records`}
          icon={DatabaseIcon}
          mono
        />
      ) : null}
      {intel.tlsVersion ? (
        <StatCard
          helpId="intel.tls_version"
          label="TLS version"
          value={intel.tlsVersion}
          icon={ShieldIcon}
          badge={intel.ocspStapled ? "OCSP stapled" : undefined}
        />
      ) : null}
      {intel.dnssecStatus ? (
        <StatCard
          helpId="intel.dnssec"
          label="DNSSEC"
          value={intel.dnssecStatus}
          icon={DatabaseIcon}
          badge={intel.dnssecStatus === "signed" ? "Signed" : undefined}
        />
      ) : null}
      {intel.ctCertCount ? (
        <StatCard
          helpId="intel.ct_certs"
          label="CT certificates"
          value={String(intel.ctCertCount)}
          icon={ScrollTextIcon}
        />
      ) : null}
      {intel.hstsPreloaded ? (
        <StatCard
          label="HSTS preload"
          value="Preloaded"
          icon={ShieldIcon}
        />
      ) : null}
      {intel.cookieNameCount ? (
        <StatCard
          label="Cookie names"
          value={`${intel.cookieNameCount} detected`}
          icon={PuzzleIcon}
          mono
        />
      ) : null}
      {intel.dnsSoaZone ? (
        <StatCard
          helpId="intel.dns_soa"
          label="DNS SOA zone"
          value={
            intel.dnsSoaSerial
              ? `${intel.dnsSoaZone} · serial ${intel.dnsSoaSerial}`
              : intel.dnsSoaZone
          }
          icon={DatabaseIcon}
          mono
        />
      ) : null}
      {intel.dnsSoaMname && !intel.dnsSoaZone ? (
        <StatCard
          label="DNS SOA mname"
          value={intel.dnsSoaMname}
          icon={DatabaseIcon}
          mono
        />
      ) : null}
      <StatCard
        label="Favicon MMH3"
        value={intel.faviconMMH3}
        icon={ImageIcon}
        mono
      />
      <StatCard helpId="intel.jarm" label="JARM" value={intel.jarm} icon={FingerprintIcon} mono />
      <StatCard helpId="intel.ja3s" label="JA3S" value={intel.ja3s} icon={FingerprintIcon} mono />
      <StatCard
        label="BGP hijack risk"
        value={intel.hijackRisk}
        icon={ShieldIcon}
      />
      {intel.bgpPathStability ? (
        <StatCard
          helpId="intel.bgp_path_stability"
          label="BGP path stability"
          value={intel.bgpPathStability}
          icon={NetworkIcon}
        />
      ) : null}
      {intel.securityTxtContact ? (
        <StatCard
          label="Security contact"
          value={intel.securityTxtContact}
          icon={ShieldIcon}
          mono
        />
      ) : null}
      {intel.mtaStsMode ? (
        <StatCard
          label="MTA-STS"
          value={intel.mtaStsMode}
          icon={ShieldIcon}
          badge={intel.mtaStsMode === "enforce" ? "Strict" : undefined}
        />
      ) : null}
      {intel.caaRecordCount ? (
        <StatCard
          label="CAA records"
          value={String(intel.caaRecordCount)}
          icon={ShieldIcon}
        />
      ) : null}
      {intel.redirectHopCount ? (
        <StatCard
          label="HTTP redirects"
          value={`${intel.redirectHopCount} hop${intel.redirectHopCount === 1 ? "" : "s"}`}
          icon={GlobeIcon}
        />
      ) : null}
      {intel.mailPostureGrade ? (
        <StatCard
          helpId="intel.mail_posture"
          label="Email posture"
          value={
            intel.mailPostureScore !== undefined
              ? `${intel.mailPostureGrade} (${intel.mailPostureScore}/100)`
              : intel.mailPostureGrade
          }
          icon={ShieldIcon}
          badge={
            intel.mailPostureGrade === "F" || intel.mailPostureGrade === "D"
              ? "Weak"
              : undefined
          }
          className={
            intel.mailPostureGrade === "F"
              ? "border-destructive/40"
              : intel.mailPostureGrade === "D"
                ? "border-amber-500/40"
                : undefined
          }
        />
      ) : null}
      {intel.jsAssetCount ? (
        <StatCard
          helpId="intel.js_assets"
          label="JS assets"
          value={`${intel.jsAssetCount} loaded`}
          icon={PuzzleIcon}
          badge={
            intel.jsLibHints?.length
              ? intel.jsLibHints.slice(0, 2).join(", ")
              : undefined
          }
        />
      ) : null}
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
          badge={
            intel.newCtSubdomainCount
              ? `${intel.newCtSubdomainCount} new`
              : undefined
          }
        />
      ) : null}
      {intel.wpPluginCount ? (
        <StatCard
          label="WP plugins"
          value={`${intel.wpPluginCount} detected`}
          icon={PuzzleIcon}
          badge={
            intel.newWpPluginCount
              ? `${intel.newWpPluginCount} new`
              : undefined
          }
        />
      ) : null}
      {intel.wpThemeSlug ? (
        <StatCard
          label="WP theme"
          value={intel.wpThemeSlug}
          icon={LayoutTemplateIcon}
          badge={
            intel.newWpThemeCount
              ? `${intel.newWpThemeCount} new`
              : intel.wpThemeVersion
                ? `v${intel.wpThemeVersion}`
                : undefined
          }
          mono
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