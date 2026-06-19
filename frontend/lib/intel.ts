import type { Snapshot } from "./types"

export interface RawIntel {
  host?: string
  scanned_at?: string
  whois?: Record<string, unknown>
  asn?: Record<string, unknown>
  web?: Record<string, unknown>
  tls?: Record<string, unknown>
  dns?: Record<string, unknown>
  favicon?: Record<string, unknown>
  crawl?: Record<string, unknown>
  storage?: Record<string, unknown>
  ct?: Record<string, unknown>
  traceroute?: Record<string, unknown>
  screenshot?: Record<string, unknown>
  enrichment?: Record<string, unknown>
  errors?: string[]
}

export interface IntelHighlights {
  host: string
  resolvedIp?: string
  asn?: string
  asName?: string
  prefix?: string
  country?: string
  registry?: string
  webTitle?: string
  webUrl?: string
  registrar?: string
  domain?: string
  expirationDate?: string
  nameServers?: string[]
  domainStatus?: string[]
  clientIp?: string
  pwhoisOrg?: string
  pwhoisCity?: string
  pwhoisPrefix?: string
  pwhoisOriginAs?: string
  certExpires?: string
  certIssuer?: string
  dnsARecords?: string[]
  dnsSoaZone?: string
  dnsSoaMname?: string
  dnsSoaSerial?: number
  faviconMMH3?: string
  jarm?: string
  ja3s?: string
  hijackRisk?: string
  bgpPathStability?: string
  mailPostureGrade?: string
  mailPostureScore?: number
  securityTxtContact?: string
  mtaStsMode?: string
  caaRecordCount?: number
  redirectHopCount?: number
  dnsAAAARecords?: string[]
  ptrRecords?: string[]
  cdnProvider?: string
  mailProvider?: string
  bimiRecord?: string
  metaDescription?: string
  providerHintCount?: number
  humansTxtLines?: number
  adsTxtLines?: number
  rdapSource?: string
  enrichmentStatus?: string
  shodanHostOrg?: string
  shodanOpenPorts?: number
  censysServiceCount?: number
  hibpBreachedEmails?: number
  waybackUrlCount?: number
  virustotalRecordCount?: number
  tlsVersion?: string
  ocspStapled?: boolean
  dnssecStatus?: string
  ctCertCount?: number
  hstsPreloaded?: boolean
  cookieNameCount?: number
  jsAssetCount?: number
  jsLibHints?: string[]
  bucketCount?: number
  crawlPathCount?: number
  ctSubdomainCount?: number
  newCtSubdomainCount?: number
  tracerouteHopCount?: number
  certDaysRemaining?: number
  certExpired?: boolean
  certStatus?: "ok" | "warning" | "critical"
  wpPluginCount?: number
  wpThemeSlug?: string
  wpThemeVersion?: string
  newWpPluginCount?: number
  newWpThemeCount?: number
  errorCount: number
  changeCount: number
}

function asString(value: unknown): string | undefined {
  if (typeof value === "string" && value.trim()) return value.trim()
  return undefined
}

function formatDateField(value: unknown): string | undefined {
  if (typeof value === "string" && value.trim()) {
    const parsed = new Date(value)
    if (!Number.isNaN(parsed.getTime())) {
      return parsed.toLocaleDateString()
    }
    return value.trim()
  }
  return undefined
}

function asStringArray(value: unknown): string[] | undefined {
  if (!Array.isArray(value)) return undefined
  const items = value
    .map((item) => (typeof item === "string" ? item.trim() : ""))
    .filter(Boolean)
  return items.length > 0 ? items : undefined
}

function wordPressItemCount(value: unknown): number | undefined {
  if (!Array.isArray(value) || value.length === 0) return undefined
  return value.length
}

function wordPressPrimaryItem(
  value: unknown
): { slug?: string; version?: string } {
  if (!Array.isArray(value) || value.length === 0) return {}
  const first = value[0]
  if (!first || typeof first !== "object") return {}
  const item = first as Record<string, unknown>
  return {
    slug: asString(item.slug),
    version: asString(item.version),
  }
}

function collectJSLibHints(
  techStack: unknown,
  jsAssets: unknown[]
): string[] {
  const hints = new Set<string>()
  const add = (value: unknown) => {
    const text = typeof value === "string" ? value.trim() : ""
    if (
      text &&
      (text.startsWith("lib:") ||
        text.startsWith("framework:") ||
        text.startsWith("ui:") ||
        text.startsWith("bundler:") ||
        text.startsWith("cdn:") ||
        text.startsWith("analytics:"))
    ) {
      hints.add(text)
    }
  }

  if (Array.isArray(techStack)) {
    for (const item of techStack) add(item)
  }
  for (const asset of jsAssets) {
    if (!asset || typeof asset !== "object") continue
    add((asset as Record<string, unknown>).hint)
  }

  return Array.from(hints).sort()
}

export function extractIntel(
  raw?: RawIntel | null,
  snapshot?: Pick<
    Snapshot,
    | "client_ip"
    | "pwhois_org_name"
    | "pwhois_city"
    | "pwhois_prefix"
    | "pwhois_origin_as"
    | "pwhois_country_code"
    | "changes"
    | "change_details"
  >
): IntelHighlights {
  const whois = raw?.whois
  const asn = raw?.asn
  const web = raw?.web
  const tls = raw?.tls
  const dns = raw?.dns
  const favicon = raw?.favicon
  const crawl = raw?.crawl
  const storage = raw?.storage
  const ct = raw?.ct
  const traceroute = raw?.traceroute
  const ctSubdomains = asStringArray(ct?.subdomains)
  const tracerouteHops = Array.isArray(traceroute?.hops) ? traceroute.hops : []
  const routing = asn?.routing as Record<string, unknown> | undefined
  const pathProfile = routing?.path_profile as Record<string, unknown> | undefined
  const mailPosture = dns?.MAIL_POSTURE as Record<string, unknown> | undefined
  const jsAssets = Array.isArray(web?.js_assets) ? web.js_assets : []
  const jsHints = collectJSLibHints(web?.tech_stack, jsAssets)
  const certDays =
    typeof tls?.days_remaining === "number" ? tls.days_remaining : undefined
  const certExpired = tls?.expired === true
  let certStatus: IntelHighlights["certStatus"]
  if (certExpired) {
    certStatus = "critical"
  } else if (certDays !== undefined) {
    if (certDays <= 7) certStatus = "critical"
    else if (certDays <= 30) certStatus = "warning"
    else certStatus = "ok"
  }

  const newCtSubdomains = (snapshot?.change_details || []).filter(
    (change) => change.type === "new_ct_subdomain"
  ).length
  const newWpPlugins = (snapshot?.change_details || []).filter(
    (change) => change.type === "wp_plugin_added"
  ).length
  const newWpThemes = (snapshot?.change_details || []).filter(
    (change) => change.type === "wp_theme_added"
  ).length
  const primaryTheme = wordPressPrimaryItem(web?.wordpress_themes)

  const buckets = Array.isArray(storage?.buckets) ? storage.buckets : []
  const sitemapURLs = Array.isArray(crawl?.sitemap_urls) ? crawl.sitemap_urls : []
  const robots = crawl?.robots as Record<string, unknown> | undefined
  const disallow = Array.isArray(robots?.disallow) ? robots.disallow : []
  const securityTxt = crawl?.security_txt as Record<string, unknown> | undefined
  const mtaSts = dns?.MTA_STS as Record<string, unknown> | undefined
  const caaRecords = Array.isArray(dns?.CAA) ? dns.CAA : []
  const redirectChain = Array.isArray(web?.redirect_chain) ? web.redirect_chain : []
  const infraLabels = dns?.INFRA_LABELS as Record<string, unknown> | undefined
  const bimi = dns?.BIMI as Record<string, unknown> | undefined
  const meta = web?.meta as Record<string, unknown> | undefined
  const humansTxt = crawl?.humans_txt as Record<string, unknown> | undefined
  const adsTxt = crawl?.ads_txt as Record<string, unknown> | undefined
  const providerHints = Array.isArray(storage?.provider_hints)
    ? storage.provider_hints
    : []
  const enrichment = raw?.enrichment as Record<string, unknown> | undefined
  const shodan = enrichment?.shodan as Record<string, unknown> | undefined
  const shodanHost = shodan?.host as Record<string, unknown> | undefined
  const censys = enrichment?.censys as Record<string, unknown> | undefined
  const censysHost = censys?.host as Record<string, unknown> | undefined
  const hibp = enrichment?.hibp as Record<string, unknown> | undefined
  const wayback = enrichment?.wayback as Record<string, unknown> | undefined
  const virustotal = enrichment?.virustotal as Record<string, unknown> | undefined
  const hibpResults = Array.isArray(hibp?.results) ? hibp.results : []
  const hibpBreached = hibpResults.filter(
    (item) =>
      item &&
      typeof item === "object" &&
      (item as Record<string, unknown>).pwned === true
  ).length
  const shodanPorts = Array.isArray(shodanHost?.ports) ? shodanHost.ports : []
  const censysServices = Array.isArray(censysHost?.services)
    ? censysHost.services
    : []
  const dnssec = dns?.DNSSEC as Record<string, unknown> | undefined
  const hstsPreload = web?.hsts_preload as Record<string, unknown> | undefined
  const cookieNames = Array.isArray(web?.cookie_names) ? web.cookie_names : []
  const ctCerts = Array.isArray(ct?.certificates) ? ct.certificates : []

  return {
    host: asString(raw?.host) || "Unknown host",
    resolvedIp: asString(asn?.ip),
    asn: asString(asn?.asn),
    asName: asString(asn?.as_name),
    prefix: asString(asn?.prefix) || snapshot?.pwhois_prefix || undefined,
    country:
      asString(asn?.country) || snapshot?.pwhois_country_code || undefined,
    registry: asString(asn?.registry),
    webTitle: asString(web?.title),
    webUrl: asString(web?.url),
    registrar: asString(whois?.registrar),
    domain: asString(whois?.domain),
    expirationDate: asString(whois?.expiration_date),
    nameServers: asStringArray(whois?.name_servers),
    domainStatus: asStringArray(whois?.status),
    clientIp:
      snapshot?.client_ip && snapshot.client_ip !== "unknown"
        ? snapshot.client_ip
        : undefined,
    pwhoisOrg: snapshot?.pwhois_org_name || undefined,
    pwhoisCity: snapshot?.pwhois_city || undefined,
    pwhoisPrefix: snapshot?.pwhois_prefix || undefined,
    pwhoisOriginAs: snapshot?.pwhois_origin_as || undefined,
    certExpires: formatDateField(tls?.not_after),
    certIssuer: asString(tls?.issuer),
    dnsARecords: asStringArray(dns?.A),
    dnsAAAARecords: asStringArray(dns?.AAAA),
    ptrRecords: asStringArray(dns?.PTR),
    cdnProvider: asString(infraLabels?.cdn_provider),
    mailProvider: asString(infraLabels?.mail_provider),
    bimiRecord: asString(bimi?.record),
    metaDescription: asString(meta?.description),
    providerHintCount:
      providerHints.length > 0 ? providerHints.length : undefined,
    humansTxtLines: Array.isArray(humansTxt?.lines)
      ? humansTxt.lines.length
      : undefined,
    adsTxtLines: Array.isArray(adsTxt?.lines) ? adsTxt.lines.length : undefined,
    rdapSource: asString(whois?.rdap_source),
    enrichmentStatus: asString(enrichment?.status),
    shodanHostOrg: asString(shodanHost?.org),
    shodanOpenPorts: shodanPorts.length > 0 ? shodanPorts.length : undefined,
    censysServiceCount:
      censysServices.length > 0 ? censysServices.length : undefined,
    hibpBreachedEmails: hibpBreached > 0 ? hibpBreached : undefined,
    waybackUrlCount:
      typeof wayback?.total === "number" ? wayback.total : undefined,
    virustotalRecordCount:
      typeof virustotal?.total === "number" ? virustotal.total : undefined,
    tlsVersion: asString(tls?.tls_version),
    ocspStapled: tls?.ocsp_stapled === true ? true : undefined,
    dnssecStatus: asString(dnssec?.status),
    ctCertCount:
      typeof ct?.certificate_count === "number"
        ? ct.certificate_count
        : ctCerts.length > 0
          ? ctCerts.length
          : undefined,
    hstsPreloaded: hstsPreload?.preloaded === true ? true : undefined,
    cookieNameCount: cookieNames.length > 0 ? cookieNames.length : undefined,
    dnsSoaZone: asString((dns?.SOA as Record<string, unknown> | undefined)?.zone),
    dnsSoaMname: asString((dns?.SOA as Record<string, unknown> | undefined)?.mname),
    dnsSoaSerial:
      typeof (dns?.SOA as Record<string, unknown> | undefined)?.serial === "number"
        ? ((dns?.SOA as Record<string, unknown>).serial as number)
        : undefined,
    faviconMMH3: asString(favicon?.mmh3) || asString(favicon?.shodan),
    jarm: asString(tls?.jarm),
    ja3s: asString(tls?.ja3s),
    hijackRisk: asString(routing?.hijack_risk),
    bgpPathStability: asString(pathProfile?.stability),
    mailPostureGrade: asString(mailPosture?.grade),
    mailPostureScore:
      typeof mailPosture?.score === "number" ? mailPosture.score : undefined,
    securityTxtContact: asStringArray(securityTxt?.contacts)?.[0],
    mtaStsMode: asString(mtaSts?.mode),
    caaRecordCount: caaRecords.length > 0 ? caaRecords.length : undefined,
    redirectHopCount:
      redirectChain.length > 1 ? redirectChain.length - 1 : undefined,
    jsAssetCount: jsAssets.length > 0 ? jsAssets.length : undefined,
    jsLibHints: jsHints.length > 0 ? jsHints : undefined,
    bucketCount: buckets.length > 0 ? buckets.length : undefined,
    crawlPathCount:
      sitemapURLs.length + disallow.length > 0
        ? sitemapURLs.length + disallow.length
        : undefined,
    ctSubdomainCount: ctSubdomains?.length,
    newCtSubdomainCount: newCtSubdomains > 0 ? newCtSubdomains : undefined,
    tracerouteHopCount:
      tracerouteHops.length > 0 ? tracerouteHops.length : undefined,
    certDaysRemaining: certDays,
    certExpired,
    certStatus,
    wpPluginCount: wordPressItemCount(web?.wordpress_plugins),
    wpThemeSlug: primaryTheme.slug,
    wpThemeVersion: primaryTheme.version,
    newWpPluginCount: newWpPlugins > 0 ? newWpPlugins : undefined,
    newWpThemeCount: newWpThemes > 0 ? newWpThemes : undefined,
    errorCount: raw?.errors?.length ?? 0,
    changeCount: snapshot?.changes?.length ?? 0,
  }
}

export function formatFieldLabel(key: string): string {
  return key
    .replace(/_/g, " ")
    .replace(/\b\w/g, (char) => char.toUpperCase())
}

export function formatValue(value: unknown): string {
  if (value === null || value === undefined) return "—"
  if (typeof value === "string") return value
  if (typeof value === "number" || typeof value === "boolean") return String(value)
  if (Array.isArray(value)) {
    if (value.every((item) => typeof item === "string")) {
      return value.join(", ")
    }
    return JSON.stringify(value, null, 2)
  }
  return JSON.stringify(value, null, 2)
}

export const WHOIS_FIELDS = [
  "domain",
  "registrar",
  "registrant_email",
  "abuse_email",
  "expiration_date",
  "name_servers",
  "status",
  "rdap_source",
  "rdap",
] as const

export const ASN_CORE_FIELDS = [
  "asn",
  "as_name",
  "ip",
  "prefix",
  "country",
  "registry",
  "allocated",
] as const

export const ASN_ROUTING_FIELDS = [
  "hijack_risk",
  "visible_origins",
  "first_seen",
  "last_seen",
  "path_profile",
  "notes",
] as const

export const PEERINGDB_FIELDS = [
  "name",
  "website",
  "ix_count",
  "fac_count",
] as const

/** @deprecated Use ASN_CORE_FIELDS plus routing/peering sections. */
export const ASN_FIELDS = [
  ...ASN_CORE_FIELDS,
  "routing",
  "peeringdb",
] as const

export const WEB_FIELDS = [
  "title",
  "url",
  "header_url",
  "redirect_chain",
  "contact_emails",
  "meta",
  "copyrights",
  "tech_stack",
  "js_assets",
  "wordpress_plugins",
  "wordpress_themes",
  "security_headers",
  "headers",
  "detected_buckets",
  "provider_hints",
  "cookie_names",
  "hsts_preload",
] as const

export const PWHOIS_FIELDS = [
  "origin_as",
  "org_name",
  "country_code",
  "city",
  "prefix",
] as const

export const TLS_FIELDS = [
  "issuer",
  "subject",
  "dns_names",
  "ip_sans",
  "not_before",
  "not_after",
  "days_remaining",
  "expired",
  "serial",
  "tls_version",
  "negotiated_cipher",
  "ocsp_stapled",
  "ocsp_response_bytes",
  "chain_length",
  "chain",
  "version",
  "signature_algorithm",
  "cipher_suite",
  "jarm",
  "ja3s",
] as const

export const FAVICON_FIELDS = ["url", "mmh3", "shodan", "sha256", "size"] as const

export const CRAWL_FIELDS = [
  "robots",
  "security_txt",
  "humans_txt",
  "ads_txt",
  "app_ads_txt",
  "sitemap_urls",
  "sitemap_url_count",
  "errors",
] as const

export const STORAGE_FIELDS = ["buckets", "provider_hints"] as const

export const CT_FIELDS = [
  "domain",
  "source",
  "count",
  "subdomains",
  "certificate_count",
  "certificates",
] as const

export const TRACEROUTE_FIELDS = [
  "destination",
  "hop_count",
  "vantages",
  "warning",
  "skipped",
] as const

export const SCREENSHOT_FIELDS = [
  "url",
  "captured_at",
  "width",
  "height",
  "format",
  "error",
] as const

export const ENRICHMENT_FIELDS = [
  "status",
  "completed_at",
  "shodan",
  "censys",
  "hibp",
  "riskiq",
  "wayback",
  "virustotal",
  "shodan_error",
  "censys_error",
  "hibp_error",
  "riskiq_error",
  "wayback_error",
  "virustotal_error",
] as const

export const DNS_FIELDS = [
  "A",
  "AAAA",
  "PTR",
  "MX",
  "NS",
  "TXT",
  "CNAME",
  "SOA",
  "CAA",
  "DMARC",
  "SPF",
  "DKIM",
  "BIMI",
  "DMARC_PARSED",
  "MTA_STS",
  "TLS_RPT",
  "INFRA_LABELS",
  "DNSSEC",
  "MAIL_POSTURE",
] as const