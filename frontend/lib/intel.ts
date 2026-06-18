import type { Snapshot } from "./types"

export interface RawIntel {
  host?: string
  scanned_at?: string
  whois?: Record<string, unknown>
  asn?: Record<string, unknown>
  web?: Record<string, unknown>
  tls?: Record<string, unknown>
  dns?: Record<string, unknown>
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
  >
): IntelHighlights {
  const whois = raw?.whois
  const asn = raw?.asn
  const web = raw?.web
  const tls = raw?.tls
  const dns = raw?.dns

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
  "expiration_date",
  "name_servers",
  "status",
] as const

export const ASN_FIELDS = [
  "asn",
  "as_name",
  "ip",
  "prefix",
  "country",
  "registry",
  "allocated",
] as const

export const WEB_FIELDS = ["title", "url", "copyrights", "tech_stack", "security_headers", "headers"] as const

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
  "version",
  "signature_algorithm",
] as const

export const DNS_FIELDS = [
  "A",
  "MX",
  "NS",
  "TXT",
  "CNAME",
  "DMARC",
] as const